package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pradyb/zsh-history-cleaner/internal/history"
)

const defaultHistoryFile = "~/.zsh_history"

// Injected at build time via -ldflags by .github/workflows/release.yml.
var (
	version   = "dev"
	commitSHA = "none"
	buildDate = "unknown"
)

func main() {
	var (
		histFile      = flag.String("file", defaultHistoryFile, "Path to the zsh history file")
		dryRun        = flag.Bool("dry-run", false, "Preview changes without writing")
		noDeduplicate = flag.Bool("no-dedup", false, "Skip deduplication")
		removeSecrets = flag.Bool("remove-secrets", false, "Remove lines that look like credentials/tokens (opt-in)")
		noTypos       = flag.Bool("no-typos", false, "Skip typo filtering")
		noSort        = flag.Bool("no-sort", false, "Skip sorting entries by timestamp")
		minLen        = flag.Int("min-len", 2, "Drop commands shorter than this many characters (0 to disable)")
		stats         = flag.Bool("stats", false, "Print removal statistics")
		showVersion   = flag.Bool("version", false, "Print version information and exit")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: zhc [flags]\n\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  zhc --dry-run --stats
  zhc --file ~/.zsh_history --min-len 3
  zhc --remove-secrets --stats
  zhc --no-typos --no-sort --stats
`)
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("zhc %s (commit %s, built %s)\n", version, commitSHA, buildDate)
		return
	}

	path := expandHome(*histFile)

	opts := history.CleanOptions{
		Deduplicate:     !*noDeduplicate,
		MinLength:       *minLen,
		RemoveSecrets:   *removeSecrets,
		RemoveTypos:     !*noTypos,
		SortByTimestamp: !*noSort,
	}

	entries, err := history.Parse(path)
	if err != nil {
		fatalf("parse: %v", err)
	}
	before := len(entries)

	cleaned := history.Clean(entries, opts)
	after := len(cleaned)

	if *stats || *dryRun {
		printStats(before, after, opts)
	}

	if *dryRun {
		fmt.Println("\n--dry-run: no changes written.")
		return
	}

	backup, err := history.Write(path, cleaned)
	if err != nil {
		fatalf("write: %v", err)
	}
	fmt.Printf("Written %d entries to %s\n", after, path)
	fmt.Printf("Backup saved to %s\n", backup)
}

func printStats(before, after int, opts history.CleanOptions) {
	removed := before - after
	fmt.Printf("Entries before : %d\n", before)
	fmt.Printf("Entries after  : %d\n", after)
	fmt.Printf("Removed        : %d (%.1f%%)\n", removed, percent(removed, before))
	fmt.Println()
	fmt.Println("Passes enabled:")
	fmt.Printf("  %-20s %v\n", "deduplication", opts.Deduplicate)
	fmt.Printf("  %-20s %v\n", "secret filter", opts.RemoveSecrets)
	fmt.Printf("  %-20s %v\n", "typo filter", opts.RemoveTypos)
	fmt.Printf("  %-20s %v (min %d chars)\n", "length filter", opts.MinLength > 0, opts.MinLength)
	fmt.Printf("  %-20s %v\n", "sort by timestamp", opts.SortByTimestamp)
}

func percent(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			fatalf("home dir: %v", err)
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
