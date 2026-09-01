![Build](https://github.com/alessio/unixtools/workflows/Build/badge.svg)
[![license](https://img.shields.io/github/license/alessio/unixtools.svg)](https://github.com/alessio/unixtools/blob/main/LICENSE)
[![codecov](https://codecov.io/github/alessio/unixtools/graph/badge.svg?token=XG71JUFEFN)](https://codecov.io/github/alessio/unixtools)

# unixtools

A collection of UNIX convenience tools written in Go.

## Installation

### Using `go install`

You can install individual tools directly with `go install`:

```bash
go install al.essio.dev/cmd/elvoke@latest
go install al.essio.dev/cmd/mcd@latest
go install al.essio.dev/cmd/refiles@latest
go install al.essio.dev/cmd/seq@latest
```

### Building from Source

Clone the repository and build all binaries using `make`:

```bash
git clone https://github.com/alessio/unixtools.git
cd unixtools
make build
```

Compiled binaries will be available in the `build/` directory.

---

## Tools

### `elvoke`

A Go implementation of [Jakub Wilk's elvoke](https://github.com/jwilk/elvoke).

Run or postpone a command, depending on how much time has elapsed since the last successful run.

#### Usage

```bash
elvoke [OPTIONS] -- COMMAND [ARGS]...
```

#### Options

| Option | Description | Default |
|---|---|---|
| `-interval` | Minimum interval between invocations of the same command | `1h0m0s` |
| `-id` | Identifier to distinguish between different commands | *(auto-generated)* |
| `-file` | Custom stamp file path | `$USERCACHEDIR/elvoke/<IDENT>.stamp` |
| `-fail-on-postpone` | Exit with a non-zero exit code when postponing | `false` |
| `-debug` | Print debug information to stderr | `false` |
| `-version` | Output version information and exit | `false` |

#### Example

```bash
# Run apt-get update only if at least 24 hours have passed since the last run
elvoke -interval 24h -- apt-get update
```

---

### `mcd`

Change current directory to `DIR`, creating any required intermediate directories along the way.

`mcd` creates `DIR` and outputs `cd DIR` on standard output so that it can be evaluated in shell environments.

#### Usage

```bash
mcd DIR
```

#### Shell Integration

Add the following wrapper function to your `~/.bashrc` or `~/.zshrc`:

```bash
mcd() {
  eval "$(command mcd "$@")"
}
```

---

### `refiles`

A regular expression batch file renaming tool inspired by Gustavo Niemeyer's [remv](http://niemeyer.net/remv).

#### Usage

```bash
refiles [OPTIONS] PATTERN REPLACE [DIRECTORY]...
```

If no `DIRECTORY` is specified, `refiles` operates on the current working directory.

#### Options

| Option | Description |
|---|---|
| `-m` | Move mode: match against full filename and replace entire name using capture groups |
| `-I` | Prompt before every file overwrite |
| `-R` | Search files under each directory recursively |
| `-simulate` | Print planned renames without executing any filesystem modifications |
| `-verbose` | Enable verbose logging |

#### Examples

```bash
# Replace spaces with underscores in filenames:
refiles ' ' '_'

# Move/rename files matching a pattern using regular expression capture groups:
refiles -m '^6.1.(\d{3})$' 'vim-6.1-$1.patch'

# Preview changes recursively before applying:
refiles -simulate -R '([a-z]+)_old' '$1_new'
```

---

### `seq`

A Go implementation of the standard UNIX `seq` command. Prints sequences of numbers.

#### Usage

```bash
seq [OPTIONS] LAST
seq [OPTIONS] FIRST LAST
seq [OPTIONS] FIRST INCREMENT LAST
```

#### Options

| Option | Description | Default |
|---|---|---|
| `-separator` | String to separate printed numbers | `\n` |
| `-width` | Equalize width by padding numbers with leading zeroes | `0` (off) |
| `-help` | Display usage help and exit | `false` |
| `-version` | Output version information and exit | `false` |

#### Examples

```bash
# Print numbers from 1 to 5:
seq 5

# Print numbers from 1 to 10 with step 2:
seq 1 2 10

# Print formatted numbers separated by commas and padded with zeroes:
seq -separator ", " -width 3 1 10
```

---

## License

This project is licensed under the MIT License. See [LICENSE](file:///Users/alessio/Documents/src/unixtools/LICENSE) for details.

