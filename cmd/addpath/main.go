package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	intpath "al.essio.dev/pkg/tools/internal/path"
)

var (
	helpMode   bool
	appendMode bool
	progName   = "addpath"
)

func init() {
	detectAppendMode()

	if !appendMode {
		flag.BoolVar(&appendMode, "append", false, "append DIR to ENVVAR.")
	}

	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.Usage = usage
}

func detectAppendMode() {
	appendMode = path.Base(os.Args[0]) == "appendpath"
	if appendMode {
		progName = "appendpath"
	}
}

func main() {
	var (
		envvar string
		dir    string
	)

	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", progName))
	log.SetOutput(os.Stderr)
	flag.Parse()

	if helpMode {
		usage()
		os.Exit(0)
	}

	switch flag.NArg() {
	case 0:
		fallthrough
	case 1:
		log.Fatal("missing operand(s)")
	case 2:
		envvar = flag.Arg(0)
		dir = flag.Arg(1)
		oldVal := os.Getenv(envvar)
		newVal := intpath.AddDir(oldVal, dir, appendMode)
		fmt.Println(newVal)
	default:
		log.Fatal("invalid arguments")
	}
}

func usage() {
	usageString := fmt.Sprintf(`Usage: %s ENVVAR DIR
Add directory to a PATH-like environment variable and
print the new contents of ENVVAR.

If the program's filename is appendpath, the append
mode is turned on.
`, progName)
	_, _ = fmt.Fprintln(os.Stderr, usageString)

	flag.PrintDefaults()
}
