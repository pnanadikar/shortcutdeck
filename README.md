# shortcutdeck

`shortcutdeck` is a desktop-first, keyboard-first flashcard tool for learning keyboard shortcuts. (Although the entire tool code and documentatation talks about shortcuts, it is a generic flashcards tool that was created for remembering shortcuts :-)

It runs as a single local Go binary with an embedded web UI. Your data stays on your machine in a SQLite database. There is no account system, cloud sync, or external service dependency.

## What It Does

v1 currently supports:

- creating, editing, and deleting decks
- creating, editing, and deleting cards inside a deck
- optional deck-level tag namespaces for structured filtering
- Ongoing Review with SM-2 scheduling
- Quick Refresher for non-scheduled browsing/review
- reverse card presentation modes: `prompt_first`, `answer_first`, `both`
- JSON deck export and import
- a lightweight About dialog with app version and short product description

Out of scope for v1:

- cloud sync
- multi-user support
- mobile apps
- dashboards, analytics, streaks, or gamification
- learning-path CRUD
- multi-deck study sessions

## Supported Platforms

The app is intended to run on:

- macOS
- Linux
- Windows

The runtime is pure Go plus SQLite via `modernc.org/sqlite`, so no CGo setup is required.

## Build And Run

Build the local binary:

```bash
make build
```

Run from source:

```bash
make run
```

Run the built binary:

```bash
./shortcuts
```

Show runtime help:

```bash
./shortcuts --help
```

## Runtime Flags

`shortcutdeck` currently supports:

- `--port`
  Controls the local HTTP port. Default: `7432`.
- `--db`
  Controls the SQLite database file path. By default the app uses your OS config directory plus `shortcutdeck/cards.db`.
- `--auto-open`
  Controls whether the app opens your default browser on startup. Default: `true`.

Examples:

```bash
./shortcuts --port 8123
./shortcuts --db /tmp/shortcutdeck.db
./shortcuts --auto-open=false
./shortcuts --auto-open false
./shortcuts --db "$HOME/Library/Application Support/shortcutdeck/cards.db"
```

Notes:

- quote `--db` values that contain spaces
- `--db` must point to a file path, not just a directory
- if the database file does not exist yet, the app creates it

## Default Data Location

Default database location by platform:

- macOS: `~/Library/Application Support/shortcutdeck/cards.db`
- Linux: `~/.config/shortcutdeck/cards.db`
- Windows: `%APPDATA%\shortcutdeck\cards.db`

You can override this with `--db`.

## Browser Behavior

On startup, the app serves locally on `http://127.0.0.1:<port>` and, by default, attempts to open your default browser automatically.

- browser-open failure is non-fatal
- the server keeps running even if auto-open fails
- pass `--auto-open=false` when running as a service or when you do not want a browser launched
- startup logging prints the URL and active DB path

## Keyboard Workflow

The UI is designed for keyboard-first use.

Deck and card management:

- `n` focuses the create form in the current view
- `e` edits the focused deck or card
- `d` or `Delete` deletes the focused deck or card, with confirmation
- `Esc` cancels edit/back when appropriate
- `Up` / `Down` move through deck rows and card rows
- `Enter` opens or edits the focused row in list contexts

Study and review:

- `Space` or `Enter` reveals the answer
- after reveal, use the review controls to rate recall in Ongoing Review
- `Esc` leaves the current session flow
- Quick Refresher supports sequential navigation after reveal with `j` / `k` or arrow keys

Study modes:

- `Quick Refresher` does not update scheduling
- `Ongoing Review` uses SM-2 and due dates

## Decks, Tags, And Review

A deck represents one app or context, such as `Vim`, `tmux`, or `Finder`.

Each card stores:

- a prompt
- an answer
- optional notes
- free-form tags

Decks may also define tag namespaces to guide filtering, for example:

```json
{
  "mode": ["normal", "insert", "visual"],
  "topic": ["navigation", "editing", "search"]
}
```

Ongoing Review is deck-specific:

- starting review makes the deck active
- due cards appear based on SM-2 scheduling
- exiting review resets scheduling state for that deck's cards
- review history is retained

## Export And Import

Export:

- exports one deck at a time as JSON
- export is available from the UI

Import:

- imports deck JSON through the UI
- imported decks and cards are stored locally in SQLite
- the app preserves the current top-level context where practical

The JSON shape is intended to stay stable because it is also the contract for future companion tooling.

## About Dialog

The `About...` menu item shows:

- the current app version
- a short product description

Those values are sourced from Go, not duplicated in frontend-only strings. The current source of truth is [`internal/appinfo/appinfo.go`](./internal/appinfo/appinfo.go).

## Development Checks

Common local checks:

```bash
make build
make test
make lint
```

Focused package checks also exist:

- `make test-store`
- `make test-scheduler`
- `make test-server`

## Release Builds

Build the cross-platform release set locally:

```bash
make release
```

Artifacts are written to `dist/` with these names:

- `dist/shortcutdeck-linux-amd64`

## Acknowledgements

This project was developed with substantial assistance from Claude.ai for design and OpenAI Codex for implementation, refactoring, testing support, and documentation drafting. Final product and release decisions were reviewed and curated by the repository owner.

- `dist/shortcutdeck-darwin-amd64`
- `dist/shortcutdeck-windows-amd64.exe`
