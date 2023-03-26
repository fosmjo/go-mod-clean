# go-mod-cleaner

`go-mod-cleaner` is a cli tool to remove stale go modules. Due to the [side-effect of go module-cache](https://go.dev/ref/mod#module-cache), you need to run this tool with administrator privileges.


## Install

```sh
go install github.com/fosmjo/go-mod-cleaner@latest
```

## Usage
Stale go modules are all mods under `$GOPATH/pkg/mod` except mods in use, you need to specify mods in use via `go.mod` files or dirs contain `go.mod` files.

```sh
# show help doc
$ go-mod-cleaner -h
Usage:
  go-mod-cleaner [flags]

Flags:
  -h, --help              help for go-mod-cleaner
  -m, --modfile strings   modfile paths or dirs, mods referenced by these modfiles will not be removed
```

## Use case

```sh
$ sudo -E go-mod-cleaner -m ~/coding -m ~/work -m ~/study
Found 37 stale mods, occupied 32 MB disk space.

You can:
(1) Remove them (need administrator privileges).
(2) View them.
(3) Quit.

Type one of the numbers in parentheses:1
```