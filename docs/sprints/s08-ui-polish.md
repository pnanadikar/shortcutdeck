# Sprint 8 — UI Polish + UX Cleanup

## Goal

Polish the Sprint 6 and Sprint 7 frontend so the app feels visually cohesive, keyboard-friendly, and ready for day-to-day use before the final ship/release sprint.

## Inputs

- [`docs/sprints/sprint-plan.md`](./sprint-plan.md) — update Sprint 8 to be the UI polish sprint and use it as the source of truth for branch naming and deliverable framing.
- [`docs/prd.md`](../prd.md) — use the desktop-first, keyboard-first, minimal-noise UX principles and the existing keyboard behavior requirements as product constraints.
- [`docs/dd.md`](../dd.md) — preserve the existing API contracts, package boundaries, and runtime architecture; this sprint is frontend polish, not a backend redesign.
- [`docs/open-issues.md`](../open-issues.md) — use the active frontend bugs and UX questions to guide polish priorities, especially:
  - `B18` sidebar focus/selection double-highlight
  - `B20` constrained tag behavior gap when decks define tag namespaces
  - `B21` default landing view should likely be `Study`, not `Create`
  - `B22` study-page framing/layout has drifted from the design reference
  - `B23` study search should likely support `/` as a keyboard shortcut
  - `B24` dark theme needs more tonal variation and hierarchy
- [`docs/sprints/s06-frontend-deck-mgmt.md`](./s06-frontend-deck-mgmt.md) — preserve established deck/card management behavior while improving its visual and interaction quality.
- [`docs/sprints/s07-study.md`](./s07-study.md) — preserve the study/browse/session flows while improving consistency, clarity, and polish.
- [`docs/design/shortcutdeck_deck_cards.html`](../design/shortcutdeck_deck_cards.html) — primary visual reference for the deck/card management shell and component language.
- [`docs/design/shortcutdeck_tree_dropdown.html`](../design/shortcutdeck_tree_dropdown.html) — primary visual reference for the study/browse deck actions and session layout.
- [`docs/design/shortcutdeck_session_summary.html`](../design/shortcutdeck_session_summary.html) — primary visual reference for the summary screen and session wrap-up visual language.
- [`web/index.html`](../../web/index.html)
- [`web/app.css`](../../web/app.css)
- [`web/app.js`](../../web/app.js)

## Outputs

- `web/index.html`
- `web/app.css`
- `web/app.js`
- `docs/open-issues.md`
- `docs/sprints/s08-ui-polish.md`

Change Go/backend files only if a genuine blocker prevents a necessary polish fix. This sprint should stay frontend-focused by default.

## Acceptance Criteria

- The frontend remains functionally equivalent to the end of Sprint 7:
  - deck CRUD still works
  - card CRUD still works
  - import/export still works
  - Ongoing Review still works
  - Quick Refresher still works
- Visual polish is applied consistently across deck management, study, browse, and summary views:
  - spacing, typography, and component density feel intentional and consistent
  - selected, focused, hovered, and disabled states are visually distinct and non-conflicting
  - status-bar messaging remains readable and visually consistent in all app modes
  - dark mode remains supported and visually coherent
  - theme-related styles are routed through semantic CSS variables so the frontend is easier to retheme later without a large CSS rewrite
- Sidebar and navigation state clarity are improved:
  - only one item appears selected at a time
  - keyboard focus is visually distinct from selected/default action state
  - `B18` is fixed or deliberately narrowed with a documented follow-up if a partial fix is the safest sprint outcome
- Study-page information architecture is cleaned up:
  - app startup/default landing state is intentionally chosen and no longer feels accidental
  - `Study` is the default landing view for v1 polish
  - sidebar/default selection and initial content agree with the study-first startup model
  - when decks exist, the first deck is preselected on the study deck list rather than auto-focusing the search field
  - when no decks exist, the empty state clearly guides the user toward `Create…`
  - breadcrumb treatment is simplified to better match the design reference and avoid redundant page framing
  - the study deck-list screen no longer relies on an oversized `Study` heading if that heading conflicts with the design language
  - deck descriptions remain visible in the study deck list even while the rest of the layout is realigned to the design
- Tag editing UX is improved where deck namespaces exist:
  - card editing no longer silently allows tag values that violate a deck’s declared tag vocabulary
  - if full constrained selection is implemented, it must use clear UI affordances
  - if only validation is implemented in this sprint, invalid tags must be blocked with explicit user feedback
  - `B20` is fixed or reduced to a clearly documented remainder in [`docs/open-issues.md`](../open-issues.md)
- Session and browse interactions are clearer and more polished:
  - reveal state vs rating state is visually obvious
  - current card position/progress remains easy to scan
  - browse navigation affordances and keyboard hints are clear
  - summary metrics are easier to read at a glance
- Study deck-list search becomes more keyboard-friendly:
  - `/` focuses the search field when the user is on the study deck list and not currently typing in another input or textarea
  - search is available on demand rather than being the default focus target on page load
  - the status/help text makes the shortcut discoverable without adding visual noise
- The frontend does not regress keyboard-first usage:
  - Sprint 6 and Sprint 7 keyboard behavior still works
  - polish changes do not trap focus or hide focus indication
  - shortcuts still do not interfere while typing into form controls
  - selected/default state and live keyboard focus are visually distinct without producing a confusing double-highlight effect
- Any active polish bugs found during manual verification are recorded in [`docs/open-issues.md`](../open-issues.md).
- `make build` succeeds.
- `make lint` is clean.
- Manual verification succeeds across deck management and study flows.

## Manual Test Checklist

This checklist is the stopping condition for Sprint 8. Failed or blocked items should be recorded in [`docs/open-issues.md`](../open-issues.md), not inline in this sprint doc.

- [ ] Run `make run`; verify the app still loads cleanly with the polished UI.
- [ ] Verify the default landing view is `Study`, not an obviously secondary action such as `Create`.
- [ ] With one or more decks present, verify the first deck is preselected on the study deck list and the search field is not auto-focused on load.
- [ ] Verify the left sidebar no longer shows conflicting selected/focused highlight states while tabbing and arrow-keying through actions.
- [ ] Open the study deck list; verify the layout aligns more closely with the design reference, does not rely on a redundant oversized `Study` heading, and still preserves deck descriptions.
- [ ] On the study deck list, press `/`; verify search focus moves to the deck search field on demand without interfering when another text input is already active.
- [ ] Create a deck and edit deck metadata; verify visual spacing, button states, and form focus treatment feel consistent.
- [ ] Open card editing for a deck; verify card-list selection, hover, and keyboard focus states remain clear and non-conflicting.
- [ ] In a deck with tag namespaces defined, create or edit a card; verify invalid free-form tags are prevented or clearly validated.
- [ ] Start Ongoing Review; verify deck-entry actions, session card layout, reveal state, rating state, and status bar all feel visually cohesive.
- [ ] Complete a review session; verify the summary screen is readable and polished in both layout and emphasis.
- [ ] Start Quick Refresher; verify browse navigation affordances and current-card context remain clear with both mouse and keyboard use.
- [ ] Repeat key flows in dark mode; verify contrast and component states remain readable.
- [ ] Run `make build`; verify the app still builds successfully.
- [ ] Run `make lint`; verify lint remains clean.

## Prompt

Implement Sprint 8 for the `shortcutdeck` repository.

Context and constraints to restate inline:

- This is the UI polish and UX cleanup sprint. The product features from Sprint 6 and Sprint 7 already exist; improve quality, consistency, and usability without turning this into a new feature branch.
- Preserve the existing app architecture:
  - `web/` owns the frontend
  - `internal/server` serves the frontend and existing API routes
  - `internal/store` and `internal/scheduler` are not to be redesigned for polish work
- Source of truth for behavior remains:
  - `docs/prd.md` for desktop-first and keyboard-first UX requirements
  - `docs/dd.md` for API and architectural invariants
  - `docs/sprints/s06-frontend-deck-mgmt.md` for deck/card management behavior
  - `docs/sprints/s07-study.md` for study/browse/session behavior
- Source of truth for visuals remains:
  - `docs/design/shortcutdeck_deck_cards.html`
  - `docs/design/shortcutdeck_tree_dropdown.html`
  - `docs/design/shortcutdeck_session_summary.html`
- Active polish/UX issues to address include:
  - `B18` sidebar focus/selection double-highlight
  - `B20` invalid card tags when a deck defines tag namespaces
  - `B21` default landing view / study-first startup behavior
  - `B22` study-page framing drift from the design reference
  - `B23` `/` shortcut for study deck search
  - `B24` dark-theme monotony / lack of tonal variation

Create or modify exactly these paths as needed for this sprint:

- `web/index.html`
- `web/app.css`
- `web/app.js`
- `docs/open-issues.md`
- `docs/sprints/s08-ui-polish.md`

Implementation requirements:

- Improve the visual consistency of the existing frontend across:
  - deck list
  - deck form
  - card editor
  - study session
  - browse mode
  - session summary
- Keep the current shell and information architecture intact unless a small targeted cleanup materially improves clarity.
- Revisit the app's default landing state:
  - make `Study` the default landing view for v1 polish
  - when decks exist, preselect the first deck row rather than auto-focusing search
  - when no decks exist, use an empty-state path that points clearly to `Create…`
  - ensure sidebar state, breadcrumbs, and initial content all agree with the study-first startup model
- Realign the study deck-list screen to the design reference more closely:
  - avoid an unnecessary large `Study` page title if the mockup does not use one
  - keep the breadcrumb minimal and non-redundant
  - preserve deck descriptions even though the mockup does not currently show them
- Ensure selected, focused, hovered, and active states are visually distinct everywhere.
- Fix or materially improve the sidebar selection/focus problem described in `B18`.
- Fix or materially improve the tag-vocabulary problem described in `B20`:
  - when a deck defines tag namespaces / allowed values, card tags should not silently drift outside that vocabulary
  - use either constrained selection or explicit validation with clear user feedback
  - do not invent a large new metadata-management flow in this sprint
- Add a keyboard shortcut on the study deck list so `/` focuses the search field when the user is not already typing in an input.
- Keep search as an explicit jump action, not the default initial focus target, since the intended primary flow is deck selection and session start.
- Improve readability and polish for study/browse screens:
  - reveal vs rate states should be immediately understandable
  - progress and navigation context should remain easy to scan
  - summary metrics should be clearer and better grouped
- Improve the dark theme so it has more tonal separation and hierarchy:
  - avoid a flat, same-value dark wash across all surfaces
  - use restrained but noticeable variation between shell, panels, cards, controls, and highlights
  - keep the tone calm and desktop-focused rather than neon or high-saturation
- Make the frontend theme-ready without shipping full theme configurability:
  - centralize colors and key visual-state values behind semantic CSS variables
  - avoid scattering raw hex values through component-level rules where semantic tokens would be clearer
  - do not add user-editable theme config in v1; this sprint should prepare for that possibility rather than implement it
- Be careful with selection/focus styling:
  - a default-preselected deck should read as the current selection
  - live keyboard focus should still be visible when it moves elsewhere
  - the two states must not collapse into an ambiguous double-highlight

Suggested minimum semantic token set for v1 polish:

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

If the CSS naturally needs a second pass, add only narrowly justified tokens rather than expanding the token set preemptively.
- Preserve all existing Sprint 6 and Sprint 7 functionality and keyboard interactions.
- Record any remaining polish bugs or intentional deferrals in `docs/open-issues.md`.

Do not do any of the following in this sprint:

- do not add new API routes
- do not redesign the entire application layout
- do not add learning-path CRUD
- do not add multi-deck study
- do not change the scheduler algorithm
- do not move ship/release work into this sprint
- do not add localStorage/sessionStorage
- do not replace the plain HTML/CSS/JS frontend architecture

Stop only when all of the following pass:

- `make build`
- `make lint`

And complete the full `Manual Test Checklist` in this sprint doc before merging the feature branch.

## Out of Scope

- CLI/runtime flags such as `--port` and `--db`
- `cmd/shortcutdeck` logging cleanup
- `make release` and release packaging
- README/release onboarding work
- New backend routes or store/schema changes unless a true blocker forces a minimal fix
- New product features beyond polish-level UX improvements

## Notes

- Sprint 8 intentionally absorbs frontend polish that would otherwise make the ship sprint vague and unstable.
- If the constrained-tagging work for `B20` turns out to be too large for a polish sprint, prefer the smallest sound fix that preserves deck tag-vocabulary integrity and document the remainder.
- Keep this sprint disciplined: improve clarity and finish quality, but avoid reopening product decisions that belong in the PRD or a later feature sprint.
- Theme configurability itself is deferred, but Sprint 8 should leave the frontend in a state where a future manual theme config file can map onto stable semantic tokens rather than requiring a full visual-system rewrite.
