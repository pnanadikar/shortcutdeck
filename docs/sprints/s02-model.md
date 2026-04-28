# Sprint 2 — Model

## Goal

Complete `internal/model` with all shared data types for `shortcutdeck`, while keeping the package pure: no logic, no I/O, and no dependencies on any other internal package.

## Inputs

- [`docs/prd.md`](../prd.md) — use the product data model and naming from Card Format, Deck Format, Spaced Repetition, and Learning Paths.
- [`docs/dd.md`](../dd.md) — use `internal/model` in Package Inventory and the full type definitions in Data Types (`internal/model`) as the source of truth.
- [`docs/sprints/sprint-plan.md`](./sprint-plan.md) — use Sprint 2 as the source of truth for goal, outputs, and acceptance criteria.
- [`internal/model/model.go`](../../internal/model/model.go) — replace the Sprint 1 stub with the real shared type definitions for this sprint.

## Outputs

- `internal/model/model.go`

The output in this sprint is intentionally limited to the shared model package. Do not modify other packages unless a compile-only import adjustment is strictly required by the model implementation.

## Acceptance Criteria

- `internal/model/model.go` defines all shared types from `docs/dd.md` §2:
  - `Card`
  - `Deck`
  - `TagNamespaces`
  - `ReverseMode`
  - `PromptFirst`, `AnswerFirst`, `Both`
  - `ReviewEntry`
  - `LearningPath`
- The package compiles.
- The package imports nothing beyond the standard library.
- The package has no `net/http`, `database/sql`, filesystem, JSON, or other I/O concerns.
- The package has no methods, constructors, validators, helpers, or business logic.
- `go test ./...` exits with status 0.
- `go vet ./...` is clean.

## Prompt

Implement Sprint 2 for the `shortcutdeck` repository.

Context and constraints to restate inline:

- This is the model-only sprint. Define only the shared data types in `internal/model`; do not implement business logic, persistence logic, scheduling logic, server wiring, or UI behavior.
- The intended package rule for this sprint is:
  - `internal/model` is shared by all other internal packages
  - `internal/model` imports no other internal package
  - `internal/model` contains data definitions only
- Sprint 2 deliverable from the sprint index: `internal/model` is complete — all shared types defined, zero logic.
- The source of truth for the exact type shapes is `docs/dd.md` §2. Match those types exactly.
- Existing scaffold files from Sprint 1 must be preserved unless a compile-only change is strictly necessary.
- Keep the public surface minimal and limited to the type definitions and constants required by the design doc.

Create or modify exactly these paths as needed for this sprint:

- `internal/model/model.go`

Implementation requirements:

- Replace the stub `internal/model/model.go` with the full shared model definitions from `docs/dd.md` §2.
- Import only standard library packages needed for the type definitions. `time` is expected because several fields are `time.Time`.
- Define `Card` with:
  - `ID string`
  - `DeckID string`
  - `Prompt string`
  - `Answer string`
  - `Notes string`
  - `Tags []string`
  - `Source string`
  - `CreatedAt time.Time`
  - `Interval int`
  - `EaseFactor float64`
  - `Repetitions int`
  - `DueDate time.Time`
  - `LastReviewedAt time.Time`
- Define `Deck` with:
  - `ID string`
  - `Name string`
  - `Description string`
  - `TagNamespaces TagNamespaces`
  - `DefaultReverseMode ReverseMode`
  - `ReviewActive bool`
  - `CreatedAt time.Time`
- Define `TagNamespaces` as `map[string][]string`.
- Define `ReverseMode` as `string`.
- Define these constants exactly:
  - `PromptFirst ReverseMode = "prompt_first"`
  - `AnswerFirst ReverseMode = "answer_first"`
  - `Both ReverseMode = "both"`
- Define `ReviewEntry` with:
  - `ID string`
  - `CardID string`
  - `ReviewedAt time.Time`
  - `Grade int`
  - `IntervalAfter int`
  - `EaseAfter float64`
- Define `LearningPath` with:
  - `ID string`
  - `Name string`
  - `Description string`
  - `DeckIDs []string`
  - `CreatedAt time.Time`
- Add brief, high-signal doc comments where helpful so the package is readable, but do not add narrative comments unrelated to the model definitions.
- Keep field names and exported identifiers aligned with `docs/dd.md` §2.

Do not do any of the following in this sprint:

- do not add methods on any model type
- do not add constructor functions
- do not add validation helpers
- do not add JSON tags unless they are explicitly required by the design doc for this sprint
- do not add database tags, ORM metadata, or SQL helpers
- do not add scheduler types from `internal/scheduler`
- do not add store interfaces, migrations, or SQLite wiring
- do not add HTTP handlers, request/response types, or server code
- do not change `cmd/shortcutdeck/main.go`, `internal/store`, `internal/server`, or `internal/scheduler` unless required for compilation

Stop only when all of the following pass:

- `go test ./...`
- `go vet ./...`

## Out of Scope

- Scheduler interfaces, grades, or algorithm code in `internal/scheduler`
- Store interfaces, migrations, SQL schema, or persistence code
- HTTP handlers, routing, JSON APIs, or embed wiring
- CLI flags, runtime config, or browser-opening behavior
- Frontend implementation in `web/`
- Any validation or transformation logic on model types

## Notes

- `internal/model` is the dependency root for the internal packages. Keep it intentionally small and dependency-free.
- The design doc shows comments alongside the types; matching the type shapes exactly matters more than comment wording.
- This sprint should leave the repository ready for Sprint 3 and Sprint 4 to import these shared types without having to reshape them.
