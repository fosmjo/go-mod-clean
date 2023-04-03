/*
Copyright © 2023 fosmjo <imefangjie@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"os"
	"path/filepath"

	cleaner "github.com/fosmjo/go-mod-cleaner"
	"github.com/spf13/cobra"
)

var modfilePaths []string

var rootCmd = &cobra.Command{
	Use:   "go-mod-cleaner",
	Short: "Clean up unused Go modules.",
	Long: `Clean unused Go modules. To be specific, it cleans up all modules within $GOPATH/pkg/mod, except for currently used modules.
To specify the modules in use, you need to indicate them via go.mod files or directories that contain go.mod files. `,
	RunE: func(cmd *cobra.Command, args []string) error {
		modCachePath := filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
		cleaner := cleaner.New(modCachePath, modfilePaths)
		return cleaner.Clean()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringSliceVarP(
		&modfilePaths, "modfile", "m", nil,
		"modfile paths or dirs, modules referenced by these modfiles are considered in use, and won't be cleaned",
	)
	err := rootCmd.MarkFlagRequired("modfile")
	if err != nil {
		panic(err)
	}
}
