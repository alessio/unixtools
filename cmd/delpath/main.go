package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/alessio/unixtools/internal/path"
)

var (
	helpMode bool
)

func init() {
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.Usage = usage
}

func main() {
	var (
		envvar string
		dir    string
	)

	log.SetFlags(0)
	log.SetPrefix("delpath: ")
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
		newVal := path.RemoveDir(oldVal, dir)
		fmt.Println(newVal)
	default:
		log.Fatal("invalid arguments")
	}
}

func usage() {
	usageString := `Usage: delpath ENVVAR DIR
Remove a directory from a PATH-like environment variable and
print the new contents of ENVVAR.
`
	_, _ = fmt.Fprintln(os.Stderr, usageString)

	flag.PrintDefaults()
}
