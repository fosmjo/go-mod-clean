package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"golang.org/x/mod/modfile"
)

type Cleaner struct {
	modCachePath string
	modfilePaths []string
}

func NewCleaner(modCachePath string, filepaths []string) *Cleaner {
	return &Cleaner{
		modCachePath: modCachePath,
		modfilePaths: modfilePaths,
	}
}

func (c *Cleaner) Clean() error {
	cachedMods, err := c.allCachedMods()
	if err != nil {
		return err
	}

	modfiles, err := c.modfiles()
	if err != nil {
		return err
	}

	inUseMods, err := c.allInUseMods(modfiles)
	if err != nil {
		return err
	}

	uselessMods := make([]string, 0, max(0, len(cachedMods)-len(inUseMods)))

	for _, mod := range cachedMods {
		if _, ok := inUseMods[mod]; !ok {
			uselessMods = append(uselessMods, mod)
		}
	}

	totalSize, err := c.calculateSize(uselessMods)
	if err != nil {
		return err
	}

	fmt.Printf(
		`Found %d stale mods, occupied %s disk space.

You can:
(1) Remove them (need admistrator permission).
(2) View them.
(3) Quit.

Type one of the numbers in parentheses:`,
		len(uselessMods),
		humanize.Bytes(uint64(totalSize)),
	)
	var input string
	_, err = fmt.Scanln(&input)
	if err != nil {
		return err
	}

	switch input {
	case "1":
		return c.removeMods(uselessMods)
	case "2":
		return c.viewMods(uselessMods)
	default:
		return nil
	}
}

func (c *Cleaner) viewMods(mods []string) error {
	for _, mod := range mods {
		path := c.modAbsPath(mod)
		fmt.Println(path)
	}
	return nil
}
func (c *Cleaner) removeMods(mods []string) error {
	for _, mod := range mods {
		path := c.modAbsPath(mod)
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Cleaner) allCachedMods() ([]string, error) {
	mods := make([]string, 0, 128)
	err := filepath.WalkDir(c.modCachePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if strings.HasPrefix(d.Name(), "cache") {
				return filepath.SkipDir
			}

			if strings.Contains(d.Name(), "@") {
				mod, err := filepath.Rel(c.modCachePath, path)
				if err != nil {
					return err
				}

				mods = append(mods, mod)
				return filepath.SkipDir
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return mods, nil
}

func (c *Cleaner) modAbsPath(mod string) string {
	return filepath.Join(c.modCachePath, mod)
}

func (c *Cleaner) calculateSize(mods []string) (int64, error) {
	var size int64
	for _, mod := range mods {
		s, err := c.calculateModSize(mod)
		if err != nil {
			return 0, err
		}

		size += s
	}

	return size, nil
}

func (c *Cleaner) calculateModSize(mod string) (int64, error) {
	var size int64
	dir := c.modAbsPath(mod)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		size += info.Size()
		return nil
	})

	return size, err
}

func (c *Cleaner) allInUseMods(modfiles []string) (map[string]struct{}, error) {
	result := make(map[string]struct{}, 128)

	for _, path := range modfiles {
		mods, err := c.parseModFile(path)
		if err != nil {
			return nil, err
		}

		for _, m := range mods {
			result[m] = struct{}{}
		}
	}

	return result, nil
}

func (c *Cleaner) modfiles() ([]string, error) {
	var modfiles []string

	for _, p := range c.modfilePaths {
		if filepath.Base(p) == "go.mod" {
			modfiles = append(modfiles, p)
			continue
		}

		files, err := filepath.Glob(filepath.Join(p, "**", "go.mod"))
		if err != nil {
			return nil, err
		}

		modfiles = append(modfiles, files...)
	}

	return modfiles, nil
}

func (c *Cleaner) parseModFile(modfilepath string) ([]string, error) {
	data, err := os.ReadFile(modfilepath)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(modfilepath)
	return c.retriveMods(filename, data)
}

func (c *Cleaner) retriveMods(filename string, data []byte) ([]string, error) {
	file, err := modfile.ParseLax(filename, data, nil)
	if err != nil {
		return nil, err
	}

	mods := make([]string, 0, len(file.Require)+2*len(file.Replace))

	for _, r := range file.Require {
		mods = append(mods, r.Mod.String())
	}

	for _, r := range file.Replace {
		mods = append(mods, r.Old.String(), r.New.String())
	}

	return mods, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
