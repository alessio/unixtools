package hdiutil_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"al.essio.dev/cmd/mkdmg/pkg/hdiutil"
)

// mockExecutor records executed commands and allows configuring responses.
type mockExecutor struct {
	commands []executedCommand
	runErr   error
	// runOutputFn allows dynamic responses based on the command.
	runOutputFn func(name string, args ...string) (string, error)
}

type executedCommand struct {
	Name string
	Args []string
}

func (m *mockExecutor) Run(name string, args ...string) error {
	m.commands = append(m.commands, executedCommand{Name: name, Args: args})
	return m.runErr
}

func (m *mockExecutor) RunOutput(name string, args ...string) (string, error) {
	m.commands = append(m.commands, executedCommand{Name: name, Args: args})
	if m.runOutputFn != nil {
		return m.runOutputFn(name, args...)
	}
	return "", m.runErr
}

func (m *mockExecutor) lastCommand() (executedCommand, bool) {
	if len(m.commands) == 0 {
		return executedCommand{}, false
	}
	return m.commands[len(m.commands)-1], true
}

//func (m *mockExecutor) reset() {
//	m.commands = nil
//	m.runErr = nil
//	m.runOutputFn = nil
//}

func newRunner(t *testing.T, cfg *hdiutil.Config, exec *mockExecutor) *hdiutil.Runner {
	t.Helper()
	r := hdiutil.New(cfg, hdiutil.WithExecutor(exec))
	t.Cleanup(r.Cleanup)
	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	return r
}

func TestAttachDiskImage_ParsesMountPoint(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{
		runOutputFn: func(name string, args ...string) (string, error) {
			return "/dev/disk4          \tGUID_partition_scheme          \t\n/dev/disk4s1        \tApple_HFS                     \t/Volumes/TestVolume\n", nil
		},
	}
	cfg := &hdiutil.Config{
		SourceDir:  t.TempDir(),
		OutputPath: "test.dmg",
	}

	r := newRunner(t, cfg, mock)

	if err := r.AttachDiskImage(); err != nil {
		t.Fatalf("AttachDiskImage() error = %v", err)
	}

	// Verify hdiutil attach was called with correct args
	cmd, ok := mock.lastCommand()
	if !ok {
		t.Fatal("expected a command to be executed")
	}
	if cmd.Name != "hdiutil" {
		t.Errorf("expected command 'hdiutil', got %q", cmd.Name)
	}
	if cmd.Args[0] != "attach" {
		t.Errorf("expected first arg 'attach', got %q", cmd.Args[0])
	}
}

func TestAttachDiskImage_NoMountPointInOutput(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{
		runOutputFn: func(name string, args ...string) (string, error) {
			return "/dev/disk4\tGUID_partition_scheme\n", nil
		},
	}
	cfg := &hdiutil.Config{
		SourceDir:  t.TempDir(),
		OutputPath: "test.dmg",
	}

	r := newRunner(t, cfg, mock)

	err := r.AttachDiskImage()
	if !errors.Is(err, hdiutil.ErrMountImage) {
		t.Errorf("AttachDiskImage() error = %v, want %v", err, hdiutil.ErrMountImage)
	}
}

func TestAttachDiskImage_ExecutorError(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{
		runOutputFn: func(name string, args ...string) (string, error) {
			return "attach failed", errors.New("exit status 1")
		},
	}
	cfg := &hdiutil.Config{
		SourceDir:  t.TempDir(),
		OutputPath: "test.dmg",
	}

	r := newRunner(t, cfg, mock)

	err := r.AttachDiskImage()
	if !errors.Is(err, hdiutil.ErrMountImage) {
		t.Errorf("AttachDiskImage() error = %v, want %v", err, hdiutil.ErrMountImage)
	}
	if !strings.Contains(err.Error(), "attach failed") {
		t.Errorf("error should contain output, got: %v", err)
	}
}

func TestDetachDiskImage_CallsFixPermissionsAndDetach(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{
		runOutputFn: func(name string, args ...string) (string, error) {
			return "/dev/disk4s1\tApple_HFS\t/Volumes/Test\n", nil
		},
	}
	cfg := &hdiutil.Config{
		SourceDir:  t.TempDir(),
		OutputPath: "test.dmg",
	}

	r := newRunner(t, cfg, mock)

	if err := r.AttachDiskImage(); err != nil {
		t.Fatalf("AttachDiskImage() error = %v", err)
	}

	mock.commands = nil // reset to track only detach commands
	if err := r.DetachDiskImage(); err != nil {
		t.Fatalf("DetachDiskImage() error = %v", err)
	}

	// Should have executed: chmod (fixPermissions) then hdiutil detach
	if len(mock.commands) != 2 {
		t.Fatalf("expected 2 commands, got %d: %+v", len(mock.commands), mock.commands)
	}
	if mock.commands[0].Name != "chmod" {
		t.Errorf("first command should be 'chmod', got %q", mock.commands[0].Name)
	}
	if mock.commands[1].Name != "hdiutil" {
		t.Errorf("second command should be 'hdiutil', got %q", mock.commands[1].Name)
	}
	if mock.commands[1].Args[0] != "detach" {
		t.Errorf("expected 'detach' arg, got %q", mock.commands[1].Args[0])
	}
}

func TestDetachDiskImage_FixPermissionsError(t *testing.T) {
	t.Parallel()
	chmodErr := errors.New("permission denied")
	mock := &mockExecutor{
		runOutputFn: func(name string, args ...string) (string, error) {
			return "/dev/disk4s1\tApple_HFS\t/Volumes/Test\n", nil
		},
	}
	cfg := &hdiutil.Config{
		SourceDir:  t.TempDir(),
		OutputPath: "test.dmg",
	}

	r := newRunner(t, cfg, mock)

	if err := r.AttachDiskImage(); err != nil {
		t.Fatalf("AttachDiskImage() error = %v", err)
	}

	// Make chmod fail
	mock.runErr = chmodErr
	err := r.DetachDiskImage()
	if err == nil {
		t.Fatal("DetachDiskImage() should fail when fixPermissions fails")
	}
	if !strings.Contains(err.Error(), "chmod failed") {
		t.Errorf("error should mention chmod, got: %v", err)
	}
}

func TestFixPermissions_Idempotent(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{
		runOutputFn: func(name string, args ...string) (string, error) {
			return "/dev/disk4s1\tApple_HFS\t/Volumes/Test\n", nil
		},
	}
	cfg := &hdiutil.Config{
		SourceDir:  t.TempDir(),
		OutputPath: "test.dmg",
	}

	r := newRunner(t, cfg, mock)

	if err := r.AttachDiskImage(); err != nil {
		t.Fatalf("AttachDiskImage() error = %v", err)
	}

	// First detach calls fixPermissions + detach
	mock.commands = nil
	if err := r.DetachDiskImage(); err != nil {
		t.Fatalf("first DetachDiskImage() error = %v", err)
	}
	firstCallCount := len(mock.commands)

	// Second Bless should NOT re-run chmod (fixPermissions is idempotent)
	mock.commands = nil
	if err := r.Bless(); err != nil {
		t.Fatalf("Bless() error = %v", err)
	}
	// Bless with Bless=false just returns nil after fixPermissions (which is a no-op now)
	if len(mock.commands) != 0 {
		t.Errorf("fixPermissions should be a no-op on second call, but %d commands were executed", len(mock.commands))
	}

	_ = firstCallCount // suppress unused
}

func TestBless_WithMockExecutor(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{
		runOutputFn: func(name string, args ...string) (string, error) {
			return "/dev/disk4s1\tApple_HFS\t/Volumes/Test\n", nil
		},
	}
	cfg := &hdiutil.Config{
		SourceDir:  t.TempDir(),
		OutputPath: "test.dmg",
		Bless:      true,
	}

	r := newRunner(t, cfg, mock)

	if err := r.AttachDiskImage(); err != nil {
		t.Fatalf("AttachDiskImage() error = %v", err)
	}

	mock.commands = nil
	if err := r.Bless(); err != nil {
		t.Fatalf("Bless() error = %v", err)
	}

	// Should run chmod (fixPermissions) then bless
	if len(mock.commands) != 2 {
		t.Fatalf("expected 2 commands, got %d: %+v", len(mock.commands), mock.commands)
	}
	if mock.commands[0].Name != "chmod" {
		t.Errorf("first command should be 'chmod', got %q", mock.commands[0].Name)
	}
	if mock.commands[1].Name != "bless" {
		t.Errorf("second command should be 'bless', got %q", mock.commands[1].Name)
	}
}

func TestBless_ErrorFromFixPermissions(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{
		runOutputFn: func(name string, args ...string) (string, error) {
			return "/dev/disk4s1\tApple_HFS\t/Volumes/Test\n", nil
		},
	}
	cfg := &hdiutil.Config{
		SourceDir:  t.TempDir(),
		OutputPath: "test.dmg",
		Bless:      true,
	}

	r := newRunner(t, cfg, mock)

	if err := r.AttachDiskImage(); err != nil {
		t.Fatalf("AttachDiskImage() error = %v", err)
	}

	mock.runErr = errors.New("chmod denied")
	err := r.Bless()
	if err == nil {
		t.Fatal("Bless() should fail when fixPermissions fails")
	}
	if !strings.Contains(err.Error(), "chmod failed") {
		t.Errorf("error should mention chmod, got: %v", err)
	}
}

func TestCodesign_SuccessWithMockExecutor(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{}
	cfg := &hdiutil.Config{
		SourceDir:       t.TempDir(),
		OutputPath:      "test.dmg",
		SigningIdentity: "Developer ID Application: Test",
	}

	r := newRunner(t, cfg, mock)

	if err := r.Codesign(); err != nil {
		t.Fatalf("Codesign() error = %v", err)
	}

	// Should call codesign twice: sign + verify
	if len(mock.commands) != 2 {
		t.Fatalf("expected 2 codesign commands, got %d", len(mock.commands))
	}
	if mock.commands[0].Name != "codesign" || mock.commands[0].Args[0] != "-s" {
		t.Errorf("first command should be 'codesign -s ...', got %+v", mock.commands[0])
	}
	if mock.commands[1].Name != "codesign" || mock.commands[1].Args[0] != "--verify" {
		t.Errorf("second command should be 'codesign --verify ...', got %+v", mock.commands[1])
	}
}

func TestCodesign_SigningFails(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{runErr: errors.New("no identity found")}
	cfg := &hdiutil.Config{
		SourceDir:       t.TempDir(),
		OutputPath:      "test.dmg",
		SigningIdentity: "Invalid Identity",
	}

	r := newRunner(t, cfg, mock)

	err := r.Codesign()
	if !errors.Is(err, hdiutil.ErrCodesignFailed) {
		t.Errorf("Codesign() error = %v, want %v", err, hdiutil.ErrCodesignFailed)
	}
}

func TestCodesign_VerificationFails(t *testing.T) {
	t.Parallel()
	callCount := 0
	mock := &mockExecutor{}
	// First call (sign) succeeds, second (verify) fails
	origRun := mock.Run
	_ = origRun
	mock.runErr = nil
	cfg := &hdiutil.Config{
		SourceDir:       t.TempDir(),
		OutputPath:      "test.dmg",
		SigningIdentity: "Developer ID",
	}

	r := hdiutil.New(cfg, hdiutil.WithExecutor(&verifyFailExecutor{callCount: &callCount}))
	t.Cleanup(r.Cleanup)
	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	err := r.Codesign()
	if !errors.Is(err, hdiutil.ErrCodesignFailed) {
		t.Errorf("Codesign() error = %v, want %v", err, hdiutil.ErrCodesignFailed)
	}
	if !strings.Contains(err.Error(), "signature seems invalid") {
		t.Errorf("error should mention invalid signature, got: %v", err)
	}
}

// verifyFailExecutor fails on the second Run call (signature verification).
type verifyFailExecutor struct {
	callCount *int
}

func (e *verifyFailExecutor) Run(name string, args ...string) error {
	*e.callCount++
	if *e.callCount >= 2 {
		return errors.New("verification failed")
	}
	return nil
}

func (e *verifyFailExecutor) RunOutput(name string, args ...string) (string, error) {
	return "", nil
}

func TestNotarize_SuccessWithMockExecutor(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{}
	cfg := &hdiutil.Config{
		SourceDir:           t.TempDir(),
		OutputPath:          "test.dmg",
		NotarizeCredentials: "my-profile",
	}

	r := newRunner(t, cfg, mock)

	if err := r.Notarize(); err != nil {
		t.Fatalf("Notarize() error = %v", err)
	}

	// Should call xcrun notarytool submit, then xcrun stapler staple
	if len(mock.commands) != 2 {
		t.Fatalf("expected 2 commands, got %d: %+v", len(mock.commands), mock.commands)
	}
	if mock.commands[0].Name != "xcrun" || mock.commands[0].Args[0] != "notarytool" {
		t.Errorf("first command should be 'xcrun notarytool ...', got %+v", mock.commands[0])
	}
	if mock.commands[1].Name != "xcrun" || mock.commands[1].Args[0] != "stapler" {
		t.Errorf("second command should be 'xcrun stapler ...', got %+v", mock.commands[1])
	}
}

func TestNotarize_SubmissionFails(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{runErr: errors.New("submission rejected")}
	cfg := &hdiutil.Config{
		SourceDir:           t.TempDir(),
		OutputPath:          "test.dmg",
		NotarizeCredentials: "my-profile",
	}

	r := newRunner(t, cfg, mock)

	err := r.Notarize()
	if !errors.Is(err, hdiutil.ErrNotarizeFailed) {
		t.Errorf("Notarize() error = %v, want %v", err, hdiutil.ErrNotarizeFailed)
	}
}

func TestNotarize_StaplerFails(t *testing.T) {
	t.Parallel()
	callCount := 0
	cfg := &hdiutil.Config{
		SourceDir:           t.TempDir(),
		OutputPath:          "test.dmg",
		NotarizeCredentials: "my-profile",
	}

	r := hdiutil.New(cfg, hdiutil.WithExecutor(&staplerFailExecutor{callCount: &callCount}))
	t.Cleanup(r.Cleanup)
	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	err := r.Notarize()
	if !errors.Is(err, hdiutil.ErrNotarizeFailed) {
		t.Errorf("Notarize() error = %v, want %v", err, hdiutil.ErrNotarizeFailed)
	}
	if !strings.Contains(err.Error(), "stapler failed") {
		t.Errorf("error should mention stapler, got: %v", err)
	}
}

// staplerFailExecutor succeeds on Run (notarytool submit) but fails on RunOutput (stapler staple).
type staplerFailExecutor struct {
	callCount *int
}

func (e *staplerFailExecutor) Run(name string, args ...string) error {
	return nil
}

func (e *staplerFailExecutor) RunOutput(name string, args ...string) (string, error) {
	return "staple error output", errors.New("stapler failed")
}

func TestFinalizeDMG_WithMockExecutor(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{}
	cfg := &hdiutil.Config{
		SourceDir:   t.TempDir(),
		OutputPath:  "test.dmg",
		ImageFormat: "UDBZ",
	}

	r := newRunner(t, cfg, mock)

	if err := r.FinalizeDMG(); err != nil {
		t.Fatalf("FinalizeDMG() error = %v", err)
	}

	cmd, ok := mock.lastCommand()
	if !ok {
		t.Fatal("expected a command to be executed")
	}
	if cmd.Name != "hdiutil" {
		t.Errorf("expected 'hdiutil', got %q", cmd.Name)
	}
	if cmd.Args[0] != "convert" {
		t.Errorf("expected 'convert' arg, got %q", cmd.Args[0])
	}
	// Verify format args are present
	argsStr := strings.Join(cmd.Args, " ")
	if !strings.Contains(argsStr, "-format UDBZ") {
		t.Errorf("expected '-format UDBZ' in args, got: %s", argsStr)
	}
}

func TestFinalizeDMG_WithVerbosity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		verbosity int
		wantFlag  string
	}{
		{"quiet", 1, "-quiet"},
		{"verbose", 2, "-verbose"},
		{"debug", 3, "-debug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockExecutor{}
			cfg := &hdiutil.Config{
				SourceDir:        t.TempDir(),
				OutputPath:       "test.dmg",
				HDIUtilVerbosity: tt.verbosity,
			}

			r := newRunner(t, cfg, mock)

			if err := r.FinalizeDMG(); err != nil {
				t.Fatalf("FinalizeDMG() error = %v", err)
			}

			cmd, _ := mock.lastCommand()
			argsStr := strings.Join(cmd.Args, " ")
			if !strings.Contains(argsStr, tt.wantFlag) {
				t.Errorf("expected %q in args, got: %s", tt.wantFlag, argsStr)
			}
			// Verbosity flag should come after "convert"
			if cmd.Args[0] != "convert" || cmd.Args[1] != tt.wantFlag {
				t.Errorf("verbosity flag should be inserted after 'convert', got args: %v", cmd.Args)
			}
		})
	}
}

func TestStart_CreateTempImage(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{}
	cfg := &hdiutil.Config{
		SourceDir:    t.TempDir(),
		OutputPath:   "test.dmg",
		VolumeName:   "TestVol",
		VolumeSizeMb: 50,
	}

	r := newRunner(t, cfg, mock)

	if err := r.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	cmd, ok := mock.lastCommand()
	if !ok {
		t.Fatal("expected a command to be executed")
	}
	if cmd.Name != "hdiutil" {
		t.Errorf("expected 'hdiutil', got %q", cmd.Name)
	}
	if cmd.Args[0] != "create" {
		t.Errorf("expected 'create' arg, got %q", cmd.Args[0])
	}

	argsStr := strings.Join(cmd.Args, " ")
	if !strings.Contains(argsStr, "-volname TestVol") {
		t.Errorf("expected '-volname TestVol' in args, got: %s", argsStr)
	}
	if !strings.Contains(argsStr, "-size 50m") {
		t.Errorf("expected '-size 50m' in args, got: %s", argsStr)
	}
	if !strings.Contains(argsStr, "-format UDRW") {
		t.Errorf("expected '-format UDRW' in args, got: %s", argsStr)
	}
}

func TestStart_SandboxSafeMode(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{}
	cfg := &hdiutil.Config{
		SourceDir:   t.TempDir(),
		OutputPath:  "test.dmg",
		SandboxSafe: true,
		FileSystem:  "HFS+",
	}

	r := newRunner(t, cfg, mock)

	if err := r.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Sandbox-safe mode should run makehybrid + convert
	if len(mock.commands) != 2 {
		t.Fatalf("expected 2 commands for sandbox-safe, got %d: %+v", len(mock.commands), mock.commands)
	}
	if mock.commands[0].Args[0] != "makehybrid" {
		t.Errorf("first command should be 'makehybrid', got %q", mock.commands[0].Args[0])
	}
	if mock.commands[1].Args[0] != "convert" {
		t.Errorf("second command should be 'convert', got %q", mock.commands[1].Args[0])
	}
}

func TestStart_SandboxSafeMakehybridFails(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{runErr: errors.New("makehybrid failed")}
	cfg := &hdiutil.Config{
		SourceDir:   t.TempDir(),
		OutputPath:  "test.dmg",
		SandboxSafe: true,
		FileSystem:  "HFS+",
	}

	r := newRunner(t, cfg, mock)

	err := r.Start()
	if err == nil {
		t.Fatal("Start() should fail when makehybrid fails")
	}
	// Should only have attempted 1 command (makehybrid), not proceed to convert
	if len(mock.commands) != 1 {
		t.Errorf("should stop after makehybrid failure, got %d commands", len(mock.commands))
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	t.Parallel()
	_, err := hdiutil.LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("LoadConfig() should fail for non-existent file")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	t.Parallel()
	tmpFile := fmt.Sprintf("%s/invalid.json", t.TempDir())
	if err := writeTestFile(t, tmpFile, "not valid json{{{"); err != nil {
		t.Fatal(err)
	}

	_, err := hdiutil.LoadConfig(tmpFile)
	if err == nil {
		t.Error("LoadConfig() should fail for invalid JSON")
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	t.Parallel()
	tmpFile := fmt.Sprintf("%s/empty.json", t.TempDir())
	if err := writeTestFile(t, tmpFile, ""); err != nil {
		t.Fatal(err)
	}

	_, err := hdiutil.LoadConfig(tmpFile)
	if err == nil {
		t.Error("LoadConfig() should fail for empty file")
	}
}

func TestDetachDiskImage_SimulateMode(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{}
	cfg := &hdiutil.Config{
		SourceDir:  t.TempDir(),
		OutputPath: "test.dmg",
		Simulate:   true,
	}

	r := newRunner(t, cfg, mock)

	if err := r.AttachDiskImage(); err != nil {
		t.Fatalf("AttachDiskImage() error = %v", err)
	}

	if err := r.DetachDiskImage(); err != nil {
		t.Fatalf("DetachDiskImage() in simulate mode error = %v", err)
	}

	// No commands should be executed in simulate mode
	if len(mock.commands) != 0 {
		t.Errorf("expected no commands in simulate mode, got %d", len(mock.commands))
	}
}

func TestAttachDiskImage_MountPointWithWhitespace(t *testing.T) {
	t.Parallel()
	mock := &mockExecutor{
		runOutputFn: func(name string, args ...string) (string, error) {
			return "/dev/disk4s1\tApple_HFS\t/Volumes/My Application   \n", nil
		},
	}
	cfg := &hdiutil.Config{
		SourceDir:  t.TempDir(),
		OutputPath: "test.dmg",
	}

	r := newRunner(t, cfg, mock)

	if err := r.AttachDiskImage(); err != nil {
		t.Fatalf("AttachDiskImage() error = %v", err)
	}

	// Verify the mount dir was extracted and trailing whitespace trimmed.
	// We confirm the runner works by calling DetachDiskImage.
	mock.commands = nil
	if err := r.DetachDiskImage(); err != nil {
		t.Fatalf("DetachDiskImage() error = %v", err)
	}

	// The detach command should reference the trimmed mount path
	if len(mock.commands) < 2 {
		t.Fatalf("expected at least 2 commands, got %d", len(mock.commands))
	}
	detachCmd := mock.commands[1]
	lastArg := detachCmd.Args[len(detachCmd.Args)-1]
	if strings.HasSuffix(lastArg, " ") {
		t.Errorf("mount dir should be trimmed, got %q", lastArg)
	}
}

func writeTestFile(t *testing.T, path, content string) error {
	t.Helper()
	return os.WriteFile(path, []byte(content), 0644)
}
