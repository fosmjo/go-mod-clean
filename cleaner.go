package cleaner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/dustin/go-humanize"
	"golang.org/x/mod/modfile"
)

type Cleaner struct {
	modCachePath    string
	modDownloadPath string
	modfilePaths    []string
}

func New(modCachePath string, modfilePaths []string) *Cleaner {
	return &Cleaner{
		modCachePath:    modCachePath,
		modDownloadPath: filepath.Join(modCachePath, "cache", "download"),
		modfilePaths:    modfilePaths,
	}
}

func (c *Cleaner) Clean() error {
	extractedMods, err := c.allExtractedMods()
	if err != nil {
		return err
	}

	downloadedMods, err := c.allDownloadedMods()
	if err != nil {
		return err
	}

	modfiles, err := c.allModfiles()
	if err != nil {
		return err
	}

	inUseMods, err := c.allInUseMods(modfiles)
	if err != nil {
		return err
	}

	unusedExtractedMods := c.unusedMods(extractedMods, inUseMods)
	unusedDownloadedMods := c.unusedMods(downloadedMods, inUseMods)

	totalSize, err := c.calculateSize(unusedExtractedMods, unusedDownloadedMods)
	if err != nil {
		return err
	}

	fmt.Printf(
		`Found %d unused mods, occupied %s disk space.

You can:
(1) Remove them (require admistrator privileges).
(2) View them.
(3) Quit.

Type one of the numbers in parentheses:`,
		len(unusedExtractedMods)+len(unusedDownloadedMods),
		humanize.Bytes(totalSize),
	)
	var input string
	_, err = fmt.Scanln(&input)
	if err != nil {
		return err
	}

	switch input {
	case "1":
		return c.removeMods(unusedExtractedMods, unusedDownloadedMods)
	case "2":
		return c.viewMods(unusedExtractedMods, unusedDownloadedMods)
	default:
		return nil
	}
}

func (c *Cleaner) viewMods(extractedMods []string, downloadedMods []string) error {
	for _, mod := range extractedMods {
		path := c.extractedModAbsPath(mod)
		fmt.Println(path)
	}

	for _, mod := range downloadedMods {
		files, err := c.downloadedModFiles(mod)
		if err != nil {
			return err
		}

		for _, file := range files {
			fmt.Println(file)
		}
	}

	return nil
}

func (c *Cleaner) removeMods(extractedMods []string, downloadedMods []string) error {
	for _, mod := range extractedMods {
		path := c.extractedModAbsPath(mod)
		fmt.Printf("Removing %s\n", path)
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}

	for _, mod := range downloadedMods {
		files, err := c.downloadedModFiles(mod)
		if err != nil {
			return err
		}

		for _, file := range files {
			fmt.Printf("Removing %s\n", file)
			err := os.Remove(file)
			if err != nil {
				return err
			}
		}
	}

	return c.rewriteVersionListFiles(downloadedMods)
}

func (c *Cleaner) rewriteVersionListFiles(removedMods []string) error {
	mod2versions := make(map[string][]string, len(removedMods))

	for _, mod := range removedMods {
		parts := strings.Split(mod, "@")
		if len(parts) != 2 {
			return fmt.Errorf("invalid mod: %s", mod)
		}

		mod2versions[parts[0]] = append(mod2versions[parts[0]], parts[1])
	}

	for mod, removedVersions := range mod2versions {
		err := c.rewriteVersionListFile(mod, removedVersions)
		if err != nil {
			log.Printf("failed to rewrite version list for %s: %v", mod, err)
			continue
		}
	}

	return nil
}

func (c *Cleaner) rewriteVersionListFile(mod string, removedVersions []string) error {
	filepath := filepath.Join(c.modDownloadPath, mod, "@v", "list")
	f, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return nil // file does not exist, nothing to do
	}

	defer f.Close()

	allVersions, err := c.parseVersionListFile(f)
	if err != nil {
		return err
	}

	err = f.Truncate(0)
	if err != nil {
		return err
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	remainedVersions := diffSlice(allVersions, removedVersions)
	for _, version := range remainedVersions {
		fmt.Fprintf(f, "%s\n", version)
	}

	return nil
}

func (c *Cleaner) allExtractedMods() ([]string, error) {
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

	return mods, err
}

// allDownloadedMods returns the list of all mods under directory $GOPATH/pkg/mod/cache/download/
func (c *Cleaner) allDownloadedMods() ([]string, error) {
	store := make(map[string]struct{}, 256)

	err := filepath.WalkDir(c.modDownloadPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() == "sumdb" {
			return filepath.SkipDir
		}

		if !d.IsDir() && strings.Contains(path, "@v") && strings.Contains(d.Name(), ".") {
			// e.g.: path => go.uber.org/fx/@v/v1.17.0.info
			relpath, err := filepath.Rel(c.modDownloadPath, path)
			if err != nil {
				return err
			}

			// e.g. go.uber.org/fx/@v/v1.17.0.info => go.uber.org/fx
			modpath := filepath.Dir(filepath.Dir(relpath))

			filename := filepath.Base(relpath)
			version := strings.TrimSuffix(filename, filepath.Ext(filename))

			mod := modpath + "@" + version
			store[mod] = struct{}{}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	mods := make([]string, 0, len(store))

	for mod := range store {
		mods = append(mods, mod)
	}

	return mods, nil
}

func (c *Cleaner) unusedMods(mods []string, inUseMods map[string]struct{}) []string {
	unused := make([]string, 0, max(len(mods)-len(inUseMods), 0))

	for _, mod := range mods {
		if _, ok := inUseMods[mod]; !ok {
			unused = append(unused, mod)
		}
	}

	return unused
}

func (c *Cleaner) extractedModAbsPath(mod string) string {
	return filepath.Join(c.modCachePath, mod)
}

func (c *Cleaner) calculateSize(extractedMods, downloadedMods []string) (uint64, error) {
	var size atomic.Uint64
	var wg sync.WaitGroup
	errCh := make(chan error)

	wg.Add(len(extractedMods) + len(downloadedMods))

	calculateModSize := func(mod string, calculate func(mod string) (int64, error)) {
		defer wg.Done()

		s, err := calculate(mod)
		if err != nil {
			errCh <- err
			return
		}

		size.Add(uint64(s))
	}

	for _, mod := range extractedMods {
		go calculateModSize(mod, c.calculateExtractedModSize)
	}

	for _, mod := range downloadedMods {
		go calculateModSize(mod, c.calculateDownloadedModSize)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return 0, errors.Join(errs...)
	}

	return size.Load(), nil
}

func (c *Cleaner) calculateExtractedModSize(mod string) (int64, error) {
	var size int64

	err := filepath.Walk(
		c.extractedModAbsPath(mod),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			size += info.Size()
			return nil
		},
	)

	return size, err
}

func (c *Cleaner) calculateDownloadedModSize(mod string) (int64, error) {
	files, err := c.downloadedModFiles(mod)
	if err != nil {
		return 0, err
	}

	var size int64

	for _, file := range files {
		fi, err := os.Stat(file)
		if err != nil {
			return 0, err
		}

		size += fi.Size()
	}

	return size, err
}

func (c *Cleaner) downloadedModFiles(mod string) ([]string, error) {
	parts := strings.Split(mod, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid mod %s", mod)
	}

	modpath, version := parts[0], parts[1]

	files := make([]string, 0, 6)

	err := filepath.WalkDir(
		filepath.Join(c.modDownloadPath, modpath, "@v"),
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if strings.Contains(d.Name(), version) {
				files = append(files, path)
			}

			return nil
		},
	)

	return files, err
}

func (c *Cleaner) allInUseMods(modfiles []string) (map[string]struct{}, error) {
	result := make(map[string]struct{}, len(modfiles)*32)

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

func (c *Cleaner) allModfiles() ([]string, error) {
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

func (c *Cleaner) parseModFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(path)
	return c.retriveMods(filename, data)
}

// parseVersionListFile parses a version list file into a slice of strings
func (c *Cleaner) parseVersionListFile(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	versions := make([]string, 0, 3)

	for scanner.Scan() {
		versions = append(versions, scanner.Text())
	}

	return versions, scanner.Err()
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

// diffSlice returns a slice of strings that are in a but not in b
func diffSlice(a, b []string) []string {
	bmap := make(map[string]struct{}, len(b))
	for _, v := range b {
		bmap[v] = struct{}{}
	}

	result := make([]string, 0, max(len(a)-len(b), 0))
	for _, v := range a {
		if _, ok := bmap[v]; !ok {
			result = append(result, v)
		}
	}

	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
