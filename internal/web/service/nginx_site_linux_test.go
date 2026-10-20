package service

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestWriteSiteFileUnderServiceUmask(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	previous := syscall.Umask(0o077)
	defer syscall.Umask(previous)
	file := filepath.Join(root, "new", "nested", "index.html")
	if err := writeSiteFile(file, []byte("website"), 0o644); err != nil {
		t.Fatal(err)
	}
	for path, mode := range map[string]os.FileMode{
		root:                       0o700,
		filepath.Join(root, "new"): 0o755,
		filepath.Dir(file):         0o755,
		file:                       0o644,
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != mode {
			t.Fatalf("%s permissions: got %o, want %o", path, info.Mode().Perm(), mode)
		}
	}
}
