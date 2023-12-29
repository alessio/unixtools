package main

import (
	"al.essio.dev/pkg/tools/version"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	PortExe = "/opt/local/bin/port"

	programName = "portup"
)

var (
	helpMode    bool
	versionMode bool
	runReclaim  bool

	//	cwd string
)

func init() {
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.BoolVar(&helpMode, "h", false, "")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")
	flag.BoolVar(&versionMode, "v", false, "")
	flag.BoolVar(&runReclaim, "with-reclaim", false, "run reclaim after 'port upgrade outdated'.")

	flag.Usage = usage
	flag.CommandLine.SetOutput(os.Stderr)
}

func main() {
	log.SetFlags(log.LstdFlags)
	log.SetPrefix(fmt.Sprintf("%s: ", programName))
	log.SetOutput(os.Stderr)
	flag.Parse()

	handleHelpAndVersionModes()

	if flag.NArg() > 1 {
		log.Fatalf("unexpected number of arguments: want 0 or 1, got %d", flag.NArg())
	}

	if flag.NArg() == 1 {
		logfile, err := openLogFile(flag.Arg(0))
		if err != nil {
			log.Fatalf("couldn't open the file %s: %v", flag.Arg(0), err)
		}

		defer logfile.Close()
		log.SetOutput(logfile)

	}

	runPortFailOnError("-N", "-v", "selfupdate")
	runPortFailOnError("-N", "outdated")
	runPortFailOnError("-N", "-v", "-R", "-u", "-c", "upgrade", "outdated")

	if runReclaim {
		runPortFailOnError("-N", "-v", "reclaim")
	}
}

func runPortFailOnError(args ...string) {
	log.Printf("RUN: %s %s\n", PortExe, strings.Join(args, " "))
	cmd := exec.Command(PortExe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("command exited with error: %v", err)
	}
}

func openLogFile(filename string) (io.WriteCloser, error) {
	fp, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	return fp, nil
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, `Usage: %s [PATH]
This command is a simple and convenient shorcut
to update and upgrade the packages installed
with MacPorts.

Options:
`, programName)
	flag.PrintDefaults()
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
