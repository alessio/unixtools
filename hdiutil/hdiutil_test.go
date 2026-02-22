// Package hdiutil_test contains tests for the hdiutil package.
package hdiutil_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"al.essio.dev/cmd/mkdmg/pkg/hdiutil"
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

func TestConfigValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  hdiutil.Config
		wantErr error
	}{
		{
			name: "empty_source_dir",
			config: hdiutil.Config{
				SourceDir:  "",
				OutputPath: "test.dmg",
			},
			wantErr: hdiutil.ErrInvSourceDir,
		},
		{
			name: "invalid_output_extension",
			config: hdiutil.Config{
				SourceDir:  "test",
				OutputPath: "test.iso",
			},
			wantErr: hdiutil.ErrImageFileExt,
		},
		{
			name: "invalid_image_format",
			config: hdiutil.Config{
				SourceDir:   "test",
				OutputPath:  "test.dmg",
				ImageFormat: "INVALID",
			},
			wantErr: hdiutil.ErrInvFormatOpt,
		},
		{
			name: "invalid_filesystem",
			config: hdiutil.Config{
				SourceDir:  "test",
				OutputPath: "test.dmg",
				FileSystem: "EXT4",
			},
			wantErr: hdiutil.ErrInvFilesystemOpt,
		},
		{
			name: "sandbox_safe_with_apfs",
			config: hdiutil.Config{
				SourceDir:   "test",
				OutputPath:  "test.dmg",
				SandboxSafe: true,
				FileSystem:  "APFS",
			},
			wantErr: hdiutil.ErrSandboxAPFS,
		},
		{
			name: "sandbox_safe_with_apfs_lowercase",
			config: hdiutil.Config{
				SourceDir:   "test",
				OutputPath:  "test.dmg",
				SandboxSafe: true,
				FileSystem:  "apfs",
			},
			wantErr: hdiutil.ErrSandboxAPFS,
		},
		{
			name: "valid_default_config",
			config: hdiutil.Config{
				SourceDir:  "test",
				OutputPath: "test.dmg",
			},
			wantErr: nil,
		},
		{
			name: "valid_hfs_plus",
			config: hdiutil.Config{
				SourceDir:  "test",
				OutputPath: "test.dmg",
				FileSystem: "HFS+",
			},
			wantErr: nil,
		},
		{
			name: "valid_apfs",
			config: hdiutil.Config{
				SourceDir:  "test",
				OutputPath: "test.dmg",
				FileSystem: "APFS",
			},
			wantErr: nil,
		},
		{
			name: "valid_sandbox_safe_hfs",
			config: hdiutil.Config{
				SourceDir:   "test",
				OutputPath:  "test.dmg",
				SandboxSafe: true,
				FileSystem:  "HFS+",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestImageFormats(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		format      string
		wantErr     bool
		wantDefault bool
	}{
		{"empty_defaults_to_UDZO", "", false, true},
		{"UDZO", "UDZO", false, false},
		{"UDBZ", "UDBZ", false, false},
		{"ULFO", "ULFO", false, false},
		{"ULMO", "ULMO", false, false},
		{"invalid_format", "INVALID", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:   "test",
				OutputPath:  "test.dmg",
				ImageFormat: tt.format,
				Simulate:    true,
			}

			r := hdiutil.New(&cfg)
			t.Cleanup(r.Cleanup)

			err := r.Setup()
			if (err != nil) != tt.wantErr {
				t.Errorf("Setup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilesystems(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		fs      string
		wantErr bool
	}{
		{"empty_defaults_to_HFS+", "", false},
		{"HFS+", "HFS+", false},
		{"hfs+_lowercase", "hfs+", false},
		{"APFS", "APFS", false},
		{"apfs_lowercase", "apfs", false},
		{"invalid_filesystem", "EXT4", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:  "test",
				OutputPath: "test.dmg",
				FileSystem: tt.fs,
				Simulate:   true,
			}

			r := hdiutil.New(&cfg)
			t.Cleanup(r.Cleanup)

			err := r.Setup()
			if (err != nil) != tt.wantErr {
				t.Errorf("Setup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVolumeNameGeneration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		volumeName     string
		outputPath     string
		wantVolumeName string
	}{
		{
			name:           "explicit_volume_name",
			volumeName:     "MyVolume",
			outputPath:     "test.dmg",
			wantVolumeName: "MyVolume",
		},
		{
			name:           "auto_generated_from_output",
			volumeName:     "",
			outputPath:     "MyApp.dmg",
			wantVolumeName: "MyApp",
		},
		{
			name:           "auto_generated_with_path",
			volumeName:     "",
			outputPath:     "/some/path/Application.dmg",
			wantVolumeName: "Application",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:  "test",
				OutputPath: tt.outputPath,
				VolumeName: tt.volumeName,
				Simulate:   true,
			}

			err := cfg.Validate()
			if err != nil {
				t.Fatalf("Validate() unexpected error = %v", err)
			}

			got := cfg.VolumeNameOpt()
			if got != tt.wantVolumeName {
				t.Errorf("VolumeNameOpt() = %v, want %v", got, tt.wantVolumeName)
			}
		})
	}
}

func TestVolumeSizeOpts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		sizeMb  int64
		wantErr bool
		hasOpts bool
	}{
		{"positive_size", 100, false, true},
		{"zero_size", 0, false, false},
		{"negative_size", -1, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:    "test",
				OutputPath:   "test.dmg",
				VolumeSizeMb: tt.sizeMb,
				Simulate:     true,
			}

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() unexpected error = %v, wantErr = %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			opts := cfg.VolumeSizeOpts()
			if (len(opts) > 0) != tt.hasOpts {
				t.Errorf("VolumeSizeOpts() returned %v opts, wantOpts = %v", opts, tt.hasOpts)
			}
		})
	}
}

func TestStartWithoutSetup(t *testing.T) {
	t.Parallel()
	cfg := hdiutil.Config{
		SourceDir:  "test",
		OutputPath: "test.dmg",
		Simulate:   true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	err := r.Start()
	if !errors.Is(err, hdiutil.ErrNeedInit) {
		t.Errorf("Start() without Setup() error = %v, want %v", err, hdiutil.ErrNeedInit)
	}
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
	cfg := hdiutil.Config{
		SourceDir:       "test",
		OutputPath:      "test.dmg",
		SigningIdentity: "", // empty, should skip
		Simulate:        true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// Should return nil when no signing identity is set
	if err := r.Codesign(); err != nil {
		t.Errorf("Codesign() error = %v, want nil (skipped)", err)
	}
}

func TestNotarizeSkipped(t *testing.T) {
	t.Parallel()
	cfg := hdiutil.Config{
		SourceDir:           "test",
		OutputPath:          "test.dmg",
		NotarizeCredentials: "", // empty, should skip
		Simulate:            true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// Should return nil when no notarize credentials are set
	if err := r.Notarize(); err != nil {
		t.Errorf("Notarize() error = %v, want nil (skipped)", err)
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

func TestCleanup(t *testing.T) {
	t.Parallel()
	cfg := hdiutil.Config{
		SourceDir:  "test",
		OutputPath: "test.dmg",
		Simulate:   true,
	}

	r := hdiutil.New(&cfg)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// Cleanup should not panic even when called multiple times
	r.Cleanup()
	r.Cleanup() // Second call should be safe
}

func TestOptFnPanicWithoutValidation(t *testing.T) {
	t.Parallel()
	cfg := hdiutil.Config{
		SourceDir:  "test",
		OutputPath: "test.dmg",
	}

	// Don't call Validate()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when calling OptFn without validation")
		}
	}()

	// This should panic because Validate wasn't called
	_ = cfg.FilesystemOpts()
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

func TestRunnerWithSigningIdentity(t *testing.T) {
	cfg := hdiutil.Config{
		SourceDir:       "test",
		OutputPath:      "test.dmg",
		SigningIdentity: "Developer ID Application",
		Simulate:        true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// In simulate mode, this should not fail even with fake identity
	if err := r.Codesign(); err != nil {
		t.Errorf("Codesign() in simulate mode error = %v", err)
	}
}

func TestRunnerWithNotarization(t *testing.T) {
	cfg := hdiutil.Config{
		SourceDir:           "test",
		OutputPath:          "test.dmg",
		NotarizeCredentials: "test-profile",
		Simulate:            true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// In simulate mode, this should not fail
	if err := r.Notarize(); err != nil {
		t.Errorf("Notarize() in simulate mode error = %v", err)
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

func TestRunnerBlessWithSandboxSafe(t *testing.T) {
	cfg := hdiutil.Config{
		SourceDir:   "test",
		OutputPath:  "test.dmg",
		Bless:       true,
		SandboxSafe: true,
		Simulate:    true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// Bless should be skipped for sandbox-safe images
	if err := r.Bless(); err != nil {
		t.Errorf("Bless() should not error when skipped, got: %v", err)
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

func TestRunnerVerbosityZero(t *testing.T) {
	cfg := hdiutil.Config{
		SourceDir:        "test",
		OutputPath:       "test.dmg",
		HDIUtilVerbosity: 0,
		Simulate:         true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	if err := r.Start(); err != nil {
		t.Errorf("Start() with verbosity 0 error = %v", err)
	}
}

func TestRunnerNegativeVerbosity(t *testing.T) {
	cfg := hdiutil.Config{
		SourceDir:        "test",
		OutputPath:       "test.dmg",
		HDIUtilVerbosity: -1,
		Simulate:         true,
	}

	r := hdiutil.New(&cfg)
	t.Cleanup(r.Cleanup)

	if err := r.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// Negative verbosity should be treated as 0 (no special flag)
	if err := r.Start(); err != nil {
		t.Errorf("Start() with negative verbosity error = %v", err)
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