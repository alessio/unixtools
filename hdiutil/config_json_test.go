package hdiutil_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"al.essio.dev/pkg/tools/hdiutil"
)

func TestConfig_JSON(t *testing.T) {
	t.Parallel()

	original := &hdiutil.Config{
		VolumeName:          "MyVolume",
		VolumeSizeMb:        100,
		SandboxSafe:         true,
		Bless:               true,
		FileSystem:          "HFS+",
		SigningIdentity:     "Developer ID Application: Test",
		NotarizeCredentials: "test-profile",
		ImageFormat:         "UDZO",
		HDIUtilVerbosity:    2,
		OutputPath:          "test.dmg",
		SourceDir:           "src",
		Simulate:            true,
		Checksum:            "SHA256",
		ExcludePatterns:     []string{".DS_Store", "*.tmp"},
	}

	// Test ToJSON
	var buf bytes.Buffer
	if err := original.ToJSON(&buf); err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Test FromJSON
	decoded := &hdiutil.Config{}
	if err := decoded.FromJSON(&buf); err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	// Compare exported fields individually to avoid brittleness with unexported state.
	if decoded.VolumeName != original.VolumeName {
		t.Errorf("VolumeName mismatch: expected %q, got %q", original.VolumeName, decoded.VolumeName)
	}
	if decoded.VolumeSizeMb != original.VolumeSizeMb {
		t.Errorf("VolumeSizeMb mismatch: expected %d, got %d", original.VolumeSizeMb, decoded.VolumeSizeMb)
	}
	if decoded.SandboxSafe != original.SandboxSafe {
		t.Errorf("SandboxSafe mismatch: expected %v, got %v", original.SandboxSafe, decoded.SandboxSafe)
	}
	if decoded.Bless != original.Bless {
		t.Errorf("Bless mismatch: expected %v, got %v", original.Bless, decoded.Bless)
	}
	if decoded.FileSystem != original.FileSystem {
		t.Errorf("FileSystem mismatch: expected %q, got %q", original.FileSystem, decoded.FileSystem)
	}
	if decoded.SigningIdentity != original.SigningIdentity {
		t.Errorf("SigningIdentity mismatch: expected %q, got %q", original.SigningIdentity, decoded.SigningIdentity)
	}
	if decoded.NotarizeCredentials != original.NotarizeCredentials {
		t.Errorf("NotarizeCredentials mismatch: expected %q, got %q", original.NotarizeCredentials, decoded.NotarizeCredentials)
	}
	if decoded.ImageFormat != original.ImageFormat {
		t.Errorf("ImageFormat mismatch: expected %q, got %q", original.ImageFormat, decoded.ImageFormat)
	}
	if decoded.HDIUtilVerbosity != original.HDIUtilVerbosity {
		t.Errorf("HDIUtilVerbosity mismatch: expected %d, got %d", original.HDIUtilVerbosity, decoded.HDIUtilVerbosity)
	}
	if decoded.OutputPath != original.OutputPath {
		t.Errorf("OutputPath mismatch: expected %q, got %q", original.OutputPath, decoded.OutputPath)
	}
	if decoded.SourceDir != original.SourceDir {
		t.Errorf("SourceDir mismatch: expected %q, got %q", original.SourceDir, decoded.SourceDir)
	}
	if decoded.Simulate != original.Simulate {
		t.Errorf("Simulate mismatch: expected %v, got %v", original.Simulate, decoded.Simulate)
	}
	if decoded.Checksum != original.Checksum {
		t.Errorf("Checksum mismatch: expected %q, got %q", original.Checksum, decoded.Checksum)
	}
	if len(decoded.ExcludePatterns) != len(original.ExcludePatterns) {
		t.Fatalf("ExcludePatterns length mismatch: expected %d, got %d", len(original.ExcludePatterns), len(decoded.ExcludePatterns))
	}
	for i, p := range original.ExcludePatterns {
		if decoded.ExcludePatterns[i] != p {
			t.Errorf("ExcludePatterns[%d] mismatch: expected %q, got %q", i, p, decoded.ExcludePatterns[i])
		}
	}
}

func TestConfig_FromJSON_Partial(t *testing.T) {
	t.Parallel()

	jsonStr := `{"volume_name": "Test", "output_path": "out.dmg", "source_dir": "src"}`
	buf := bytes.NewBufferString(jsonStr)

	cfg := &hdiutil.Config{}
	if err := cfg.FromJSON(buf); err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if cfg.VolumeName != "Test" {
		t.Errorf("Expected VolumeName 'Test', got '%s'", cfg.VolumeName)
	}
	if cfg.OutputPath != "out.dmg" {
		t.Errorf("Expected OutputPath 'out.dmg', got '%s'", cfg.OutputPath)
	}
	if cfg.SourceDir != "src" {
		t.Errorf("Expected SourceDir 'src', got '%s'", cfg.SourceDir)
	}
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "config.json")
	jsonStr := `{"volume_name": "TestFile", "output_path": "file.dmg", "source_dir": "src"}`
	if err := os.WriteFile(tmpFile, []byte(jsonStr), 0644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	cfg, err := hdiutil.LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.VolumeName != "TestFile" {
		t.Errorf("Expected VolumeName 'TestFile', got '%s'", cfg.VolumeName)
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

func writeTestFile(t *testing.T, path, content string) error {
	t.Helper()
	return os.WriteFile(path, []byte(content), 0644)
}
