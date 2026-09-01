# cmd

## Installation

```shell
go install al.essio.dev/cmd/mcd@latest
```

## Usage

```shell
mcd DIR
```

**mcd** create the directory DIR and all intermediate directories.
Its output can be passed to the builtin **eval** as command prints
`cd DIR` if it succeeds or `:` if it fails.

See **mcd -help**.
