# Repository Guidelines

## Project Structure

- `cmd/memelink/`: CLI entrypoint
- `internal/`: implementation
  - `cmd/`: command routing (kong CLI framework)
  - `api/`: Memegen.link API client
  - `actions/`: meme generation actions
  - `cache/`: template caching
  - `config/`: YAML configuration
  - `encoding/`: text encoding utilities
  - `outfmt/`: output formatting
  - `preview/`: meme preview
  - `tui/`: Bubbletea/bubbles interactive UI
  - `ui/`: UI components
- `setup-token/`: helper for auth setup
- `bin/`: build outputs

## Build, Test, and Development Commands

- `make build`: compile to `bin/memelink` (symbols stripped for size)
- `make run ARGS="..."`: build + run with arguments
- `make install`: install to GOPATH or /usr/local/bin
- `make fmt` / `make lint` / `make test` / `make ci`: format, lint, test, full local gate
- `make tools`: install pinned dev tools into `.tools/`
- `make clean`: remove bin/ and .tools/

## Coding Style & Naming Conventions

- Formatting: `make fmt` (goimports local prefix `github.com/dedene/memelink-cli` + gofumpt)
- Output: keep stdout parseable; send human hints/progress to stderr
- Linting: golangci-lint v2.8.0 with 13 linters + formatters
- TUI: use Charmbracelet ecosystem (bubbletea + bubbles + lipgloss)

## Testing Guidelines

- Unit tests: stdlib `testing` + `stretchr/testify`
- 15 test files; comprehensive coverage
- CI gate: fmt-check, lint, test

## Config & Secrets

- **No authentication**: Memegen.link is stateless
- **Config file**: YAML-based configuration
- **Caching**: template cache for performance

## Key Features

- Generate memes from templates
- Interactive TUI for meme creation
- Clipboard integration
- Browser preview
- JSON5 parsing support

## Commit & Pull Request Guidelines

- Conventional Commits: `feat|fix|refactor|build|ci|chore|docs|style|perf|test`
- Group related changes; avoid bundling unrelated refactors
- PR review: use `gh pr view` / `gh pr diff`; don't switch branches

## Security Tips

- No credentials to manage (stateless API)
- Clipboard may contain generated meme URLs
