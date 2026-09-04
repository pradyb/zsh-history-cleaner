# Security Policy

## Supported Versions

Only the latest commit on `main` is actively supported with security fixes.

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report security issues privately by emailing: **pradeep.devlabs@gmail.com**

Include:
- A description of the vulnerability and its potential impact
- Steps to reproduce (proof-of-concept if possible)
- Your suggested fix or mitigation (optional)

You can expect an acknowledgement within **72 hours**.

## Handling of Shell History

`zsh-history-cleaner` reads and rewrites `~/.zsh_history`, which can contain
real credentials pasted into past commands. Take these precautions:

- The tool always writes a backup to `~/.zsh-history-cleaner/backup/` before
  modifying the original file — never delete backups until you've verified
  the cleaned output.
- `--remove-secrets` is heuristic (pattern-based) and opt-in — it will not
  catch every credential shape, so review `--dry-run --stats` output before
  trusting it to scrub a file you intend to share.
- Never paste real history content (especially matches from the secret
  filter) into a public GitHub issue.

## Scope

In scope: vulnerabilities in the parsing, filtering, or file-writing logic
that could corrupt history data, escape the intended file path, or leak
credentials during processing.

Out of scope: the completeness of the secret-detection regexes as a security
control — they are a convenience heuristic, not a guarantee.
