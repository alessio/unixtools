package dirsnapshots

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDirname    = "dirsnapshots"
	snapshotsDirName = "snapshots"
	version          = 0
)

func Load() (*Backups, error) {
	configDir := ensureConfigDir()
	filename := filepath.Join(configDir, "config.json")
	snapshotsDir := filepath.Join(configDir, snapshotsDirName)

	file, err := os.Open(filename)
	if err != nil && os.IsNotExist(err) {
		return newConfig(snapshotsDir), nil
	} else if err != nil {
		return nil, fmt.Errorf("couldn't load Backups: %w", err)
	}

	defer file.Close()

	var b Backups
	if err := json.NewDecoder(file).Decode(&b); err != nil {
		return nil, fmt.Errorf("couldn't decode configuration: %w", err)
	}

	if b.Version != version {
		return nil, fmt.Errorf("incompatbile configuration format: %d", b.Version)
	}

	b.snapshotsDir = snapshotsDir

	return &b, nil
}

func Save(b *Backups) error {
	configDir := ensureConfigDir()
	filename := filepath.Join(configDir, "b.json")

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("couldn't save configuration file %q: %w", filename, err)
	}

	if err := json.NewEncoder(file).Encode(b); err != nil {
		return fmt.Errorf("couldn't save configuration: %w", err)
	}

	return nil
}

type Backups struct {
	Snapshots    map[string][]string
	Version      uint8
	snapshotsDir string
}

func newConfig(snapshotsDir string) *Backups {
	return &Backups{
		Snapshots:    make(map[string][]string),
		Version:      version,
		snapshotsDir: snapshotsDir,
	}
}

func (b *Backups) PushDir(orig, bak string) {
	b.Snapshots[orig] = append(b.Snapshots[orig], bak)
}

func (b *Backups) PopDir(orig string) (string, bool) {
	if snapshots := b.Snapshots[orig]; len(snapshots) == 0 {
		return "", false
	}

	var elem string
	elem, b.Snapshots[orig] = b.Snapshots[orig][len(b.Snapshots[orig])-1], b.Snapshots[orig][:len(b.Snapshots[orig])-1]

	return elem, true
}

func (b *Backups) SnapshotsDir() string { return b.snapshotsDir }

// ensureConfigDir ensures that the user's Backups directory
// is created and returns its absolute path.
func ensureConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	snapshotsDir := filepath.Join(configDir, configDirname, snapshotsDirName)
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		panic(fmt.Errorf("couldn't create %q: %w", snapshotsDir, err))
	}

	return configDir
}
