//go:build !windows

package osutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommitTempFilePreservesPrivatePermissions(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cache.bin")
	tmp, err := CreateTempSibling(target)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write([]byte("private cache")); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	if before.Mode().Perm()&0o077 != 0 {
		t.Fatalf("temporary file is not private: %04o", before.Mode().Perm())
	}
	if err := CommitTempFile(tmp.Name(), target); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if after.Mode().Perm() != before.Mode().Perm() {
		t.Fatalf("commit changed permissions: before=%04o after=%04o", before.Mode().Perm(), after.Mode().Perm())
	}
}
