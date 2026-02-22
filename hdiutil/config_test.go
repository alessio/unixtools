package hdiutil_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"al.essio.dev/pkg/tools/hdiutil"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  hdiutil.Config
		wantErr error
	}{
		// SourceDir validation
		{
			name:    "empty_source_dir_returns_error",
			config:  hdiutil.Config{SourceDir: "", OutputPath: "test.dmg"},
			wantErr: hdiutil.ErrInvSourceDir,
		},

		// OutputPath validation
		{
			name:    "missing_dmg_extension_returns_error",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test"},
			wantErr: hdiutil.ErrImageFileExt,
		},
		{
			name:    "wrong_extension_returns_error",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.iso"},
			wantErr: hdiutil.ErrImageFileExt,
		},
		{
			name:    "uppercase_DMG_extension_returns_error",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.DMG"},
			wantErr: hdiutil.ErrImageFileExt,
		},

		// ImageFormat validation
		{
			name:    "invalid_image_format_returns_error",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", ImageFormat: "INVALID"},
			wantErr: hdiutil.ErrInvFormatOpt,
		},
		{
			name:    "valid_with_lowercase_format",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", ImageFormat: "udzo"},
			wantErr: nil,
		},

		// FileSystem validation
		{
			name:    "invalid_filesystem_returns_error",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", FileSystem: "EXT4"},
			wantErr: hdiutil.ErrInvFilesystemOpt,
		},
		{
			name:    "ntfs_filesystem_returns_error",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", FileSystem: "NTFS"},
			wantErr: hdiutil.ErrInvFilesystemOpt,
		},

		// SandboxSafe + APFS mutual exclusion
		{
			name:    "sandbox_safe_with_apfs_returns_error",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", SandboxSafe: true, FileSystem: "APFS"},
			wantErr: hdiutil.ErrSandboxAPFS,
		},
		{
			name:    "sandbox_safe_with_apfs_lowercase_returns_error",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", SandboxSafe: true, FileSystem: "apfs"},
			wantErr: hdiutil.ErrSandboxAPFS,
		},
		{
			name:    "sandbox_safe_with_apfs_mixed_case_returns_error",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", SandboxSafe: true, FileSystem: "Apfs"},
			wantErr: hdiutil.ErrSandboxAPFS,
		},

		// Valid configurations
		{
			name:    "minimal_valid_config",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg"},
			wantErr: nil,
		},
		{
			name:    "valid_with_hfs_plus",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", FileSystem: "HFS+"},
			wantErr: nil,
		},
		{
			name:    "valid_with_hfs_plus_lowercase",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", FileSystem: "hfs+"},
			wantErr: nil,
		},
		{
			name:    "valid_with_apfs",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", FileSystem: "APFS"},
			wantErr: nil,
		},
		{
			name:    "valid_sandbox_safe_with_hfs",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", SandboxSafe: true, FileSystem: "HFS+"},
			wantErr: nil,
		},
		{
			name:    "valid_sandbox_safe_with_default_fs",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", SandboxSafe: true},
			wantErr: nil,
		},
		{
			name:    "valid_with_all_options",
			config:  hdiutil.Config{SourceDir: "src", OutputPath: "out.dmg", VolumeName: "Vol", VolumeSizeMb: 100, FileSystem: "HFS+", ImageFormat: "UDBZ", Bless: true},
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

func TestConfig_Validate_SetsValidFlag(t *testing.T) {
	t.Parallel()
	cfg := hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg"}

	// Before validation, OptFn fields should be nil
	if cfg.FilesystemOpts != nil {
		t.Error("FilesystemOpts should be nil before Validate()")
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// After validation, OptFn fields should be set
	if cfg.FilesystemOpts == nil {
		t.Error("FilesystemOpts should not be nil after Validate()")
	}
	if cfg.ImageFormatOpts == nil {
		t.Error("ImageFormatOpts should not be nil after Validate()")
	}
	if cfg.VolumeSizeOpts == nil {
		t.Error("VolumeSizeOpts should not be nil after Validate()")
	}
	if cfg.VolumeNameOpt == nil {
		t.Error("VolumeNameOpt should not be nil after Validate()")
	}
}

func TestConfig_FilesystemOpts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fs       string
		wantOpts []string
	}{
		{
			name:     "empty_defaults_to_hfs_plus",
			fs:       "",
			wantOpts: []string{"-fs", "HFS+", "-fsargs", "-c c=64,a=16,e=16"},
		},
		{
			name:     "hfs_plus",
			fs:       "HFS+",
			wantOpts: []string{"-fs", "HFS+", "-fsargs", "-c c=64,a=16,e=16"},
		},
		{
			name:     "hfs_plus_lowercase",
			fs:       "hfs+",
			wantOpts: []string{"-fs", "HFS+", "-fsargs", "-c c=64,a=16,e=16"},
		},
		{
			name:     "apfs",
			fs:       "APFS",
			wantOpts: []string{"-fs", "APFS"},
		},
		{
			name:     "apfs_lowercase",
			fs:       "apfs",
			wantOpts: []string{"-fs", "APFS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", FileSystem: tt.fs}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			got := cfg.FilesystemOpts()
			if !reflect.DeepEqual(got, tt.wantOpts) {
				t.Errorf("FilesystemOpts() = %v, want %v", got, tt.wantOpts)
			}
		})
	}
}

func TestConfig_ImageFormatOpts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		format   string
		wantOpts []string
	}{
		{
			name:     "empty_defaults_to_udzo",
			format:   "",
			wantOpts: []string{"-format", "UDZO", "-imagekey", "zlib-level=9"},
		},
		{
			name:     "udzo",
			format:   "UDZO",
			wantOpts: []string{"-format", "UDZO", "-imagekey", "zlib-level=9"},
		},
		{
			name:     "udbz",
			format:   "UDBZ",
			wantOpts: []string{"-format", "UDBZ", "-imagekey", "bzip2-level=9"},
		},
		{
			name:     "ulfo",
			format:   "ULFO",
			wantOpts: []string{"-format", "ULFO"},
		},
		{
			name:     "ulmo",
			format:   "ULMO",
			wantOpts: []string{"-format", "ULMO"},
		},
		{
			name:     "lowercase_udzo",
			format:   "udzo",
			wantOpts: []string{"-format", "UDZO", "-imagekey", "zlib-level=9"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", ImageFormat: tt.format}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			got := cfg.ImageFormatOpts()
			if !reflect.DeepEqual(got, tt.wantOpts) {
				t.Errorf("ImageFormatOpts() = %v, want %v", got, tt.wantOpts)
			}
		})
	}
}

func TestConfig_VolumeSizeOpts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		sizeMb  int64
		wantErr error
	}{
		{
			name:   "positive_size",
			sizeMb: 100,
		},
		{
			name:   "large_size",
			sizeMb: 4096,
		},
		{
			name:   "zero_size",
			sizeMb: 0,
		},
		{
			name:    "negative_size_returns_error",
			sizeMb:  -50,
			wantErr: hdiutil.ErrVolumeSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", VolumeSizeMb: tt.sizeMb}
			err := cfg.Validate()
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			got := cfg.VolumeSizeOpts()
			if tt.sizeMb > 0 {
				want := []string{"-size", fmt.Sprintf("%dm", tt.sizeMb)}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("VolumeSizeOpts() = %v, want %v", got, want)
				}
			} else {
				if got != nil {
					t.Errorf("VolumeSizeOpts() = %v, want nil", got)
				}
			}
		})
	}
}

func TestConfig_VolumeNameOpt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		volumeName string
		outputPath string
		want       string
	}{
		{
			name:       "explicit_volume_name",
			volumeName: "MyVolume",
			outputPath: "test.dmg",
			want:       "MyVolume",
		},
		{
			name:       "auto_from_simple_filename",
			volumeName: "",
			outputPath: "MyApp.dmg",
			want:       "MyApp",
		},
		{
			name:       "auto_from_path_with_directories",
			volumeName: "",
			outputPath: "/Users/dev/builds/Application.dmg",
			want:       "Application",
		},
		{
			name:       "auto_from_relative_path",
			volumeName: "",
			outputPath: "./dist/output.dmg",
			want:       "output",
		},
		{
			name:       "volume_name_with_spaces",
			volumeName: "My Application",
			outputPath: "test.dmg",
			want:       "My Application",
		},
		{
			name:       "volume_name_with_special_chars",
			volumeName: "App-v1.2.3",
			outputPath: "test.dmg",
			want:       "App-v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{SourceDir: "src", OutputPath: tt.outputPath, VolumeName: tt.volumeName}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			got := cfg.VolumeNameOpt()
			if got != tt.want {
				t.Errorf("VolumeNameOpt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_OptFn_PanicWithoutValidation(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		invoke func(cfg *hdiutil.Config)
	}{
		{
			name:   "FilesystemOpts_panics",
			invoke: func(cfg *hdiutil.Config) { _ = cfg.FilesystemOpts() },
		},
		{
			name:   "ImageFormatOpts_panics",
			invoke: func(cfg *hdiutil.Config) { _ = cfg.ImageFormatOpts() },
		},
		{
			name:   "VolumeSizeOpts_panics",
			invoke: func(cfg *hdiutil.Config) { _ = cfg.VolumeSizeOpts() },
		},
		{
			name:   "VolumeNameOpt_panics",
			invoke: func(cfg *hdiutil.Config) { _ = cfg.VolumeNameOpt() },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg"}
			// Intentionally NOT calling cfg.Validate()

			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Expected panic when calling OptFn without validation")
				}
			}()

			tc.invoke(&cfg)
		})
	}
}

func TestConfig_Validate_Idempotent(t *testing.T) {
	t.Parallel()
	cfg := hdiutil.Config{SourceDir: "src", OutputPath: "test.dmg", VolumeName: "Test"}

	// Call Validate multiple times
	for i := 0; i < 3; i++ {
		if err := cfg.Validate(); err != nil {
			t.Fatalf("Validate() call %d error = %v", i+1, err)
		}
	}

	// Verify options still work correctly
	if got := cfg.VolumeNameOpt(); got != "Test" {
		t.Errorf("VolumeNameOpt() = %v, want Test", got)
	}
}

func TestConfig_ValidationOrder(t *testing.T) {
	t.Parallel()
	// Test that validation checks are performed in the expected order
	// Empty SourceDir should be caught before other validations

	cfg := hdiutil.Config{
		SourceDir:   "",         // Invalid
		OutputPath:  "test.iso", // Also invalid
		ImageFormat: "INVALID",  // Also invalid
		FileSystem:  "EXT4",     // Also invalid
	}

	err := cfg.Validate()
	// Should return the first error encountered (ErrInvSourceDir)
	if !errors.Is(err, hdiutil.ErrInvSourceDir) {
		t.Errorf("Validate() error = %v, want %v (first validation check)", err, hdiutil.ErrInvSourceDir)
	}
}

func TestConfig_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  hdiutil.Config
		wantErr bool
	}{
		{
			name: "whitespace_source_dir_is_valid",
			config: hdiutil.Config{
				SourceDir:  "   ",
				OutputPath: "test.dmg",
			},
			wantErr: false, // Whitespace is technically non-empty
		},
		{
			name: "output_path_only_extension",
			config: hdiutil.Config{
				SourceDir:  "src",
				OutputPath: ".dmg",
			},
			wantErr: false, // Valid extension, though unusual
		},
		{
			name: "very_long_volume_name",
			config: hdiutil.Config{
				SourceDir:  "src",
				OutputPath: "test.dmg",
				VolumeName: "ThisIsAVeryLongVolumeNameThatMightCauseIssues",
			},
			wantErr: false, // Validation doesn't enforce length limits
		},
		{
			name: "max_int64_volume_size",
			config: hdiutil.Config{
				SourceDir:    "src",
				OutputPath:   "test.dmg",
				VolumeSizeMb: 9223372036854775807, // Max int64
			},
			wantErr: false, // Validation doesn't enforce size limits
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_MultipleFormatsValidation(t *testing.T) {
	t.Parallel()
	formats := []string{"UDZO", "UDBZ", "ULFO", "ULMO", ""}

	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:   "src",
				OutputPath:  "test.dmg",
				ImageFormat: format,
			}

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Validate() with format %s error = %v", format, err)
			}
		})
	}
}

func TestConfig_PathWithSpaces(t *testing.T) {
	t.Parallel()
	cfg := hdiutil.Config{
		SourceDir:  "source with spaces",
		OutputPath: "output with spaces.dmg",
		VolumeName: "Volume With Spaces",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() with spaces in paths error = %v", err)
	}

	volumeName := cfg.VolumeNameOpt()
	if volumeName != "Volume With Spaces" {
		t.Errorf("VolumeNameOpt() = %v, want 'Volume With Spaces'", volumeName)
	}
}

func TestConfig_SpecialCharactersInPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		outputPath string
		wantErr    bool
	}{
		{
			name:       "hyphen_in_name",
			outputPath: "my-app.dmg",
			wantErr:    false,
		},
		{
			name:       "underscore_in_name",
			outputPath: "my_app.dmg",
			wantErr:    false,
		},
		{
			name:       "dot_in_middle",
			outputPath: "my.app.dmg",
			wantErr:    false,
		},
		{
			name:       "multiple_extensions",
			outputPath: "my.app.v1.dmg",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:  "src",
				OutputPath: tt.outputPath,
			}

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ValidateAfterModification(t *testing.T) {
	t.Parallel()
	cfg := hdiutil.Config{
		SourceDir:  "src",
		OutputPath: "test.dmg",
	}

	// First validation
	if err := cfg.Validate(); err != nil {
		t.Fatalf("First Validate() error = %v", err)
	}

	// Modify config
	cfg.VolumeName = "NewVolume"
	cfg.VolumeSizeMb = 100

	// Second validation should also succeed
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Second Validate() error = %v", err)
	}

	// Verify new values are used
	if got := cfg.VolumeNameOpt(); got != "NewVolume" {
		t.Errorf("VolumeNameOpt() = %v, want NewVolume", got)
	}

	opts := cfg.VolumeSizeOpts()
	if len(opts) == 0 {
		t.Error("VolumeSizeOpts() should return options after setting size")
	}
}

func TestConfig_ZeroValueConfig(t *testing.T) {
	t.Parallel()
	var cfg hdiutil.Config

	// Zero value config should fail validation
	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() on zero value Config should return error")
	}

	if !errors.Is(err, hdiutil.ErrInvSourceDir) {
		t.Errorf("Validate() error = %v, want %v", err, hdiutil.ErrInvSourceDir)
	}
}

func TestConfig_BothBlessAndSandboxSafe(t *testing.T) {
	t.Parallel()
	cfg := hdiutil.Config{
		SourceDir:   "src",
		OutputPath:  "test.dmg",
		Bless:       true,
		SandboxSafe: true,
		FileSystem:  "HFS+",
	}

	// This combination should be valid (bless just gets skipped for sandbox-safe)
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() with both Bless and SandboxSafe error = %v", err)
	}
}

func TestConfig_AllImageFormatsWithAllFilesystems(t *testing.T) {
	t.Parallel()
	formats := []string{"UDZO", "UDBZ", "ULFO", "ULMO"}
	filesystems := []string{"HFS+", "APFS"}

	for _, format := range formats {
		for _, fs := range filesystems {
			t.Run("format_"+format+"_fs_"+fs, func(t *testing.T) {
				t.Parallel()
				cfg := hdiutil.Config{
					SourceDir:   "src",
					OutputPath:  "test.dmg",
					ImageFormat: format,
					FileSystem:  fs,
				}

				err := cfg.Validate()
				if err != nil {
					t.Errorf("Validate() with format %s and fs %s error = %v", format, fs, err)
				}

				formatOpts := cfg.ImageFormatOpts()
				if len(formatOpts) == 0 {
					t.Errorf("ImageFormatOpts() returned empty for format %s", format)
				}

				fsOpts := cfg.FilesystemOpts()
				if len(fsOpts) == 0 {
					t.Errorf("FilesystemOpts() returned empty for fs %s", fs)
				}
			})
		}
	}
}

func TestConfig_OutputPathExtensionCaseSensitive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		outputPath string
		wantErr    bool
	}{
		{"lowercase_dmg", "test.dmg", false},
		{"uppercase_DMG", "test.DMG", true},
		{"mixed_case_Dmg", "test.Dmg", true},
		{"mixed_case_dMg", "test.dMg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:  "src",
				OutputPath: tt.outputPath,
			}

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_VolumeNameFromComplexPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		outputPath string
		want       string
	}{
		{
			name:       "absolute_path",
			outputPath: "/usr/local/bin/myapp.dmg",
			want:       "myapp",
		},
		{
			name:       "relative_path_with_dots",
			outputPath: "../../output/myapp.dmg",
			want:       "myapp",
		},
		{
			name:       "path_with_multiple_dots",
			outputPath: "/path/to/app.v1.2.3.dmg",
			want:       "app.v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := hdiutil.Config{
				SourceDir:  "src",
				OutputPath: tt.outputPath,
			}

			if err := cfg.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			got := cfg.VolumeNameOpt()
			if got != tt.want {
				t.Errorf("VolumeNameOpt() = %v, want %v", got, tt.want)
			}
		})
	}
}
