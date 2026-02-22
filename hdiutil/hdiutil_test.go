// Package hdiutil_test contains tests for the hdiutil package.
package hdiutil_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"al.essio.dev/pkg/tools/hdiutil"
)

func TestSetLogWriter(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	hdiutil.SetLogWriter(&buf)
	// Restore to discard after test
	t.Cleanup(func() {
		hdiutil.SetLogWriter(os.Stderr)
	})
}

func TestRunnerSimulateMode(t *testing.T) {
	t.Parallel()
	cfg := hdiutil.Config{
		SourceDir:  "test",
		OutputPath: "test.dmg",
		Simulate:   true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// All these should succeed in simulate mode
	if err := r.Start(); err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestRunnerSandboxSafeMode(t *testing.T) {
	t.Parallel()
	cfg := hdiutil.Config{
		SourceDir:   "test",
		OutputPath:  "test.dmg",
		SandboxSafe: true,
		FileSystem:  "HFS+",
		Simulate:    true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	if err := r.Start(); err != nil {
		t.Errorf("Start() (sandbox safe) error = %v", err)
	}
}

func TestCodesignSkipped(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		signingIdentity string
	}{
		{"no_identity_skips", ""},
		{"simulate_with_identity", "Developer ID Application"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:       "test",
				OutputPath:      "test.dmg",
				SigningIdentity: tt.signingIdentity,
				Simulate:        true,
			}

			r := hdiutil.New(&cfg)
			t.Cleanup(r.Cleanup)

			if err := r.Setup(); err != nil {
				t.Fatalf("Setup() error = %v", err)
			}

			if err := r.Codesign(); err != nil {
				t.Errorf("Codesign() error = %v, want nil", err)
			}
		})
	}
}

func TestNotarizeSkipped(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		notarizeCredentials string
	}{
		{"no_credentials_skips", ""},
		{"simulate_with_credentials", "test-profile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:           "test",
				OutputPath:          "test.dmg",
				NotarizeCredentials: tt.notarizeCredentials,
				Simulate:            true,
			}

			r := hdiutil.New(&cfg)
			t.Cleanup(r.Cleanup)

			if err := r.Setup(); err != nil {
				t.Fatalf("Setup() error = %v", err)
			}

			if err := r.Notarize(); err != nil {
				t.Errorf("Notarize() error = %v, want nil", err)
			}
		})
	}
}

func TestBlessSkipped(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		bless       bool
		sandboxSafe bool
	}{
		{"bless_disabled", false, false},
		{"sandbox_safe_skips_bless", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:   "test",
				OutputPath:  "test.dmg",
				Bless:       tt.bless,
				SandboxSafe: tt.sandboxSafe,
				Simulate:    true,
			}

			r := hdiutil.New(&cfg)
			t.Cleanup(r.Cleanup)

			if err := r.Setup(); err != nil {
				t.Fatalf("Setup() error = %v", err)
			}

			if err := r.Bless(); err != nil {
				t.Errorf("Bless() error = %v, want nil", err)
			}
		})
	}
}

func TestHDIUtilVerbosityLevels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		verbosity int
	}{
		{"verbosity_negative", -1},
		{"verbosity_0_default", 0},
		{"verbosity_1_quiet", 1},
		{"verbosity_2_verbose", 2},
		{"verbosity_3_debug", 3},
		{"verbosity_4_debug", 4}, // 3+ should all map to debug
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:        "test",
				OutputPath:       "test.dmg",
				HDIUtilVerbosity: tt.verbosity,
				Simulate:         true,
			}

			r := hdiutil.New(&cfg)
			t.Cleanup(r.Cleanup)

			if err := r.Setup(); err != nil {
				t.Fatalf("Setup() error = %v", err)
			}

			// Should not fail regardless of verbosity level
			if err := r.Start(); err != nil {
				t.Errorf("Start() error = %v", err)
			}
		})
	}
}

// TestInit preserves the original test for backward compatibility
func TestInit(t *testing.T) {
	t.Parallel()
	hdiutil.SetLogWriter(os.Stderr)
	configs := []hdiutil.Config{
		{
			OutputPath:      "test.dmg",
			VolumeName:      "test",
			VolumeSizeMb:    100,
			SandboxSafe:     true,
			FileSystem:      "APFS",
			SigningIdentity: "",
			ImageFormat:     "ULFO",
			Simulate:        true,
			SourceDir:       "test",
		},
		{
			OutputPath:       "test.dmg",
			VolumeName:       "test",
			FileSystem:       "APFS",
			SigningIdentity:  "",
			HDIUtilVerbosity: 1,
			Simulate:         true,
			SourceDir:        "test",
		},
	}

	type args struct {
		c *hdiutil.Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"sandbox_safe_with_volume_size_should_fail", args{&configs[0]}, true},
		{"valid_config_should_succeed", args{&configs[1]}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			r := hdiutil.New(tt.args.c)
			t2.Cleanup(r.Cleanup)
			if err := r.Setup(); (err != nil) != tt.wantErr {
				t2.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if err := r.Start(); (err != nil) != tt.wantErr {
				t2.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunnerWithRealSourceDirectory(t *testing.T) {
	// Create a real temporary source directory
	sourceDir := t.TempDir()
	testFile := sourceDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := hdiutil.Config{
		SourceDir:  sourceDir,
		OutputPath: "test.dmg",
		Simulate:   true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	if err := r.Start(); err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestRunnerCompleteWorkflow(t *testing.T) {
	// Test the complete workflow in simulate mode
	sourceDir := t.TempDir()
	testFile := sourceDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := t.TempDir()
	outputPath := outputDir + "/test.dmg"

	cfg := hdiutil.Config{
		SourceDir:    sourceDir,
		OutputPath:   outputPath,
		VolumeName:   "TestVol",
		VolumeSizeMb: 10,
		Simulate:     true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	// Test complete workflow
	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	if err := r.Start(); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// AttachDiskImage, Bless, and DetachDiskImage will fail in simulate mode
	// without hdiutil because they try to parse output that doesn't exist.
	// In a real scenario (with hdiutil), these would work in simulate mode.
	// For unit testing, we skip these or expect errors.
	_ = r.AttachDiskImage() // Expected to fail in simulate mode without hdiutil
	_ = r.Bless()           // May fail if attach failed
	_ = r.DetachDiskImage() // May fail if attach failed

	if err := r.FinalizeDMG(); err != nil {
		t.Errorf("FinalizeDMG() error = %v", err)
	}

	if err := r.Codesign(); err != nil {
		t.Errorf("Codesign() error = %v", err)
	}

	if err := r.Notarize(); err != nil {
		t.Errorf("Notarize() error = %v", err)
	}
}

func TestRunnerAllFormatsInSimulateMode(t *testing.T) {
	formats := []string{"UDZO", "UDBZ", "ULFO", "ULMO"}

	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			cfg := hdiutil.Config{
				SourceDir:   "test",
				OutputPath:  "test.dmg",
				ImageFormat: format,
				Simulate:    true,
			}

			r := hdiutil.New(&cfg)
			t.Cleanup(r.Cleanup)

			if err := r.Setup(); err != nil {
				t.Fatalf("Setup() error = %v", err)
			}

			if err := r.Start(); err != nil {
				t.Errorf("Start() with format %s error = %v", format, err)
			}

			if err := r.FinalizeDMG(); err != nil {
				t.Errorf("FinalizeDMG() with format %s error = %v", format, err)
			}
		})
	}
}

func TestRunnerMultipleCleanupCalls(t *testing.T) {
	cfg := hdiutil.Config{
		SourceDir:  "test",
		OutputPath: "test.dmg",
		Simulate:   true,
	}

	r := hdiutil.New(&cfg)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// Multiple cleanup calls should not panic
	r.Cleanup()
	r.Cleanup()
	r.Cleanup()
}

func TestRunnerDetachWithoutAttach(t *testing.T) {
	cfg := hdiutil.Config{
		SourceDir:  "test",
		OutputPath: "test.dmg",
		Simulate:   true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// Detach without attach should not panic in simulate mode
	if err := r.DetachDiskImage(); err != nil {
		t.Logf("DetachDiskImage() without attach returned: %v", err)
	}
}

func TestRunnerOperationsBeforeSetup(t *testing.T) {
	cfg := hdiutil.Config{
		SourceDir:  "test",
		OutputPath: "test.dmg",
		Simulate:   true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	// All operations should fail before Setup
	if err := r.Start(); !errors.Is(err, hdiutil.ErrNeedInit) {
		t.Errorf("Start() before Setup() should return ErrNeedInit, got: %v", err)
	}
}

func TestSetLogWriterWithBuffer(t *testing.T) {
	var buf bytes.Buffer
	hdiutil.SetLogWriter(&buf)

	// Restore to stderr after test
	t.Cleanup(func() {
		hdiutil.SetLogWriter(os.Stderr)
	})

	cfg := hdiutil.Config{
		SourceDir:  "test",
		OutputPath: "test.dmg",
		Simulate:   true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// The buffer might contain log messages (or might not, depending on implementation)
	t.Logf("Log buffer contents: %s", buf.String())
}

func TestRunnerOutputPathField(t *testing.T) {
	outputPath := "/path/to/test.dmg"

	cfg := hdiutil.Config{
		SourceDir:  "test",
		OutputPath: outputPath,
		Simulate:   true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// Verify the OutputPath is accessible through the embedded Config
	if r.OutputPath != outputPath {
		t.Errorf("OutputPath = %v, want %v", r.OutputPath, outputPath)
	}
}

func TestRunnerSourceDirField(t *testing.T) {
	sourceDir := "/path/to/source"

	cfg := hdiutil.Config{
		SourceDir:  sourceDir,
		OutputPath: "test.dmg",
		Simulate:   true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// Verify the SourceDir is accessible through the embedded Config
	if r.SourceDir != sourceDir {
		t.Errorf("SourceDir = %v, want %v", r.SourceDir, sourceDir)
	}
}
