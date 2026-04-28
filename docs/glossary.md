# shortcutdeck — Glossary

> Status: v1.1
> Last updated: 2026-03-16
> Companion docs: prd.md, dd.md

---

## Core concepts

**Shortcut**
A single flashcard representing one keyboard shortcut. Has a prompt (what the shortcut does), an answer (the key combination), optional notes, and tags.

**Deck**
A collection of shortcuts for one application or context (e.g. Vim, tmux, Finder). The primary organisational unit of the app.

**Tag**
A free-form label attached to a shortcut card. Tags are the mechanism for sub-categorising shortcuts within a deck (e.g. `normal`, `insert`, `folding`, `basic`).

**Tag namespace**
A named grouping of related tags defined at the deck level. Tells the UI which tags belong to which dimension, enabling structured filtering. Example for a Vim deck:

- `mode` → normal, insert, visual, command
- `topic` → folding, buffers, motions, search
- `level` → basic, advanced

A card can carry tags from any or all namespaces. Decks without namespaces fall back to flat tag display.

---

## Study modes

**Ongoing Review**
The SM-2 spaced repetition study mode. The algorithm owns card selection and scheduling. Each day a queue of due cards is generated; the user works through them and rates each one. Runs indefinitely once started. Never namespace-filtered — always operates on the full deck.

**Quick Refresher**
A free-form, unscheduled browse mode. No ratings, no scheduling updates. The user pages through cards at their own pace. Supports namespace filtering so the user can focus on a specific topic or level.

**Due cards**
Cards whose SM-2 scheduled review date falls on or before today. Shown as a count on each deck entry on the home screen. Only shown when `review_active` is `true` for that deck.

**Start Review**
The action that initiates Ongoing Review for a deck. Sets `review_active = true`. Always a cold start: all card scheduling state is at defaults (interval=0, ease=2.5, repetitions=0, due=today), so every card is immediately due. The deck entry button label when `review_active` is `false`.

**Exit Review**
The action that stops Ongoing Review for a deck. Sets `review_active = false` and resets all SM-2 scheduling fields on the deck's cards to defaults. The `review_log` is never deleted. The deck entry returns to showing the Start Review button, and due counts disappear.

**Session**
A single sitting of Ongoing Review. Ends when all due cards for today in the deck are rated, or when the user explicitly ends it via Esc. A session summary is shown on completion. Completing or abandoning a session does not change `review_active`; only Start Review / Exit Review actions do.

---

## Rating scale (Ongoing Review)

| UI gesture | Needs effort | SM-2 grade | Meaning |
| --- | --- | --- | --- |
| Missed it | unchecked | 0 — Again | Complete blank — interval resets to 1 day |
| Missed it | checked | 1 — Partial | Saw answer, recognised it — still a miss, less severe |
| Got it | checked | 3 — Hard | Recalled but struggled — interval grows slowly |
| Got it | unchecked | 5 — Easy | Clean recall — interval grows aggressively |

The "Needs effort" checkbox appears after the answer is revealed, before the Got it / Missed it buttons.

---

## UI regions

**Sidebar**
The left navigation panel. Contains deck actions (Import, Create, Delete) and app actions (Settings, About). Shows a summary stat block at the bottom (deck count, shortcut count, global due status). No deck names or content.

**Main pane**
The central content area. Shows the deck list on the home screen, and the active session or summary screen during study.

**Status bar**
The single line at the bottom of the main pane. Used for system messages, keyboard hints, and error notifications. Never used for persistent data display.

---

## Home screen elements

**Deck entry**
A list item in the main pane representing one deck. Shows deck name, shortcut count, due count (if Ongoing Review is active), the scope dropdown, and the Quick Refresher / Ongoing Review (or Start Review) buttons.

**Scope dropdown**
A tree-style dropdown on each deck entry. Defaults to "All". Expands to show tag namespaces, each of which expands to show its values. The selected scope filters Quick Refresher sessions. Has no effect on Ongoing Review.

**Sidebar summary block**
The stat block at the bottom of the sidebar. Shows total decks, total shortcuts, and the global due status. Due status shows a red count if any active-review deck has due cards remaining today, and "all reviewed" in green only when every active-review deck has zero due cards for today.

---

## Session summary screen

Shown automatically when all due cards for today are rated, or when the user ends a session early via Esc. Displays:

- Cards reviewed this session
- Got it / Missed it counts
- Coming up: due card counts bucketed by horizon (tomorrow, in 3 days, in 7 days)
- Exit Review button (destructive, left-aligned)
- Back to decks button (right-aligned)

---

## Data model terms

**review_active**
A boolean flag on the Deck record. Controls whether due cards are generated and whether the due count is shown on the deck entry. The three-state lifecycle:

| State | `review_active` | Deck entry |
| --- | --- | --- |
| Never started | `false` (initial) | "Start Review" button; no due count |
| Active | `true` | "Ongoing Review →" button; due count visible |
| Exited | `false` (after exit) | "Start Review" button; no due count; cold start on next click |

**Review log**
An append-only record of every review event. Retained permanently even after Exit Review. Schema: `(id, card_id, reviewed_at, grade, interval_after, ease_after)`.

**Scheduling state**
The SM-2 fields stored on each card: `interval`, `ease_factor`, `repetitions`, `due_date`, `last_reviewed_at`. Cleared to defaults on Exit Review. Not affected by Quick Refresher sessions.
