package hdiutil

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// OptFn is a function type that returns a value of type T when called.
// It is used to lazily compute configuration options after validation.
type OptFn[T string | []string] func() T

// Config holds the configuration for creating a DMG disk image.
type Config struct {
	// VolumeName is the name of the mounted volume. If empty, it defaults to the output filename without extension.
	VolumeName string `json:"volume_name,omitempty"`
	// VolumeSizeMb specifies the volume size in megabytes. If zero, hdiutil determines the size automatically.
	VolumeSizeMb int64 `json:"volume_size_mb,omitempty"`
	// SandboxSafe enables sandbox-safe mode. Cannot be used with APFS filesystem.
	SandboxSafe bool `json:"sandbox_safe,omitempty"`
	// Bless marks the volume as bootable.
	Bless bool `json:"bless,omitempty"`
	// FileSystem specifies the filesystem type (e.g., "HFS+", "APFS"). Defaults to "HFS+".
	FileSystem string `json:"filesystem,omitempty"`
	// SigningIdentity specifies the signing identity to use.
	SigningIdentity string `json:"signing_identity,omitempty"`
	// NotarizeCredentials contains credentials for Apple notarization.
	NotarizeCredentials string `json:"notarize_credentials,omitempty"`
	// ImageFormat specifies the DMG format (e.g., "UDZO", "UDBZ", "ULFO", "ULMO"). Defaults to "UDZO".
	ImageFormat string `json:"image_format,omitempty"`

	// HDIUtilVerbosity controls the verbosity level of hdiutil output.
	HDIUtilVerbosity int `json:"hdiutil_verbosity,omitempty"`

	// OutputPath is the destination path for the created DMG file. Must have .dmg extension.
	OutputPath string `json:"output_path,omitempty"`
	// SourceDir is the directory containing files to include in the DMG.
	SourceDir string `json:"source_dir,omitempty"`

	// Simulate enables dry-run mode without actually creating the DMG.
	Simulate bool `json:"simulate,omitempty"`

	// Checksum specifies the hash algorithm for generating a checksum file alongside the DMG.
	// Supported values: "SHA256", "SHA512". If empty, no checksum is generated.
	Checksum string `json:"checksum,omitempty"`

	// ExcludePatterns is a list of glob patterns for files to exclude from the DMG.
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`

	valid bool

	// FilesystemOpts returns the hdiutil arguments for the configured filesystem.
	// Only available after calling Validate.
	FilesystemOpts OptFn[[]string] `json:"-"`
	// ImageFormatOpts returns the hdiutil arguments for the configured image format.
	// Only available after calling Validate.
	ImageFormatOpts OptFn[[]string] `json:"-"`
	// VolumeSizeOpts returns the hdiutil arguments for the configured volume size.
	// Only available after calling Validate.
	VolumeSizeOpts OptFn[[]string] `json:"-"`
	// VolumeNameOpt returns the resolved volume name.
	// Only available after calling Validate.
	VolumeNameOpt OptFn[string] `json:"-"`
}

// FromJSON populates the Config from a JSON reader.
func (c *Config) FromJSON(r io.Reader) error {
	var tmp Config
	if err := json.NewDecoder(r).Decode(&tmp); err != nil {
		return err
	}
	// Ensure validation is required after (re)loading.
	tmp.valid = false
	tmp.FilesystemOpts = nil
	tmp.ImageFormatOpts = nil
	tmp.VolumeSizeOpts = nil
	tmp.VolumeNameOpt = nil
	*c = tmp
	return nil
}

// ToJSON writes the Config to a JSON writer.
func (c *Config) ToJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(c)
}

// LoadConfig reads the configuration from a JSON file.
func LoadConfig(path string) (*Config, error) {
	// Clean the path to ensure it is normalized.
	// G304: Potential file inclusion via variable.
	// This is a CLI tool and the user is expected to provide a path to the config file.
	// #nosec G304
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	cfg := &Config{}
	if err := cfg.FromJSON(f); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks the configuration for errors and initializes the option functions.
// It must be called before using FilesystemOpts, ImageFormatOpts, VolumeSizeOpts, or VolumeNameOpt.
// Returns an error if:
//   - SourceDir is empty
//   - OutputPath does not have a .dmg extension
//   - ImageFormat is invalid
//   - FileSystem is invalid
//   - SandboxSafe is enabled with APFS filesystem
func (c *Config) Validate() error {
	c.valid = false
	if len(c.SourceDir) == 0 {
		return ErrInvSourceDir
	}

	if c.VolumeSizeMb < 0 {
		return ErrVolumeSize
	}

	if filepath.Ext(c.OutputPath) != ".dmg" {
		return ErrImageFileExt
	}

	if len(c.imageFormatToOpts()) == 0 {
		return ErrInvFormatOpt
	}

	if len(c.filesystemToOpts()) == 0 {
		return ErrInvFilesystemOpt
	}

	// sandbox safe and APFS are mutually exclusive
	if c.SandboxSafe && strings.ToUpper(c.FileSystem) == "APFS" {
		return ErrSandboxAPFS
	}

	if c.Checksum != "" {
		switch strings.ToUpper(c.Checksum) {
		case "SHA256", "SHA512":
			// valid
		default:
			return ErrInvChecksumAlgo
		}
	}

	c.valid = true

	c.FilesystemOpts = c.validWrapper(c.filesystemToOpts)
	c.ImageFormatOpts = c.validWrapper(c.imageFormatToOpts)
	c.VolumeSizeOpts = c.validWrapper(c.volumeSizeToOpts)
	c.VolumeNameOpt = c.validWrapperStr(c.volumeNameToOpt)

	return nil
}

// volumeNameToOpt returns the volume name, defaulting to the output filename without extension.
func (c *Config) volumeNameToOpt() string {
	if len(c.VolumeName) == 0 {
		return strings.TrimSuffix(filepath.Base(c.OutputPath), ".dmg")
	}

	return c.VolumeName
}

// validWrapper wraps a function to ensure Validate has been called before execution.
// Panics if called before validation.
func (c *Config) validWrapper(fn func() []string) OptFn[[]string] {
	return func() []string {
		if !c.valid {
			panic("state is corrupted, Validate() must be called first")
		}
		return fn()
	}
}

// validWrapperStr wraps a string-returning function to ensure Validate has been called before execution.
// Panics if called before validation.
func (c *Config) validWrapperStr(fn func() string) OptFn[string] {
	return func() string {
		if !c.valid {
			panic("state is corrupted, Validate() must be called first")
		}
		return fn()
	}
}

// filesystemToOpts returns the hdiutil arguments for the configured filesystem.
// Supports "HFS+" (default) and "APFS". Returns nil for unsupported filesystems.
func (c *Config) filesystemToOpts() []string {
	switch strings.ToUpper(c.FileSystem) {
	case "", "HFS+":
		return []string{"-fs", "HFS+", "-fsargs", "-c c=64,a=16,e=16"}
	case "APFS":
		return []string{"-fs", "APFS"}
	default:
		return nil
	}
}

// imageFormatToOpts returns the hdiutil arguments for the configured image format.
// Supports "UDZO" (default), "UDBZ", "ULFO", and "ULMO". Returns nil for unsupported formats.
func (c *Config) imageFormatToOpts() []string {
	format := strings.ToUpper(c.ImageFormat)
	switch format {
	case "", "UDZO":
		return []string{"-format", "UDZO", "-imagekey", "zlib-level=9"}
	case "UDBZ":
		return []string{"-format", "UDBZ", "-imagekey", "bzip2-level=9"}
	case "ULFO", "ULMO":
		return []string{"-format", format}
	default:
		return nil
	}
}

// volumeSizeToOpts returns the hdiutil arguments for the configured volume size.
// Returns nil if VolumeSizeMb is zero or negative.
func (c *Config) volumeSizeToOpts() []string {
	if c.VolumeSizeMb > 0 {
		return []string{"-size", fmt.Sprintf("%dm", c.VolumeSizeMb)}
	}

	return nil
}
