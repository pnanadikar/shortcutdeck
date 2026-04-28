# Design Document: shortcutdeck

> Status: Draft v0.6  
> Last updated: 2026-04-28  
> Companion to: prd.md v0.7

---

## 1. Package Inventory

### `internal/model`

Shared data types used across all packages. No logic, no I/O, no dependencies on any other internal package. Every other package imports this; it imports nothing internal.

### `internal/scheduler`

Implements the SM-2 spaced repetition algorithm. Exposes a `Scheduler` interface so the algorithm can be swapped (e.g. to FSRS-5 via `go-fsrs`) without touching any other package. No database access, no HTTP, no file I/O. Pure functions operating on `SchedulingState`.

### `internal/store`

SQLite persistence layer. All database access is isolated here. Accepts and returns `internal/model` types. No HTTP concepts (no `http.Request`, no `http.ResponseWriter`), no terminal concepts, no business logic beyond query construction. Owns schema migrations.

### `internal/server`

HTTP handlers and routing. Translates between HTTP requests/responses and calls to `internal/store` and `internal/scheduler`. No SQL, no direct file I/O. Embeds and serves the web frontend via `embed.FS`.

### `cmd/shortcutdeck`

Entry point. Parses CLI flags, resolves the database path, initialises the store, wires dependencies, starts the HTTP server, optionally opens the browser, and owns runtime logging for startup, shutdown, fatal runtime errors, browser-open warnings, and server-level operational events. Minimal logic — delegates product behavior to `internal/`.

### `web/`

Single embedded HTML/CSS/JS frontend. Served by `internal/server` via `embed.FS`. No build step. Plain HTML, CSS, and JavaScript.

The frontend should be theme-ready without making theming a v1 feature. In practice, this means visual styling should flow through a small semantic CSS-variable layer rather than scattering raw color literals throughout component rules. User-editable theme configuration is deferred to v2, but v1 should leave behind a stable token surface that can support it later.

Recommended minimum semantic token set for the frontend:

- surfaces:
  - `--bg-app`
  - `--bg-sidebar`
  - `--bg-panel`
  - `--bg-card`
- text:
  - `--text-primary`
  - `--text-secondary`
  - `--text-muted`
- chrome:
  - `--border-subtle`
  - `--hover-bg`
- interaction:
  - `--accent`
  - `--focus-ring`
  - `--selected-bg`
- semantic feedback:
  - `--danger-fg`
  - `--danger-bg`
  - `--success-fg`
  - `--success-bg`

Status-bar behavior should also stay semantically disciplined:

- the status bar is the persistent mode/context channel
- transient action prompts may temporarily replace the current status message
- once that transient action completes or is cancelled, the status bar should return to the steady-state message for the current view/pane
- short-lived result feedback such as success/failure of import/export should prefer the toast channel, while the status bar returns to describing the current mode

Startup/landing behavior should be explicit rather than accidental:

- when one or more decks exist, the app should open on the study deck list
- the first deck should be preselected and should receive live focus on startup so keyboard navigation works immediately
- the study search field should not steal initial focus; `/` is the explicit jump-to-search action
- when no decks exist, the app should open directly to deck creation rather than showing an empty study shell first

Dirty-state behavior should follow the same principle of only interrupting when user input is genuinely at risk:

- prompt when an action would discard or overwrite unsaved deck/card edits
- this includes switching selected decks/cards if that selection change repopulates the edit form from another record
- this includes leaving the current editing context entirely while the form is dirty
- do not prompt for harmless focus movement, search input, scope changes, or normal study-session navigation that does not destroy form state

Import should follow the same editing-context discipline:

- if `Import...` is triggered from inside deck/card editing context, run the dirty-state check first when appropriate
- do not tear down the current view merely because the file picker was opened
- if the file picker is cancelled, restore the steady-state status message and leave the current view/context unchanged
- if import is launched from a nested deck/card editing workspace and a real import attempt begins, leave the nested editing workspace before continuing
- after an actual import attempt, preserve the current top-level context when it is already stable (for example top-level `Study` or top-level `Create`)
- only normalize to a top-level non-editing view when the import was launched from a nested editing context; in that case, prefer the study deck list

---

## 2. Data Types (`internal/model`)

```go
// Card represents a single keyboard shortcut flashcard.
type Card struct {
    ID             string    // UUID
    DeckID         string    // parent deck UUID
    Prompt         string    // what the shortcut does
    Answer         string    // the key combination
    Notes          string    // optional context
    Tags           []string  // free-form tags (stored as JSON array)
    Source         string    // "manual" | "ai-generated" | "imported"
    CreatedAt      time.Time

    // Scheduling state (SM-2)
    Interval       int       // days until next review
    EaseFactor     float64   // SM-2 ease factor (min 1.3)
    Repetitions    int       // consecutive successful reviews
    DueDate        time.Time
    LastReviewedAt time.Time
}

// Deck represents a set of cards for one application or context.
type Deck struct {
    ID                 string          // UUID
    Name               string          // e.g. "Vim", "Finder"
    Description        string          // optional
    TagNamespaces      TagNamespaces   // optional structured tag groupings
    DefaultReverseMode ReverseMode     // pre-selected option when starting a session; overridable per session
    ReviewActive       bool            // true when Ongoing Review has been started and not yet exited
    CreatedAt          time.Time
}

// ReviewActive lifecycle:
//   false (never started) → "Start Review" button; no due counts shown
//   true  (active)        → "Ongoing Review →" button; due counts visible
//   false (after exit)    → "Start Review" button again; scheduling state reset; review log retained

// TagNamespaces maps namespace names to their known tag values.
// Stored as JSON on the Deck. Example:
//   {"mode": ["normal","insert"], "level": ["basic","advanced"]}
type TagNamespaces map[string][]string

// ReverseMode controls which side of a card is shown first during a session.
// Stored as the deck's default; the user may override it at session start.
type ReverseMode string
const (
    PromptFirst  ReverseMode = "prompt_first"  // default
    AnswerFirst  ReverseMode = "answer_first"
    Both         ReverseMode = "both"           // two-pass session: all prompt_first, then all answer_first
)

// ReviewEntry is a single review event, stored in review_log.
type ReviewEntry struct {
    ID             string    // UUID
    CardID         string
    ReviewedAt     time.Time
    Grade          int       // 0=Again, 1=Partial, 3=Hard, 5=Easy
    IntervalAfter  int       // days, after this review
    EaseAfter      float64   // ease factor after this review
}

// LearningPath is an ordered sequence of decks representing a suggested
// study progression (e.g. "Vim — Zero to Fluent": Core → Motions → Buffers).
// v1: table exists, paths are importable and displayed read-only in the UI.
// v2: full create/edit/delete UI.
type LearningPath struct {
    ID          string    // UUID
    Name        string    // e.g. "Vim — Zero to Fluent"
    Description string    // optional narrative
    DeckIDs     []string  // ordered list of deck UUIDs (stored as JSON array)
    CreatedAt   time.Time
}
```

---

## 3. Scheduler Interface (`internal/scheduler`)

```go
// SchedulingState is the algorithm-agnostic representation of a card's
// current scheduling position. SM-2 and FSRS both map onto this struct.
type SchedulingState struct {
    Interval       int
    EaseFactor     float64
    Repetitions    int
    DueDate        time.Time
    LastReviewedAt time.Time
}

// Grade represents the user's self-assessment of recall quality.
// The four UI gestures (Missed it / Got it × Needs effort checkbox) map
// onto SM-2 grades as follows:
//
//   UI gesture          | Needs effort | Grade | SM-2 meaning
//   --------------------|--------------|-------|-------------------------------
//   Missed it           | unchecked    |   0   | Complete blank — interval resets
//   Missed it           | checked      |   1   | Saw answer, recognised it — still a miss, less severe
//   Got it              | checked      |   3   | Recalled but struggled — interval grows slowly
//   Got it              | unchecked    |   5   | Clean recall — interval grows aggressively
type Grade int
const (
    GradeAgain   Grade = 0  // complete blackout
    GradePartial Grade = 1  // saw answer, recognised it; still counts as a miss
    GradeHard    Grade = 3  // recalled with serious difficulty
    GradeEasy    Grade = 5  // perfect recall
)

// Scheduler is the interface all algorithm implementations must satisfy.
type Scheduler interface {
    // Schedule computes the new scheduling state given the current state,
    // the user's grade, and the current time.
    Schedule(current SchedulingState, grade Grade, now time.Time) SchedulingState
}

// SM2 is the v1 implementation of Scheduler using the SM-2 algorithm.
type SM2 struct{}

// Schedule implements Scheduler for SM2.
// Pure function: no side effects, no I/O.
func (SM2) Schedule(current SchedulingState, grade Grade, now time.Time) SchedulingState

// NewState returns a SchedulingState with SM-2 defaults for a brand-new card.
func NewState() SchedulingState
```

---

## 4. SM-2 Algorithm Testing

`internal/scheduler/scheduler_test.go` provides unit-test coverage for the current SM-2 implementation. The tests are intentionally focused on deterministic algorithm behavior rather than integration concerns.

Current test coverage includes:

- `NewState()` defaults: verifies `interval=0`, `ease_factor=2.5`, `repetitions=0`, `due_date=today`, and zero `last_reviewed_at`
- Miss behavior: verifies `GradeAgain` resets interval and repetitions, and `GradePartial` is still a miss but less severe than `GradeAgain`
- Early review progression: verifies the first successful review produces interval `1` and the second successful review produces interval `6`
- Relative grade behavior: verifies `GradeEasy` grows intervals faster than `GradeHard`, and cross-grade ordering is consistent for a fixed starting state
- Long-run behavior: verifies successive successful reviews compound correctly, repeated misses do not push ease below the floor, and relapse after a long good run resets progress correctly
- Ease-factor floor: verifies the scheduler never drops below `1.3`, including repeated failure paths and starting exactly at the floor
- Due-date and timestamp handling: verifies `DueDate` is derived from the provided `now`, is timezone-neutral after normalization to UTC, and updates `LastReviewedAt` correctly
- Determinism: verifies fixed input state plus fixed review time produces fixed output state

These tests are meant to pin down the algorithm contract tightly enough that later store, server, and UI layers can depend on scheduler behavior without reinterpreting SM-2 rules.

---

## 5. Store Interface (`internal/store`)

```go
// Store is the persistence interface for decks, cards, review history,
// and learning paths.
// The SQLite implementation satisfies this interface.
// Exposing an interface (rather than a concrete struct) makes the store
// testable via a mock without a real database.
type Store interface {
    // --- Decks ---

    // CreateDeck inserts a new deck. Sets ID and CreatedAt if not provided.
    CreateDeck(d model.Deck) (model.Deck, error)

    // GetDeck returns a single deck by ID.
    GetDeck(id string) (model.Deck, error)

    // ListDecks returns all decks ordered by name.
    ListDecks() ([]model.Deck, error)

    // UpdateDeck updates name, description, tag_namespaces, default_reverse_mode.
    UpdateDeck(d model.Deck) (model.Deck, error)

    // DeleteDeck deletes a deck and all its cards. Cascades to review_log.
    // Does not delete any LearningPath that references this deck; missing
    // deck IDs are filtered at display time.
    DeleteDeck(id string) error

    // --- Cards ---

    // CreateCard inserts a new card. Sets ID, CreatedAt, and default
    // scheduling state (interval=0, ease=2.5, repetitions=0, due=today).
    CreateCard(c model.Card) (model.Card, error)

    // GetCard returns a single card by ID.
    GetCard(id string) (model.Card, error)

    // ListCards returns all cards in a deck, optionally filtered by tags.
    // tags is ANDed: a card must carry all specified tags to be returned.
    ListCards(deckID string, tags []string) ([]model.Card, error)

    // GetDueCards returns cards in a deck whose DueDate <= today,
    // ordered by DueDate ascending.
    GetDueCards(deckID string, today time.Time) ([]model.Card, error)

    // UpdateCard updates prompt, answer, notes, tags, and scheduling state.
    UpdateCard(c model.Card) (model.Card, error)

    // DeleteCard deletes a card and its review_log entries.
    DeleteCard(id string) error

    // --- Review Log ---

    // LogReview appends a review event to review_log.
    LogReview(entry model.ReviewEntry) error

    // GetReviewLog returns all review entries for a card, ordered by
    // reviewed_at ascending.
    GetReviewLog(cardID string) ([]model.ReviewEntry, error)

    // --- Learning Paths ---

    // CreateLearningPath inserts a new learning path. Sets ID and CreatedAt
    // if not provided.
    CreateLearningPath(p model.LearningPath) (model.LearningPath, error)

    // GetLearningPath returns a single learning path by ID.
    GetLearningPath(id string) (model.LearningPath, error)

    // ListLearningPaths returns all learning paths ordered by name.
    ListLearningPaths() ([]model.LearningPath, error)

    // --- Import / Export ---

    // ExportDeck returns a deck and all its cards as an ExportPayload.
    ExportDeck(deckID string) (ExportPayload, error)

    // ImportDeck inserts or merges a deck from an ExportPayload.
    // mode: "merge" upserts cards by ID; "replace" deletes existing cards first.
    // If the payload includes a LearningPath, it is upserted by ID.
    ImportDeck(payload ExportPayload, mode string) error
}

// ExportPayload is the JSON structure used for deck export/import.
// LearningPath is optional: present when the AI companion generates a path
// alongside a deck, absent for standalone deck exports.
type ExportPayload struct {
    Deck         model.Deck          `json:"deck"`
    Cards        []model.Card        `json:"cards"`
    LearningPath *model.LearningPath `json:"learning_path,omitempty"`
}

// SQLiteStore is the SQLite-backed implementation of Store.
type SQLiteStore struct { /* unexported fields */ }

// New opens (or creates) the SQLite database at dbPath,
// runs any pending migrations, and returns a ready Store.
func New(dbPath string) (Store, error)
```

---

## 5. SQL Schema

```sql
CREATE TABLE IF NOT EXISTS decks (
    id                   TEXT PRIMARY KEY,
    name                 TEXT NOT NULL,
    description          TEXT NOT NULL DEFAULT '',
    tag_namespaces       TEXT NOT NULL DEFAULT '{}',      -- JSON
    default_reverse_mode TEXT NOT NULL DEFAULT 'prompt_first',
    review_active        INTEGER NOT NULL DEFAULT 0,      -- boolean: 1 when Ongoing Review is active
    created_at           DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS cards (
    id               TEXT PRIMARY KEY,
    deck_id          TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    prompt           TEXT NOT NULL,
    answer           TEXT NOT NULL,
    notes            TEXT NOT NULL DEFAULT '',
    tags             TEXT NOT NULL DEFAULT '[]',  -- JSON array
    source           TEXT NOT NULL DEFAULT 'manual',
    created_at       DATETIME NOT NULL,

    -- SM-2 scheduling state
    interval         INTEGER NOT NULL DEFAULT 0,
    ease_factor      REAL    NOT NULL DEFAULT 2.5,
    repetitions      INTEGER NOT NULL DEFAULT 0,
    due_date         DATETIME NOT NULL,
    last_reviewed_at DATETIME
);

CREATE TABLE IF NOT EXISTS review_log (
    id              TEXT PRIMARY KEY,
    card_id         TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    reviewed_at     DATETIME NOT NULL,
    grade           INTEGER NOT NULL,   -- 0, 1, 3, or 5
    interval_after  INTEGER NOT NULL,
    ease_after      REAL    NOT NULL
);

-- Learning paths: ordered deck sequences. v1: importable, read-only in UI.
-- deck_ids is a JSON array of deck UUIDs in study order.
-- Missing deck IDs (not yet imported) are filtered at display time.
CREATE TABLE IF NOT EXISTS learning_paths (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    deck_ids    TEXT NOT NULL DEFAULT '[]',  -- JSON array of deck UUIDs
    created_at  DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cards_deck_id  ON cards(deck_id);
CREATE INDEX IF NOT EXISTS idx_cards_due_date ON cards(due_date);
CREATE INDEX IF NOT EXISTS idx_review_log_card_id ON review_log(card_id);
```

---

## 6. HTTP API

All endpoints are JSON. Base path: `/api/v1`.

### Decks

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/v1/decks` | List all decks |
| `POST` | `/api/v1/decks` | Create a deck |
| `GET` | `/api/v1/decks/:id` | Get a deck |
| `PUT` | `/api/v1/decks/:id` | Update a deck |
| `DELETE` | `/api/v1/decks/:id` | Delete a deck and all its cards |

### Cards

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/v1/decks/:id/cards` | List cards (optional `?tags=a,b`) |
| `POST` | `/api/v1/decks/:id/cards` | Create a card |
| `GET` | `/api/v1/cards/:id` | Get a card |
| `PUT` | `/api/v1/cards/:id` | Update a card |
| `DELETE` | `/api/v1/cards/:id` | Delete a card |

### Study Session

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/v1/decks/:id/due` | Get due cards for today (optional `?tags=a,b`); only meaningful when `review_active` is true |
| `POST` | `/api/v1/cards/:id/review` | Submit a review grade; updates scheduling state and appends to review_log |
| `POST` | `/api/v1/decks/:id/start-review` | Set `review_active = true`; no body required |
| `POST` | `/api/v1/decks/:id/exit-review` | Set `review_active = false`; resets scheduling state on all cards in deck; retains review_log |

Request body for `POST /review`:

```json
{ "grade": 3 }
```

Valid grades: `0` (Again / complete blank), `1` (Partial / missed but recognised), `3` (Hard / recalled with effort), `5` (Easy / clean recall).

### Learning Paths

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/v1/paths` | List all learning paths (with missing deck IDs filtered out) |
| `GET` | `/api/v1/paths/:id` | Get a single learning path |

_Note: No POST/PUT/DELETE for learning paths in v1. Paths enter the system via `POST /api/v1/import` only._

### Import / Export

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/v1/decks/:id/export` | Export deck as JSON (triggers file download) |
| `POST` | `/api/v1/import` | Import a deck JSON (`?mode=merge` or `?mode=replace`); upserts any embedded learning path |

_Note: The API supports both `merge` and `replace` modes. The v1 frontend uses `replace` only. `merge` is reserved for future AI companion tooling and non-frontend workflows._

### Frontend

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/` | Serves the embedded `web/index.html` |

---

## 7. Key Invariants

- **`internal/model`, `internal/scheduler`, `internal/store` must never import `net/http` or any terminal/TUI package.** All UI concepts are confined to `internal/server` and `cmd/`.

- **`internal/scheduler` is a pure function.** `Schedule()` has no side effects and does not write to the database. The caller (`internal/server`) is responsible for persisting the updated scheduling state via `store.UpdateCard()` and logging the review via `store.LogReview()`.

- **Scheduling state lives on the `Card` struct.** There is no separate `scheduling` table. The fields `interval`, `ease_factor`, `repetitions`, `due_date`, and `last_reviewed_at` are columns on the `cards` table.

- **`review_log` is append-only.** No update or delete operations are exposed on review log entries. Cascade delete on card deletion is the only removal path.

- **`DeleteDeck` cascades.** Deleting a deck deletes all its cards and, via cascade, all their review log entries. This is enforced at the SQLite level (`ON DELETE CASCADE`) and must not be worked around in application code.

- **Learning paths are not cascade-deleted with decks.** A path's `deck_ids` array may contain IDs of decks that have been deleted. Missing IDs are filtered silently at display time; the path itself is not removed.

- **Learning paths are read-only in v1.** The only write path for learning paths is `ImportDeck`. No API endpoint creates, updates, or deletes a path directly in v1.

- **Algorithm swap is a one-line change.** The `Scheduler` interface is the only coupling point between the algorithm and the rest of the system. Swapping SM-2 for FSRS-5 means replacing `scheduler.SM2{}` with `scheduler.FSRS{}` at the injection site in `cmd/shortcutdeck/main.go`.

- **`due_date` for a new card is set to today.** A newly created card is immediately due, so it appears in the first study session.

- **`review_active` governs due-card visibility and button state.** `GetDueCards` must only be called (and due counts only shown) when `review_active` is `true` for that deck. The home screen button label is "Start Review" when `review_active` is `false`, and "Ongoing Review →" when `true`.

- **Exit Review clears scheduling state but retains review log.** When the user exits Ongoing Review, all SM-2 fields on the deck's cards (`interval`, `ease_factor`, `repetitions`, `due_date`, `last_reviewed_at`) are reset to their defaults. `review_active` is set to `false`. The `review_log` rows are never deleted. The next "Start Review" is a cold start.

- **`review_active` is scoped to the deck, not the session.** A session is a single sitting within an active review. Starting and completing a session does not change `review_active`; only explicit Start Review / Exit Review actions do.

- **`ExportPayload` is the stable contract for the AI generation companion.** The JSON structure of deck export must not change without a versioning strategy, as the future AI tool targets this format.

- **The frontend should use semantic theme tokens, not ad hoc color sprawl.** v1 ships a fixed polished light/dark presentation, not user-configurable theming, but the CSS should be organized around a stable semantic token layer so future config-file theming can map onto it without a full rewrite.

- **The status bar should describe the current mode, not retain stale transient prompts.** Temporary guidance such as "select a file to import" may appear briefly, but once the action completes or is cancelled the status bar should return to the baseline message for the active view. Short-lived action outcomes should prefer the toast channel.

- **Dirty-state guards should trigger on destructive/replacing transitions, not on ordinary movement.** If a transition would overwrite a dirty deck/card form with another record's data or abandon the editing context entirely, the user should be prompted. Pure focus movement and non-destructive study/navigation interactions should remain prompt-free.

- **Import should preserve stable top-level context and only normalize nested editing flows.** Canceling the file picker should not navigate. Top-level `Study` or top-level `Create` may remain in place after import, but imports launched from nested deck/card editing should resolve dirty-state first and then continue from a top-level non-editing view, preferably the study deck list.

- **Startup focus should support the primary keyboard path.** With decks present, startup focus should land on the first study deck so `↑` / `↓` work immediately. With no decks present, startup should land in the create-deck flow instead.
