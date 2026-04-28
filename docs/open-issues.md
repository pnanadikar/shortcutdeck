# Open Issues

## Guiding Principles (stable)

- No vibe coding. Break work into small sprints. Use git feature branches. Make code testable.
- Prefer established open-source libraries over custom implementations, except where the spec is trivial (SM-2).
- Internal packages must be UI-agnostic — no HTTP or terminal concepts in `internal/store`, `internal/sm2`, or `internal/model`.

---

## Open Questions

### Architecture

### Product / UX

1. `[ ]` **User-facing term for `tag namespace`:** `Tag namespace` is acceptable in PRD/DD, but may be too implementation-flavored for the user-facing UI. Change request: choose clearer UI wording before the constrained-tagging work lands in the card/deck editor.
2. `[ ]` **Editing a card from an active session/browse flow:** Possible new feature - Should the user be allowed to edit from Quick Refresher or Ongoing Review? If yes, should the app open the current card in the card editor and then return to the same session context? (This behavior is not currently specified in the PRD.)
3. `[ ]` **Reveal-answer control affordance:** The current session UI shows a passive `Space to reveal answer` hint below the card. Should v1 instead show an explicit focused `Reveal the answer` button while keeping `Space` as the keyboard shortcut that activates it? If so, decide whether this applies to both Quick Refresher and Ongoing Review and whether the status-bar copy should stay unchanged.
4. `[ ]` **Quick Refresher skip/jump affordance:** Quick Refresher currently supports sequential navigation only. Should it gain an explicit skip/jump control (for example a dropdown or jump-to-card affordance), or is linear keyboard navigation the intended v1 behavior?
5. `[ ]` **Dark theme visual direction:** Current dark mode is usable but visually monotonous. Sprint 8 should choose a clearer dark-theme direction with more tonal separation and component hierarchy while staying calm/minimal rather than flashy.
6. `[ ]` **Configurable theming:** Likely a v2 feature, not a v1 ship requirement. Preferred direction: make the v1 frontend theme-ready by expressing visual styling through stable semantic tokens/CSS variables, then add manual user-editable theme configuration later once the token schema is proven and documented.

---

## Resolved (for reference)

| Issue | Decision | PRD ref |
| --- | --- | --- |
| Runtime | **Single binary vs two binaries (Option C vs D):** Deferred to v2. v1 ships the web frontend only. Decision: Decide if the TUI ships as a separate `shortcutdeck-tui` binary (Option C) or as `shortcutdeck --tui` (Option D). (Both share 100% of `internal/`; difference is ~20 lines of entry-point code.) in v2 | §3 |
| Runtime | Go CLI + embedded web server (`net/http`) | §10 |
| Storage | SQLite via `modernc/sqlite` (pure Go, no CGo) | §10 |
| SM-2 implementation | Implement directly in Go (~40 lines), fully unit-tested | §6 |
| UI framework (v1) | Single embedded HTML/JS/CSS file via `embed.FS` | §10 |
| Export/Import | JSON via browser file download / `<input type="file">`; Import is replace-only in v1; merge deferred | §8 |
| Browse/review mode | Supported as a separate non-SM-2 mode (Quick Refresher) | §7 |
| Keyboard navigation | Space/Enter to reveal; Got it / Missed it buttons; Needs effort checkbox; Esc to quit | §7 |
| Reverse cards | Deck-level setting: `prompt_first` / `answer_first` / `both` | §5 |
| `review_log` scope | Full history retained, not just current scheduling state | §8 |
| TUI frontend | Out of scope for v1; web frontend first; TUI targets study session + browse mode in v2 | — |
| AI coding approach | Hybrid: detailed design via chat, code generation in VS Code | — |
| Runtime port | Fixed default `7432`; overridable via `--port` flag | §12 |
| Browser auto-open | v1 keeps browser launch enabled by default, but runtime supports `--auto-open=false` for service-style or non-interactive starts. | §12 |
| Runtime logging | `cmd/shortcutdeck` owns startup, shutdown, fatal runtime, browser-open warning, and server-level operational logging; internal packages remain UI-agnostic and avoid terminal/runtime concepts. | design |
| SM-2 vs FSRS-5 | SM-2; implement directly in Go (~40 lines, fully unit-tested) | §12 |
| `reverse_mode: both` behaviour | Each card shown twice per session, interleaved | §12 |
| Frontend framework | Plain HTML/JS/CSS, single embedded file, no build step | §10 |
| Rating scale / grade mapping | Two buttons (Got it / Missed it) + "Needs effort" checkbox → 4 SM-2 grades: 0 (Again), 1 (Partial), 3 (Hard), 5 (Easy) | §6 |
| `review_active` lifecycle | Deck-level boolean; false=never started or exited; true=active. Start Review is always a cold start. Exit Review resets scheduling state, retains review_log | §7 |
| Sidebar navigation | Study… (top) · gap · Create… / Import… / Export… (grouped) · Delete… | design |
| Import mode | Replace-only in v1; merge deferred — users manage merge externally via JSON | §8 |
| Import confirmation | Parse file client-side; show deck name, card count, namespaces+tags summary, replace warning before committing | §8 |
| Export placement | Sidebar (Import… and Export… grouped together as infrequent actions) | design |
| Default landing view / Study IA | v1 is study-first: when decks exist, startup opens on the study deck list; creation/import/export remain secondary actions. When no decks exist, startup opens directly to deck creation. | — |
| Study breadcrumbs / page framing | Keep study framing minimal and close to the mockup: no oversized `Study` heading; breadcrumb stays lightweight and only adds context when needed. | design |
| Deck-list study screen fidelity | Realign the study deck-list layout and hierarchy toward [`docs/design/shortcutdeck_tree_dropdown.html`](./design/shortcutdeck_tree_dropdown.html) while explicitly retaining deck descriptions. | design |
| Study deck-list quick search shortcut | `/` should focus the study search field on demand; search is an explicit jump action rather than the default initial focus target. | design |
| Dirty-state guard coverage for deck/card forms | Prompt only when an action would discard or overwrite unsaved deck/card edits, including record-switch/repopulate transitions; do not prompt for harmless focus movement or normal study navigation. | — |
| Import from editing context | `Import...` should respect dirty-state checks and should not remain nested inside deck/card editing after a real import attempt begins. | — |
| Transient import status message reset | Import picker guidance is transient; once the picker flow completes or is cancelled, the status bar returns to the steady-state message for the active view and result feedback uses toast. | — |
| Card-tag structure | Card tags are no longer treated as truly free-form. Deck metadata owns the namespace → allowed values mapping; card editing uses constrained selection only. Decks without explicit namespaces should use a default namespace rather than unscoped tags. | — |
| Session header metadata hierarchy | In Quick Refresher and Ongoing Review, the deck name is the primary session heading; the session type is secondary context, with scoped Quick Refresher filters shown alongside that secondary context. | design |
| About menu item and version surface | v1 should include a working `About...` menu item with the app version number and a short product description as the minimum version surface. | — |
| Default card ordering within a deck | Cards should use the app's existing default order rather than introducing a new sort rule or explicit ordering model for v1. | — |

## Bugs

### Manual Test Findings

#### Active

| ID | Sprint found | Description | Status |
| ---- | ------------- | ------------- | -------- |
| B24 | S08 planning | Dark theme currently lacks enough tonal variation, making the UI feel flatter and more monotonous than intended. | Open. Improve dark-mode hierarchy, contrast grouping, and accent usage in Sprint 8. |

#### Resolved

| ID | Sprint found | Sprint fixed | Description | Fix |
| ---- | ------------- | ------------- | ------------- | ----- |
| B01 | S06 | S06 | Export... menu option missing in the sidebar. | Added the missing `Export...` sidebar action and matched the intended sidebar grouping/spacing. |
| B02 | S06 | S06 | After creating a deck, clicking the deck row on the left does not populate the form on the right | Updated deck-row selection so clicking a deck populates the deck metadata form for editing. |
| B03 | S06 | S06 | `Ctrl-C` in the VS Code integrated terminal does not stop `make run` or the built binary cleanly. | Added explicit signal handling, graceful HTTP server shutdown, and a clear `shutting down...` terminal message. Normal terminal behavior is now correct. |
| B04 | S06 | S06 | Clicking `Edit cards` in a deck row does not transition cleanly into the card editor flow. | Removed the deck-row `focus` rerender path so Safari no longer tears down the row before the button click is delivered. |
| B05 | S06 | S06 | Card editor has no visible back action other than breadcrumbs. | Added a visible `← Back` control in the card editor header, wired to the existing `showDeckView()` navigation flow. |
| B06 | S06 | S06 | Status/tag pills have poor contrast in dark theme. | Updated pill and tag colors to use the new readable blue palette in the app CSS and synced the design mockup to the same dark-theme-friendly treatment. |
| B07 | S06 | S06 | In card editor, pressing `Tab` after entering one or more tags did not land on `Add card`, and creating a card from the keyboard did not return focus to the first field for the next entry. | Intercepted forward `Tab` in the tag input to commit any pending tag and focus `Add card`, then returned focus to the `Prompt` field after successful card creation. |
| B08 | S06 | S06 | After reopening the app, the Vim deck showed `0 cards` in the deck row until `Edit cards` was opened, even though the cards were present and loaded correctly once the card editor fetch ran. | Updated the initial deck load to fetch each deck's cards up front, so deck-row counts and overall shortcut stats are correct immediately after startup. |
| B09 | S06 | S06 | After clicking `Edit cards` on the Vim deck and seeing the card list, clicking a card row did not populate the card editor on the right or switch the action button into edit mode. | Changed single-click and Enter on card rows to enter edit mode, and removed the card-row focus rerender path that could interrupt click handling before the form was populated. |
| B10 | S06 | S06 | In the card editor list, clicking a card row without moving the mouse away did not show the selected highlight immediately, and hovering the selected row could hide the highlight again. | Kept the selected card border color on hover so the visual selection state remains stable under the mouse pointer. |
| B11 | S06 | S06 | After clicking a card row to edit, focus landed on `Save changes` instead of the first editable field. | Moved focus to the `Prompt` field after click-to-edit populates the card form, making immediate keyboard editing the default. |
| B12 | S06 | S06 | The `← Back` control in the card editor header returned to deck view immediately and did not check whether the card form had unsaved changes. | Added a card-form dirty-state guard to the back button so it prompts before leaving when there are unsaved edits. |
| B13 | S06 | S06 | The card-editor unsaved-changes check needed to ignore trailing whitespace in text fields while still treating leading whitespace as a real change. | Normalized dirty-state comparison with trailing-whitespace-only trimming and a stable form baseline snapshot so the back-button guard matches expected editing semantics. |
| B14 | S06 | S06 | The card editor had no visible `Delete card` action in the form actions row while editing an existing card. | Added a conditional `Delete card` button to the form actions row, shown only in edit mode, and updated the action labels/layout so add and edit states read clearly. |
| B15 | S06 | S06 | After deleting a card from the card editor, focus/selection was left hanging instead of moving to the next sensible editing target. | Updated card deletion to select the previous card when possible, otherwise the next card, and fall back to the empty add-card form only when no cards remain. |
| B16 | S06 | S06 | `Esc` did not yet follow the PRD keyboard rule for Sprint 6 forms: unchanged forms should cancel immediately, while dirty deck/card forms should prompt before discarding. | Added stable deck/card form baselines plus dirty-aware `Esc` handling so clean forms cancel immediately and dirty forms confirm before discarding. |
| B17 | S06 | S06 | The deck list and card list did not yet support `↑` / `↓` keyboard traversal between rows, despite the PRD keyboard interaction now calling that out for Sprint 6 deck/card management. | Added explicit arrow-key focus movement for deck and card rows, preserving row selection while moving through the list from the keyboard. |
| B18 | S06 | S07 | When keyboard tabbing reaches the left sidebar actions, the currently focused action can be highlighted at the same time as the preselected/default action (for example `Create`), resulting in two highlighted menu items. The double-highlight is an app issue; Safari's Advanced → "Press Tab to highlight each item on a webpage" setting only affects whether sidebar tabbing is exposed during testing. | Open. Needs a single clear visual distinction between focus and selected/default sidebar action state. |
| B19 | S06 | S06 | Importing an exported deck could fail in the frontend with an undefined-object error because the client expected `response.Deck` while the API returned the JSON-tagged `deck` payload. | Updated the import flow to read the actual response shape (`deck`), keep a defensive fallback for `Deck`, and surface a clear error if the imported deck is missing. |
| B20 | S07 | S08 | Card creation/editing currently allows arbitrary free-form tags even when the deck defines tag namespaces. This lets cards be saved with tags that do not belong to any declared namespace value, which breaks the deck's tag vocabulary contract and can lead to inconsistent Study/Review filtering. | Card editing now validates or constrains tags against the deck's declared namespace values so invalid free-form tags are blocked with clear feedback. |
| B21 | S08 planning | S08 | On app startup, the default primary view is `Create`, but the product flow appears study-first and the study deck list is a more natural home screen. | Startup now uses the study-first flow: existing decks open on the study deck list, while the no-deck empty state guides users toward `Create...`. |
| B22 | S08 planning | S08 | The current study page framing has drifted from the design: redundant breadcrumb/title treatment (`All Decks > Study`, large `Study` heading) and overall deck-list presentation no longer match the intended mockup hierarchy. | The study deck-list framing was realigned toward the design reference, removing redundant page-title treatment while preserving deck descriptions. |
| B23 | S08 planning | S08 | The study deck-list search affordance is missing the `/` keyboard shortcut hinted at in the design footer, reducing keyboard-first discoverability. | Added `/` as an on-demand shortcut to focus the study search field when the user is not already typing in another control. |
| B25 | S08 | S08 | Dirty-state protection was incomplete when switching selected decks/cards or otherwise repopulating an edit form from another record, risking silent loss of unsaved deck/card edits. | Added dirty-state guards for deck/card record-switch transitions, top-level workspace exits, and browser/tab close, with targeted focus restoration after canceled prompts. |
| B26 | S08 planning | S08 | `Import...` is currently available from editing contexts, which makes the flow semantically muddy and risks conflicting with dirty form state. | `Import...` now respects dirty-state checks and normalizes nested editing launches out to a top-level context once a real import begins. |
| B27 | S08 planning | S08 | The import status-bar message could read like a sticky state if it was not restored after the native file-picker flow ended. | Treated the import prompt as transient UI guidance, restored the mode-appropriate steady-state status after picker cancel or import completion, and kept success/failure feedback in toast. |
| B28 | S08 | S08 | Starting a partial new-deck entry and then clicking an existing deck could trigger the dirty-state confirmation, but pressing `Cancel` could leave focus on the attempted deck row and retrigger the prompt in a loop. | Added blocked-selection focus restoration so cancel returns focus to the active deck form when there is no prior selected deck row to restore. |
| B29 | S08 | S08 | While editing an existing card with unsaved changes, clicking `Import...` and then canceling the dirty-state prompt could restore saved card data and clear the dirty state because focus restoration went through the selected card row. | Changed blocked top-level action cancel paths to restore focus directly to the active form field instead of via row-focus repopulation. |
| B30 | S08 | S08 | Clicking `Quick Refresher` on a deck that is not already selected only changes selection on the first click and requires a second click to actually start the session. | Stopped study-row focus from rerendering the deck list before the action click is delivered, so `Quick Refresher` starts on the first click. |
| B31 | S08 | S09 | Reverse-tabbing from the study search field appears to make focus "float" because the `All decks` breadcrumb stop is too visually subtle to read as the current keyboard target. | Removed the breadcrumb home control from the sequential tab order while preserving mouse activation, so reverse-tabbing no longer lands on the subtle breadcrumb stop. |
| B32 | S08 | S08 | The study deck-list view showed overlapping helper guidance in two places at once (`Press / to search...` in the main pane and separate session-start guidance in the status bar), which made the messaging feel redundant and unfocused. | Simplified the helper copy down to a single status-bar message so study guidance is focused and non-redundant. |
| B33 | S08 | S08 | Deleting a deck did not refresh the study deck list, leaving stale deck rows visible until some other rerender happened. | Deleting a deck now rerenders the study list and recalculates study selection immediately. |
| B34 | S08 | S08 | The study deck-row action label `Edit cards` was misleading because activating it first opened the broader deck editor workspace, including deck metadata, rather than a card-only editing surface. | Changed the deck-row action copy to `Edit deck` so it matches the broader deck editor workspace it opens. |
| B35 | S08 | S08 | Deck-editor breadcrumbs were ordered incorrectly as `All decks › Finder › Edit deck` instead of expressing the mode first, e.g. `All decks › Edit deck - Finder`. | Updated deck-editor breadcrumb copy to express the edit mode before the deck name. |
| B36 | S08 | S08 | Invoking `Import...` from a clean top-level view such as `Create` should preserve that top-level context if the file picker is canceled and after import completes, rather than eagerly rehoming the user to top-level `Study`. | Changed import to capture origin context, avoid navigation while the picker is open, preserve stable top-level views after cancel/success, and normalize only nested editing launches to top-level `Study` once a real import begins. |
| B37 | S09 planning | S09 | In active study sessions, the visual hierarchy currently emphasizes the session type (`Quick Refresher` / `Ongoing Review`) while the deck name is demoted into the smaller subtitle, which makes the current deck/context harder to scan quickly than the mode label. | Session rendering now promotes the deck name to the primary heading and demotes the session mode to the secondary subtitle/context line. |
| B38 | S09 planning | S09 | Multi-line deck descriptions are not preserved in presentation: the deck description field is a `<textarea>`, but the study-list rendering outputs the description into a plain `<div>` without preserving newline formatting, so line breaks collapse visually after save. | Description rendering now preserves saved line breaks in the study deck list and the card-editor deck metadata bar. |
| B39 | S09 | S09 | Clicking `Create...` switches the deck form into add/new-deck mode, but the deck list can still show the last selected deck row as highlighted. This leaves the screen presenting two conflicting states at once: create-new form on the right and existing-deck selection on the left. | Entering create mode now clears the prior deck selection highlight so the deck list and form communicate the same active context. |
| B40 | S09 | S09 | Import failures can be hard to diagnose or miss in the UI because server error payloads use lowercase `error` while the frontend only reads `Error`, and import failures only show a transient toast while restoring the status bar immediately. | Import now reads server `error` payloads correctly, shows failures in both toast and persistent status text, and logs metadata-only server import failures/successes without logging deck/card payload contents. |
| B41 | S09 | S09 | During manual testing of the new runtime flag, `./shortcuts --db /tmp/shortcuts-s09.db --auto-open false` failed with a misleading positional-argument error because the CLI only accepted the boolean form `--auto-open=false`. | Runtime argument normalization now accepts both `--auto-open=false` and `--auto-open false`, preserving the natural spaced form and avoiding the misleading DB-path guidance. |

Template:

```md
### Bug N — Short title

- Step: Name the checklist step that failed or was blocked
- Expected: What should have happened
- Actual: What actually happened
- Repro:
  1. Step one
  2. Step two
  3. Step three
- Severity: Low | Medium | High
- Environment: Browser + OS, if relevant
- Notes: Any extra context, screenshots, or suspected cause
```
