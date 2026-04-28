# Sprint 9 — Polish + Ship

## Goal

Make `shortcutdeck` feel ready to ship as a real v1 binary: runtime flags are in place, startup/shutdown logging is intentional and testable, release packaging is automated, and the README is complete enough for first-time use without reading the design docs.

## Inputs

- [`docs/sprints/sprint-plan.md`](./sprint-plan.md) — use Sprint 9 as the source of truth for branch name (`feat/s09-ship`), release intent, and the standing expectation that `make release` exists by the end of the sprint.
- [`docs/prd.md`](../prd.md) — use §11 and §12 for runtime/distribution expectations:
  - Go CLI + embedded local web server
  - single binary per platform
  - default port `7432`, overridable via `--port`
  - browser auto-open enabled by default, overridable via `--auto-open=false`
- [`docs/dd.md`](../dd.md) — use the current `cmd/shortcutdeck` package role as the source of truth:
  - parses CLI flags
  - resolves the database path
  - initialises the store
  - wires dependencies
  - starts the HTTP server
  - opens the browser
- [`docs/open-issues.md`](../open-issues.md) — Sprint 9 must absorb the deferred logging work called out for `cmd/shortcutdeck`.
- [`cmd/shortcutdeck/main.go`](../../cmd/shortcutdeck/main.go) — current runtime entrypoint; presently hardcodes the default address and default DB path and uses package-level `log.Printf` / `log.Fatalf`.
- [`Makefile`](../../Makefile) — extend the current local dev targets with a release target; preserve the existing cache-aware target style.
- [`README.md`](../../README.md) — currently only a placeholder and should become the primary user-facing v1 guide.
- Existing `internal/server`, `internal/store`, and `web/` behavior — use the already-completed Sprint 5-8 work as the shipped product surface; Sprint 9 is primarily runtime, packaging, docs, and polish.

## Outputs

- `cmd/shortcutdeck/main.go`
- `cmd/shortcutdeck/main_test.go`
- `Makefile`
- `README.md`
- `docs/sprints/s09-ship.md`

Add additional files only if they are directly required to support Sprint 9 verification or packaging. Avoid widening the sprint into backend or frontend feature work.

## Acceptance Criteria

- `cmd/shortcutdeck/main.go` supports explicit runtime flags:
  - `--port`
  - `--db`
  - `--auto-open`
- Running the binary without flags preserves the current v1 defaults:
  - server listens on port `7432`
  - database path defaults to the OS user config directory plus `shortcutdeck/cards.db`
- `--help` documents the runtime flags clearly enough for a first-time user to discover:
  - what `--port` controls
  - what `--db` controls
  - what `--auto-open` controls
  - what the default values are
- Startup/runtime logging in `cmd/shortcutdeck` is cleaned up and intentional:
  - one clear startup message with the URL being served
  - browser auto-open failures are logged as warnings/non-fatal events
  - shutdown is logged clearly
  - fatal startup failures still terminate with actionable error messages
  - logging behavior is testable or otherwise structured enough to avoid being trapped in ad hoc package-global side effects
- `cmd/shortcutdeck/main_test.go` covers the Sprint 9 runtime logic that is practical to unit test, such as:
  - flag/default resolution
  - URL/address construction from the chosen port
  - default DB path resolution helper behavior
  - any extracted logging/config helpers introduced to make `main.go` less brittle
- `Makefile` includes a `release` target that produces release binaries under `dist/`.
- `make release` produces at least the Sprint-plan release set:
  - `dist/shortcutdeck-linux-amd64`
  - `dist/shortcutdeck-darwin-amd64`
  - `dist/shortcutdeck-windows-amd64.exe`
- Release builds are produced with the existing embedded frontend included; the release target must build the real app, not a reduced CLI-only binary.
- `README.md` is expanded from placeholder text into a usable v1 guide covering:
  - what the tool is
  - supported platforms
  - how to build and run it
  - default browser-launch behavior
  - how to disable browser auto-open for service/non-interactive use
  - default DB location and `--db` override
  - `--port` usage
  - keyboard-first study and deck/card management summary
  - export/import behavior
  - release artifact names or local release build instructions
- The README reflects current product reality only; it must not describe unbuilt v2 features as available.
- Existing Sprint 8 product behavior remains intact:
  - deck/card management still works
  - study session flow still works
  - Quick Refresher still works
  - no API contract changes are introduced
- `make build` succeeds.
- `make test` succeeds.
- `make lint` is clean.
- `make release` succeeds.

## Manual Test Checklist

This checklist is the stopping condition for Sprint 9. Failed or blocked items should be recorded in [`docs/open-issues.md`](../open-issues.md), not inline in this sprint doc.

- [x] Run `make build`; verify the binary builds successfully from the current tree.
- [x] Run `./shortcuts --help`; verify `--port`, `--db`, and `--auto-open` are documented with clear descriptions.
- [x] Run `./shortcuts`; verify the app starts on `http://127.0.0.1:7432`, logs a single clear startup message, and still auto-opens the browser.
- [x] Run `./shortcuts --auto-open=false`; verify the app serves normally without launching a browser.
- [x] Stop the app with `Ctrl-C`; verify shutdown logging is clean and non-noisy.
- [x] Run `./shortcuts --port 8123`; verify the app serves on `http://127.0.0.1:8123`.
- [x] Run `./shortcuts --db /tmp/shortcutdeck-s09.db`; verify the app starts successfully and creates/uses the overridden database path.
- [x] With the overridden DB, create a small deck and one card; restart with the same `--db` value and verify the data persists.
- [x] Trigger a browser-open failure path in a controlled way if practical for the platform; verify the app keeps serving and logs a non-fatal warning.
- [x] Run `make test`; verify the full automated test suite remains green.
- [x] Run `make lint`; verify lint is clean.
- [x] Run `make release`; verify the expected files appear in `dist/`.
- [x] Read through `README.md` from top to bottom as if onboarding fresh; verify it is sufficient to build, run, and use the app without consulting the PRD/DD.

## Prompt

Implement Sprint 9 for the `shortcutdeck` repository.

Context and constraints to restate inline:

- This is the polish-and-ship sprint. Do not add new product features beyond what is required for a releasable v1 runtime, packaging flow, and user-facing documentation.
- Existing product functionality already exists from prior sprints:
  - Sprint 6: deck/card management frontend
  - Sprint 7: study session + Quick Refresher
  - Sprint 8: UI polish and UX cleanup
  - Sprint 9 must preserve those flows rather than redesign them
- Source of truth for Sprint 9 scope is:
  - `docs/sprints/sprint-plan.md` Sprint 9 summary: `feat/s09-ship`, `--port` / `--db`, cross-platform builds, README, `make release`
  - `docs/prd.md` §11 and §12: single-binary distribution, embedded local web server, default port `7432`, `--port` override, browser auto-open defaulting on with a disable flag
  - `docs/dd.md` Section 1 package inventory: `cmd/shortcutdeck` parses CLI flags, resolves DB path, wires dependencies, starts the server, and optionally opens the browser
  - `docs/open-issues.md`: Sprint 9 must absorb the deferred logging work in `cmd/shortcutdeck`
- Keep `internal/model`, `internal/store`, `internal/scheduler`, and the HTTP API contracts stable unless a genuine Sprint 9 blocker is discovered.
- Do not turn Sprint 9 into a frontend redesign or a new feature sprint.

Create or modify exactly these paths as needed for this sprint:

- `cmd/shortcutdeck/main.go`
- `cmd/shortcutdeck/main_test.go`
- `Makefile`
- `README.md`
- `docs/open-issues.md`
- `docs/prd.md`
- `docs/dd.md`
- `docs/sprints/s09-ship.md`

Implementation requirements:

- Add CLI/runtime flag support in `cmd/shortcutdeck/main.go`:
  - `--port` overrides the listen port while preserving `7432` as the default
  - `--db` overrides the SQLite database file path while preserving the current OS-specific default config-dir path as the default
  - `--auto-open` defaults to `true` and allows service/non-interactive starts to suppress browser launch
- Keep browser auto-open enabled by default, but make failure to open the browser non-fatal.
- Refactor `main.go` only as much as needed to make runtime configuration and logging explicit and testable.
- Improve logging in `cmd/shortcutdeck`:
  - startup log should clearly state where the app is listening
  - shutdown log should be explicit
  - browser-open failures should be visible but non-fatal
  - fatal errors should remain actionable
  - avoid scattered package-global logging that makes behavior hard to test
- Add focused tests in `cmd/shortcutdeck/main_test.go` for the extracted runtime/config helpers.
- Update `Makefile` with a `release` target that builds release artifacts into `dist/`.
- The release target must produce the Sprint 9 artifact set:
  - `dist/shortcutdeck-linux-amd64`
  - `dist/shortcutdeck-darwin-amd64`
  - `dist/shortcutdeck-windows-amd64.exe`
- Preserve the existing local developer targets and cache settings in the `Makefile`.
- Expand `README.md` into real end-user documentation that covers:
  - what `shortcutdeck` is
  - build/run instructions
  - runtime flags (`--port`, `--db`, `--auto-open`)
  - where data is stored by default
  - browser-launch behavior
  - keyboard workflow summary
  - export/import notes
  - release build usage or artifact overview
- Keep the README strictly aligned with implemented v1 behavior. Do not document learning-path CRUD, multi-deck study, or other v2 ideas as current features.

Do not do any of the following in this sprint:

- do not add new API routes
- do not redesign the web UI
- do not add learning-path CRUD
- do not add multi-deck “all due cards” study
- do not change the scheduler algorithm
- do not add client-side persistence
- do not replace the embedded-frontend architecture
- do not broaden the sprint into installer packaging, auto-update support, or code signing
- do not rewrite unrelated server/store/frontend files unless a real blocker requires a minimal targeted fix

Stop only when all of the following pass:

- `make build`
- `make test`
- `make lint`
- `make release`

And complete the full `Manual Test Checklist` in this sprint doc before merging the feature branch.

## Out of Scope

- New study features or UI redesign
- Tag-vocabulary enforcement / constrained tag selection if still not complete after Sprint 8
- Sidebar focus/selection polish if still not complete after Sprint 8
- Learning-path CRUD or richer learning-path presentation
- Multi-deck study sessions
- Telemetry, analytics, crash reporting, or remote logging
- Installers, package managers, code signing, notarization, or auto-update flows
- Any API, store, or scheduler redesign not directly required for Sprint 9 shipping work

## Notes

- The current repo already contains a `README.md`, but it is only a placeholder; Sprint 9 should treat README completion as a real deliverable, not a minor cleanup.
- The current runtime entrypoint already logs basic events, but the logging behavior is still bundled into `main.go` and was explicitly deferred in [`docs/open-issues.md`](../open-issues.md) for this sprint.
- The About dialog now exists as a lightweight shipped UI surface. Its user-facing version and short description are sourced from Go (`internal/appinfo`) and injected into the root HTML at serve time so release builds can override version metadata without duplicating strings in the frontend.
- Sprint 9 intentionally follows the UI polish sprint so release acceptance is not blurred by ongoing visual cleanup.
- Keep the release scope pragmatic: reproducible cross-platform binaries via `make release` are in scope; distribution channels and installer ecosystems are not.
