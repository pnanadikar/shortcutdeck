# Sprint 1 — Scaffold

## Goal

Create only the project scaffold for `shortcutdeck`: a valid Go module, a minimal executable entrypoint, empty internal package layout, a `Makefile` with `build`, `test`, and `run` targets, and no business logic.

## Inputs

- [`docs/prd.md`](../prd.md) — use the runtime and stack constraints from Technical Stack / Architecture, especially Go CLI + embedded local web server, single embedded frontend file, and SQLite via `modernc/sqlite`.
- [`docs/dd.md`](../dd.md) — use the intended package layout and package responsibilities for `cmd/shortcutdeck`, `internal/model`, `internal/scheduler`, `internal/store`, `internal/server`, and `web/`.
- [`docs/sprints/sprint-plan.md`](./sprint-plan.md) — use Sprint 1 as the source of truth for goal, outputs, and acceptance criteria.
- [`README.md`](../../README.md) and [`.gitignore`](../../.gitignore) — preserve or reconcile existing root repo files rather than replacing them blindly.

## Outputs

- `go.mod`
- `go.sum`
- `cmd/shortcutdeck/main.go`
- `internal/model/model.go`
- `internal/scheduler/scheduler.go`
- `internal/store/`
- `internal/server/`
- `web/`
- `web/.keep`
- `web/index.html`
- `Makefile`
- `.gitignore`

All outputs in this sprint are scaffold-only. File .keep created in scaffold to track empty directory in git; deleted when index.html is added. They must contain the minimum files and code required for the repository to compile, test, and run a stub executable cleanly.

## Acceptance Criteria

- `make build` succeeds and produces a binary.
- `make test` exits with status 0.
- `go vet ./...` is clean.
- `go list ./...` includes all planned packages.
- `make run` starts the stub app without immediate failure from missing scaffold wiring.
- No real product logic, SQL schema, store implementation, scheduler logic, HTTP handlers, flags, or frontend behavior exists yet.

## Prompt

Implement Sprint 1 for the `shortcutdeck` repository.

Context and constraints to restate inline:

- This is a scaffold-only  (compile-only) sprint; do not implement any product logic (later sprints own all real interfaces and logic).
- Public surface allowed in this sprint should be limited to `package main` with `func main()` and empty internal packages that compile.
- The intended architecture is:
  - `cmd/shortcutdeck/main.go` — executable entrypoint
  - `internal/model` — shared types, but do not define them in this sprint unless strictly required for compilation
  - `internal/scheduler` — scheduler package stub only
  - `internal/store` — store package stub only
  - `internal/server` — server package stub only
  - `web/` — future embedded frontend location
- Sprint 1 deliverable from the sprint index: repo compiles, Makefile works, empty packages in place.
- Required stack constraints:
  - Go project
  - local embedded web-server architecture later
  - frontend served from embedded files later
  - SQLite planned later via `modernc/sqlite`
- Existing root files must be preserved or reconciled:
  - preserve `.gitignore` and extend it only if the scaffold introduces missing build/test artifact paths
  - do not replace existing repo metadata unnecessarily
- Use the repo’s canonical git remote path as the default Go module path.

Create or modify exactly these paths as needed for the scaffold:

- `go.mod`
- `go.sum`
- `cmd/shortcutdeck/main.go`
- `internal/model/`
- `internal/scheduler/`
- `internal/store/`
- `internal/server/`
- `web/`
- `web/index.html`
- `Makefile`
- `.gitignore`

Implementation requirements:

- Add a valid `go.mod`.
- Add only the minimum dependencies required for scaffold compilation.
- Create `cmd/shortcutdeck/main.go` with `package main` and a minimal `func main()` that runs safely as a stub.
- Create stub Go files for `internal/model`, `internal/scheduler`, `internal/store`, and `internal/server` so all planned packages compile.
- Create `web/` only as a placeholder location for future embedded assets. Do not implement real UI.
- Create `web/index.html` as a one-line HTML comment placeholder so a future `embed.FS` pattern has a real file to include at compile time.
- Add a `Makefile` with these targets only as first-class sprint requirements:
  - `build`
  - `test`
  - `run`
- Keep the repository intentionally empty of business behavior.

Do not do any of the following in this sprint:

- do not define real domain types in `internal/model`
- do not add store interfaces, migrations, or SQLite wiring
- do not implement SM-2 or scheduler interfaces
- do not add HTTP routes, handlers, JSON APIs, or embed wiring
- do not add CLI flags such as `--port` or `--db`
- do not add real HTML, CSS, or JavaScript behavior
- do not add release automation or cross-platform packaging

Stop only when all of the following pass:

- `make build`
- `make test`
- `go vet ./...`

## Out of Scope

- All domain types from `internal/model`
- Store interfaces and migrations
- Scheduler interfaces or SM-2 logic
- HTTP handlers, routes, and embed wiring
- Real HTML, CSS, or JavaScript
- CLI flags such as `--port` and `--db`
- Release targets and cross-platform packaging
- README expansion except minimal scaffold-related touchups if strictly necessary

## Notes

- Default Go module path should come from the repo remote path unless intentionally overridden later.
- Existing `.gitignore` must be preserved, not replaced.
- `web/` may contain only a placeholder artifact if needed for future embedding.
