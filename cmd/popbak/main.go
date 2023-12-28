package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"al.essio.dev/pkg/tools/internal/dirsnapshots"
)

func main() {
	log.SetPrefix("popbak: ")
	log.SetFlags(0)
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatalf("invalid arguments")
	}

	target, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}

	backups, err := dirsnapshots.Load()
	if err != nil {
		log.Fatalln(err)
	}

	if err := restoreDirectory(target, backups); err != nil {
		log.Fatalln(err)
	}

	if err := dirsnapshots.Save(backups); err != nil {
		log.Fatalln(err)
	}
}

func restoreDirectory(target string, backups *dirsnapshots.Backups) error {
	orig, ok := backups.PopDir(target)
	if !ok {
		return fmt.Errorf("no backups available")
	}

	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("couldn't remove %q: %v", target, err)
	}

	if err := os.Rename(orig, target); err != nil {
		return fmt.Errorf("couldn't rename %q: %v", orig, err)
	}

	return nil
}
