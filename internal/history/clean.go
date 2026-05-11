package history

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// CleanOptions controls which cleaning passes to run.
type CleanOptions struct {
	// Deduplicate removes earlier occurrences, keeping the most recent run.
	Deduplicate bool
	// MinLength drops commands shorter than this (e.g. bare "ls", "cd").
	MinLength int
	// RemoveSecrets strips lines that look like they contain credentials.
	// Off by default — must be explicitly opted in.
	RemoveSecrets bool
	// RemoveTypos strips commands whose first word looks like a typo of a
	// known program (transposed letters, common mis-spellings).
	RemoveTypos bool
	// SortByTimestamp re-orders entries by timestamp after all other passes.
	// Useful because extended history can get out of order across sessions.
	SortByTimestamp bool
}

// DefaultOptions returns a sensible default configuration.
func DefaultOptions() CleanOptions {
	return CleanOptions{
		Deduplicate:     true,
		MinLength:       2,
		RemoveSecrets:   false, // opt-in
		RemoveTypos:     true,
		SortByTimestamp: true,
	}
}

// Clean applies all enabled passes and returns the cleaned list.
func Clean(entries []Entry, opts CleanOptions) []Entry {
	if opts.RemoveSecrets {
		entries = filterSecrets(entries)
	}
	if opts.RemoveTypos {
		entries = filterTypos(entries)
	}
	if opts.MinLength > 0 {
		entries = filterShort(entries, opts.MinLength)
	}
	if opts.Deduplicate {
		entries = deduplicate(entries)
	}
	if opts.SortByTimestamp {
		entries = sortByTimestamp(entries)
	}
	return entries
}

// ---------------------------------------------------------------------------
// Sort by timestamp
// ---------------------------------------------------------------------------

// sortByTimestamp re-orders entries ascending by Timestamp.
// Entries without a timestamp (plain history) are placed at the front.
func sortByTimestamp(entries []Entry) []Entry {
	out := make([]Entry, len(entries))
	copy(out, entries)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Timestamp < out[j].Timestamp
	})
	return out
}

// ---------------------------------------------------------------------------
// Deduplication
// ---------------------------------------------------------------------------

// deduplicate keeps only the last occurrence of each unique command,
// preserving final order.
func deduplicate(entries []Entry) []Entry {
	seen := make(map[string]int, len(entries)) // command -> last index
	for i, e := range entries {
		seen[e.Command] = i
	}
	out := make([]Entry, 0, len(seen))
	for i, e := range entries {
		if seen[e.Command] == i {
			out = append(out, e)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Short command filter
// ---------------------------------------------------------------------------

func filterShort(entries []Entry, minLen int) []Entry {
	out := entries[:0]
	for _, e := range entries {
		if len([]rune(strings.TrimSpace(e.Command))) >= minLen {
			out = append(out, e)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Secret filter
// ---------------------------------------------------------------------------

var secretPatterns = []*regexp.Regexp{
	// export FOO=bar  /  FOO=bar
	regexp.MustCompile(`(?i)(password|passwd|secret|token|api[_-]?key|auth|credential)\s*=\s*\S+`),
	// curl -H "Authorization: Bearer ..."
	regexp.MustCompile(`(?i)(authorization|bearer|x-api-key)\s*[:=]\s*\S+`),
	// AWS/GCP-style keys (long random strings after =)
	regexp.MustCompile(`(?i)(aws_secret|aws_access|private_key)\s*=`),
}

func filterSecrets(entries []Entry) []Entry {
	out := entries[:0]
	for _, e := range entries {
		if !looksLikeSecret(e.Command) {
			out = append(out, e)
		}
	}
	return out
}

func looksLikeSecret(cmd string) bool {
	for _, re := range secretPatterns {
		if re.MatchString(cmd) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Typo filter
// ---------------------------------------------------------------------------

// knownCommands is the set of programs we watch for typos of.
var knownCommands = []string{
	"git", "grep", "curl", "sudo", "ssh", "cat", "echo", "make",
	"docker", "kubectl", "go", "npm", "yarn", "python", "python3",
	"ls", "cd", "cp", "mv", "rm", "find", "awk", "sed",
}

func filterTypos(entries []Entry) []Entry {
	out := entries[:0]
	for _, e := range entries {
		if !isTypo(e.Command) {
			out = append(out, e)
		}
	}
	return out
}

func isTypo(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	// Extract the first word (the program name)
	word := firstWord(cmd)
	if word == "" {
		return false
	}
	// If it's a known command it's definitely not a typo
	for _, k := range knownCommands {
		if word == k {
			return false
		}
	}
	// Check edit distance against every known command
	for _, k := range knownCommands {
		if levenshtein(word, k) == 1 {
			return true
		}
	}
	return false
}

func firstWord(s string) string {
	f := strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r)
	})
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min3(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
