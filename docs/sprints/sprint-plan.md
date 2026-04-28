# Sprint Plan — shortcutdeck v1

> Last updated: 2026-04-04
> Status: Active
> Companion docs: prd.md (v0.6), dd.md (v0.5)

---

## Folder Structure

```plaintext
shortcutdeck/
├── cmd/
│   └── shortcutdeck/
│       └── main.go
├── internal/
│   ├── model/
│   │   └── model.go
│   ├── scheduler/
│   │   ├── scheduler.go
│   │   └── scheduler_test.go
│   ├── store/
│   │   ├── store.go
│   │   ├── store_test.go
│   │   └── migrations/
│   │       └── 001_init.sql
│   └── server/
│       ├── server.go
│       ├── handlers_decks.go
│       ├── handlers_cards.go
│       ├── handlers_review.go
│       ├── handlers_paths.go
│       └── handlers_import_export.go
├── web/
│   └── index.html
├── docs/
│   ├── prd.md
│   ├── dd.md
│   └── sprints/
│       ├── sprint-plan.md          ← this file
│       ├── s01-scaffold.md
│       ├── s02-model.md
│       ├── s03-store.md
│       ├── s04-scheduler.md
│       ├── s05-server-api.md
│       ├── s06-frontend-deck-mgmt.md
│       ├── s07-study.md
│       ├── s08-ui-polish.md
│       └── s09-ship.md
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Sprint Document Template

Each `docs/sprints/sNN-name.md` follows this structure:

```plaintext
# Sprint N — <Title>

## Goal
One sentence. What is true at the end of this sprint that wasn't true before.

## Inputs
Files and interfaces this sprint reads/depends on.

## Outputs
Files created or modified. Specific paths.

## Acceptance Criteria
Concrete, verifiable statements. "go test ./... passes" counts. "looks good" does not.

## Prompt
Verbatim prompt to paste into Cursor / VS Code Copilot Chat.
Self-contained: embeds all relevant context. Assumes no AI memory of prior sessions.

## Out of Scope
Explicit list of things NOT to do in this sprint.

## Notes
Decisions made during the sprint, gotchas, follow-up issues.
```

---

## Sprint Index

| # | Name | Branch | Deliverable | Tooling Note |
| --- | ------ | -------- | ------------- | --- |
| 1 | Scaffold | `feat/s01-scaffold` | Repo compiles, Makefile works, empty packages in place | — |
| 2 | Model | `feat/s02-model` | `internal/model` — all shared types, zero logic | — |
| 3 | Store | `feat/s03-store` | `internal/store`, SQL schema, store tests | Introduces `make lint` and `make test-store` |
| 4 | Scheduler | `feat/s04-scheduler` | `internal/scheduler`, SM-2 implementation + unit tests | Adds `make test-scheduler` |
| 5 | HTTP Server + API | `feat/s05-server-api` | All JSON endpoints, curl-testable | Existing lint/test targets required |
| 6 | Frontend — Deck Mgmt | `feat/s06-deck-mgmt` | Deck list, card CRUD, tag namespaces, export/import | Existing lint/test targets required |
| 7 | Frontend — Study | `feat/s07-study` | Study session + browse mode, full keyboard nav | Existing lint/test targets required |
| 8 | UI Polish + UX Cleanup | `feat/s08-ui-polish` | Frontend polish, state clarity, styling consistency, targeted UX fixes | Existing lint/build checks required |
| 9 | Polish + Ship | `feat/s09-ship` | `--port`/`--db` flags, cross-platform builds, README | Adds `make release` |

---

## Sprint Summaries

### Sprint 1 — Scaffold

**Goal:** A repo that compiles, has correct module layout, and a Makefile with `build`, `test`, `run` targets. No logic.

**Outputs:** `go.mod`, stub `main.go`, empty package stubs for all `internal/` packages, `Makefile`, `.gitignore`.

**Acceptance:** `make build` produces a binary. `make test` exits 0. `go vet ./...` is clean.

---

### Sprint 2 — Model

**Goal:** `internal/model` is complete — all shared data types defined, no logic, no I/O, no dependencies on any other internal package.

**Outputs:** `internal/model/model.go` with all types from dd.md §2: `Card`, `Deck`, `TagNamespaces`, `ReverseMode` constants, `ReviewEntry`, `LearningPath`.

**Acceptance:** Package compiles. Imports nothing beyond the standard library. No `net/http`, `database/sql`, or I/O packages. Types exactly match dd.md §2.

---

### Sprint 3 — Store

**Goal:** `internal/store` complete: schema created on startup, all CRUD methods implemented, tested against a real in-memory SQLite instance.

**Outputs:** `internal/store/store.go`, `internal/store/migrations/001_init.sql`, `internal/store/store_test.go`. Introduces `golangci-lint`, adds `.golangci.yml`, adds `make lint`, and adds `make test-store`.

**Note:** `internal/scheduler` does not exist yet. The store has no scheduler dependency — scheduling state fields are plain columns on `cards`. A stub `internal/scheduler/scheduler.go` (empty package declaration only) is sufficient for the repo to compile.

**Key test cases:** CreateDeck round-trips; CreateCard sets due=today and default SM-2 state; GetDueCards filters correctly; DeleteDeck cascades to cards and review_log; LogReview is append-only; ImportDeck merge mode upserts cards by ID without touching others; ImportDeck replace mode deletes existing cards first; an ExportPayload containing a LearningPath upserts the path on import.

**Acceptance:** `go test ./internal/store/...` green. `make test-store` is green. `make lint` is clean. Uses `modernc.org/sqlite` (no CGo). Schema matches dd.md §5 exactly. No `net/http` import.

---

### Sprint 4 — Scheduler

**Goal:** `internal/scheduler` complete and fully tested. Zero database, zero HTTP. By this point the store tests have exercised the scheduling state fields concretely, so the algorithm can be implemented against a well-understood data shape.

**Outputs:** `internal/scheduler/scheduler.go` (replaces the Sprint 3 stub) with `Scheduler` interface, `SM2` struct, `NewState()`, `Grade` constants. `internal/scheduler/scheduler_test.go` with ≥ 10 test cases. Adds `make test-scheduler`.

**Key test cases:** new-card defaults, Again resets repetitions to zero, Hard/Good/Easy interval progression, ease-factor floor at 1.3, successive correct reviews compound interval correctly.

**Acceptance:** `go test ./internal/scheduler/... -v` green. `make test-scheduler` is green. Package imports nothing beyond the standard library (`time`).

---

### Sprint 5 — HTTP Server + API

**Goal:** All endpoints from dd.md §6 implemented and curl-testable. Frontend placeholder returns `"hello"` at `/`.

**Outputs:** `internal/server/server.go` + all handler files. `cmd/shortcutdeck/main.go` wires dependencies and starts on port 7432.

**Acceptance:** Every endpoint in dd.md §6 returns correct status codes for happy-path and basic error cases (404 on missing ID, 400 on bad body). `go vet ./...` clean. Sprint doc includes a curl smoke-test script.

---

### Sprint 6 — Frontend: Deck Management

**Goal:** `web/index.html` renders a working deck list, deck CRUD, card list, card CRUD, tag namespace editor, export, and import.

**Outputs:** `web/index.html`, `web/app.css`, `web/app.js`, `web/embed.go`, plus minimal static-serving changes in `internal/server/server.go` and focused asset-serving coverage in `internal/server/server_test.go` if needed.

**Keyboard map:** `n` new item, `e` edit focused item, `d`/`Delete` delete, `Esc` cancel/back.

**Acceptance:** Create a Vim deck with 5 cards using only the keyboard. Export is valid `ExportPayload` JSON. Import via frontend `replace` mode round-trips cleanly into a fresh db.

---

### Sprint 7 — Frontend: Study Session + Browse Mode

**Goal:** Study session and browse mode fully functional from the UI.

**Outputs:** additions to `web/index.html` only.

**Study flow:** deck → session-options (reverse mode, tag filter) → card-by-card → summary.
**Keyboard:** `Space`/`Enter` reveal, `1`–`4` rate, `e` edit in-session, `Esc` quit.
**`both` mode:** client interleaves two passes over the due-card list (each card gets a `side` flag).

**Browse flow:** deck → browse screen → `j`/`k` navigate, `e` edit, `Esc` back.

**Acceptance:** Complete a full study session on the Sprint 6 Vim deck. Verify SM-2 intervals updated (check via export). Complete a browse-mode pass through all cards.

---

### Sprint 8 — UI Polish + UX Cleanup

**Goal:** The existing Sprint 6 and Sprint 7 frontend feels cohesive and polished enough that the release sprint can focus on shipping rather than ongoing UI cleanup.

**Outputs:** Primarily `web/index.html`, `web/app.css`, `web/app.js`, plus targeted documentation updates in `docs/open-issues.md` and `docs/sprints/s08-ui-polish.md`.

**Focus:** visual consistency, focus/selection state clarity, dark-mode cleanup, study/browse polish, and targeted UX fixes such as sidebar highlighting and tag-vocabulary enforcement when decks define namespaces.

**Acceptance:** `make build` and `make lint` pass. Manual verification confirms deck management, study, browse, and summary flows still work cleanly after polish changes.

---

### Sprint 9 — Polish + Ship

**Goal:** v1.0 is releasable. Binary works on Linux, macOS, Windows.

**Outputs:** `--port` and `--db` flags in `main.go`. `Makefile` `release` target → `dist/shortcutdeck-{linux,darwin,windows}-amd64`. `README.md` covering install, run, keyboard reference, export/import, data location.

**Acceptance:** `make release` produces three binaries. Each opens the browser on first run. `--help` documents all flags. README is accurate.

---

## Prompt Discipline Rules

These rules apply to every sprint prompt. Encode them as habits, not guidelines.

1. **Restate constraints inline.** Never say "see dd.md §7". Paste the relevant invariants into the prompt. The coding assistant has no reliable memory of prior sessions.
2. **Name every file to create or modify.** "Create `internal/store/store.go` containing..." beats "implement the store layer."
3. **State what not to do.** "Do not add any methods to the `Store` interface beyond those listed below" prevents well-intentioned additions that break the contract.
4. **End with the acceptance test.** Give the assistant a concrete stopping condition: the exact command to run and what passing looks like.
5. **One sprint = one branch = one PR.** Never carry partial work across a sprint boundary.

---

## Known Gotchas

- When a later sprint exercises an existing store method for the first time in a real end-to-end flow, persistence gaps can surface even if the store package already has passing unit tests.
- Review every `UpdateX` store method against its corresponding model fields before closing a sprint that depends on it. Confirm the SQL `UPDATE` statement persists every field the model exposes and the caller expects to change.
- Store happy-path round-trip tests are not always enough to catch partial persistence bugs. Add assertions for every mutable field that matters to later integration flows, especially booleans and state-transition fields such as `review_active`.
- Go's `//go:embed` directive cannot reference paths outside the package directory. Embed directives for frontend assets should live in `web/embed.go` and be exported as `web.FS` for consumption by `internal/server`.

---

## Makefile Targets

| Target | Command | Introduced |
| --- | --- | --- |
| `build` | `go build ./cmd/shortcutdeck` | Sprint 1 |
| `test` | `go test ./...` | Sprint 1 |
| `run` | `go run ./cmd/shortcutdeck` | Sprint 1 |
| `lint` | `golangci-lint run` | Sprint 3 |
| `test-store` | `go test ./internal/store/... -v` | Sprint 3 |
| `test-scheduler` | `go test ./internal/scheduler/... -v` | Sprint 4 |
| `release` | cross-platform builds | Sprint 9 |

---

## Linting

From Sprint 3 onward, `golangci-lint` is part of the planned development workflow.

- Sprint 3 is the introduction point for `golangci-lint`.
- Sprint 3 should add `.golangci.yml`.
- Sprint 3 should add `make lint` to the Makefile.
- Sprint 3 should add `make test-store` to the Makefile.
- Sprint 4 should add `make test-scheduler` to the Makefile.
- From Sprint 3 onward, lint cleanliness is a standing acceptance and merge requirement.

---

## Git Workflow

For every sprint, use this workflow:

1. Create and work on a dedicated feature branch for that sprint, for example `feat/s01-scaffold`.
2. Keep all sprint changes on that feature branch until the sprint acceptance criteria are clean.
3. Commit the sprint work on the feature branch with a sprint-specific message.
4. Switch to `main` only after the sprint branch is ready to merge.
5. Merge with `--no-ff` so each sprint lands as an explicit merge commit on `main`.
6. Delete the local feature branch after the merge succeeds.
7. Push `main` to `origin`.
8. From Sprint 3 onward, `make lint` must pass before merging.

If acceptance criteria fail mid-sprint:

1. Stay on the feature branch.
2. Fix the issue there and commit again on the same sprint branch.
3. Do not merge until `make test` and `go vet ./...` are clean.
4. Treat the sprint as incomplete until the branch is green.
