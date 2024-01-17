package instpkg

import (
	"errors"
	"fmt"
)

type PackageInfo struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	Version    string `json:"version"`
}

func (p PackageInfo) Validate() error {
	if p.Identifier == "" || p.Name == "" || p.Version == "" {
		return errors.New("PackageInfo: name, version, and identifier are all required")
	}

	return nil
}

type Signing struct {
	// Developer ID Application + Developer ID Installer
	// https://developer.apple.com/account/resources/certificates/list
	Identity         string `json:"identity"`
	SkipCode         bool   `json:"skip_code"`
	SkipInstaller    bool   `json:"skip_installer"`
	SkipNotarization bool   `json:"skip_notarization"`
}

type Configuration struct {
	Package PackageInfo `json:"package"`
	Signing Signing     `json:"signing"`

	// PackageOutputFilename string
	SourceDir       string `json:"source_dir"`
	InstallLocation string `json:"install_location"`
	ScriptsDir      string `json:"scripts_dir"`
}

func (c Configuration) Validate() error {
	if err := c.Package.Validate(); err != nil {
		return fmt.Errorf("configuration: %v", err)
	}

	if c.SourceDir == "" {
		return errors.New("configuration: source_dir is not set")
	}

	if c.InstallLocation == "" {
		return errors.New("configuration: install_location is not set")
	}

	return nil
}

func DefaultConfiguration() Configuration {
	return Configuration{
		SourceDir:       "build",
		InstallLocation: "./Library/",
	}
}
