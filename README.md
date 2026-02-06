# ccs - Claude Code Summary

A fast CLI tool that reads your local `~/.claude/` data and displays usage summaries in the terminal.

Unlike server-based analytics, `ccs` works entirely offline — no syncing, no accounts, no network requests.

## Install

```bash
# Build from source (requires Go 1.21+)
make build

# Install to ~/.local/bin
make install
```

## Commands

### Dashboard (default)

```bash
ccs
```

Shows an overview: total sessions, messages, token usage, model breakdown, peak hours, and longest session.

### Time Periods

```bash
ccs today
ccs week
ccs month
```

Activity summary for the given period with daily breakdown and session list.

### Projects

```bash
ccs projects
```

Ranks all projects by message count with activity bars.

### Sessions

```bash
ccs sessions                    # Last 20 sessions
ccs sessions -n 50              # Last 50 sessions
ccs sessions --project=myapp    # Filter by project name
```

### Session Detail

```bash
ccs session <id>                # Full ID or prefix
ccs session 660223fe            # Partial ID match
```

Shows messages, token usage, tool usage breakdown, and conversation prompts.

### Token Usage

```bash
ccs tokens
```

Token breakdown by model and daily output token chart.

### Help

```bash
ccs help
ccs version
```

## Data Sources

`ccs` reads from `~/.claude/` (read-only, never writes):

| File | Used By | Size |
|------|---------|------|
| `stats-cache.json` | `summary`, `today/week/month`, `tokens` | ~10 KB |
| `projects/*/sessions-index.json` | `projects`, `sessions` | ~1 KB each |
| `projects/*/*.jsonl` | `session <id>` | Varies |
| `history.jsonl` | — (reserved) | ~1 MB |

## Performance

- **Dashboard / Tokens**: Reads only `stats-cache.json` — instant
- **Projects / Sessions**: Scans index JSON files only — <100ms
- **Session Detail**: Parses one JSONL file — <2s even for large sessions

## Color Support

- Colors are auto-disabled when piping output (`ccs | cat`)
- Set `NO_COLOR=1` to force disable colors
- Works with any terminal that supports ANSI escape codes

## License

[MIT](LICENSE)
