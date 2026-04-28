# Sprint 3 — Store

## Goal

Complete `internal/store` with a SQLite-backed implementation of the store contract, including schema migrations and tests against a real in-memory SQLite database.

## Inputs

- [`docs/prd.md`](../prd.md) — use the storage model, export/import behavior, and scheduling defaults from Storage, Card Format, Deck Format, and Learning Paths.
- [`docs/dd.md`](../dd.md) — use `internal/store` in Package Inventory, the Store Interface in §4, and the SQL Schema in §5 as the source of truth.
- [`docs/sprints/sprint-plan.md`](./sprint-plan.md) — use Sprint 3 as the source of truth for goal, outputs, key test cases, and acceptance criteria.
- [`internal/model/model.go`](../../internal/model/model.go) — use the Sprint 2 shared types exactly as the store input/output surface.
- [`internal/store/store.go`](../../internal/store/store.go) — replace the Sprint 1 stub with the real store implementation for this sprint.

## Outputs

- `internal/store/store.go`
- `internal/store/store_test.go`
- `internal/store/migrations/001_init.sql`
- `.golangci.yml`
- `Makefile`
- `go.mod`
- `go.sum`

The core sprint output is the store package and its tests. Sprint 3 also introduces `golangci-lint`, adds `make lint`, and adds `make test-store`. `go.mod` and `go.sum` may change only as needed to add the minimum dependency set required for a pure-Go SQLite implementation and test execution.

## Acceptance Criteria

- `internal/store/store.go` defines the Store interface, `ExportPayload`, `SQLiteStore`, and `New(dbPath string) (Store, error)` exactly as specified in `docs/dd.md` §4.
- The schema in `internal/store/migrations/001_init.sql` matches `docs/dd.md` §5.
- The implementation uses `modernc.org/sqlite` and does not require CGo.
- The package does not import `net/http`.
- `go test ./internal/store/...` exits with status 0.
- `make test-store` exits with status 0.
- `make lint` is clean (golangci-lint).

## Prompt

Implement Sprint 3 for the `shortcutdeck` repository.

Context and constraints to restate inline:

- This is the store-only sprint. Implement SQLite persistence in `internal/store`, plus schema migration and tests. Do not implement scheduler logic, HTTP handlers, frontend behavior, or CLI wiring.
- The intended package rule for this sprint is:
  - `internal/store` accepts and returns `internal/model` types
  - `internal/store` owns all database access and schema migration
  - `internal/store` must not contain HTTP concepts or UI concerns
- Sprint 3 deliverable from the sprint index: `internal/store` complete — schema created on startup, all CRUD methods implemented, tested against a real in-memory SQLite instance.
- The source of truth for the store API is `docs/dd.md` §4. Match the Store interface and `ExportPayload` exactly.
- The source of truth for the schema is `docs/dd.md` §5. Match the tables, columns, defaults, and indexes exactly.
- Existing Sprint 1 and Sprint 2 files must be preserved unless a store implementation change requires a direct dependency update in `go.mod` or `go.sum`.
- Sprint 3 introduces `golangci-lint`, adds `.golangci.yml`, adds `make lint`, and adds `make test-store` to the Makefile.
- From Sprint 3 onward, `make lint` must pass before merge. Keep the implementation lint-clean.

Create or modify exactly these paths as needed for this sprint:

- `internal/store/store.go`
- `internal/store/store_test.go`
- `internal/store/migrations/001_init.sql`
- `.golangci.yml`
- `Makefile`
- `go.mod`
- `go.sum`

Implementation requirements:

- Replace the stub `internal/store/store.go` with the full store implementation.
- Add only the minimum dependencies required for a pure-Go SQLite implementation and tests. Use `modernc.org/sqlite`.
- Add `.golangci.yml` with the starter lint configuration for this project.
- Add `make lint` and `make test-store` to the Makefile.
- Define the Store interface exactly as shown in `docs/dd.md` §4:
  - deck CRUD methods
  - card CRUD methods plus `GetDueCards`
  - review-log methods
  - learning-path methods
  - import/export methods
- Define `ExportPayload` exactly as shown in `docs/dd.md` §4.
- Define a concrete `SQLiteStore` type with unexported fields as needed.
- Implement `New(dbPath string) (Store, error)` so it:
  - opens or creates the SQLite database
  - enables foreign keys if needed
  - runs pending migrations from `internal/store/migrations/001_init.sql`
  - returns a ready store implementation
- Implement all CRUD and import/export behavior using `internal/model` types only.
- Keep JSON storage fields aligned with the design doc:
  - `tag_namespaces` stored as JSON
  - `tags` stored as a JSON array
  - `deck_ids` stored as a JSON array
- Apply the documented scheduling defaults for newly created cards:
  - `interval = 0`
  - `ease_factor = 2.5`
  - `repetitions = 0`
  - `due_date = today`
- Implement import/export behavior exactly as described:
  - `ExportDeck` returns the deck and all of its cards as `ExportPayload`
  - `ImportDeck(..., "merge")` upserts cards by ID without touching other existing cards
  - `ImportDeck(..., "replace")` deletes existing cards for the deck before inserting imported cards
  - when a payload includes `LearningPath`, the path is upserted by ID on import
- Handle cascade delete semantics through the schema and/or implementation so deleting a deck removes its cards and related review log entries.
- Use synthetic, deterministic test data in `internal/store/store_test.go`. Do not depend on external fixtures or a preexisting database.
- Test against a real in-memory SQLite instance, not mocks.
- Cover these key test cases explicitly:
  - `CreateDeck` round-trips
  - `CreateCard` sets `due_date=today` and default SM-2 state
  - `GetDueCards` filters correctly
  - `DeleteDeck` cascades to cards and `review_log`
  - `LogReview` is append-only
  - `ImportDeck` merge mode upserts cards by ID without touching others
  - `ImportDeck` replace mode deletes existing cards first
  - an `ExportPayload` containing a `LearningPath` upserts the path on import
- Keep helper functions internal to the store package and tests. Do not leak implementation details into other packages.

Do not do any of the following in this sprint:

- do not import or use `net/http`
- do not add scheduler interfaces or algorithm code
- do not add server handlers, routes, or request/response types
- do not add frontend code or embed wiring
- do not add CLI flags or entrypoint wiring
- do not redesign the store interface beyond what `docs/dd.md` §4 specifies
- do not use CGo-backed SQLite drivers
- do not replace real database tests with mocks

Stop only when all of the following pass:

- `go test ./internal/store/...`
- `make test-store`
- `make lint`

## Out of Scope

- Scheduler interfaces and SM-2 implementation in `internal/scheduler`
- HTTP server wiring and JSON API behavior in `internal/server`
- Frontend behavior in `web/`
- CLI flags and runtime bootstrap logic in `cmd/shortcutdeck/main.go`
- Cross-package refactors unrelated to the store contract

## Notes

- Sprint 3 is the first sprint that needs real module dependency changes because SQLite support must be added.
- The tests will need synthetic test data for decks, cards, review entries, and import/export payloads.
- The store package should be the only place that knows SQL details. Everything crossing the package boundary should be `internal/model` types.
- Post-Sprint 5 correction: `UpdateDeck` originally failed to persist the `review_active` column even though `model.Deck` includes `ReviewActive`. This surfaced only when Sprint 5 exercised `start-review` and `exit-review` through the real HTTP flow. The final store implementation now includes `review_active` in the deck update statement.
