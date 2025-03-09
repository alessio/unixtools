package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	recursiveMode   bool
	interactiveMode bool
	moveMode        bool
	simulateMode    bool
	verboseMode     bool

	verboseLog *log.Logger
)

func init() {
	flag.CommandLine.SetOutput(os.Stderr)
	flag.CommandLine.Usage = usage

	flag.BoolVar(&moveMode, "m", false, "move files matching PATTERN to REPLACE")
	flag.BoolVar(&interactiveMode, "I", false, "prompt before every overwrite")
	flag.BoolVar(&recursiveMode, "R", false, "search files under each directory recursively")
	flag.BoolVar(&simulateMode, "simulate", false, "print changes that are supposed to be done, but don't actually make any")
	flag.BoolVar(&verboseMode, "verbose", false, "enable verbose output")
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("refiles: ")
	flag.Parse()

	if flag.NArg() < 2 {
		log.Fatalln("wrong number of arguments")
	}

	pattern, err := regexp.Compile(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}

	replace := flag.Arg(1)
	verboseWriter := io.Discard

	if verboseMode || simulateMode {
		verboseWriter = os.Stderr
	}

	verboseLog = log.New(verboseWriter, "refiles: ", 0)

	var dirs = []string{filepath.Dir(".")}
	if flag.NArg() > 2 {
		dirs = flag.Args()[2:]
	}

	var wg sync.WaitGroup

	for _, dir := range dirs {
		wg.Add(1)

		go func(d string, pattern *regexp.Regexp, replace string) {
			defer wg.Done()
			walkDirectory(d, pattern, replace)
		}(dir, pattern, replace)
	}

	wg.Wait()
}

func walkDirectory(dir string, pattern *regexp.Regexp, replace string) {
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("cannot access %q: %v", path, err)
			return nil
		}

		if path == dir || (recursiveMode && info.IsDir()) {
			// nil instead of SkipDir as contents of the root directory and
			// directories in recursive mode must be recursively processed
			return nil
		}

		if info.IsDir() && !recursiveMode {
			verboseLog.Printf("skipping %q", path)
			return filepath.SkipDir
		}

		rename(path, filepath.Join(filepath.Dir(path),
			replaceFilename(pattern, info.Name(), replace)), interactiveMode, simulateMode)

		return nil
	}); err != nil {
		verboseLog.Printf("error walking the path %q: %v", dir, err)
	}
}

func replaceFilename(pattern *regexp.Regexp, filename, replace string) string {
	if !moveMode {
		return pattern.ReplaceAllString(filename, replace)
	}

	if !pattern.MatchString(filename) {
		return filename
	}

	result := []byte{}
	for _, submatches := range pattern.FindAllStringSubmatchIndex(filename, -1) {
		result = pattern.ExpandString(result, replace, filename, submatches)
	}

	return string(result)
}

func rename(orig, new string, interactive, simulate bool) {
	if orig == new { // skip if noop
		return
	}

	verboseLog.Printf("%q -> %q", orig, new)

	if interactive {
		if _, err := os.Stat(new); err == nil && !confirmPrompt(orig, new) {
			return
		}
	}

	if simulate {
		return
	}

	if err := os.Rename(orig, new); err != nil {
		log.Printf("couldn't rename %s: %v", orig, err)
	}
}

func confirmPrompt(from, to string) bool {
	reader := bufio.NewReader(os.Stdin)
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "rename %q to %q?", from, to)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	if response := strings.ToLower(strings.TrimSpace(response)); response == "y" || response == "yes" {
		return true
	}

	return false
}

func usage() {
	fmt.Fprintln(flag.CommandLine.Output(), `Usage: refiles [OPTIONS] PATTERN REPLACE [DIRECTORY]...
Rename files that match a given pattern.`)
	flag.PrintDefaults()
	fmt.Fprintln(flag.CommandLine.Output(), `
This program replaces the matched patten in filenames with the
replace pattern. The '-m' option replaces the complete filename
with the replace pattern.

With no DIRECTORY, it runs over the current working directory.

Examples:

Replace spaces in filenames with underlines:
  refiles ' ' '_'

Move files like 6.1.001 to vim-6.1-001.patch:
  refiles -m '^6.1.(\d{3})$' 'vim-6.1-$1.patch'

Written by Alessio Treglia <alessio@debian.org>.
Inspired by Gustavo Niemeyer's remv: http://niemeyer.net/remv.`)
}
