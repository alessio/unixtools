package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"al.essio.dev/pkg/tools/internal/version"
)

var (
	stampfile      string
	ident          string
	cachedir       string
	failOnPostpone bool
	debug          bool
	versionMode    bool
	interval       time.Duration

	printDebug = func(_ string, _ ...interface{}) {}
)

func init() {
	flag.StringVar(&ident, "id", "", "identifier to distinguish between different commands.")
	flag.StringVar(&stampfile, "file", "", "stamp file (default: $USERCACHEDIR/IDENT.stamp)")
	flag.BoolVar(&failOnPostpone, "fail-on-postpone", false, "exit with non-zero code when postponing.")
	flag.BoolVar(&debug, "debug", false, "print debug information.")
	flag.DurationVar(&interval, "interval", 1*time.Hour, "minimum interval between invocations of the same command.")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")
	flag.Usage = usage
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("elvoke: ")
	log.SetOutput(os.Stderr)
	flag.Parse()

	if versionMode {
		version.PrintWithCopyright()
		os.Exit(0)
	}

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

	// Look up ELVOKE_HOME first for compatibility with upstream's.
	if envHome := os.Getenv("ELVOKE_HOME"); envHome != "" {
		if err := validateEnvDir(envHome); err != nil {
			log.Fatal(err)
		}
		cachedir, err = normalizeDir(envHome)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Use $HOME/.elvoke if it exists
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	cachedir, err = normalizeDir(filepath.Join(homedir, ".elvoke"))
	if err != nil {
		log.Fatal(err)
	}

	info, err := os.Stat(cachedir)
	if err == nil && info.IsDir() {
		return
	}

	// Fallback to $XDG_CACHE_DIR/elvoke
	// It creates the directory if it does not exist.
	cacheBase, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	cachedir, err = normalizeDir(filepath.Join(cacheBase, "elvoke"))
	if err != nil {
		log.Fatal(err)
	}
	mustMkDirAll(cachedir)
}

func mustMkDirAll(s string) {
	if s == "" {
		log.Fatal("directory path is empty")
	}

	normalized, err := normalizeDir(s)
	if err != nil {
		log.Fatal(err)
	}

	if !filepath.IsAbs(normalized) {
		log.Fatal("directory path is not absolute")
	}

	if err := os.MkdirAll(normalized, os.ModePerm); err != nil {
		log.Fatal(err)
	}
}

func normalizeDir(dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("empty directory path")
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	// Basic sanity check: reject paths with parent directory traversals.
	if strings.Contains(absDir, "..") {
		return "", fmt.Errorf("directory path contains invalid components")
	}

	return absDir, nil
}

func validateEnvDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("environment directory path is empty")
	}

	// Normalize the directory to an absolute path.
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to normalize environment directory: %w", err)
	}

	// Reject obvious traversal patterns in ELVOKE_HOME.
	if strings.Contains(absDir, "..") {
		return fmt.Errorf("environment directory path contains invalid components")
	}

	return nil
}

func stampFilename(ident string) string {
	if stampfile != "" {
		return stampfile
	}

	return filepath.Clean(filepath.Join(cachedir, fmt.Sprintf("%s.stamp", ident)))
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
