# go-mod-cleaner

`go-mod-cleaner` is a cli tool to clean unused Go modules. To be specific, it cleans up all modules within `$GOPATH/pkg/mod`, except for currently used modules. To specify the modules in use, you need to indicate them via `go.mod` files or directories that contain `go.mod` files. Because of [the side-effect of the go module-cache](https://go.dev/ref/mod#module-cache), administrator privileges are necessary when running this tool.

## Install

```sh
go install github.com/fosmjo/go-mod-cleaner/cmd/go-mod-cleaner@latest
```

## Usage

```sh
# show help doc
$ go-mod-cleaner -h
Clean up outdated Go modules.

Usage:
  go-mod-cleaner [flags]

Flags:
  -h, --help              help for go-mod-cleaner
  -m, --modfile strings   modfile paths or dirs, modules referenced by these modfiles are considered in use, and won't be cleaned
```

## Use case

```sh
$ sudo -E go-mod-cleaner -m ~/coding -m ~/work -m ~/study
Found 37 unused mods, occupied 32 MB disk space.

You can:
(1) Remove them (need administrator privileges).
(2) View them.
(3) Quit.

Type one of the numbers in parentheses:1
```
