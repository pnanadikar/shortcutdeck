# Sprint 7 — Frontend: Study Session + Browse Mode

## Goal

Build the study-facing frontend for `shortcutdeck`: from the deck list, the user can start an Ongoing Review or Quick Refresher session, move through cards with full keyboard support, submit review grades in Ongoing Review, browse cards without scheduling changes in Quick Refresher, and reach a session summary/exit flow that matches the existing PRD, DD, and design references.

## Inputs

- [`docs/prd.md`](../prd.md) — use the Study Session Flow in §7, the grade mapping in §6, the `review_active` lifecycle, browse-mode semantics, and the Technical Stack / UI principles as product source of truth.
- [`docs/dd.md`](../dd.md) — use the package inventory, model semantics, HTTP API in §6, and key invariants in §7 as the implementation source of truth.
- [`docs/sprints/sprint-plan.md`](./sprint-plan.md) — use Sprint 7 as the source of truth for branch name, deliverable, study flow, browse flow, and keyboard expectations.
- [`docs/sprints/s06-frontend-deck-mgmt.md`](./s06-frontend-deck-mgmt.md) — preserve the established Sprint 6 app shell, deck/card management behavior, keyboard fixes, and frontend file split.
- [`docs/design/shortcutdeck_tree_dropdown.html`](../design/shortcutdeck_tree_dropdown.html) — authoritative study/browse UI reference for deck-entry actions, scope selector, session layout, reveal flow, browse navigation affordances, and status-bar language.
- [`docs/design/shortcutdeck_session_summary.html`](../design/shortcutdeck_session_summary.html) — authoritative summary-screen reference for session completion, next-due breakdown, and exit/back actions.
- [`docs/design/shortcutdeck_deck_cards.html`](../design/shortcutdeck_deck_cards.html) — supporting reference for keeping the established shell, typography, spacing, and deck-list visual language consistent with Sprint 6.
- [`web/index.html`](../../web/index.html) — extend the current Sprint 6 shell rather than replacing it.
- [`web/app.css`](../../web/app.css) — extend the existing style system and components rather than starting over.
- [`web/app.js`](../../web/app.js) — add the Sprint 7 session/browse state machine on top of the existing Sprint 6 deck/card management frontend.
- [`internal/server/server.go`](../../internal/server/server.go) and existing handler files — use the already-implemented Sprint 5 API routes as-is; only change Go code if a genuine blocking bug is found.

## Outputs

- `web/index.html`
- `web/app.css`
- `web/app.js`
- `docs/sprints/s07-study.md`

The core sprint output is frontend-only. `internal/server` and `internal/store` must remain unchanged unless a real backend bug is discovered while integrating the study UI. If that happens, change only the minimum required backend file(s) and document the reason in this sprint doc.

## Acceptance Criteria

- The frontend adds deck-level study actions that visually follow [`docs/design/shortcutdeck_tree_dropdown.html`](../design/shortcutdeck_tree_dropdown.html):
  - scope selector / filter affordance on deck entries
  - `Quick Refresher` action
  - `Start Review` or `Ongoing Review →` action depending on `review_active`
  - due-count visibility only when `review_active` is true
- Deck list study affordances use real backend state from the existing API; no mock due counts and no hardcoded review state.
- Ongoing Review flow is fully functional from the UI:
  - `POST /api/v1/decks/:id/start-review` is called when starting review from an inactive deck
  - `GET /api/v1/decks/:id/due` is used to load due cards, with optional `?tags=...` filter based on the selected scope
  - `POST /api/v1/cards/:id/review` is called for each review rating
  - review grading matches the PRD/DD four-grade mapping exactly:
    - Missed it + unchecked effort → `0`
    - Missed it + checked effort → `1`
    - Got it + checked effort → `3`
    - Got it + unchecked effort → `5`
  - cards advance until the due list is exhausted or the user exits early
- Quick Refresher / browse flow is fully functional from the UI:
  - cards are loaded from `GET /api/v1/decks/:id/cards`, with optional `?tags=...` filter
  - no review grade is submitted
  - no scheduling state is changed
  - navigation works card-to-card until the user exits
- Reverse-mode handling is implemented for sessions:
  - session options are pre-filled from the deck’s `default_reverse_mode`
  - user may override per session
  - `prompt_first` and `answer_first` show the correct initial side
  - `both` presents each card twice in two passes: first all cards `prompt_first`, then all cards `answer_first`
- Session summary screen is implemented and visually follows [`docs/design/shortcutdeck_session_summary.html`](../design/shortcutdeck_session_summary.html):
  - cards reviewed count
  - Got it count
  - Missed it count
  - next-due breakdown based on the updated deck export / card schedule data
  - `Exit Review` action when applicable
  - `Back to decks` action
- Exiting Ongoing Review from the summary screen or explicit session exit uses `POST /api/v1/decks/:id/exit-review`.
- Exiting Quick Refresher never calls `start-review`, `review`, or `exit-review`.
- Existing Sprint 6 deck/card management views still work unchanged after adding Sprint 7 views.
- Keyboard behavior for Sprint 7 is implemented:
  - `Space` / `Enter` reveals the answer during a session
  - `e` edits the current card from a session or browse view
  - `Esc` exits the active session/view according to the current mode
  - browse navigation supports `j` / `k` and `↑` / `↓`
  - keyboard shortcuts do not interfere while typing into form inputs unless explicitly part of the input interaction
- Status-bar text is updated for session start, reveal/rate steps, browse mode, summary completion, and exit flows; errors remain visible and are not swallowed.
- `make build` succeeds.
- `make lint` is clean.
- Existing Sprint 6 checks continue to pass.
- Manual verification succeeds:
  - start Ongoing Review for the Sprint 6 Vim deck and complete a full session
  - verify review ratings affect later due scheduling via export or reloaded UI state
  - run Quick Refresher on the same deck and confirm no scheduling changes occur
  - verify `both` reverse mode runs in two passes: all prompt-first cards, then all answer-first cards
  - verify session summary and exit/back actions behave correctly

## Manual Test Checklist

This checklist is the stopping condition for Sprint 7. Failed or blocked items should be recorded in [`docs/open-issues.md`](../open-issues.md), not inline in this sprint doc.

- [x] Run `make run`; verify the Sprint 6 deck list still renders and no existing deck/card management regressions appear.
- [x] Open a deck with `review_active = false`; verify the deck row shows `Start Review` and no due count.
- [x] Start Ongoing Review for the deck; verify the button state changes to `Ongoing Review →` and due counts become visible.
- [x] Start a review session with default reverse mode and no filter; reveal and rate at least 5 cards using keyboard only.
- [x] Start a session with reverse mode override `answer_first`; verify the answer side is shown first.
- [x] Start a session with reverse mode override `both`; verify the session shows all cards prompt-first, then repeats them answer-first.
- [x] Start a session with a tag scope selected; verify only scoped due cards are shown.
- [x] Press `Esc` during Ongoing Review; verify the intended exit flow and summary behavior.
- [x] Complete a full Ongoing Review session; verify the summary screen shows reviewed / got-it / missed-it counts and next-due breakdown.
- [x] Use `Exit Review` from the summary screen; verify the deck returns to inactive review state.
- [x] Start Quick Refresher for the same deck; verify cards load, `j` / `k` and `↑` / `↓` navigate, and no review grade is submitted.
- [x] Export the deck before and after Quick Refresher; verify scheduling fields are unchanged by Quick Refresher.
- [x] Export the deck after Ongoing Review; verify scheduling fields changed as expected.
- [ ] Verify dark mode still renders correctly across deck list, session view, browse view, and summary view.

## Prompt

Implement Sprint 7 for the `shortcutdeck` repository.

Context and constraints to restate inline:

- This is the frontend study-session sprint. Build the study and browse UI on top of the existing Sprint 6 frontend. Do not redesign deck/card management.
- Source of truth for behavior is:
  - `docs/prd.md` §7 for study lifecycle, browse semantics, reverse modes, and summary expectations
  - `docs/dd.md` §6 for the existing study API routes
  - `docs/dd.md` §7 for invariants such as `review_active`, due-card visibility, and exit-review reset behavior
- Source of truth for visuals is:
  - `docs/design/shortcutdeck_tree_dropdown.html` for the deck-entry study actions, scope selector, session card layout, reveal/rate flow, and browse-mode feel
  - `docs/design/shortcutdeck_session_summary.html` for the summary screen
  - `docs/design/shortcutdeck_deck_cards.html` as supporting reference for preserving the Sprint 6 shell language
- Preserve the current frontend file split exactly:
  - `web/index.html` for structure
  - `web/app.css` for styles
  - `web/app.js` for state, rendering, API calls, and keyboard handling
- Use plain JavaScript only. No frameworks, no bundlers, no generated assets, no localStorage/sessionStorage.
- Existing backend API routes already exist from Sprint 5 and must be reused as-is:
  - `GET /api/v1/decks/:id/due`
  - `POST /api/v1/cards/:id/review`
  - `POST /api/v1/decks/:id/start-review`
  - `POST /api/v1/decks/:id/exit-review`
  - `GET /api/v1/decks/:id/cards`
  - existing deck/card/export/import endpoints as needed for deck list state refresh
- Do not add new API endpoints for Sprint 7.
- Do not change `internal/store`, `internal/scheduler`, or server contracts unless a genuine blocking bug is discovered.
- Keep Sprint 6 functionality intact. The deck/card CRUD UI and its keyboard behavior must continue to work after Sprint 7 changes.

Create or modify exactly these paths as needed for this sprint:

- `web/index.html`
- `web/app.css`
- `web/app.js`
- `docs/sprints/s07-study.md`

Implementation requirements:

- Add deck-entry study controls to the existing deck list UI:
  - scope selector matching the study design reference
  - `Quick Refresher`
  - `Start Review` when `review_active = false`
  - `Ongoing Review →` when `review_active = true`
  - due counts only when `review_active = true`
- Implement a session-options/start flow that allows:
  - using the deck’s `default_reverse_mode` as the preselected reverse mode
  - overriding it per session
  - optional scope/tag filtering from the selected namespace value
- Implement Ongoing Review session flow:
  - if the deck is inactive, call `POST /api/v1/decks/:id/start-review` first
  - load due cards via `GET /api/v1/decks/:id/due`, applying tag filter when selected
  - present cards one by one
  - on reveal, show the effort checkbox plus `Got it` / `Missed it`
  - map button + checkbox state to grades `0`, `1`, `3`, `5` exactly
  - submit grades via `POST /api/v1/cards/:id/review`
  - advance until no due cards remain or the user exits
- Implement Quick Refresher / browse mode:
  - load cards with `GET /api/v1/decks/:id/cards`
  - apply selected tag filter when present
  - no rating UI
  - no scheduling changes
  - navigate freely through cards
- Implement reverse-mode behavior:
  - `prompt_first`
  - `answer_first`
  - `both` as a two-pass presentation: one prompt-first pass, then one answer-first pass
- Implement the session summary view:
  - reviewed count
  - Got it count
  - Missed it count
  - next-due breakdown derived from the updated deck/card schedule state
  - `Exit Review` action for Ongoing Review
  - `Back to decks`
- Implement exit behavior:
  - `Esc` during Ongoing Review should leave the session in the intended Sprint 7 flow and allow the user to reach the summary / exit path
  - `Exit Review` must call `POST /api/v1/decks/:id/exit-review`
  - Quick Refresher exit must not call review start/review/exit endpoints
- Implement Sprint 7 keyboard behavior:
  - `Space` / `Enter` reveal answer
  - `e` edits current card
  - `Esc` exits current mode
  - browse supports `j` / `k` and `↑` / `↓`
  - shortcuts must not interfere with typing in form inputs
- Keep status-bar messaging explicit across:
  - deck ready-to-study state
  - session start
  - reveal prompt
  - rating step
  - browse navigation
  - summary completion
  - errors

Do not do any of the following in this sprint:

- do not redesign the Sprint 6 deck/card management layout
- do not add learning-path CRUD
- do not add “All decks” combined review
- do not add new backend routes
- do not change the scheduler algorithm
- do not add persistent client-side storage
- do not replace browser-native confirm dialogs unless that is already required by the existing frontend architecture
- do not fix unrelated Sprint 6 polish issues unless they directly block the study flow

Stop only when all of the following pass:

- `make build`
- `make lint`
- manual verification using the full Sprint 7 checklist in this doc

## Out of Scope

- Learning-path management UI
- Multi-deck “all due cards” sessions
- New backend endpoints or API contract changes
- Stats/history dashboards beyond the single-session summary screen
- Reworking Sprint 6 deck/card management beyond integration needed for study actions
- Polishing unrelated sidebar selection/focus issues such as [`B18`](../open-issues.md)

## Notes

- Sprint 7 should build directly on the stable Sprint 6 frontend rather than introducing a new shell.
- `review_active` semantics are already enforced by the backend and must remain the single source of truth for review button state and due-count visibility.
- Quick Refresher is a UI mode only; it must never mutate scheduling state.
- If a backend bug is discovered while implementing the study UI, fix the minimum required backend behavior on the Sprint 7 branch and record the reason here.
