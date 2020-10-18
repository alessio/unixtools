package dirbaks

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

const (
	configDirname    = "dirbaks"
	snapshotsDirName = "snapshots"
	version          = 0
)

func Load() *Config {
	configDir := ensureConfigDir()
	filename := filepath.Join(configDir, "config.json")
	snapshotsDir := filepath.Join(configDir, snapshotsDirName)

	file, err := os.Open(filename)
	if err != nil && os.IsNotExist(err) {
		return new(snapshotsDir)
	} else if err != nil {
		log.Fatalf("couldn't load Config: %v", err)
	}

	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		log.Fatalf("couldn't decode configuration: %v", err)
	}

	if config.Version != version {
		log.Fatalf("incompatbile configuration format: %d", config.Version)
	}

	config.snapshotsDir = snapshotsDir

	return &config
}

func Save(config *Config) {
	configDir := ensureConfigDir()
	filename := filepath.Join(configDir, "config.json")

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("couldn't save configuration file %q: %v", filename, err)
	}

	if err := json.NewEncoder(file).Encode(config); err != nil {
		log.Fatalf("couldn't save configuration: %v", err)
	}
}

type Config struct {
	Snapshots    map[string][]string
	Version      uint8
	snapshotsDir string
}

func new(snapshotsDir string) *Config {
	return &Config{
		Snapshots:    make(map[string][]string),
		Version:      version,
		snapshotsDir: snapshotsDir,
	}
}

func (c *Config) PushDir(orig, bak string) {
	c.Snapshots[orig] = append(c.Snapshots[orig], bak)
}

func (c *Config) PopDir(orig string) (string, bool) {
	if snapshots := c.Snapshots[orig]; len(snapshots) == 0 {
		return "", false
	}

	var elem string
	elem, c.Snapshots[orig] = c.Snapshots[orig][len(c.Snapshots[orig])-1], c.Snapshots[orig][:len(c.Snapshots[orig])-1]

	return elem, true
}

func (c *Config) SnapshotsDir() string { return c.snapshotsDir }

// ensureConfigDir ensures that the user's Config directory
// is created and returns its absolute path.
func ensureConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	configDir = filepath.Join(configDir, configDirname)
	if err := os.MkdirAll(filepath.Join(configDir, snapshotsDirName), 0755); err != nil {
		log.Fatalf("couldn't create %q: %v", configDir, err)
	}

	return configDir
}
