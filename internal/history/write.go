package history

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Write serializes entries back to path.
// Before writing it creates a backup under ~/.zsh-history-cleaner/backup/<timestamp>.
// Returns the backup path so callers can report it.
func Write(path string, entries []Entry) (backupPath string, err error) {
	backupPath, err = backup(path)
	if err != nil {
		return "", err
	}

	// Write to a temp file first, then rename — atomic on POSIX.
	tmp, err := os.CreateTemp(os.TempDir(), "zsh-history-*.tmp")
	if err != nil {
		return backupPath, fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	w := bufio.NewWriter(tmp)
	for _, e := range entries {
		if _, err := fmt.Fprintln(w, e.Serialize()); err != nil {
			tmp.Close()
			os.Remove(tmpName)
			return backupPath, fmt.Errorf("write entry: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return backupPath, fmt.Errorf("flush: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return backupPath, fmt.Errorf("close temp: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return backupPath, fmt.Errorf("replace history file: %w", err)
	}
	return backupPath, nil
}

// backup copies the history file into ~/.zsh-history-cleaner/backup/<timestamp>
// and returns the destination path.
func backup(path string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	baseDir := filepath.Join(home, ".zsh-history-cleaner", "backup")
	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}
	timestamp := time.Now().Format("20060102-150405")
	dest := filepath.Join(baseDir, timestamp)

	src, err := os.Open(path)
	if err != nil {
		// No existing file — nothing to back up, that's fine.
		if os.IsNotExist(err) {
			return dest, nil
		}
		return "", fmt.Errorf("open for backup: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("create backup: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("write backup: %w", err)
	}
	return dest, nil
}
