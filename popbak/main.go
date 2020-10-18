package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alessio/tools/internal/dirbaks"
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

	config := dirbaks.Load()
	err = restoreDirectory(target, config)

	dirbaks.Save(config)

	if err != nil {
		log.Fatalln(err)
	}
}

func restoreDirectory(target string, config *dirbaks.Config) error {
	orig, ok := config.PopDir(target)
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
