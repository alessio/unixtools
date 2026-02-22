package hdiutil

import (
	"errors"
	"fmt"
	"io"
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
	// ErrUnsafeArg indicates a config value contains characters unsafe for command arguments.
	ErrUnsafeArg = errors.New("argument contains unsafe characters")
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
// Each method corresponds to a specific allowed command, ensuring that only
// known binaries can be invoked and satisfying static analysis requirements.
type CommandExecutor interface {
	Hdiutil(args ...string) error
	HdiutilOutput(args ...string) (string, error)
	Codesign(args ...string) error
	Xcrun(args ...string) error
	XcrunOutput(args ...string) (string, error)
	Chmod(args ...string) error
	Bless(args ...string) error
}

type realCommandExecutor struct{}

func (e *realCommandExecutor) Hdiutil(args ...string) error {
	cmd := exec.Command("hdiutil", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *realCommandExecutor) HdiutilOutput(args ...string) (string, error) {
	output, err := exec.Command("hdiutil", args...).CombinedOutput()
	return string(output), err
}

func (e *realCommandExecutor) Codesign(args ...string) error {
	cmd := exec.Command("codesign", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *realCommandExecutor) Xcrun(args ...string) error {
	cmd := exec.Command("xcrun", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *realCommandExecutor) XcrunOutput(args ...string) (string, error) {
	output, err := exec.Command("xcrun", args...).CombinedOutput()
	return string(output), err
}

func (e *realCommandExecutor) Chmod(args ...string) error {
	cmd := exec.Command("chmod", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *realCommandExecutor) Bless(args ...string) error {
	cmd := exec.Command("bless", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

	return r.runBless("--folder", r.mountDir)
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

	if err := r.runCodesign("-s", r.signOpt, r.finalDmg); err != nil {
		return fmt.Errorf("%w: codesign command failed: %v", ErrCodesignFailed, err)
	}

	if err := r.runCodesign(
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
	if err := r.runXcrun("notarytool", "submit",
		r.finalDmg, "--keychain-profile", r.notarizeOpt,
	); err != nil {
		return fmt.Errorf("%w: notarization failed: %v", ErrNotarizeFailed, err)
	}

	verboseLog.Println("Stapling the notarization ticket")
	if output, err := r.runXcrunOutput(
		"stapler", "staple", r.finalDmg); err != nil {
		return fmt.Errorf("%w: stapler failed: %v (output: %s)", ErrNotarizeFailed, err, output)
	}

	verboseLog.Println("Notarization complete")

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
	if len(args) == 0 || r.HDIUtilVerbosity <= 0 {
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
	r.finalDmg = filepath.Clean(r.OutputPath)

	r.volNameOpt = r.VolumeNameOpt()
	r.formatOpts = r.ImageFormatOpts()
	r.fsOpts = r.FilesystemOpts()
	r.sizeOpts = r.VolumeSizeOpts()

	// create a working directory
	tmpDir, err := os.MkdirTemp("", "mkdmg-")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateDir, err)
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

	return nil
}

// fixPermissions removes group and other write permissions from the mounted volume.
// This is called automatically before detaching the image and is idempotent.
func (r *Runner) fixPermissions() error {
	if r.permFixed {
		return nil
	}

	verboseLog.Println("Fixing permissions")
	if err := r.runChmod("-Rf", "go-w", r.mountDir); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	r.permFixed = true
	return nil
}

// Command runner helpers — each logs, checks simulate mode, and delegates
// to the corresponding typed CommandExecutor method.

func (r *Runner) runHdiutil(args ...string) error {
	verboseLog.Println("Running 'hdiutil", args)
	if r.Simulate {
		return nil
	}
	return r.executor.Hdiutil(args...)
}

func (r *Runner) runHdiutilOutput(args ...string) (string, error) {
	verboseLog.Println("Running 'hdiutil", args)
	if r.Simulate {
		return "", nil
	}
	return r.executor.HdiutilOutput(args...)
}

func (r *Runner) runCodesign(args ...string) error {
	verboseLog.Println("Running 'codesign", args)
	if r.Simulate {
		return nil
	}
	return r.executor.Codesign(args...)
}

func (r *Runner) runXcrun(args ...string) error {
	verboseLog.Println("Running 'xcrun", args)
	if r.Simulate {
		return nil
	}
	return r.executor.Xcrun(args...)
}

func (r *Runner) runXcrunOutput(args ...string) (string, error) {
	verboseLog.Println("Running 'xcrun", args)
	if r.Simulate {
		return "", nil
	}
	return r.executor.XcrunOutput(args...)
}

func (r *Runner) runChmod(args ...string) error {
	verboseLog.Println("Running 'chmod", args)
	if r.Simulate {
		return nil
	}
	return r.executor.Chmod(args...)
}

func (r *Runner) runBless(args ...string) error {
	verboseLog.Println("Running 'bless", args)
	if r.Simulate {
		return nil
	}
	return r.executor.Bless(args...)
}
