package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"al.essio.dev/pkg/tools/internal/dirsnapshots"
	"al.essio.dev/pkg/tools/internal/file"
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

func validateSnapshotBaseDir(baseDir, expectedRoot string) (string, error) {
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	baseEval, err := filepath.EvalSymlinks(baseAbs)
	if err != nil {
		return "", err
	}

	rootAbs, err := filepath.Abs(expectedRoot)
	if err != nil {
		return "", err
	}
	rootEval, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(baseEval)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", os.ErrInvalid
	}

	rel, err := filepath.Rel(rootEval, baseEval)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", os.ErrPermission
	}

	return baseEval, nil
}

func backupDirectory(target string, backups *dirsnapshots.Backups) error {
	snapshotsBaseAbs, err := backups.ResolveSnapshotPath(".")
	if err != nil {
		return err
	}

	snapshotsBaseAbs, err = validateSnapshotBaseDir(snapshotsBaseAbs, backups.SnapshotsDir())
	if err != nil {
		return err
	}

	backupDir, err := os.MkdirTemp(snapshotsBaseAbs, "")
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
