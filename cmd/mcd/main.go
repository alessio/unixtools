package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/alessio/unixtools/internal/version"
)

var (
	helpMode    bool
	debugMode   bool
	versionMode bool

	dbgLog *log.Logger
)

func init() {
	flag.CommandLine.SetOutput(os.Stderr)
	flag.CommandLine.Usage = usage

	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")
	flag.BoolVar(&debugMode, "D", false, "produce debug output.")

	log.SetFlags(0)
	log.SetPrefix("mcd: ")

	dbgLog = log.New(io.Discard, "mcd: ", log.LstdFlags)
}

func main() {
	flag.Parse()

	handleHelpAndVersionModes()
	if debugMode {
		dbgLog.SetOutput(os.Stderr)
	}

	if _, err := os.Getwd(); err != nil {
		log.Fatal(err)
	}

	if flag.Parse(); flag.NArg() != 1 {
		log.Fatal("invalid arguments")
	}

	dbgLog.Printf("Abs(%s)\n", flag.Arg(0))
	newDir, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalf("couldn't find an absolute path: %v", err)
	}
	dbgLog.Println("Abs =", newDir)

	if err := os.MkdirAll(newDir, os.ModePerm); err != nil {
		dbgLog.Printf("mkdir: err = %v", err)
		log.Fatalf("couldn't create the directory: %v", err)
	}

	if err := os.Chdir(newDir); err != nil {
		dbgLog.Printf("chdir: err = %v", err)
		log.Fatalf("couldn't cd: %v", err)
	}
}

func handleHelpAndVersionModes() {
	switch {
	case helpMode:
		usage()
		os.Exit(0)
	case versionMode:
		version.PrintWithCopyright()
		os.Exit(0)
	}

}

func usage() {
	usageString := `Usage: mcd DIR
Change the current directory to DIR. Also, create intermediate directories as required.

Options:`
	_, _ = fmt.Fprintln(os.Stderr, usageString)

	flag.PrintDefaults()
}
