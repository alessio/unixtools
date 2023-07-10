package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var helpMode bool

func init() {
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.Usage = usage
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("mcd: ")
	log.SetOutput(os.Stderr)

	if helpMode {
		usage()
		os.Exit(0)
	}

	if _, err := os.Getwd(); err != nil {
		log.Fatal(err)
	}

	if 	flag.Parse(); flag.NArg() != 1 {
		log.Fatal("invalid arguments")
	}

	if err := os.MkdirAll(flag.Arg(0), os.ModePerm); err != nil {
		log.Fatal(err)
	}

	if err := os.Chdir(flag.Arg(0)); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	usageString := `Usage: mcd DIR
Change the current directory DIR. Also, create intermediate directories as required.

Options:`
	_, _ = fmt.Fprintln(os.Stderr, usageString)

	flag.PrintDefaults()
}

