package history

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Entry represents a single zsh extended history line.
// Format: ": <timestamp>:<elapsed>;<command>"
// Plain (non-extended) lines are stored with Timestamp == 0.
type Entry struct {
	Timestamp int64
	Elapsed   int
	Command   string
}

// Parse reads a zsh history file and returns all entries.
// Handles both plain and EXTENDED_HISTORY formats.
// Multi-line commands (continuation with `\`) are joined.
func Parse(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open history file: %w", err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	// zsh history lines can be long with heredocs etc.
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var current *Entry

	for scanner.Scan() {
		line := scanner.Text()

		// Continuation of a multi-line command
		if current != nil {
			if strings.HasSuffix(current.Command, "\\") {
				current.Command = strings.TrimSuffix(current.Command, "\\") + "\n" + line
				if !strings.HasSuffix(line, "\\") {
					entries = append(entries, *current)
					current = nil
				}
				continue
			}
		}

		// Extended history format: ": 1715000000:0;command"
		if strings.HasPrefix(line, ": ") {
			e, ok := parseExtended(line)
			if ok {
				if strings.HasSuffix(e.Command, "\\") {
					current = &e
				} else {
					entries = append(entries, e)
				}
				continue
			}
		}

		// Plain history line
		if line != "" {
			e := Entry{Command: line}
			if strings.HasSuffix(line, "\\") {
				current = &e
			} else {
				entries = append(entries, e)
			}
		}
	}

	// Flush any dangling entry (file ended mid-continuation)
	if current != nil {
		entries = append(entries, *current)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan history file: %w", err)
	}
	return entries, nil
}

// Serialize converts an entry back to its zsh history line.
func (e Entry) Serialize() string {
	if e.Timestamp > 0 {
		return fmt.Sprintf(": %d:%d;%s", e.Timestamp, e.Elapsed, e.Command)
	}
	return e.Command
}

// parseExtended parses a single extended history line.
func parseExtended(line string) (Entry, bool) {
	// Strip leading ": "
	rest := line[2:]
	// Split on ";" — everything after first ";" is the command
	semicolon := strings.Index(rest, ";")
	if semicolon < 0 {
		return Entry{}, false
	}
	meta := rest[:semicolon]
	cmd := rest[semicolon+1:]

	parts := strings.SplitN(meta, ":", 2)
	if len(parts) != 2 {
		return Entry{}, false
	}
	ts, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return Entry{}, false
	}
	elapsed, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		elapsed = 0
	}
	return Entry{Timestamp: ts, Elapsed: elapsed, Command: cmd}, true
}
