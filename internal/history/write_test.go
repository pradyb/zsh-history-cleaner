package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteBacksUpAndReplacesFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path := filepath.Join(t.TempDir(), "zsh_history")
	original := "old command\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("seed original file: %v", err)
	}

	entries := []Entry{
		{Command: "git status"},
		{Timestamp: 1715000000, Elapsed: 0, Command: "ls -la"},
	}
	backupPath, err := Write(path, entries)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(backupContent) != original {
		t.Errorf("backup content = %q, want %q", backupContent, original)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	want := "git status\n: 1715000000:0;ls -la\n"
	if string(got) != want {
		t.Errorf("written content = %q, want %q", got, want)
	}
}

func TestWriteNoExistingFileSkipsBackupCopy(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path := filepath.Join(t.TempDir(), "zsh_history")

	backupPath, err := Write(path, []Entry{{Command: "echo hi"}})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Errorf("expected no backup file since original didn't exist, got err=%v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if want := "echo hi\n"; string(got) != want {
		t.Errorf("written content = %q, want %q", got, want)
	}
}

func TestBackupDirHasRestrictedPermissions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path := filepath.Join(t.TempDir(), "zsh_history")
	if err := os.WriteFile(path, []byte("x\n"), 0o600); err != nil {
		t.Fatalf("seed original file: %v", err)
	}

	backupPath, err := Write(path, []Entry{{Command: "x"}})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	info, err := os.Stat(filepath.Dir(backupPath))
	if err != nil {
		t.Fatalf("stat backup dir: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("backup dir perm = %o, want %o", perm, 0o700)
	}
}
