package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alessio/unixtools/internal/path"
	"github.com/alessio/unixtools/internal/version"
)

const (
	programme = "pathctl"
)

var (
	helpMode     bool
	versionMode  bool
	listMode     bool
	noprefixMode bool
	pathListSep  string
)

var (
	envVar string
	paths  path.DirList
)

var cmdHandlers map[string]func()

func init() {
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")
	flag.BoolVar(&noprefixMode, "noprefix", false, "output the variable contents only.")
	flag.StringVar(&pathListSep, "sep", string(filepath.ListSeparator), "path list separator.")
	flag.StringVar(&envVar, "E", "PATH", "input environment variable")
	flag.Usage = usage
	flag.CommandLine.SetOutput(os.Stderr)

	cmdHandlers = func() map[string]func() {
		hdlrList := func() { list() }
		hdlrAppend := func() { appendPath(flag.Arg(1)) }
		hdlrDrop := func() { drop(flag.Arg(1)) }
		hdlrPrepend := func() { prependPath(flag.Arg(1)) }

		return map[string]func(){
			"list":              hdlrList,
			"apppend":           hdlrAppend,
			"drop":              hdlrDrop,
			"prepend":           hdlrPrepend,
			"appendPathctlDir":  func() { appendPath(exePath()) },
			"prependPathctlDir": func() { prependPath(exePath()) },

			// aliases
			"a": hdlrAppend,
			"d": hdlrDrop,
			"p": hdlrPrepend,
			"l": hdlrList,
		}
	}()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", programme))
	log.SetOutput(os.Stderr)
	flag.Parse()

	handleHelpAndVersionModes()

	paths := path.NewDirList()
	paths.SetEnvvar(envVar)

	if flag.NArg() < 1 {
		printPathList()
		os.Exit(0)
	}

	if handler, ok := cmdHandlers[flag.Arg(0)]; ok {
		handler()
	} else {
		log.Fatalf("unrecognized command: %s", flag.Arg(0))
	}

	printPathList()
}

func printPathList() {
	var sb strings.Builder
	switch {
	case listMode:
		for _, p := range paths.Slice() {
			fmt.Println(p)
		}
	case noprefixMode:
		goto printout
	default:
		sb.WriteString(fmt.Sprintf("%s=", envVar))
	}

printout:
	sb.WriteString(paths.String())
	fmt.Println(sb.String())
}

func list() {
	for _, p := range paths.Slice() {
		fmt.Println(p)
	}
}

func prependPath(p string) {
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
	if !helpMode && !versionMode {
		return
	}

	switch {
	case helpMode:
		usage()
	case versionMode:
		version.PrintWithCopyright()
	}

	os.Exit(0)
}

func usage() {
	s := fmt.Sprintf(`Usage: %s COMMAND [PATH]
Make the management of the PATH environment variable
simple, fast, and predictable.

Commands:

   append, a       append a path to the end
   drop, d         drop a path
   list, l         list the paths
   prepend, p      prependPath a path to the list

Options:
`, programme)
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
