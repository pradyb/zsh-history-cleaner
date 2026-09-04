package history

import "testing"

func TestClean(t *testing.T) {
	entries := []Entry{
		{Timestamp: 3, Command: "l"},                       // too short, dropped
		{Timestamp: 1, Command: "gits status"},             // typo (1-edit of git), dropped
		{Timestamp: 2, Command: "git status"},              // duplicate, kept (last)
		{Timestamp: 4, Command: "git status"},              // duplicate, kept
		{Timestamp: 5, Command: "export PASSWORD=hunter2"}, // secret, dropped
		{Timestamp: 0, Command: "echo hello"},              // plain, kept
	}

	opts := DefaultOptions()
	opts.RemoveSecrets = true // opt-in, off by default
	got := Clean(entries, opts)

	want := []string{"echo hello", "git status"}
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Command != w {
			t.Errorf("entry %d = %q, want %q", i, got[i].Command, w)
		}
	}
}
