# diskwhy

Your disk is full. But why?

diskwhy is a developer-focused CLI tool that scans your disk, identifies what is consuming space, and safely removes the things you no longer need — node_modules, Python caches, stale Git objects, Docker leftovers, and more.

---

## Install

### From source

```bash
git clone https://github.com/dhananjay6561/diskwhy
cd diskwhy
make install
```

Requires Go 1.21 or later.

### Build locally

```bash
make build
./diskwhy --help
```

---

## Quick start

```bash
# Interactive shell (default when no subcommand is given)
diskwhy

# Quick scan of your home directory
diskwhy scan

# Deep scan of a specific path
diskwhy scan --deep --path /Users/you/projects

# Preview what would be cleaned, no changes
diskwhy clean --all --dry-run

# Clean everything, move to Trash, no prompt
diskwhy clean --all --trash --yes
```

---

## Commands

### `diskwhy scan`

Scans for space-consuming items and shows a ranked list with size and staleness.

```
diskwhy scan [flags]

Flags:
  -p, --path string   Scan a specific directory instead of home
      --deep          Also scan /usr/local, /opt/homebrew, /var, etc.
      --json          Output as JSON (schema_version: 1)
      --verbose       Show per-file timing, resolved paths, and diagnostic info
```

**Scan modes**

| Mode    | Trigger                  | What is scanned                                     |
|---------|--------------------------|-----------------------------------------------------|
| `quick` | no `--path`, no `--deep` | Home directory only                                 |
| `deep`  | `--deep`                 | Home + system dirs (/usr/local, Homebrew, /var, …)  |
| `path`  | `--path <dir>`           | The specified directory only                        |

**Categories detected**

| Category            | Description                                       |
|---------------------|---------------------------------------------------|
| `node_modules`      | npm/yarn/pnpm dependency trees                    |
| `git_objects`       | Loose objects and pack files in `.git`            |
| `docker`            | Dangling images, unused volumes (via Docker API)  |
| `pycache`           | Python `__pycache__` and `.pyc` bytecode          |
| `pip_cache`         | pip download/wheel cache                          |
| `npm_cache`         | npm's global cache (`~/.npm`)                     |
| `brew_cache`        | Homebrew downloads cache (macOS)                  |
| `xcode_derived`     | Xcode DerivedData (macOS)                         |
| `apt_cache`         | apt package cache (Linux)                         |
| `snap_cache`        | Snap package cache (Linux)                        |
| `logs`              | Log files (`.log`, `*.log.*`)                     |
| `trash`             | Items in system Trash / `.Trash`                  |

**Staleness scores**

| Score     | Meaning                                        |
|-----------|------------------------------------------------|
| `unused`  | Not accessed in > 90 days (default threshold)  |
| `stale`   | Not accessed in 30–90 days                     |
| `recent`  | Accessed in the last 30 days                   |
| `active`  | Modified within the last 7 days                |
| `unknown` | Access time not available on this platform     |

---

### `diskwhy clean`

Runs a deep scan then removes the selected categories.

```
diskwhy clean [flags]

Category flags:
      --all       All cleanable categories
      --node      node_modules
      --cache     All cache categories (brew, npm, pip, apt, pycache)
      --git       Git objects (runs git gc)
      --logs      Log files
      --trash     Empty system Trash

Safety flags:
      --dry-run   Show what would be deleted, make no changes
      --trash     Move files to Trash instead of permanent delete
  -y, --yes       Skip confirmation prompt

Output flags:
      --json      Output as JSON (schema_version: 1)
      --verbose   Show per-item detail
```

**Safety rules**

- `active` items are never deleted unless the category is always-safe (caches, trash)
- `git_objects` runs `git gc --prune=now` rather than deleting `.git` directly
- `docker` pruning uses the Docker API — no `exec.Command("docker", ...)` shell-out
- Paths outside your home directory and known system dirs are blocked
- `os.Remove` / `os.RemoveAll` are confined to a single file (`internal/clean/safe.go`)

---

### `diskwhy shell`

Opens an interactive REPL with slash commands.

```
diskwhy shell
```

Available slash commands inside the shell:

```
/scan [--deep] [--path <dir>]
/clean [--all|--node|--cache|--git|--logs|--trash] [--dry-run] [--yes]
/version
/home
/help
/clear
/exit
```

---

### Shell completion

```bash
# Bash
diskwhy completion bash > /usr/local/etc/bash_completion.d/diskwhy

# Zsh
diskwhy completion zsh > "${fpath[1]}/_diskwhy"

# Fish
diskwhy completion fish > ~/.config/fish/completions/diskwhy.fish

# PowerShell
diskwhy completion powershell | Out-String | Invoke-Expression
```

---

## JSON output

Both `scan` and `clean` support `--json`. The output schema is documented in [SCHEMA_CHANGELOG.md](SCHEMA_CHANGELOG.md).

```bash
diskwhy scan --json | jq '.summary'
diskwhy clean --all --dry-run --json | jq '.results[] | select(.outcome == "dry_run")'
```

---

## Configuration

diskwhy reads config from `~/.config/diskwhy/config.yaml` (XDG) or `~/.diskwhy.yaml`.

```yaml
no_color: false
json: false
verbose: false
```

All config keys correspond to the persistent flags (`--no-color`, `--json`, `--verbose`).

---

## Exit codes

| Code | Meaning                                |
|------|----------------------------------------|
| `0`  | Success                                |
| `1`  | General error                          |
| `2`  | Bad arguments / unknown flag           |
| `3`  | Permission denied                      |
| `4`  | Docker unavailable (non-fatal warning) |
| `130`| Cancelled by SIGINT / Ctrl-C           |

---

## Development

```bash
make build    # build the binary
make test     # run tests
make vet      # go vet
make lint     # vet + staticcheck
make install  # go install to $GOPATH/bin
```

CI runs on Go 1.21, 1.22, 1.23 across ubuntu-latest and macos-latest with a 70% coverage gate.

---

## Project layout

```
cmd/            cobra commands (root, scan, clean, shell)
internal/
  scan/         directory walker, category detection, staleness scoring
  clean/        outcome engine, git gc, safe delete, trash integration
  docker/       Docker SDK client (images, volumes, prune)
  tui/          terminal capability detection
  jsonout/      JSON schema types and writers
  config/       viper config loading
  errtype/      typed exit-code errors
  exitcode/     exit code constants
  trash/        XDG trash (Linux) and ~/.Trash bridge (macOS)
```
