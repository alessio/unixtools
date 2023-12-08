package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alessio/unixtools/internal/path"
	"github.com/alessio/unixtools/internal/version"
)

const (
	progName = "pathctl"
)

var (
	helpMode    bool
	versionMode bool
	pathListSep string
)

var (
	envVar string
	paths  path.List
)

func init() {
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")
	flag.StringVar(&pathListSep, "s", string(os.PathListSeparator), "path list separator.")
	flag.StringVar(&envVar, "e", "PATH", "input environment variable")
	flag.Usage = usage
	flag.CommandLine.SetOutput(os.Stderr)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", progName))
	log.SetOutput(os.Stderr)
	flag.Parse()

	handleHelpAndVersionModes()

	paths = path.NewPathList(envVar)

	if flag.NArg() < 1 {
		list()
		os.Exit(0)
	}

	if flag.NArg() == 1 {
		switch flag.Arg(0) {
		case "list", "l":
			list()
		case "appendPathctlDir", "apd":
			appendPath(exePath())
		case "prependPathctlDir", "ppd":
			prepend(exePath())
		default:
			log.Fatalf("unrecognized command: %s", flag.Arg(0))
		}
	} else {
		switch flag.Arg(0) {
		case "prepend", "p":
			prepend(flag.Arg(1))
		case "drop", "d":
			drop(flag.Arg(1))
		case "append", "a":
			appendPath(flag.Arg(1))
		}
	}

	fmt.Println(paths.String())
}

func list() {
	for _, p := range paths.StringSlice() {
		fmt.Println(p)
	}
}

func prepend(p string) {
	//oldPath := pathEnvvar
	if ok := paths.Prepend(p); !ok {
		log.Println("the path already exists")
	}
}

func drop(p string) {
	if ok := paths.Drop(p); !ok {
		log.Println("the path already exists")
	}
}

func appendPath(p string) {
	if ok := paths.Append(p); !ok {
		log.Println("the path already exists")
	}
}

func handleHelpAndVersionModes() {
	if helpMode {
		usage()
		os.Exit(0)
	}

	if versionMode {
		version.PrintWithCopyright()
		os.Exit(0)
	}
}

func usage() {
	s := fmt.Sprintf(`Usage: %s COMMAND [PATH]
Make the management of the PATH environment variable
simple, fast, and predictable.

Commands:

   append, a       append a path to the end
   drop, d         drop a path
   list, l         list the paths
   prepend, p      prepend a path to the list

Options:
`, progName)
	_, _ = fmt.Fprintln(os.Stderr, s)

	flag.PrintDefaults()

	_, _ = fmt.Fprintln(os.Stderr, `
If COMMAND is not provided, it prints the contents of the PATH
environment variable; the default output format is one path per
line.`)
}

func exePath() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	return filepath.Dir(exePath)
}
