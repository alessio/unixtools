package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alessio/unixtools/internal/file"
)

func TestCopyDir(t *testing.T) {
	if err := file.CopyDir("testdata", "testdata"); err == nil {
		t.Fatal("expected error, got nil: src and dst must be different")
	}

	if err := file.CopyDir("non-existing", filepath.Join(t.TempDir(), "testdata")); err == nil {
		t.Fatal("expected error, got nil: non-existing dir can not be copied")
	}

	if err := file.CopyDir(filepath.Join("testdata", "regular"), filepath.Join(t.TempDir(), "testdata")); err == nil {
		t.Fatal("expected error, got nil: src must be a directory")
	}

	if err := file.CopyDir("testdata", t.TempDir()); err == nil {
		t.Fatal("expected error, got nil: dst directory must not exist")
	}

	destDir := filepath.Join(t.TempDir(), "testdata")
	if err := file.CopyDir("testdata", destDir); err != nil {
		t.Fatalf("couldn't copy testdata: %v", err)
	}

	// check symlinks are copied correctly
	symlink := filepath.Join(destDir, "symlink")
	if mode := mustLstat(t, symlink); mode&os.ModeSymlink == 0 {
		t.Fatalf("%q is not a symbolic link", symlink)
	}
}

func mustLstat(t *testing.T, path string) os.FileMode {
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("err: want nil, got: %v", err)
	}

	return info.Mode()
}
