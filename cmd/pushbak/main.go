package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/alessio/unixtools/internal/dirsnapshots"
	"github.com/alessio/unixtools/internal/file"
)

var shelveMode bool

func init() {
	flag.BoolVar(&shelveMode, "shelve", false, "shelve the directory once the backup copy is done")
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

	backups, err := dirsnapshots.Load()
	if err != nil {
		log.Fatalln(err)
	}

	if err := backupDirectory(target, backups); err != nil {
		log.Fatalln(err)
	}

	if err := dirsnapshots.Save(backups); err != nil {
		log.Fatalln(err)
	}
}

func backupDirectory(target string, backups *dirsnapshots.Backups) error {
	backupDir, err := ioutil.TempDir(backups.SnapshotsDir(), "")
	if err != nil {
		return err
	}

	defer backups.PushDir(target, backupDir)

	if shelveMode {
		return os.Rename(target, backupDir)
	}

	if err := os.Remove(backupDir); err != nil {
		return err
	}

	return file.CopyDir(target, backupDir)
}
