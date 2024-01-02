# pathctl

## Installation

```shell
go install al.essio.dev/pkg/tools/cmd/pathctl@latest
```

## Usage

```shell
pathctl [[append|prepend|drop] DIR]
```

**pathctl** streamlines operations over PATH-like environment variables.
It provides idempotent operations add and remove directories from lists
of directories typically separated by the  character ':' on POSIX
systems (or the OS-specific path list separator). Also, it automatically
removes duplicate elements as it encounters them.

The append/prepend commands do nothing if the argument exists already
in the path list; if invoked with the option -D, the already existing
duplicate would be removed and the argument would be appended/prepended.

See **pathctl -help** for more information.
