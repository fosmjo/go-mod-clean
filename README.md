# go-mod-cleaner

`go-mod-cleaner` is a cli tool to clean up unused Go modules. To be specific, it cleans up all modules within `$GOPATH/pkg/mod`, except for currently used modules. To specify the modules in use, you need to indicate them via `go.mod` files or directories that contain `go.mod` files. Because of [the side-effect of the go module-cache](https://go.dev/ref/mod#module-cache), administrator privileges are required if you want to removed unused modules.

## Install

```sh
go install github.com/fosmjo/go-mod-cleaner/cmd/go-mod-cleaner@latest
```

## Usage

```sh
# show help doc
$ go-mod-cleaner -h
Clean up unused Go modules. To be specific, it cleans up all modules within $GOPATH/pkg/mod,
except for currently used modules. To specify the modules in use, you need to indicate them
via go.mod files or directories that contain go.mod files.

Usage:
  go-mod-cleaner [flags]

Flags:
  -h, --help              help for go-mod-cleaner
  -m, --modfile strings   go.mod files or directories that contain go.mod files,
                          modules referenced by these files are considered in use
```

## Use case
### View unused modules

```sh
# no administrator privileges required
$ go-mod-cleaner -m ~/coding -m ~/work -m ~/study
Found 37 unused mods, occupied 32 MB disk space.

You can:
(1) View them.
(2) Remove them (require admistrator privileges).
(3) Quit.

Type one of the numbers in parentheses:1
```

### Remove unused modules

```sh
# require administrator privileges
$ sudo -E go-mod-cleaner -m ~/coding -m ~/work -m ~/study
Found 37 unused mods, occupied 32 MB disk space.

You can:
(1) View them.
(2) Remove them (require admistrator privileges).
(3) Quit.

Type one of the numbers in parentheses:2
```
