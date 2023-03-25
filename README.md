# go-mod-cleaner

`go-mod-cleaner` is a cli tool to remove stale go modules. Due to [side-effect of go module-cache](https://go.dev/ref/mod#module-cache), you need run this tool as an administrator.


## Install

```sh
go install github.com/fosmjo/go-mod-cleaner@latest
```

## Usage

```sh
# show help doc
$ go-mod-cleaner -h
Usage:
  go-mod-cleaner [flags]

Flags:
  -h, --help              help for go-mod-cleaner
  -m, --modfile strings   modfile paths or dirs, mods referenced by these modfiles will not be removed

# remove or view mods to be deleted
$ sudo -E go-mod-cleaner -m <path1> -m <path2> 
```

## Look & Feel

```sh
sudo -E go-mod-cleaner -m ~/coding -m ~/work -m ~/study
Found 37 stale mods, occupied 32 MB disk space.

You can:
(1) Remove them (need administrator permission).
(2) View them.
(3) Quit.

Type one of the numbers in parentheses:1
```