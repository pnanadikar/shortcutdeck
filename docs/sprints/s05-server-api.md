# Sprint 5 — HTTP Server + API

## Goal

Complete `internal/server` and wire the application entrypoint so `shortcutdeck` exposes the full JSON HTTP API from `docs/dd.md` §6, with a placeholder frontend response at `/`, while reusing the existing `internal/store` and `internal/scheduler` packages without redesigning them.

## Inputs

- [`docs/prd.md`](../prd.md) — use the runtime model, study-session semantics, review lifecycle, import/export behavior, and learning-path constraints from Technical Stack, Spaced Repetition, Study Session Flow, Learning Paths, and Storage.
- [`docs/dd.md`](../dd.md) — use `internal/server` in Package Inventory, the store and scheduler contracts in §§3 and 5, the full HTTP API in §6, and the invariants in §7 as the source of truth.
- [`docs/sprints/sprint-plan.md`](./sprint-plan.md) — use Sprint 5 as the source of truth for goal, outputs, and acceptance criteria.
- [`internal/model/model.go`](../../internal/model/model.go) — use the existing shared types exactly as the JSON request/response surface where applicable.
- [`internal/store/store.go`](../../internal/store/store.go) — use the existing `Store` interface and `ExportPayload` exactly; do not redesign the persistence contract in this sprint.
- [`internal/scheduler/scheduler.go`](../../internal/scheduler/scheduler.go) — use the existing `Scheduler`, `SchedulingState`, `Grade`, `SM2`, and `NewState()` behavior exactly; do not change the algorithm in this sprint.
- [`internal/server/server.go`](../../internal/server/server.go) — replace the Sprint 1 stub with the real server wiring for this sprint.
- [`cmd/shortcutdeck/main.go`](../../cmd/shortcutdeck/main.go) — replace the Sprint 1 stub entrypoint with dependency wiring and server startup for this sprint.

## Outputs

- `internal/server/server.go`
- `internal/server/handlers_decks.go`
- `internal/server/handlers_cards.go`
- `internal/server/handlers_review.go`
- `internal/server/handlers_paths.go`
- `internal/server/handlers_import_export.go`
- `internal/server/server_test.go`
- `cmd/shortcutdeck/main.go`
- `web/index.html`
- `Makefile`
- `docs/sprints/s05-server-api.md`

The core sprint output is the server package plus entrypoint wiring. `web/index.html` may be updated only as needed to support the placeholder frontend response for `/`. Existing Sprint 1–4 files should otherwise remain unchanged unless a compile-only adjustment is strictly required.

## Acceptance Criteria

- `internal/server/server.go` defines the HTTP server wiring for all routes in `docs/dd.md` §6.
- The API base path is `/api/v1`.
- Every endpoint in `docs/dd.md` §6 is implemented and returns JSON.
- `GET /` returns the frontend placeholder response for this sprint (`"hello"` served from the embedded frontend path or equivalent server-owned placeholder behavior).
- `cmd/shortcutdeck/main.go` initialises the store, wires a scheduler implementation, constructs the server, and starts listening on port `7432`.
- Happy-path and basic error-path behavior is covered for each endpoint:
  - `400` for malformed JSON, invalid grade values, or invalid import mode
  - `404` for missing deck, card, or learning-path IDs
  - appropriate `200`/`201`/`204` responses for successful requests
- `POST /api/v1/cards/:id/review` updates card scheduling state through `internal/scheduler` and appends a review entry through `internal/store`.
- `POST /api/v1/decks/:id/start-review` sets `review_active = true`.
- `POST /api/v1/decks/:id/exit-review` sets `review_active = false` and resets scheduling state on all cards in the deck while retaining `review_log`.
- `GET /api/v1/decks/:id/due` only returns due cards when the deck has `review_active = true`; otherwise it returns an empty list.
- `GET /api/v1/paths` and `GET /api/v1/paths/:id` filter missing deck IDs from returned learning paths.
- `GET /api/v1/decks/:id/export` returns the store `ExportPayload` JSON for download.
- `go test ./internal/server/... -v` passes with `httptest`-based tests covering every endpoint.
- Handler tests cover the happy path for each endpoint, including correct status code and response body shape.
- Handler tests cover these basic error cases where applicable:
  - `404` on unknown ID
  - `400` on malformed request body
  - `400` on invalid grade value
- Sprint doc includes a curl smoke-test script covering the main API flow.
- `Makefile` includes `test-server` as:
  - `go test ./internal/server/... -v`
- `go test ./...` exits with status 0.
- `go vet ./...` is clean.
- `make lint` is clean.

## Prompt

Implement Sprint 5 for the `shortcutdeck` repository.

Context and constraints to restate inline:

- This is the HTTP server and API sprint. Implement routing, JSON handlers, dependency wiring, and the placeholder frontend response only. Do not build the real frontend in this sprint.
- The intended package rule for this sprint is:
  - `internal/server` translates HTTP requests/responses into calls to `internal/store` and `internal/scheduler`
  - `internal/server` may import `net/http`, JSON, and embed-related packages
  - `internal/server` must not contain SQL or direct database logic
  - `internal/server` must not reimplement SM-2 logic; it must call the injected scheduler
  - `cmd/shortcutdeck/main.go` owns bootstrap wiring and starts the server
- Sprint 5 deliverable from the sprint index: all JSON endpoints are implemented and curl-testable.
- The HTTP API source of truth is `docs/dd.md` §6. Match those routes exactly.
- The behavioral invariants from `docs/dd.md` §7 must be enforced exactly:
  - `internal/scheduler` is pure; the server calls `Schedule()` and then persists the resulting scheduling state with `store.UpdateCard()`
  - scheduling state lives on `model.Card`; there is no separate scheduling table or server-side shadow state
  - `review_log` is append-only; the server only appends via `store.LogReview()`
  - deleting a deck relies on store/schema cascade behavior and must not be reimplemented in HTTP handlers
  - learning paths are read-only in v1; no create/update/delete API endpoints for them
  - `review_active` governs due-card visibility; due-card APIs are only meaningful while active
  - exiting review resets card scheduling state to defaults but retains the review log
- Existing Sprint 1–4 files must be preserved unless a compile-only change is strictly required for server wiring.
- Keep the server implementation small, explicit, and testable. Favor standard-library routing and helpers over extra framework dependencies.

Create or modify exactly these paths as needed for this sprint:

- `internal/server/server.go`
- `internal/server/handlers_decks.go`
- `internal/server/handlers_cards.go`
- `internal/server/handlers_review.go`
- `internal/server/handlers_paths.go`
- `internal/server/handlers_import_export.go`
- `internal/server/server_test.go`
- `cmd/shortcutdeck/main.go`
- `web/index.html`
- `Makefile`

Implementation requirements:

- Replace the stub `internal/server/server.go` with the real server wiring.
- Create the handler files listed above and keep concerns separated by route area.
- Keep dependencies limited to the standard library plus the already-added project packages.
- Add `Makefile` target:
  - `test-server`
  - command: `go test ./internal/server/... -v`
- Define a server type with injected dependencies:
  - a `store.Store`
  - a `scheduler.Scheduler`
  - embedded frontend assets or equivalent standard-library file-serving support for `/`
- Expose or construct an `http.Handler` that registers all Sprint 5 routes.
- Implement these deck endpoints:
  - `GET /api/v1/decks`
  - `POST /api/v1/decks`
  - `GET /api/v1/decks/:id`
  - `PUT /api/v1/decks/:id`
  - `DELETE /api/v1/decks/:id`
- Implement these card endpoints:
  - `GET /api/v1/decks/:id/cards` with optional `?tags=a,b`
  - `POST /api/v1/decks/:id/cards`
  - `GET /api/v1/cards/:id`
  - `PUT /api/v1/cards/:id`
  - `DELETE /api/v1/cards/:id`
- Implement these study-session endpoints:
  - `GET /api/v1/decks/:id/due` with optional `?tags=a,b`
  - `POST /api/v1/cards/:id/review`
  - `POST /api/v1/decks/:id/start-review`
  - `POST /api/v1/decks/:id/exit-review`
- Implement these learning-path endpoints:
  - `GET /api/v1/paths`
  - `GET /api/v1/paths/:id`
- Implement these import/export endpoints:
  - `GET /api/v1/decks/:id/export`
  - `POST /api/v1/import?mode=merge`
  - `POST /api/v1/import?mode=replace`
- Implement `GET /` to return the sprint placeholder frontend response `"hello"`. Do not build real deck-management UI in this sprint.

- Request/response behavior:
  - use JSON for all API responses, including errors
  - accept JSON request bodies for create/update operations and for review submission
  - `POST /api/v1/cards/:id/review` accepts exactly:
    - `{ "grade": 0 }`
    - `{ "grade": 1 }`
    - `{ "grade": 3 }`
    - `{ "grade": 5 }`
  - reject any other review grade with `400`
  - `POST /api/v1/import` accepts a `store.ExportPayload` JSON body
  - reject import modes other than `merge` and `replace` with `400`
  - `GET /api/v1/decks/:id/export` should set headers appropriate for JSON download

- Test requirements:
  - use `net/http/httptest`; no real server process and no real port binding in handler tests
  - add `internal/server/server_test.go` with endpoint coverage for every route in `docs/dd.md` §6
  - use the real `internal/store` implementation backed by an in-memory SQLite database for handler tests
  - do not substitute the curl script for automated handler coverage; the curl script is manual convenience only
  - test each handler through the HTTP layer with realistic requests and JSON bodies
  - cover happy-path behavior for every endpoint, including correct status code and response body shape
  - cover basic error-path behavior where applicable:
    - `404` on unknown ID
    - `400` on malformed request body
    - `400` on invalid review grade
  - prefer the real in-memory SQLite store over a mock store for Sprint 5 so handler tests exercise the actual store contract and data interactions end to end within process

- Review-flow implementation details:
  - fetch the target card from the store
  - map the card’s scheduling fields onto `scheduler.SchedulingState`
  - call the injected scheduler with the provided grade and current time
  - write the returned scheduling fields back onto the card
  - persist the updated card via `store.UpdateCard()`
  - append a `model.ReviewEntry` via `store.LogReview()`
  - return the updated card (and/or another concise JSON success payload) in a way that stays consistent across the API

- Start-review and exit-review implementation details:
  - `start-review` updates the deck so `review_active = true`
  - `exit-review` updates the deck so `review_active = false`
  - on exit-review, reset every card in the deck to scheduler defaults:
    - `interval = 0`
    - `ease_factor = 2.5`
    - `repetitions = 0`
    - `due_date = today`
    - `last_reviewed_at = zero value`
  - retain all `review_log` rows

- Due-card behavior:
  - if `review_active` is false, return an empty JSON list rather than scheduled cards
  - when tags are provided, apply the same AND semantics already defined by `store.ListCards`
  - only include cards due on or before today

- Learning-path behavior:
  - use the store’s read methods only
  - before returning a path, filter out any `deck_ids` that do not correspond to an existing deck
  - do not persist that filtering back into storage

- Entry-point requirements for `cmd/shortcutdeck/main.go`:
  - stop printing the scaffold message
  - initialise the SQLite store
  - instantiate the scheduler implementation (`scheduler.SM2{}`)
  - construct the server
  - start listening on port `7432`
  - no `--port` or `--db` flags in this sprint; those belong to Sprint 8

- Add a curl smoke-test script block to this sprint doc covering at least:
  - create deck
  - create card
  - start review
  - fetch due cards
  - submit a review
  - export the deck
  - import the deck with `mode=merge`
  - exit review

Do not do any of the following in this sprint:

- do not change the `internal/store` interface or redesign persistence behavior
- do not change the scheduler API or SM-2 algorithm
- do not add CLI flags such as `--port` or `--db`
- do not implement browser-opening behavior yet unless strictly required for startup parity
- do not build the real frontend UI; Sprint 6 owns deck management and Sprint 7 owns study UI
- do not add learning-path write endpoints
- do not add authentication, sessions, templating, or third-party HTTP frameworks
- do not add cross-platform packaging or README release work

Stop only when all of the following pass:

- `go test ./internal/server/... -v`
- `go test ./...`
- `go vet ./...`
- `make test-server`
- `make lint`

And verify the API manually with a curl smoke test like this (adapt IDs from responses as needed):

```bash
BASE=http://127.0.0.1:7432

curl -s "$BASE/" 

curl -s -X POST "$BASE/api/v1/decks" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Vim","description":"Keyboard shortcuts","default_reverse_mode":"prompt_first","tag_namespaces":{"mode":["normal"],"topic":["editing"]}}'

curl -s "$BASE/api/v1/decks"

curl -s -X POST "$BASE/api/v1/decks/DECK_ID/cards" \
  -H 'Content-Type: application/json' \
  -d '{"prompt":"Delete to end of line","answer":"d$","notes":"normal mode","tags":["normal","editing"],"source":"manual"}'

curl -s -X POST "$BASE/api/v1/decks/DECK_ID/start-review"

curl -s "$BASE/api/v1/decks/DECK_ID/due"

curl -s -X POST "$BASE/api/v1/cards/CARD_ID/review" \
  -H 'Content-Type: application/json' \
  -d '{"grade":3}'

curl -s "$BASE/api/v1/decks/DECK_ID/export"

curl -s -X POST "$BASE/api/v1/import?mode=merge" \
  -H 'Content-Type: application/json' \
  -d @export.json

curl -s -X POST "$BASE/api/v1/decks/DECK_ID/exit-review"
```

## Out of Scope

- Frontend deck-management UI in `web/index.html`
- Study-session frontend flow and keyboard navigation
- CLI flags such as `--port` and `--db`
- Browser auto-open behavior and release polish
- Learning-path create/update/delete endpoints
- Any store schema or scheduler algorithm redesign
- Cross-package refactors unrelated to the HTTP server boundary

## Notes

- Sprint 5 is the first integration sprint: it composes the already-built model, store, and scheduler layers into the application’s first end-to-end user boundary.
- The main risk in this sprint is accidental contract drift between HTTP handlers and the existing store/scheduler APIs. The prompt should keep those interfaces fixed and make the server adapt to them, not the other way around.
- The placeholder root response is intentionally tiny so Sprint 5 can focus on API correctness and leave real UI work to Sprint 6 and Sprint 7.
- Handler tests in this sprint should use the real `SQLiteStore` with an in-memory SQLite database rather than a mock store. This keeps the tests closer to the actual integration boundary while still using `net/http/httptest` and avoiding a real server process.
- `UpdateDeck` in `internal/store/store.go` did not persist `ReviewActive` when Sprint 5 first exercised `start-review` and `exit-review` end to end. Fixed during this sprint by adding `review_active` to the `UPDATE decks ...` statement. Without that correction, the review-state endpoints had no effect on the deck’s persisted review lifecycle.
