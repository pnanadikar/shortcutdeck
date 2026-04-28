# Sprint 6 — Frontend: Deck Management

## Goal

Build the first real frontend for `shortcutdeck`: a keyboard-friendly deck and card management UI that matches the deck-management mockup, uses the existing HTTP API for all state, and is served from embedded static assets in the `web/` directory.

## Inputs

- [`docs/prd.md`](../prd.md) — use the desktop-first UX goals, tag namespace model, deck/card model, export/import expectations, and technical stack decisions.
- [`docs/dd.md`](../dd.md) — use the package inventory, data model, HTTP API in §6, and frontend/runtime invariants in §7 as the source of truth.
- [`docs/sprints/sprint-plan.md`](./sprint-plan.md) — use Sprint 6 as the source of truth for goal, branch name, deliverable, keyboard map, and acceptance criteria.
- [`docs/design/shortcutdeck_deck_cards.html`](../design/shortcutdeck_deck_cards.html) — authoritative UI reference for layout, component structure, styling, and interaction patterns for this sprint.
- [`docs/design/shortcutdeck_tree_dropdown.html`](../design/shortcutdeck_tree_dropdown.html) — supporting reference for namespace-driven filtering language and shared shell styling; use it only where it helps preserve design consistency without pulling Sprint 7 functionality forward.
- [`docs/design/shortcutdeck_session_summary.html`](../design/shortcutdeck_session_summary.html) — supporting reference for shared shell styling and visual consistency; do not implement summary flow in this sprint.
- [`internal/server/server.go`](../../internal/server/server.go) — currently serves a placeholder root response; update only as needed to serve embedded frontend assets from `web/`.
- [`web/index.html`](../../web/index.html) — replace the placeholder with the frontend entrypoint document.

## Outputs

- `web/index.html`
- `web/app.css`
- `web/app.js`
- `web/embed.go`
- `internal/server/server.go`
- `internal/server/server_test.go`
- `Makefile`
- `docs/sprints/s06-frontend-deck-mgmt.md`

The core sprint output is the frontend. `internal/server/server.go` may change only to embed and serve the `web/` directory cleanly so `/`, `/app.css`, and `/app.js` are available without adding complexity elsewhere. No other Go files should change unless a genuine bug blocks Sprint 6.

## Acceptance Criteria

- `web/index.html`, `web/app.css`, and `web/app.js` together render a working deck-management UI that visually follows [`docs/design/shortcutdeck_deck_cards.html`](../design/shortcutdeck_deck_cards.html).
- `GET /` serves the frontend entrypoint from embedded static assets rather than the Sprint 5 `"hello"` placeholder.
- Static asset requests such as `/app.css` and `/app.js` are served from the embedded `web/` directory.
- `internal/server/server_test.go` includes `httptest` coverage for the Sprint 6 static-serving behavior:
  - `GET /` returns `200 OK`
  - `GET /` serves `text/html`
  - `GET /app.css` returns `200 OK` and serves CSS with the correct content type
  - `GET /app.js` returns `200 OK` and serves JavaScript with the correct content type
  - these embedded asset responses are non-empty
- `Makefile` includes a Sprint 6 target:
  - `test-frontend`
  - command: `go test ./internal/server/... -v`
- The frontend uses the existing `/api/v1` API for all data. No mock data and no hardcoded application state.
- The deck list loads from `GET /api/v1/decks`.
- Deck CRUD is fully functional:
  - create via `POST /api/v1/decks`
  - read/list via `GET /api/v1/decks` and selected deck state
  - update via `PUT /api/v1/decks/:id`
  - delete via `DELETE /api/v1/decks/:id`
- Card CRUD is fully functional within the active deck:
  - list via `GET /api/v1/decks/:id/cards`
  - create via `POST /api/v1/decks/:id/cards`
  - update via `PUT /api/v1/cards/:id`
  - delete via `DELETE /api/v1/cards/:id`
- Deck export triggers a valid JSON download using `GET /api/v1/decks/:id/export`.
- Deck import uploads a file and imports it through `POST /api/v1/import?mode=replace`.
- Tag namespace editing is supported in the deck form and persisted through deck create/update.
- Card tag editing uses the chip-style input shown in the mockup and supports keyboard entry.
- The keyboard map from Sprint 6 is implemented:
  - `n` focuses the create form for the active view
  - `e` edits the focused item
  - `d` or `Delete` deletes the focused item
  - `Esc` cancels editing or navigates back from cards to decks
- Error states are visible in the status bar and are not silently swallowed.
- Dark mode support via `prefers-color-scheme: dark` is preserved.
- `make build` succeeds.
- `make lint` is clean.
- Manual verification succeeds:
  - create a Vim deck with 5 cards using only the keyboard
  - export the deck and confirm it is valid `ExportPayload` JSON
  - import the exported file into a fresh database and confirm it round-trips cleanly

## Manual Test Checklist

This checklist is the stopping condition for Sprint 6, equivalent to automated test completion in a code-only sprint. It must be worked through explicitly and checked off before merging `feat/s06-deck-mgmt`.
Failed or blocked checklist items should be recorded in [`docs/open-issues.md`](../open-issues.md), not inline in this sprint doc.

- [x] Run `make run`; verify the server starts, the browser opens, and the deck list renders empty.
- [x] Create a Vim deck with tag namespaces `mode` and `topic`.
- [x] Add 5 cards to the Vim deck with tags drawn from both namespaces.
- [x] Edit 1 card; verify the form pre-fills correctly and the saved change persists.
- [x] Delete 1 card; verify it disappears from the list.
- [x] Export the Vim deck; verify the downloaded JSON matches the `ExportPayload` shape.
- [x] Delete the Vim deck.
- [x] Import the exported JSON; verify the deck and all cards are restored.
- [ ] Repeat steps 2 through 8 using keyboard only, with no mouse.
- [ ] Switch the OS to dark mode; verify the color scheme switches correctly.

## Prompt

Implement Sprint 6 for the `shortcutdeck` repository.

Context and constraints to restate inline:

- This is the frontend deck-management sprint. Build the first real frontend for deck and card CRUD, tag namespace editing, export, and import. Do not implement study-session UI, browse mode, or scheduling flows in this sprint.
- Source of truth for UI is `docs/design/shortcutdeck_deck_cards.html`. Match its layout, component structure, visual style, spacing, and interaction patterns faithfully. Do not invent alternate UI patterns that are not present in the mockup.
- Preserve the established app shell language from the design references:
  - left sidebar
  - breadcrumb-style topbar
  - split form/list workspace
  - bottom status bar
  - compact desktop-first controls
- The intended package rule for this sprint is:
  - `web/` owns the HTML, CSS, and browser-side JavaScript
  - `internal/server` only serves embedded static assets and existing API routes
  - `internal/server` must not gain new business logic, SQL, or alternate API behavior for the sake of the frontend
  - `internal/store` and `internal/scheduler` are already complete enough for this sprint and must not be redesigned
- Sprint 6 deliverable from the sprint index: deck list, deck CRUD, card list, card CRUD, tag namespaces, export/import.

Create or modify exactly these paths as needed for this sprint:

- `web/index.html`
- `web/app.css`
- `web/app.js`
- `web/embed.go`
- `internal/server/server.go`
- `internal/server/server_test.go`
- `Makefile`
- `docs/sprints/s06-frontend-deck-mgmt.md`

Implementation requirements:

- Split the frontend into exactly three files:
  - `web/index.html` for document structure and mount points only
  - `web/app.css` for all frontend styles
  - `web/app.js` for application state, rendering, API calls, keyboard handling, and DOM updates
- Replace the placeholder frontend with a real HTML shell in `web/index.html`.
- Extract the styling from `docs/design/shortcutdeck_deck_cards.html` into `web/app.css`.
- Preserve the design tokens, color palette, spacing, compact density, and `prefers-color-scheme: dark` behavior from the mockup.
- Use plain JavaScript only. No React, Vue, Svelte, build tooling, bundlers, or generated assets.
- Do not use localStorage or sessionStorage. All durable state belongs to the API-backed backend.

- Permitted Go change:
  - `internal/server/server.go` currently returns a hardcoded `"hello"` string at `/`
  - replace that root/static serving behavior so the frontend is served from embedded assets in the `web/` directory
  - do not modify any other Go file unless an actual blocking bug surfaces
  - do not add endpoints, change handler contracts, or redesign API behavior
  - add or update focused `httptest` coverage in `internal/server/server_test.go` only for this static-serving behavior
  - add a matching `Makefile` target for Sprint 6 automated checks

- Keep the embed/static story simple by embedding the frontend assets in `web/embed.go` and serving them through `internal/server`. The intended pattern is:

```go
package web

import "embed"

//go:embed index.html app.css app.js
var FS embed.FS
```

- `internal/server` should import the `web` package and serve `web.FS`. The root route should return the frontend entrypoint, and requests such as `/app.css` and `/app.js` should also be served from the embedded assets.
- Add small, focused `httptest` coverage for this static-serving behavior:
  - `GET /` must return `200 OK` and `Content-Type` beginning with `text/html`
  - `GET /app.css` must return `200 OK` and a CSS content type
  - `GET /app.js` must return `200 OK` and a JavaScript content type
  - `/`, `/app.css`, and `/app.js` responses must not be empty
- Add a `Makefile` target:
  - `test-frontend`
  - command: `go test ./internal/server/... -v`

- Frontend behavior to implement:
  - deck list displays all decks from `GET /api/v1/decks`
  - create deck form supports:
    - name
    - description
    - default reverse mode
    - tag namespaces
  - edit deck pre-fills the same form and saves through `PUT /api/v1/decks/:id`
  - delete deck confirms before calling `DELETE /api/v1/decks/:id`
  - selecting a deck opens the card editor view for that deck
  - card list displays cards for the active deck from `GET /api/v1/decks/:id/cards`
  - create card form supports:
    - prompt
    - answer
    - notes
    - tags
  - edit card pre-fills the card form and saves through `PUT /api/v1/cards/:id`
  - delete card confirms before calling `DELETE /api/v1/cards/:id`
  - export deck calls `GET /api/v1/decks/:id/export` and triggers a file download
  - import deck uses file upload and `POST /api/v1/import?mode=replace`
  - tag namespace editor must match the `ns-editor` component style and interaction model from the mockup
  - card tag editing must match the `tag-input-wrap` chip-input style and keyboard entry model from the mockup

- Keyboard behavior to implement:
  - `n` focuses the create form for the current view:
    - deck form on the deck list screen
    - card form on the card editor screen
  - `e` edits the currently focused deck or card
  - `d` or `Delete` deletes the currently focused deck or card after confirmation
  - `Esc` cancels edit state or navigates back from card editor to deck list
  - keyboard shortcuts must not interfere while the user is typing into an input or textarea unless that shortcut is explicitly part of the input interaction

- API constraints:
  - use the existing local API at `http://localhost:7432/api/v1`
  - no mock data
  - no hardcoded deck or card state
  - on load, fetch `GET /api/v1/decks` and render the deck list
  - when a deck becomes active, fetch its cards from `GET /api/v1/decks/:id/cards`
  - surface API errors visibly in the status bar rather than swallowing them

- Design constraints:
  - follow `docs/design/shortcutdeck_deck_cards.html` closely for:
    - shell layout
    - form/list split
    - namespace editor
    - tag chips
    - breadcrumb structure
    - status bar feedback
  - use `docs/design/shortcutdeck_tree_dropdown.html` and `docs/design/shortcutdeck_session_summary.html` only as supporting visual references for consistency, not as a reason to implement Sprint 7 features early

Do not do any of the following in this sprint:

- do not implement study session UI
- do not implement browse mode UI
- do not add Start Review / Exit Review flows
- do not surface due-card counts or scheduling state in the deck-management UI
- do not add learning-path UI in this sprint
- do not add new backend endpoints
- do not change the `internal/store` interface
- do not change the scheduler API or SM-2 behavior
- do not add frontend build tooling
- do not introduce third-party frameworks
- do not make Go changes beyond the minimal embedded-static-serving fix unless a real bug blocks the sprint

Stop only when all of the following pass:

- `go test ./internal/server/... -v`
- `make test-frontend`
- `make build`
- `make lint`

And complete the full `Manual Test Checklist` in this sprint doc, checking off each item explicitly before merging the feature branch.

## Out of Scope

- Study session UI and summary flow
- Browse mode UI
- Review controls, due-card counts, and scheduling state display
- Learning-path display or management
- Browser auto-open behavior
- CLI flags such as `--port` and `--db`
- README or release packaging work
- Any store, scheduler, or API redesign

## Notes

- Sprint 6 intentionally relaxes the sprint-plan shorthand of "`web/index.html` only" to allow a minimal server-side static-serving fix. Without that change, the embedded frontend cannot replace the Sprint 5 placeholder at `/`.
- The frontend is split into `index.html`, `app.css`, and `app.js` for maintainability, but still keeps the no-build-step and embedded-static-assets approach described in the PRD.
- `web/embed.go` is part of the sprint output because Go's `//go:embed` directive cannot reference `../../web` or any other path outside the embedding package directory. The idiomatic solution is to place the embed directives in the `web` package and export `web.FS` for `internal/server` to consume.
- The design mockup in `docs/design/shortcutdeck_deck_cards.html` is the primary UI reference for this sprint and should be treated as stronger than generic frontend instincts.
- The `Manual Test Checklist` in this document is part of the sprint contract and should be treated as a merge gate, not as optional follow-up validation.
