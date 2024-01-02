package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"al.essio.dev/pkg/tools/internal/version"
)

const shortUsage = "usage: mcd DIR"

var (
	helpMode    bool
	versionMode bool
	errLog      *log.Logger
)

func init() {
	errLog = log.New(os.Stderr, "mcd: ", 0)
	errLog.SetFlags(0)
	errLog.SetOutput(os.Stderr)

	flag.Usage = func() { fmt.Fprintln(os.Stderr, shortUsage) }
	flag.ErrHelp = nil
	flag.CommandLine.SetOutput(errLog.Writer())

	flag.BoolVar(&helpMode, "help", false, "display this help and exit")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit")
}

func main() {
	flag.Parse()
	handleHelpAndVersionModes()

	if flag.NArg() != 1 {
		errLog.Fatalf("invalid arguments -- '%s'\n%s\n", strings.Join(flag.Args(), " "), shortUsage)
	}

	if _, err := os.Getwd(); err != nil {
		errLog.Fatal(err)
	}

	var newDir = flag.Arg(0)

	if err := os.MkdirAll(newDir, os.ModePerm); err != nil {
		fmt.Println(":")
		errLog.Fatal(err)
	}

	fmt.Println("cd", newDir)
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
Create DIR and all intermediate directories as required.
Also, it prints 'cd DIR' to STDOUT in case of success else
':' so that it can be passed as an argument to the shell
builtin 'eval'.

Examples:

  $ mcd ~/a/b/newdir
  cd /home/user/a/b/c
  $ mcd /root/a/b/newdir
  :
  mcd: mkdir /root/a/b/newdir: permission denied

Options:`
	_, _ = fmt.Fprintln(os.Stderr, usageString)

	flag.PrintDefaults()
}
