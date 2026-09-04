# zsh-history-cleaner

A fast, zero-dependency CLI tool written in Go to clean up your `~/.zsh_history` file.

Supports both plain and `EXTENDED_HISTORY` formats. A timestamped backup is always written before any changes are made.

## Features

| Pass | Default | Description |
|---|---|---|
| **Deduplicate** | on | Keep only the last occurrence of each command |
| **Typo filter** | on | Remove commands whose first word is a 1-edit-distance typo of a known program (`doker`, `pythn`, `crl`, …) |
| **Length filter** | on (≥2) | Drop commands shorter than N characters |
| **Secret filter** | **off** | Opt-in: strip lines that look like credentials / tokens |
| **Sort by timestamp** | **on** | Re-order entries by timestamp after all other passes (fixes out-of-order extended history) |

## Installation

### Download a release binary

Grab a prebuilt binary from the [releases page](https://github.com/pradyb/zsh-history-cleaner/releases),
or fetch the latest for your platform:

```sh
OS=$(uname -s | tr '[:upper:]' '[:lower:]')   # linux | darwin
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -fsSL -o zhc "https://github.com/pradyb/zsh-history-cleaner/releases/latest/download/zhc-${OS}-${ARCH}"
chmod +x zhc && mv zhc ~/bin/
```

Each release ships `checksums.txt` for verification.

### Build from source

```sh
git clone https://github.com/pradyb/zsh-history-cleaner
cd zsh-history-cleaner
go build -o ~/bin/zhc ./cmd/zhc
```

Make sure `~/bin` is on your `$PATH`, or install directly:

```sh
go install github.com/pradyb/zsh-history-cleaner/cmd/zhc@latest
```

## Usage

```
zhc [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--file` | `~/.zsh_history` | Path to the history file |
| `--dry-run` | false | Preview changes without writing anything |
| `--stats` | false | Print a removal summary |
| `--no-dedup` | false | Skip deduplication |
| `--no-typos` | false | Skip typo filtering |
| `--remove-secrets` | false | **Opt-in**: remove lines matching credential patterns |
| `--no-sort` | false | Skip timestamp sorting |
| `--min-len` | 2 | Drop commands shorter than N chars (0 to disable) |
| `--version` | false | Print version, commit and build date |

### Examples

```sh
# Preview what would be removed — nothing is written
zhc --dry-run --stats

# Clean with default passes (dedup + typos + length filter)
zhc --stats

# Also strip credentials (sorting is on by default)
zhc --remove-secrets --stats

# Stricter length filter, skip typo detection
zhc --min-len 4 --no-typos

# Clean a custom history file
zhc --file ~/.config/zsh/history --stats
```

### Sample output

```
Entries before : 1131
Entries after  : 331
Removed        : 800 (70.7%)

Passes enabled:
  deduplication        true
  secret filter        false
  typo filter          true
  length filter        true (min 2 chars)
  sort by timestamp    true

Written 331 entries to /Users/you/.zsh_history
Backup saved to /Users/you/.zsh-history-cleaner/backup/20260511-230614
```

## Backups

Every run (non-dry-run) writes a copy of the original history file to:

```
~/.zsh-history-cleaner/backup/<YYYYMMDD-HHMMSS>
```

The directory is created automatically with `0700` permissions. You can list or restore backups at any time:

```sh
# List all backups
ls -lh ~/.zsh-history-cleaner/backup/

# Restore a specific backup
cp ~/.zsh-history-cleaner/backup/20260511-230614 ~/.zsh_history
```

## Project structure

```
zsh-history-cleaner/
├── cmd/
│   └── zhc/
│       └── main.go          # CLI entry point (flags, output)
├── internal/
│   └── history/
│       ├── entry.go         # Parse & serialize (plain + EXTENDED_HISTORY)
│       ├── clean.go         # Dedup, typo, secret, length, sort passes
│       └── write.go         # Atomic write + backup
└── go.mod
```

## Secret patterns detected (when `--remove-secrets` is set)

- `PASSWORD=`, `SECRET=`, `TOKEN=`, `API_KEY=`, `AUTH=`, `CREDENTIAL=`
- `Authorization: Bearer …` / `x-api-key: …` (curl headers)
- `AWS_SECRET=`, `AWS_ACCESS=`, `PRIVATE_KEY=`

## Typo detection

Uses [Levenshtein distance](https://en.wikipedia.org/wiki/Levenshtein_distance) = 1 against a built-in list of common programs:

`git`, `grep`, `curl`, `sudo`, `ssh`, `cat`, `echo`, `make`, `docker`, `kubectl`, `go`, `npm`, `yarn`, `python`, `python3`, `ls`, `cd`, `cp`, `mv`, `rm`, `find`, `awk`, `sed`

Examples of commands that get filtered: `doker ps`, `pythn app.py`, `crl example.com`.

Note: distance 1 means a single insertion, deletion, or substitution — not a
transposition. Swapped-letter typos like `gti` (git) or `dockre` (docker) are
2 edits away and are *not* caught.
