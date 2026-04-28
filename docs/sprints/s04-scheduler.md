# Sprint 4 — Scheduler

## Goal

Complete `internal/scheduler` with a pure, fully tested SM-2 implementation that operates only on scheduling state and time values, with no database, HTTP, or file I/O concerns.

## Inputs

- [`docs/prd.md`](../prd.md) — use the SM-2 algorithm choice, grade mapping, and scheduling fields from Spaced Repetition and Study Session Flow.
- [`docs/dd.md`](../dd.md) — use the scheduler package contract in Scheduler Interface (`internal/scheduler`) as the source of truth.
- [`docs/sprints/sprint-plan.md`](./sprint-plan.md) — use Sprint 4 as the source of truth for goal, outputs, key test cases, and acceptance criteria.
- [`internal/model/model.go`](../../internal/model/model.go) — use the store-established scheduling field shapes as background context only; do not add a dependency from `internal/scheduler` to `internal/model`.
- [`internal/scheduler/scheduler.go`](../../internal/scheduler/scheduler.go) — replace the Sprint 1 stub with the real scheduler implementation for this sprint.

## Outputs

- `internal/scheduler/scheduler.go`
- `internal/scheduler/scheduler_test.go`
- `Makefile`

The core sprint output is the scheduler package and its tests. Sprint 4 also adds `make test-scheduler` to the Makefile.

## Acceptance Criteria

- `internal/scheduler/scheduler.go` defines `SchedulingState`, `Grade`, `GradeAgain`, `GradePartial`, `GradeHard`, `GradeEasy`, `Scheduler`, `SM2`, and `NewState()` exactly as specified in `docs/dd.md` §3.
- `SM2.Schedule` is implemented as a pure function with no side effects and no I/O.
- The package imports nothing beyond the standard library, specifically `time`.
- `go test ./internal/scheduler/... -v` exits with status 0.
- `make test-scheduler` exits with status 0.
- `make lint` is clean (golangci-lint).

## Prompt

Implement Sprint 4 for the `shortcutdeck` repository.

Context and constraints to restate inline:

- This is the scheduler-only sprint. Implement the SM-2 algorithm in `internal/scheduler` plus tests. Do not implement store logic, HTTP handlers, frontend behavior, or CLI wiring.
- The intended package rule for this sprint is:
  - `internal/scheduler` contains pure scheduling logic only
  - `internal/scheduler` must not import any other internal package
  - `internal/scheduler` must not know about databases, HTTP, terminals, files, or UI state
- Sprint 4 deliverable from the sprint index: `internal/scheduler` complete and fully tested.
- The source of truth for the scheduler API is `docs/dd.md` §3. Match the scheduler types and exported symbols exactly.
- The source of truth for grade semantics and scheduling behavior is `docs/prd.md` §6.
- Existing Sprint 1–3 files must be preserved unless a compile-only change is strictly required.
- Sprint 4 adds `make test-scheduler` to the Makefile.
- From Sprint 3 onward, `make lint` must pass before merge. Keep the implementation lint-clean.

Create or modify exactly these paths as needed for this sprint:

- `internal/scheduler/scheduler.go`
- `internal/scheduler/scheduler_test.go`
- `Makefile`

Implementation requirements:

- Replace the stub `internal/scheduler/scheduler.go` with the full scheduler implementation.
- Define `SchedulingState` exactly as shown in `docs/dd.md` §3:
  - `Interval int`
  - `EaseFactor float64`
  - `Repetitions int`
  - `DueDate time.Time`
  - `LastReviewedAt time.Time`
- Define `Grade` as `int`.
- Define these constants exactly:
  - `GradeAgain Grade = 0`
  - `GradePartial Grade = 1`
  - `GradeHard Grade = 3`
  - `GradeEasy Grade = 5`
- Define the `Scheduler` interface exactly as shown in `docs/dd.md` §3.
- Define `SM2` as the v1 implementation of `Scheduler`.
- Implement `func (SM2) Schedule(current SchedulingState, grade Grade, now time.Time) SchedulingState` as a pure function.
- Implement `func NewState() SchedulingState` returning the SM-2 default state for a brand-new card:
  - `Interval = 0`
  - `EaseFactor = 2.5`
  - `Repetitions = 0`
  - `DueDate = today`
  - `LastReviewedAt = zero value`
- Apply the PRD grade mapping semantics:
  - `GradeAgain` resets learning progress
  - `GradePartial` is still a miss but less severe than `Again`
  - `GradeHard` grows interval slowly
  - `GradeEasy` grows interval aggressively
- Enforce the SM-2 ease-factor floor at `1.3`.
- Keep date arithmetic explicit and deterministic. Normalize new due dates based on the passed `now` value rather than wall-clock time hidden inside helpers.
- Add `make test-scheduler` to the Makefile as:
  - `go test ./internal/scheduler/... -v`
- Write a test suite with at least 10 cases covering:
  - new-card defaults
  - Again resets repetitions to zero
  - Partial behaves as a miss with less severe penalty than Again
  - Hard interval progression
  - Easy interval progression
  - successive correct reviews compound interval correctly
  - ease-factor floor at `1.3`
  - `LastReviewedAt` updates on schedule
  - `DueDate` advances from the provided `now`
  - output stays deterministic for fixed input values
- Keep tests table-driven where it helps readability, but prefer clarity over cleverness.

Do not do any of the following in this sprint:

- do not import `internal/model`
- do not import or use `database/sql`
- do not import or use `net/http`
- do not add store, server, or frontend code
- do not add file I/O, JSON handling, or embed wiring
- do not pull in external SM-2 or FSRS libraries
- do not implement FSRS in this sprint
- do not change `internal/store` or `internal/server` unless required for compilation

Stop only when all of the following pass:

- `go test ./internal/scheduler/... -v`
- `make test-scheduler`
- `make lint`

## Out of Scope

- SQLite persistence and schema changes in `internal/store`
- HTTP handlers and routing in `internal/server`
- Frontend behavior in `web/`
- CLI flags and runtime wiring in `cmd/shortcutdeck/main.go`
- Alternative scheduling algorithms such as FSRS

## Notes

- Sprint 4 should stay small and mathematically focused. This package is meant to be easy to trust and easy to test.
- Store and UI layers will call into this package later, so the API should remain narrow and algorithm-focused.
- The main risk in this sprint is hidden ambiguity in SM-2 interval progression. Tests should pin down the intended behavior clearly so later layers can depend on it safely.
