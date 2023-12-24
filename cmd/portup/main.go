package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/alessio/unixtools/internal/version"
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

	if len(flag.Args()) > 0 {
		logfile, err := openLogFile(flag.Arg(0))
		if err != nil {
			log.Fatalf("couldn't open the file %s: %v", flag.Arg(0), err)
		}

		log.SetOutput(logfile)
	}

	if err := runPortCommand("-N", "-v", "selfupdate"); err != nil {
		log.Fatal(err)
	}

	if err := runPortCommand("-N", "outdated"); err != nil {
		log.Fatal(err)
	}

	if err := runPortCommand("-N", "-v", "-R", "-u", "-c", "upgrade", "outdated"); err != nil {
		log.Fatal(err)
	}

	if runReclaim {
		if err := runPortCommand("-N", "-v", "reclaim"); err != nil {
			log.Fatal(err)
		}
	}
}

func runPortCommand(args ...string) error {
	log.Println("Will run", PortExe, args)
	cmd := exec.Command(PortExe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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
