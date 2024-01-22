package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"al.essio.dev/pkg/tools/dirlist"
	"al.essio.dev/pkg/tools/internal/version"
)

const (
	program = "pathctl"
)

var (
	helpMode     bool
	versionMode  bool
	listMode     bool
	noPrefixMode bool
	dropMode     bool
)

var (
	envVar string
)

var cmdHandlers map[string]func(d dirlist.List)

func init() {
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")
	flag.BoolVar(&dropMode, "D", false, "drop the path before adding it again to the list.")
	flag.BoolVar(&noPrefixMode, "noprefix", false, "output the variable contents only.")
	flag.BoolVar(&listMode, "L", false, "use a newline character as path list separator.")
	flag.StringVar(&envVar, "E", "PATH", "input environment variable.")
	flag.Usage = usage
	flag.CommandLine.SetOutput(os.Stderr)

	cmdHandlers = func() map[string]func(dirlist.List) {
		return map[string]func(dirlist.List){
			"append":  cmdHandlerAppend,
			"drop":    cmdHandlerDrop,
			"prepend": cmdHandlerPrepend,

			// aliases
			"a": cmdHandlerAppend,
			"d": cmdHandlerDrop,
			"p": cmdHandlerPrepend,
		}
	}()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", program))
	log.SetOutput(os.Stderr)
	flag.Parse()

	handleHelpAndVersionModes()

	dirs := dirlist.New()
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

func printPathList(d dirlist.List) {
	var sb = strings.Builder{}
	sb.Reset()

	printPrefix := !noPrefixMode

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
	s := fmt.Sprintf(`Usage: %s [COMMAND [PATH]]
Make the management of the PATH environment variable
simple, fast, and predictable.

Commands:

   append, a           append a path to the end of the list.
   drop, d             drop a path.
   prepend, p          prepend a path to the list.

Options:
`, program)
	_, _ = fmt.Fprintln(os.Stderr, s)

	flag.PrintDefaults()

	_, _ = fmt.Fprintln(os.Stderr, `
When used with the -D flag, the commands append and prepend
drop PATH before adding it again to the list. This behaviour
guarantees that PATH is added as either the first or the last
element of the path list.

If COMMAND is not provided, it prints the contents of the PATH
environment variable.`)
}

func cmdHandlerAppend(d dirlist.List) {
	if dropMode {
		d.Drop(flag.Arg(1))
	}
	d.Append(flag.Arg(1))
}

func cmdHandlerDrop(d dirlist.List) {
	d.Drop(flag.Arg(1))
}

func cmdHandlerPrepend(d dirlist.List) {
	if dropMode {
		d.Drop(flag.Arg(1))
	}
	d.Prepend(flag.Arg(1))
}
