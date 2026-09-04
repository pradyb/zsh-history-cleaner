package history

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempHistory(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "zsh_history")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp history: %v", err)
	}
	return path
}

func TestParsePlain(t *testing.T) {
	path := writeTempHistory(t, "git status\n\nls -la\n")

	entries, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	want := []string{"git status", "ls -la"}
	if len(entries) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(entries), len(want), entries)
	}
	for i, w := range want {
		if entries[i].Command != w || entries[i].Timestamp != 0 {
			t.Errorf("entry %d = %+v, want Command %q, Timestamp 0", i, entries[i], w)
		}
	}
}

func TestParseExtended(t *testing.T) {
	path := writeTempHistory(t, ": 1715000000:5;git status\n: 1715000010:0;ls -la\n")

	entries, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2: %+v", len(entries), entries)
	}
	if entries[0] != (Entry{Timestamp: 1715000000, Elapsed: 5, Command: "git status"}) {
		t.Errorf("entry 0 = %+v", entries[0])
	}
	if entries[1] != (Entry{Timestamp: 1715000010, Elapsed: 0, Command: "ls -la"}) {
		t.Errorf("entry 1 = %+v", entries[1])
	}
}

func TestParseMultiLineContinuation(t *testing.T) {
	path := writeTempHistory(t, ": 1715000000:0;echo foo \\\necho bar\ngit status\n")

	entries, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2: %+v", len(entries), entries)
	}
	if want := "echo foo \necho bar"; entries[0].Command != want {
		t.Errorf("entry 0 Command = %q, want %q", entries[0].Command, want)
	}
	if entries[1].Command != "git status" {
		t.Errorf("entry 1 Command = %q, want %q", entries[1].Command, "git status")
	}
}

func TestParseMalformedExtendedLineFallsBackToPlain(t *testing.T) {
	// Starts with ": " but has no ";" separator, so parseExtended fails
	// and the whole line is kept verbatim as a plain command.
	path := writeTempHistory(t, ": not really extended\n")

	entries, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(entries) != 1 || entries[0].Command != ": not really extended" {
		t.Fatalf("got %+v", entries)
	}
}

func TestParseDanglingContinuationAtEOF(t *testing.T) {
	path := writeTempHistory(t, "echo foo \\\n")

	entries, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(entries) != 1 || entries[0].Command != "echo foo \\" {
		t.Fatalf("got %+v", entries)
	}
}

func TestParseMissingFile(t *testing.T) {
	_, err := Parse(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestSerialize(t *testing.T) {
	cases := []struct {
		entry Entry
		want  string
	}{
		{Entry{Command: "git status"}, "git status"},
		{Entry{Timestamp: 1715000000, Elapsed: 5, Command: "git status"}, ": 1715000000:5;git status"},
	}
	for _, c := range cases {
		if got := c.entry.Serialize(); got != c.want {
			t.Errorf("Serialize(%+v) = %q, want %q", c.entry, got, c.want)
		}
	}
}
