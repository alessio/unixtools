package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"al.essio.dev/pkg/tools/internal/version"
	"al.essio.dev/pkg/tools/pathlist"
)

const (
	programme = "pathctl"
)

var (
	helpMode     bool
	versionMode  bool
	listMode     bool
	noprefixMode bool
)

var (
	envVar string
)

var cmdHandlers map[string]func(d pathlist.List)

func init() {
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")
	flag.BoolVar(&noprefixMode, "noprefix", false, "output the variable contents only.")
	flag.BoolVar(&listMode, "L", false, "use a newline character as path list separator.")
	flag.StringVar(&envVar, "E", "PATH", "input environment variable")
	flag.Usage = usage
	flag.CommandLine.SetOutput(os.Stderr)

	cmdHandlers = func() map[string]func(pathlist.List) {
		hAppend := func(d pathlist.List) { d.Append(flag.Arg(1)) }
		hMoveAppend := func(d pathlist.List) {
			d.Drop(flag.Arg(1))
			d.Append(flag.Arg(1))
		}
		hDrop := func(d pathlist.List) { d.Drop(flag.Arg(1)) }
		hMovePrepend := func(d pathlist.List) {
			d.Drop(flag.Arg(1))
			d.Prepend(flag.Arg(1))
		}
		hPrepend := func(d pathlist.List) {
			d.Prepend(flag.Arg(1))
		}

		return map[string]func(pathlist.List){
			"append":       hAppend,
			"drop":         hDrop,
			"prepend":      hPrepend,
			"move-append":  hMoveAppend,
			"move-prepend": hMovePrepend,
			//"appendPathctlDir":  func() { appendPath(exePath()) },
			//"prependPathctlDir": func() { prependPath(exePath()) },

			// aliases
			"a":  hAppend,
			"d":  hDrop,
			"p":  hPrepend,
			"mv": hMoveAppend,
			"mp": hMovePrepend,
		}
	}()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", programme))
	log.SetOutput(os.Stderr)
	flag.Parse()

	handleHelpAndVersionModes()

	dirs := pathlist.New()
	dirs.LoadEnv(envVar)

	if flag.NArg() < 1 {
		printPathList(dirs)
		os.Exit(0)
	}

	if handler, ok := cmdHandlers[flag.Arg(0)]; ok {
		handler(dirs)
		printPathList(dirs)
	} else {
		log.Fatalf("unrecognized command: %s", flag.Arg(0))
	}
}

func printPathList(d pathlist.List) {
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

func list(d pathlist.List) {
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

   append, a           append a path to the end of the list
   move-append, ma     append a new path to the end of the list;
                       if the list contains the path already then
                       it will be moved to the end of the list
   drop, d             drop a path
   prepend, p          prepend a path to the list
   move-prepend, mp    prepend a new path to the top of the list;
                       if the list contains the path already then
                       it will be moved to the top of the list
Options:
`, programme)
	_, _ = fmt.Fprintln(os.Stderr, s)

	flag.PrintDefaults()

	_, _ = fmt.Fprintln(os.Stderr, `
If COMMAND is not provided, it prints the contents of the PATH
environment variable.`)
}

