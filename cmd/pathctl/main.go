package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alessio/unixtools/internal/path"
)

const (
	progName = "pathctl"
)

var (
	helpMode    bool
	pathListSep string
)

var (
	envVar string
	paths  path.List
)

func init() {
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.StringVar(&pathListSep, "s", string(filepath.ListSeparator), "path list separator.")
	flag.StringVar(&envVar, "e", "PATH", "input environment variable")
	flag.Usage = usage
	flag.CommandLine.SetOutput(os.Stderr)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", progName))
	log.SetOutput(os.Stderr)
	flag.Parse()

	if helpMode {
		usage()
		os.Exit(0)
	}

	paths = path.NewPathList(envVar)

	if flag.NArg() < 1 || flag.Arg(0) == "list" || flag.Arg(0) == "l" {
		list()
		os.Exit(0)
	}

	if flag.NArg() < 2 {
		log.Fatalf("%s: wrong arguments: %s", flag.Arg(0), flag.Args())
	}

	switch flag.Arg(0) {
	case "prepend", "p":
		push()
	case "drop", "d":
		drop()
	case "append", "a":
		appendPath()
		//case "cleanup":
	}

	fmt.Println(paths.String())
}

func list() {
	for _, p := range paths.StringSlice() {
		fmt.Println(p)
	}
}

func push() {
	//oldPath := pathEnvvar
	p := flag.Arg(1)
	if ok := paths.Prepend(p); !ok {
		_, _ = fmt.Fprint(os.Stderr, "the path exists already")
	}
}

func drop() {
	p := flag.Arg(1)
	if ok := paths.Drop(p); !ok {
		_, _ = fmt.Fprint(os.Stderr, "the path did npt exist")
	}
}

func appendPath() {
	p := flag.Arg(1)
	if ok := paths.Append(p); !ok {
		_, _ = fmt.Fprint(os.Stderr, "the path exists already")
	}
}

func usage() {
	s := fmt.Sprintf(`Usage: %s COMMAND [PATH]
Make the management of the PATH environment variable
simple, fast, and predictable.

Commands:

   append, a       append a path to the end

Options:
`, progName)
	_, _ = fmt.Fprintln(os.Stderr, s)

	flag.PrintDefaults()

	_, _ = fmt.Fprintln(os.Stderr, `
If COMMAND is not provided, it prints the contents of the PATH
environment variable; the default output format is one path per
line.`)
}
