# PRD: Keyboard Shortcut Flashcard Tool

> Status: Draft v0.7 — in progress  
> Owner: Personal use  
> Last updated: 2026-04-28

---

## 1. Problem Statement

Existing flashcard tools (Anki, SuperMemo, Quizlet, etc.) are:

- Too broad and generic — built for all learning domains, not keyboard shortcuts specifically
- Bloated with unnecessary features (goals, streaks, analytics dashboards, mobile sync)
- Heavy to install or run (Anki requires multiple packages and has a large footprint)
- Mobile-first or cross-platform in ways that add friction/bloat for a desktop power user

**Goal:** A lightweight, desktop-first, keyboard-shortcut-focused memorization tool that stays out of the way and does one thing well.

---

## 2. Target User

Personal use. Single user, desktop (Linux, macOS, Windows), power user comfortable with the terminal and config files.

---

## 3. Scope

### v1 / v2 Feature Map

#### Front-end

| Capability | v1 | v2 |
| --- | --- | --- |
| **Two binaries (Option C vs D):**  TUI ships as a separate `shortcutdeck-tui` binary or as `shortcutdeck --tui` (Both share 100% of `internal/`; difference is ~20 lines of entry-point code.) in v2 | | ✅ |

#### Manage Decks

| Capability | v1 | v2 |
| --- | --- | --- |
| Create / edit / delete decks | ✅ | |
| Create / edit / delete cards within a deck | ✅ | |
| Tag cards (free-form) | ✅ | |
| Define tag namespaces on a deck | ✅ | |
| Export a single deck as JSON | ✅ | |
| Import a deck from JSON | ✅ | |
| AI-assisted card generation (separate companion tool) | — | ✅ |
| Learning paths (ordered deck sequences, create/edit/manage) | — | ✅ |
| Import a full learning path (multiple decks + ordering) | — | ✅ |

#### Memorize Shortcuts

| Capability | v1 | v2 |
| --- | --- | --- |
| Study session — SM-2 scheduled cards | ✅ | |
| Browse mode — free review, no scheduling | ✅ | |
| Tag filtering within a session | ✅ | |
| Reverse mode — session-level option | ✅ | |
| Study summary (cards reviewed, ratings breakdown, next due date) | ✅ | |
| Study across multiple decks ("All due cards") | — | ✅ |
| Learning path — guided deck sequence in UI | — | ✅ |
| Progress views / stats | — | ✅ |

### Out of scope (v1 and v2)

- Mobile support
- Cloud sync
- User accounts / multi-user
- Goals, streaks, gamification
- Rich media cards (images, audio)
- Analytics dashboards or progress graphs

---

## 4. Card Format

Each card represents a single keyboard shortcut:

| Field | Description | Example |
| --- | --- | --- |
| `id` | UUID | `a1b2c3...` |
| `deck_id` | Parent deck | `vim` |
| `prompt` | What the shortcut does | `"Delete to end of line"` |
| `answer` | The key combination | `"d$"` |
| `notes` | Optional context | `"In normal mode only"` |
| `tags` | Free-form sub-categorisation (JSON array) | `["normal-mode", "delete", "advanced"]` |
| `created_at` | Timestamp | |
| `source` | How card was created | `manual` / `ai-generated` / `imported` |

### Tagging Model

Tags are **free-form per card** — no enforced taxonomy. This keeps the data model simple and makes the AI generation companion straightforward to implement (it emits tags naturally).

Each deck optionally defines **tag namespaces** (stored as JSON on the deck): named groupings that tell the UI which tags belong to which dimension, enabling smart filtering per deck.

Example for a Vim deck:

```json
{
  "tag_namespaces": {
    "mode":  ["normal", "insert", "visual", "command"],
    "level": ["basic", "advanced", "plugin-specific"],
    "topic": ["navigation", "editing", "buffers", "windows", "search"]
  }
}
```

A card can carry tags from any or all namespaces. The UI uses the namespace definition to offer structured filters; decks without namespaces fall back to flat tag search.

---

## 5. Deck Format

A deck maps to a single application or context:

| Field | Description |
| --- | --- |
| `id` | UUID |
| `name` | App name (e.g. `"Vim"`, `"Finder"`) |
| `description` | Optional |
| `tag_namespaces` | Optional JSON — named tag groupings for this deck (see §4) |
| `default_reverse_mode` | `prompt_first` (default) \| `answer_first` \| `both` — used as the pre-selected option when starting a session; overridable per session |
| `review_active` | Boolean — `true` when Ongoing Review has been started and not yet exited |
| `created_at` | Timestamp |

---

## 6. Spaced Repetition

- **Algorithm:** SM-2 (well-understood, proven, open specification)
- **Implementation:** Implement the ~40-line SM-2 spec directly in Go. No external library dependency. Fully unit-tested.
- **Rating scale:** Two UI controls (Got it / Missed it buttons + "Needs effort" checkbox) produce four SM-2 grades:

  | UI gesture | Needs effort | SM-2 grade | Meaning |
  | --- | --- | --- | --- |
  | Missed it | unchecked | 0 — Again | Complete blank — interval resets |
  | Missed it | checked | 1 — Partial | Saw answer, recognised it; still a miss, less severe |
  | Got it | checked | 3 — Hard | Recalled but struggled; interval grows slowly |
  | Got it | unchecked | 5 — Easy | Clean recall; interval grows aggressively |

- **Scheduling data per card:**
  - `interval` (days until next review)
  - `ease_factor`
  - `repetitions`
  - `due_date`
  - `last_reviewed_at`

---

## 7. Study Session Flow

### Ongoing Review lifecycle

Each deck independently tracks whether Ongoing Review is active via a `review_active` flag:

| State | Condition | Deck entry shows |
| --- | --- | --- |
| Never started | `review_active = false` (initial) | "Start Review" button; no due count |
| Active | `review_active = true` | "Ongoing Review →" button; due count visible |
| Exited | `review_active = false` (after exit) | "Start Review" button again; no due count |

**Start Review** sets `review_active = true`. It is a cold start: all card scheduling state is at defaults (interval=0, ease=2.5, repetitions=0, due=today), so every card is immediately due.

**Exit Review** sets `review_active = false` and resets all SM-2 scheduling fields on the deck's cards to defaults. The `review_log` is never deleted. The next Start Review is again a cold start.

### Session flow

1. User selects a deck with Ongoing Review active, optionally scoped by tag namespace
2. User sets session options:
   - **Reverse mode:** `prompt_first` / `answer_first` / `both` (pre-filled from deck's `default_reverse_mode`; overridable)
3. Tool presents cards due today (per SM-2 schedule)
4. For each card:
   - Show prompt side (determined by session reverse mode)
   - User attempts recall, then presses **Space/Enter** to reveal answer
   - After reveal: "Needs effort" checkbox appears alongside **Got it** / **Missed it** buttons
   - User rates recall; SM-2 grade is derived from button + checkbox combination (see §6)
   - SM-2 updates scheduling data; review logged
5. Session ends when no more due cards remain, or user presses **Esc**
6. Summary: cards reviewed, Got it / Missed it counts, next due dates

### Reverse Mode — `both`

When `both` is selected, each card is presented twice per session: once `prompt_first`, once `answer_first`. The session is split into two passes: first all cards `prompt_first`, then the same cards `answer_first`. This avoids immediate back-to-back reversal of the same card.

### Browse Mode (non-SM-2)

A separate mode for free review without scheduling:

- Select a deck + optional tag filter
- Page through cards (arrow keys / j/k)
- No rating, no scheduling updates
- Useful for reviewing a new plugin or cheatsheet before adding to spaced repetition

---

## 8. Learning Paths (data model — v1; UI management — v2)

A learning path is an ordered sequence of decks representing a suggested study progression (e.g. "Vim — Zero to Fluent": Vim Core → Vim Motions → Vim Buffers → Vim Macros).

**v1:** The `learning_paths` table is created and migrations run. Paths can be imported as part of a JSON payload (generated by the AI companion). The UI surfaces existing paths as a read-only suggested sequence on the deck list screen. No create/edit/delete in the UI.

**v2:** Full path management — create, reorder, attach/detach decks.

### Learning Path Format

| Field | Description |
| --- | --- |
| `id` | UUID |
| `name` | e.g. `"Vim — Zero to Fluent"` |
| `description` | Optional narrative |
| `deck_ids` | Ordered list of deck UUIDs |
| `created_at` | Timestamp |

A single deck may belong to multiple paths. Missing deck IDs (not yet imported) are silently filtered at display time; the path fills in automatically as decks are imported.

---

## 9. Storage

- **Format:** SQLite (`~/.shortcutdeck/cards.db` on Linux/macOS, `%APPDATA%\shortcutdeck\cards.db` on Windows)
- **Tables:** `decks`, `cards`, `review_log`, `learning_paths`
- **review_log:** Full history retained (not just current scheduling state)
  - Schema: `(id, card_id, reviewed_at, grade, interval_after, ease_after)`
  - Rationale: negligible storage cost; enables future stats; useful signal for AI generation companion
- **Export:** JSON (single deck), triggered via file download
- **Import:** JSON (merge or replace), via file upload in UI

---

## 10. UI / UX Principles

- Desktop-first, keyboard-navigable throughout
- Clean and minimal — no dashboards, no graphs, no noise
- Low friction: adding a card should take < 10 seconds
- Single binary, no installation beyond copying the executable
- Study session fully operable without mouse (see keyboard map in §7)

### Keyboard Interaction

Keyboard-first use is a core requirement. Deck and card management must be operable without a mouse; mouse support is additive, not required.

#### Sprint 6 — Deck and Card Management

##### Global

| Key | Action |
| --- | --- |
| `n` | Focus the create form for the current view |
| `e` | Edit the focused deck or card |
| `d` / `Delete` | Delete the focused deck or card, with confirmation |
| `Esc` | Cancel the active edit, or return from card editor to deck list |

`Esc` cancels immediately when the form is unchanged. If the form has unsaved changes, it must prompt before discarding them. Text-field comparison ignores trailing whitespace but not leading whitespace.

##### Deck list

| Key | Action |
| --- | --- |
| `↑` / `↓` | Move focus between decks |
| `Enter` | Open card editor for the focused deck |

##### Card editor

| Key | Action |
| --- | --- |
| `↑` / `↓` | Move focus between cards |
| `Enter` | Edit the focused card |

---

## 11. Technical Stack

| Concern | Decision | Notes |
| --- | --- | --- |
| Runtime | Go CLI + embedded local web server | `net/http`, serves frontend on `localhost:PORT`; browser auto-open is enabled by default but can be disabled with `--auto-open=false` |
| Frontend | Single HTML file, embedded in binary via `embed.FS` | Plain HTML/JS/CSS + Pico CSS; no build step; Alpine.js as escape hatch if needed |
| Storage | SQLite via `modernc/sqlite` | Pure Go, no CGo, cross-platform |
| SM-2 | Implemented directly in Go | ~40 lines, fully unit-tested; no external library |
| Export/Import | JSON via browser file download / `<input type="file">` | |
| Distribution | Single binary per platform | `go build` — no Electron, no WebView dependency, no Node runtime |

### Architecture

```plaintext
cmd/shortcutdeck/main.go     — entry point: starts server, optionally opens browser
internal/
  server/                    — HTTP handlers (decks, cards, review, import/export)
  store/                     — SQLite layer (decks, cards, review_log, learning_paths)
  sm2/                       — SM-2 algorithm + unit tests
  model/                     — shared types (Card, Deck, ReviewEntry, LearningPath)
web/                         — embedded frontend (HTML/JS/CSS)
```

---

## 12. Resolved Decisions

All open questions are now closed.

| Question | Decision |
| --- | --- |
| Runtime port | Fixed default `7432`; overridable via `--port` flag |
| Browser auto-open | Enabled by default; overridable via `--auto-open=false` for service/non-interactive starts |
| SM-2 vs FSRS-5 | SM-2; implement directly in Go (~40 lines, fully unit-tested) |
| Reverse mode scope | Session-level option; deck stores a `default_reverse_mode` as pre-fill |
| `reverse_mode: both` behaviour | Two-pass session: all cards `prompt_first`, then all cards `answer_first` |

---

## 13. AI Card Generation (Separate Tool — Future)

To be designed separately. This tool targets a **local AI backend** (e.g. Ollama) for privacy and offline use.

Intended workflow:

1. Provide high-level inputs:
   - App name
   - Source material: documentation / help files (e.g. `:help` in Vim), plugin list, raw text / config files (e.g. `.vimrc`), or a URL to crawl (best-effort)
   - Deck structure: tag namespaces (from `data_model.md`)
2. Local AI generates a candidate deck (and optionally a learning path)
3. Output: JSON matching the card format in §4 and the export schema in §9, importable into the main tool
4. User reviews, edits, and imports

The v1 data model (card format, tag namespaces, learning path, JSON export schema) is designed to be the stable contract this tool targets. No changes to the main tool are required to support it.
