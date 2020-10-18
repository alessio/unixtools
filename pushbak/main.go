package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/alessio/tools/internal/dirbaks"
	"github.com/alessio/tools/internal/file"
)

var shelveMode bool

func init() {
	flag.BoolVar(&shelveMode, "-shelve", false, "shelve the directory once the backup copy is done")
}

func main() {
	log.SetPrefix("pushbak: ")
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

	err = backupDirectory(target, config)

	dirbaks.Save(config)

	if err != nil {
		log.Fatalln(err)
	}
}

func backupDirectory(target string, config *dirbaks.Config) error {
	backupDir, err := ioutil.TempDir(config.SnapshotsDir(), "")
	if err != nil {
		return err
	}

	defer config.PushDir(target, backupDir)

	if shelveMode {
		return os.Rename(target, backupDir)
	}

	if err := os.Remove(backupDir); err != nil {
		return err
	}

	return file.CopyDir(target, backupDir)
}
