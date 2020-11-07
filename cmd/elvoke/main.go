package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	stampfile      string
	ident          string
	cachedir       string
	failOnPostpone bool
	debug          bool
	interval       time.Duration

	printDebug = func(_ string, _ ...interface{}) {}
)

func init() {
	flag.StringVar(&ident, "id", "", "identifier to distinguish between different commands.")
	flag.StringVar(&stampfile, "file", "", "stamp file (default: $USERCACHEDIR/IDENT.stamp)")
	flag.BoolVar(&failOnPostpone, "fail-on-postpone", false, "exit with non-zero code when postponing.")
	flag.BoolVar(&debug, "debug", false, "print debug information.")
	flag.DurationVar(&interval, "interval", 1*time.Hour, "minimum interval between invocations of the same command.")
	flag.Usage = usage
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("elvoke: ")
	log.SetOutput(os.Stderr)
	flag.Parse()

	if debug {
		printDebug = log.Printf
	}

	ensureCache()
	printDebug("cachedir = %s", cachedir)

	var (
		program string
		args    []string
		needRun bool
	)

	switch flag.NArg() {
	case 0:
		log.Fatal("missing operand")
	case 1:
		program = flag.Arg(0)
	default:
		program = flag.Arg(0)
		args = flag.Args()[1:]
	}

	cmd := exec.Command(program, args...)
	printDebug("program = %q, args = %+v", cmd.Path, cmd.Args)

	if ident == "" {
		ident = cmdIdent(cmd)
	}

	stampfilepath := stampFilename(ident)

	printDebug("ident = %s, stampfile = ", ident, stampfilepath)

	info, err := os.Stat(stampfilepath)

	switch {
	case err != nil && os.IsNotExist(err):
		printDebug("stampfile does not exist, will run")

		needRun = true
	case err != nil:
		log.Fatal(err)
	default:
		elapsed := time.Since(info.ModTime())
		needRun = elapsed > interval
		printDebug("interval = %s, elapsed = %s, mtime = %s", interval, elapsed, info.ModTime())
	}

	if !needRun {
		if !failOnPostpone {
			printDebug("no need to run, exiting with success")
			os.Exit(0)
		}

		printDebug("no need to run, failing")
		os.Exit(1)
	}

	printDebug("running")

	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatalf("the child exited with error: %v", err)
	}

	printDebug("writing %s", stampfilepath)

	if err := os.Chtimes(stampfilepath, time.Now(), time.Now()); err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}

		fp, err := os.Create(stampfilepath)
		if err != nil {
			log.Fatal(err)
		}

		defer fp.Close()
	}
}

func ensureCache() {
	var err error

	cachedir, err = os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	cachedir = filepath.Join(cachedir, "elvoke")

	if err := os.MkdirAll(cachedir, 0755); err != nil {
		log.Fatal(err)
	}
}

func stampFilename(ident string) string {
	if stampfile != "" {
		return stampfile
	}

	return filepath.Join(cachedir, fmt.Sprintf("%s.stamp", ident))
}

func cmdIdent(cmd *exec.Cmd) string {
	identHash := sha256.New()
	if _, err := identHash.Write([]byte(strings.Join(append(filepath.SplitList(cmd.Path), cmd.Args...), "_"))); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", identHash.Sum(nil))
}

func usage() {
	usageString := `Usage: elvoke [OPTION]... -- COMMAND [ARG]...
Run or postpone a command, depending on how much time elapsed from the last successful run.

Options:`
	_, _ = fmt.Fprintln(os.Stderr, usageString)

	flag.PrintDefaults()
}
