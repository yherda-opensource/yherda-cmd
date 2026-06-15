# YOS-55: Idea Documents CLI

## Summary

Add a `yherda docs` subcommand group for listing, showing, creating, and updating IdeaDocument records. All commands map to existing backend API endpoints — no backend changes required.

## API Endpoints

| Command | Method | Endpoint |
|---|---|---|
| `docs list` | GET | `/api/storyline/{id}/documents/` |
| `docs show` | GET | `/api/ideadocument/{id}/` |
| `docs create` | POST | `/api/storyline/{id}/documents/` |
| `docs update` | PATCH | `/api/ideadocument/{id}/` |

Fields: `id`, `title`, `body`, `created`, `updated`. The `body` field contains markdown; the backend handles encryption/decryption transparently.

## Commands

```
yherda docs list --idea <idea-id>
yherda docs show <doc-id>
yherda docs create --idea <idea-id> --title <title> [--file <file.md>]
yherda docs update <doc-id> [--file <file.md>]
```

- `--file` reads content from a file path; if omitted, content is read from stdin
- `--json` (global flag) outputs full JSON metadata on any command
- `docs show` prints the `body` field (markdown) to stdout by default

## Architecture

- New file: `cmd/docs.go` — four cobra commands + `readContent` helper
- New method: `client.Patch` in `internal/api/client.go` — mirrors `client.Post` with PATCH verb
- `readContent(cmd)` checks `--file` first; falls back to `io.ReadAll(os.Stdin)`

## Operational Requirements

### Healthy State
All four commands complete successfully. `list` returns tabular or JSON output scoped to the specified idea. `show` renders `body` as markdown to stdout. `create` and `update` accept a file or stdin and return the created/updated document as JSON.

### Failure State
- Non-200 API response → error on stderr, non-zero exit
- Missing required flag (`--idea`, `--title`) → usage error before any API call
- `--file` path not found → file-not-found error before API call
- Stdin with no data → blocks indefinitely (known v1 limitation, no timeout)

### Observability Signal
- Exit code (0 vs. non-zero)
- Stderr for error messages
- `DEVELOPER=1` enables per-request logging to stderr

### Gaps
None that block Ready. Stdin-hang is a known v1 limitation.
