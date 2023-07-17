![Build](https://github.com/alessio/unixtools/workflows/Build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/alessio/unixtools)](https://goreportcard.com/report/github.com/alessio/unixtools)
[![license](https://img.shields.io/github/license/alessio/unixtools.svg)](https://github.com/alessio/unixtools/blob/master/LICENSE)
[![LoC](https://tokei.rs/b1/github/alessio/unixtools)](https://github.com/alessio/unixtools)

# unixtools

alessio's UNIX Convenience Tools.

# Installation

Just run:

```
$ go get github.com/alessio/unixtools/cmd/...
```

# What's in This Repo?

## elvoke

This is a Golang implementation of [Jakub Wilk's elvoke](https://github.com/jwilk/elvoke).

Run or postpone a command, depending on how much time elapsed from the last successful run.

## mcd

Change the current directory to DIR. Also, create intermediate directories as required.

## refiles

This was inspired by @niemeyer's [remv](http://niemeyer.net/remv).

Rename files in directories that match a given pattern.

### Options

Run `refiles -help` to print the following help screen:

```
  -I	prompt before every overwrite
  -R	search files under each directory recursively
  -m	move files matching PATTERN to REPLACE
  -simulate
    	print changes that are supposed to be done, but don't actually make any
  -verbose
    	enable verbose output
```

## popbak, pushbak

Manage a stack of directories backups. **pushbak** makes backups of a directory, **popbak**
restores the last backup available.

## seq

Golang implementation of the UNIX `seq` command. It prints sequences of numbers.

This is a Go implementation of the UNIX `seq` command.
