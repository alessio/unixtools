package hdiutil

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Error variables for common failure conditions during DMG creation.
var (
	// ErrInvSourceDir indicates the source directory is empty or invalid.
	ErrInvSourceDir = errors.New("invalid source directory")
	// ErrVolumeSize indicates that a negative volume size.
	ErrVolumeSize = errors.New("volume size must be >= 0")
	// ErrInvFormatOpt indicates an unsupported image format was specified.
	ErrInvFormatOpt = errors.New("invalid image format")
	// ErrInvFilesystemOpt indicates an unsupported filesystem type was specified.
	ErrInvFilesystemOpt = errors.New("invalid image filesystem")
	// ErrCreateDir indicates a failure to create a temporary working directory.
	ErrCreateDir = errors.New("couldn't create directory")
	// ErrImageFileExt indicates the output path doesn't have a .dmg extension.
	ErrImageFileExt = errors.New("output file must have a .dmg extension")
	// ErrMountImage indicates failure to attach/mount the disk image.
	ErrMountImage = errors.New("couldn't attach disk image")
	// ErrCodesignFailed indicates the codesign command failed or signature verification failed.
	ErrCodesignFailed = errors.New("codesign command failed")
	// ErrNotarizeFailed indicates Apple notarization or stapling failed.
	ErrNotarizeFailed = errors.New("notarization failed")
	// ErrSandboxAPFS indicates an attempt to create a sandbox-safe APFS image, which is unsupported.
	ErrSandboxAPFS = errors.New("creating an APFS disk image that is sandbox safe is not supported")
	// ErrNeedInit indicates Runner.Setup was not called before attempting operations.
	ErrNeedInit = errors.New("runner not properly initialized, call Setup() first")
	// ErrChecksum indicates failure to generate the checksum file.
	ErrChecksum = errors.New("failed to generate checksum")
	// ErrInvChecksumAlgo indicates an unsupported checksum algorithm was specified.
	ErrInvChecksumAlgo = errors.New("invalid checksum algorithm, supported: SHA256, SHA512")
	// ErrExcludeCopy indicates failure to copy files while applying exclusion patterns.
	ErrExcludeCopy = errors.New("failed to copy files with exclusions")
)

var (
	verboseLog *log.Logger
)

func init() {
	verboseLog = log.New(io.Discard, "hdiutil: ", 0)
}

// SetLogWriter configures the output writer for verbose logging.
// By default, verbose logging is discarded. Pass os.Stdout or os.Stderr
// to enable logging output.
func SetLogWriter(w io.Writer) {
	verboseLog.SetOutput(w)
}

// CommandExecutor defines the interface for executing external commands.
type CommandExecutor interface {
	Run(name string, args ...string) error
	RunOutput(name string, args ...string) (string, error)
}

type realCommandExecutor struct{}

func (e *realCommandExecutor) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *realCommandExecutor) RunOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Option is a functional option for configuring a Runner.
type Option func(*Runner)

// WithExecutor sets a custom command executor for testing.
func WithExecutor(e CommandExecutor) Option {
	return func(r *Runner) {
		r.executor = e
	}
}

// New creates a new Runner with the provided configuration.
// The returned Runner must have Setup called before use.
func New(c *Config, opts ...Option) *Runner {
	r := &Runner{
		Config:   c,
		executor: &realCommandExecutor{},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Runner orchestrates the DMG creation process, including image creation,
// mounting, file copying, code signing, and notarization.
type Runner struct {
	*Config

	executor CommandExecutor

	formatOpts  []string
	sizeOpts    []string
	fsOpts      []string
	volNameOpt  string
	signOpt     string
	notarizeOpt string

	srcDir   string
	tmpDir   string
	mountDir string

	tmpDmg   string
	finalDmg string

	permFixed bool

	cleanupFuncs []func()
}

// Setup validates the configuration and initializes the Runner for use.
// It creates temporary directories and prepares internal state.
// Must be called before Start or any other operation methods.
// Returns an error if validation fails or temporary directory creation fails.
func (r *Runner) Setup() error {
	return r.init()
}

// Cleanup removes temporary files and directories created during the DMG build process.
// Should be called when the Runner is no longer needed, typically via defer.
func (r *Runner) Cleanup() {
	for _, f := range r.cleanupFuncs {
		f()
	}
}

// Start begins the DMG creation process by creating a temporary writable disk image.
// It uses either the standard or sandbox-safe creation method based on configuration.
// Returns ErrNeedInit if Setup was not called first.
func (r *Runner) Start() error {
	if r.tmpDir == "" || r.tmpDmg == "" {
		return ErrNeedInit
	}

	if r.SandboxSafe {
		return r.createTempImageSandboxSafe()
	}

	return r.createTempImage()
}

// AttachDiskImage mounts the temporary disk image and stores the mount point.
// The image is attached with -nobrowse (hidden from Finder) and -noverify flags.
// Returns ErrMountImage if it fails or the mount point cannot be determined.
// AttachDiskImage mounts the temporary disk image and stores the mount point.
// The image is attached with -nobrowse (hidden from Finder) and -noverify flags.
// Returns ErrMountImage if it fails or the mount point cannot be determined.
func (r *Runner) AttachDiskImage() error {
	if r.Simulate {
		r.mountDir = filepath.Join(r.tmpDir, "SIMULATED_MOUNT")
		return nil
	}
	output, err := r.runHdiutilOutput("attach", "-nobrowse", "-noverify", r.tmpDmg)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrMountImage, output)
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if idx := strings.Index(line, "/Volumes/"); idx != -1 {
			r.mountDir = strings.TrimSpace(line[idx:])
			return nil
		}
	}

	return fmt.Errorf("%w: couldn't find mount point: %q", ErrMountImage, output)
}

// DetachDiskImage unmounts the disk image after fixing file permissions.
// Should be called after all modifications to the mounted volume are complete.
func (r *Runner) DetachDiskImage() error {
	if r.Simulate {
		verboseLog.Println("Simulating detach of disk image")
		return nil
	}
	if err := r.fixPermissions(); err != nil {
		return err
	}
	return r.runHdiutil("detach", r.mountDir)
}

// Bless marks the mounted volume as bootable using the bless command.
// This operation is skipped if Config.Bless is false or if SandboxSafe mode is enabled.
// Bless is typically used for bootable installer images.
func (r *Runner) Bless() error {
	if err := r.fixPermissions(); err != nil {
		return err
	}
	if !r.Config.Bless {
		return nil
	}

	if r.SandboxSafe {
		verboseLog.Println("Skipping blessing on sandbox safe images")
		return nil
	}

	return r.runCommand("bless", "--folder", r.mountDir)
}

// FinalizeDMG converts the temporary writable image to the final compressed format
// specified in the configuration (e.g., UDZO, UDBZ, ULFO, ULMO).
func (r *Runner) FinalizeDMG() error {
	return r.runHdiutil(r.setHdiutilVerbosity(slices.Concat(
		[]string{"convert", r.tmpDmg},
		r.formatOpts,
		[]string{"-o", r.finalDmg}),
	)...)
}

// Codesign signs the final DMG with the specified signing identity and verifies the signature.
// If no SigningIdentity is configured, this method returns nil without action.
// Returns ErrCodesignFailed if signing or verification fails.
func (r *Runner) Codesign() error {
	if len(r.signOpt) == 0 {
		verboseLog.Println("Skipping codesign")
		return nil
	}

	if err := r.runCommand("codesign", "-s", r.signOpt, r.finalDmg); err != nil {
		return fmt.Errorf("%w: codesign command failed: %v", ErrCodesignFailed, err)
	}

	if err := r.runCommand("codesign",
		"--verify", "--deep", "--strict", "--verbose=2", r.finalDmg); err != nil {
		return fmt.Errorf("%w: the signature seems invalid: %v", ErrCodesignFailed, err)
	}

	verboseLog.Println("codesign complete")
	return nil
}

// Notarize submits the DMG to Apple's notarization service and staples the ticket.
// Requires NotarizeCredentials to be set with a valid keychain profile name.
// If no credentials are configured, this method returns nil without action.
// Returns ErrNotarizeFailed if notarization submission or stapling fails.
func (r *Runner) Notarize() error {
	if len(r.notarizeOpt) == 0 {
		verboseLog.Println("Skipping notarization")
		return nil
	}

	verboseLog.Println("Start notarization")
	if err := r.runCommand("xcrun", "notarytool", "submit",
		r.finalDmg, "--keychain-profile", r.notarizeOpt,
	); err != nil {
		return fmt.Errorf("%w: notarization failed: %v", ErrNotarizeFailed, err)
	}

	verboseLog.Println("Stapling the notarization ticket")
	if output, err := r.runCommandOutput(
		"xcrun", "stapler", "staple", r.finalDmg); err != nil {
		return fmt.Errorf("%w: stapler failed: %v (output: %s)", ErrNotarizeFailed, err, output)
	}

	verboseLog.Println("Notarization complete")

	return nil
}

// GenerateChecksum computes a hash of the final DMG and writes it to a file.
// The output file is named after the DMG with a hash-specific extension (e.g., ".sha256").
// If Config.Checksum is empty, this method returns nil without action.
func (r *Runner) GenerateChecksum() error {
	if r.Checksum == "" {
		return nil
	}

	if r.Simulate {
		verboseLog.Println("Simulating checksum generation")
		return nil
	}

	var h hash.Hash
	var ext string
	switch strings.ToUpper(r.Checksum) {
	case "SHA256":
		h = sha256.New()
		ext = ".sha256"
	case "SHA512":
		h = sha512.New()
		ext = ".sha512"
	default:
		return fmt.Errorf("%w: %s", ErrInvChecksumAlgo, r.Checksum)
	}

	f, err := os.Open(r.finalDmg)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrChecksum, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("%w: %v", ErrChecksum, err)
	}

	sum := hex.EncodeToString(h.Sum(nil))
	basename := filepath.Base(r.finalDmg)
	line := fmt.Sprintf("%s  %s\n", sum, basename)

	checksumPath := r.finalDmg + ext
	if err := os.WriteFile(checksumPath, []byte(line), 0644); err != nil {
		return fmt.Errorf("%w: %v", ErrChecksum, err)
	}

	verboseLog.Printf("Checksum written to %s\n", checksumPath)
	return nil
}

// createTempImage creates a writable temporary disk image using hdiutil create.
// The image is created with the configured filesystem, size, and volume name,
// populated with files from the source directory.
func (r *Runner) createTempImage() error {
	args := slices.Concat([]string{"create"},
		r.fsOpts,
		r.sizeOpts,
		[]string{"-format", "UDRW", "-volname", r.volNameOpt, "-srcfolder", r.srcDir, r.tmpDmg},
	)

	return r.runHdiutil(r.setHdiutilVerbosity(args)...)
}

// createTempImageSandboxSafe creates a sandbox-safe temporary disk image.
// Uses hdiutil makehybrid followed by convert, which produces images that
// can be opened in sandboxed applications.
func (r *Runner) createTempImageSandboxSafe() error {
	args1 := r.setHdiutilVerbosity([]string{"makehybrid",
		"-default-volume-name", r.volNameOpt, "-hfs", "-r", r.tmpDmg, r.srcDir})
	if err := r.runHdiutil(args1...); err != nil {
		return err
	}

	args2 := r.setHdiutilVerbosity([]string{"convert",
		r.tmpDmg, "-format", "UDRW", "-ov", "-o", r.tmpDmg})

	return r.runHdiutil(args2...)
}

// setHdiutilVerbosity inserts the appropriate verbosity flag into hdiutil arguments.
// Verbosity levels: 1 = quiet, 2 = verbose, 3+ = debug.
// Returns the original args if verbosity is 0 or args is empty.
func (r *Runner) setHdiutilVerbosity(args []string) []string {
	if len(args) == 0 || r.HDIUtilVerbosity == 0 {
		return args
	}

	var val string

	switch r.HDIUtilVerbosity {
	case 1:
		val = "-quiet"
	case 2:
		val = "-verbose"
	default:
		val = "-debug"
	}

	switch args[0] {
	case "create", "makehybrid", "convert":
		return slices.Insert(args, 1, val)
	default:
		return slices.Insert(args, 0, val)
	}
}

// init validates configuration, resolves paths, and creates the temporary working directory.
// Called by Setup to prepare the Runner for DMG creation operations.
func (r *Runner) init() error {
	if err := r.Validate(); err != nil {
		return err
	}

	r.srcDir = filepath.Clean(r.SourceDir)
	r.finalDmg = r.OutputPath

	r.volNameOpt = r.VolumeNameOpt()
	r.formatOpts = r.ImageFormatOpts()
	r.fsOpts = r.FilesystemOpts()
	r.sizeOpts = r.VolumeSizeOpts()

	// create a working directory
	tmpDir, err := os.MkdirTemp("", "mkdmg-")
	if err != nil {
		return fmt.Errorf("%v: %w", ErrCreateDir, err)
	}
	r.tmpDir = tmpDir

	r.cleanupFuncs = []func(){}
	r.cleanupFuncs = append(r.cleanupFuncs, func() {
		if r.tmpDir != "" {
			verboseLog.Println("Removing temporary directory: ", r.tmpDir)
			_ = os.RemoveAll(r.tmpDir)
		}
	})

	r.tmpDmg = filepath.Join(tmpDir, "temp.dmg")
	// signingIdentity
	r.signOpt = r.SigningIdentity
	r.notarizeOpt = r.NotarizeCredentials

	// If exclude patterns are set, copy source to a staging directory
	// skipping files that match any pattern.
	if len(r.ExcludePatterns) > 0 {
		stagingDir := filepath.Join(r.tmpDir, "staging")
		if err := r.copyWithExclusions(r.srcDir, stagingDir); err != nil {
			return fmt.Errorf("%w: %v", ErrExcludeCopy, err)
		}
		r.srcDir = stagingDir
	}

	return nil
}

// copyWithExclusions copies the source directory tree to dst, skipping files
// whose base name matches any of the configured ExcludePatterns.
func (r *Runner) copyWithExclusions(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		base := d.Name()
		for _, pattern := range r.ExcludePatterns {
			matched, matchErr := filepath.Match(pattern, base)
			if matchErr != nil {
				return fmt.Errorf("bad exclude pattern %q: %w", pattern, matchErr)
			}
			if matched {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		// Prevent path traversal: ensure target stays within dst.
		if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator), filepath.Clean(dst)+string(os.PathSeparator)) &&
			filepath.Clean(target) != filepath.Clean(dst) {
			return fmt.Errorf("path traversal detected: %q escapes destination %q", rel, dst)
		}

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		return copyFile(path, target)
	})
}

// copyFile copies a single file from src to dst, preserving permissions.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// fixPermissions removes group and other write permissions from the mounted volume.
// This is called automatically before detaching the image and is idempotent.
func (r *Runner) fixPermissions() error {
	if r.permFixed {
		return nil
	}

	verboseLog.Println("Fixing permissions")
	if err := r.runCommand("chmod", []string{
		"-Rf", "go-w", r.mountDir,
	}...); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	r.permFixed = true
	return nil
}

// runHdiutil executes hdiutil with the given arguments.
// In simulation mode, logs the command without executing it.
func (r *Runner) runHdiutil(args ...string) error {
	return r.runCommand("hdiutil", args...)
}

// runHdiutilOutput executes hdiutil with the given arguments and returns the combined output.
// In simulation mode, logs the command and returns an empty string.
func (r *Runner) runHdiutilOutput(args ...string) (string, error) {
	return r.runCommandOutput("hdiutil", args...)
}

// runCommand executes an external command.
func (r *Runner) runCommand(name string, args ...string) error {
	verboseLog.Println("Running '", name, args)
	if r.Simulate {
		return nil
	}
	return r.executor.Run(name, args...)
}

// runCommandOutput executes an external command and returns the combined output as a string.
func (r *Runner) runCommandOutput(name string, args ...string) (string, error) {
	verboseLog.Println("Running '", name, args)
	if r.Simulate {
		return "", nil
	}
	return r.executor.RunOutput(name, args...)
}
