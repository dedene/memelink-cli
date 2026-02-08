# 🔗 memelink-cli - Memes from the terminal

[![CI](https://github.com/dedene/memelink-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/dedene/memelink-cli/actions/workflows/ci.yml)
[![Go](https://img.shields.io/github/go-mod/go-version/dedene/memelink-cli)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Generate memes from the terminal using the [Memegen.link](https://memegen.link) API.

## Install

**Homebrew:**

```sh
brew install dedene/tap/memelink-cli
```

**Go:**

```sh
go install github.com/dedene/memelink-cli/cmd/memelink@latest
```

**Binary:** download from [Releases](https://github.com/dedene/memelink-cli/releases).

## Quick start

```sh
# Auto-generate — API picks the best template
memelink "One does not simply mass produce memes"

# Template-based — specify template ID + text lines
memelink drake "Writing memes by hand" "Using memelink"

# Custom background
memelink custom --background https://example.com/photo.jpg "Top text" "Bottom text"

# Browse templates interactively (TUI picker)
memelink templates
```

## Commands

| Command     | Aliases    | Description                                 |
| ----------- | ---------- | ------------------------------------------- |
| `generate`  | `gen`, `g` | Generate a meme (default command)           |
| `templates` | `ls`       | List templates or launch interactive picker |
| `fonts`     |            | List available fonts                        |
| `config`    |            | Manage configuration                        |
| `version`   |            | Print version info                          |

`generate` is the default — bare `memelink "text"` works without typing it.

## Generate flags

| Flag                         | Short | Description                                   |
| ---------------------------- | ----- | --------------------------------------------- |
| `--format`                   | `-f`  | Image format: jpg, png, gif, webp             |
| `--font`                     |       | Font ID or alias                              |
| `--layout`                   |       | Text layout: default, top                     |
| `--text-color`               |       | Text color per line (repeatable)              |
| `--style`                    |       | Style name or overlay URL (repeatable)        |
| `--width`                    |       | Image width in pixels                         |
| `--height`                   |       | Image height in pixels                        |
| `--safe`                     |       | Filter NSFW content                           |
| `--background`               |       | Background image URL (with `custom` template) |
| `--copy`                     | `-c`  | Copy URL to clipboard                         |
| `--open`                     | `-o`  | Open URL in browser                           |
| `--output`                   |       | Download image to file path                   |
| `-O`                         |       | Download with auto-generated filename         |
| `--preview` / `--no-preview` |       | Inline image preview (on by default in TTY)   |

## Configuration

Config file: `~/.config/memelink/config.json` (JSON5 readable).

```sh
memelink config set default_format png
memelink config get default_format
memelink config unset default_format
memelink config list
memelink config path
```

| Key              | Values                   | Description                           |
| ---------------- | ------------------------ | ------------------------------------- |
| `default_format` | jpg, png, gif, webp      | Default image format                  |
| `default_font`   | any font ID              | Default font                          |
| `default_layout` | default, top             | Default text layout                   |
| `safe`           | true, false              | Filter NSFW content                   |
| `auto_copy`      | true, false              | Auto-copy URL to clipboard            |
| `auto_open`      | true, false              | Auto-open URL in browser              |
| `preview`        | true, false              | Inline image preview                  |
| `cache_ttl`      | Go duration (e.g. `12h`) | Template cache lifetime (default 24h) |

## Interactive mode

When stdout is a TTY, `memelink templates` launches a fuzzy-search picker. Select a template, enter
text for each line, and the meme URL is printed to stdout.

Inline image preview renders in terminals that support it (iTerm2, Kitty, Sixel). Disable with
`--no-preview` or `memelink config set preview false`.

## Global flags

| Flag         | Description                       |
| ------------ | --------------------------------- |
| `--json`     | JSON output                       |
| `--color`    | Color output: auto, always, never |
| `--verbose`  | Verbose logging                   |
| `--no-input` | Never prompt; fail instead        |
| `--force`    | Skip confirmations                |
| `--version`  | Print version and exit            |

## Agent Skill

This CLI is available as an [open agent skill](https://skills.sh/) for AI assistants including [Claude Code](https://claude.ai/code), [OpenClaw](https://openclaw.ai/), Cursor, and GitHub Copilot:

```bash
npx skills add dedene/memelink-cli
```

## Environment

| Variable          | Description                                              |
| ----------------- | -------------------------------------------------------- |
| `MEMEGEN_API_KEY` | API key for authenticated Memegen.link access (optional) |

## License

[MIT](LICENSE)
