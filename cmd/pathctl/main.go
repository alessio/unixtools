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
)

var cmdHandlers map[string]func(d path.List)

func init() {
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")
	flag.BoolVar(&noprefixMode, "noprefix", false, "output the variable contents only.")
	flag.StringVar(&pathListSep, "sep", string(filepath.ListSeparator), "path list separator.")
	flag.StringVar(&envVar, "E", "PATH", "input environment variable")
	flag.Usage = usage
	flag.CommandLine.SetOutput(os.Stderr)

	cmdHandlers = func() map[string]func(path.List) {
		hList := func(d path.List) { list(d) }
		hAppend := func(d path.List) { d.Append(flag.Arg(1)) }
		hDrop := func(d path.List) { d.Drop(flag.Arg(1)) }
		hPrepend := func(d path.List) { d.Prepend(flag.Arg(1)) }

		return map[string]func(path.List){
			"list":    hList,
			"append":  hAppend,
			"drop":    hDrop,
			"prepend": hPrepend,
			//"appendPathctlDir":  func() { appendPath(exePath()) },
			//"prependPathctlDir": func() { prependPath(exePath()) },

			// aliases
			"a": hAppend,
			"d": hDrop,
			"p": hPrepend,
			"l": hList,
		}
	}()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", programme))
	log.SetOutput(os.Stderr)
	flag.Parse()

	handleHelpAndVersionModes()

	dirs := path.NewList()
	dirs.LoadEnv(envVar)
	//fmt.Println(Paths.Slice())

	if flag.NArg() < 1 {
		list(dirs)
		os.Exit(0)
	}

	if handler, ok := cmdHandlers[flag.Arg(0)]; ok {
		handler(dirs)
		printPathList(dirs)
	} else {
		log.Fatalf("unrecognized command: %s", flag.Arg(0))
	}
}

func printPathList(d path.List) {
	//if len(Paths.Slice()) == 0 {
	//	fmt.Println()
	//	os.Exit(0)
	//}

	var sb = strings.Builder{}
	sb.Reset()

	printPrefix := !noprefixMode

	switch {
	case listMode:
		sb.WriteString(strings.Join(d.Slice(), "\n"))
		break
	case printPrefix:
		sb.WriteString(fmt.Sprintf("%s=", envVar))
		fallthrough
	default:
		sb.WriteString(d.String())
	}

	fmt.Println(sb.String())
}

func list(d path.List) {
	printPathList(d)
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
