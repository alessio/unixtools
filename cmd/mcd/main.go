package main

import (
	"al.essio.dev/pkg/tools/version"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	helpMode    bool
	versionMode bool
)

func init() {
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")

	flag.Usage = usage
	flag.ErrHelp = nil
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("mcd: ")
	log.SetOutput(os.Stderr)
	flag.Parse()

	handleHelpAndVersionModes()

	if _, err := os.Getwd(); err != nil {
		log.Fatal(err)
	}

	if flag.Parse(); flag.NArg() != 1 {
		log.Fatal("invalid arguments")
	}

	if err := os.MkdirAll(flag.Arg(0), os.ModePerm); err != nil {
		log.Fatal(err)
	}

	if err := os.Chdir(flag.Arg(0)); err != nil {
		log.Fatal(err)
	}
}

func handleHelpAndVersionModes() {
	switch {
	case helpMode:
		usage()
	case versionMode:
		version.PrintWithCopyright()
	}

	os.Exit(0)
}

func usage() {
	usageString := `Usage: mcd DIR
Change the current directory to DIR. Also, create intermediate directories as required.

Options:`
	_, _ = fmt.Fprintln(os.Stderr, usageString)

	flag.PrintDefaults()
}
