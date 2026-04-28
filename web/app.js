const API_BASE = "/api/v1";
const DOT_COLORS = ["#378ADD", "#BA7517", "#1D9E75", "#534AB7", "#D85A30", "#D4537E", "#639922"];
const DECK_CREATE_STATUS = "Fill in the essential details to create an empty deck";
const APP_INFO = window.__SHORTCUTDECK__ || {
  version: "dev",
  description: "A desktop-first, keyboard-first flashcard tool for learning keyboard shortcuts without dashboard clutter.",
};

const state = {
  decks: [],
  cardsByDeck: new Map(),
  view: "study",
  activeDeckId: null,
  selectedDeckId: null,
  editingDeckId: null,
  deckFormBaseline: null,
  selectedCardId: null,
  editingCardId: null,
  cardTags: [],
  cardFormBaseline: null,
  studySearch: "",
  studyScopeByDeck: new Map(),
  dueCountsByDeck: new Map(),
  studyOptions: null,
  session: null,
  importPromptActive: false,
  importContext: null,
  aboutOpen: false,
  aboutReturnFocus: null,
};

const els = {
  deckList: document.getElementById("deck-list"),
  cardList: document.getElementById("card-list"),
  cardListHeader: document.getElementById("card-list-header"),
  deckMetaBar: document.getElementById("deck-meta-bar"),
  deckForm: {
    name: document.getElementById("d-name"),
    desc: document.getElementById("d-desc"),
    mode: document.getElementById("d-mode"),
    namespaces: document.getElementById("d-ns-editor"),
    submit: document.getElementById("d-submit"),
  },
  cardForm: {
    prompt: document.getElementById("c-prompt"),
    answer: document.getElementById("c-answer"),
    notes: document.getElementById("c-notes"),
    tagInput: document.getElementById("c-tag-input"),
    tagWrap: document.getElementById("tag-input-wrap"),
    tagHints: document.getElementById("tag-hint-wrap"),
    source: document.getElementById("c-source-badge"),
    submit: document.getElementById("c-submit"),
  },
  status: document.getElementById("status-msg"),
  toast: document.getElementById("toast"),
  statsDecks: document.getElementById("stat-decks"),
  statsCards: document.getElementById("stat-cards"),
  views: {
    study: document.getElementById("view-study"),
    session: document.getElementById("view-session"),
    decks: document.getElementById("view-decks"),
    cards: document.getElementById("view-cards"),
  },
  study: {
    search: document.getElementById("study-search"),
    list: document.getElementById("study-list"),
  },
  session: {
    back: document.getElementById("session-back-btn"),
    title: document.getElementById("session-title"),
    subtitle: document.getElementById("session-subtitle"),
    body: document.querySelector(".session-body"),
    summary: document.getElementById("session-summary"),
    progressFill: document.getElementById("session-progress-fill"),
    progressText: document.getElementById("session-progress-text"),
    cardLabel: document.getElementById("session-card-label"),
    cardFace: document.getElementById("session-card-face"),
    cardValue: document.getElementById("session-card-value"),
    cardTag: document.getElementById("session-card-tag"),
    cardAnswer: document.getElementById("session-card-answer"),
    cardNotes: document.getElementById("session-card-notes"),
    revealHint: document.getElementById("session-reveal-hint"),
    answerActions: document.getElementById("session-answer-actions"),
    effortCheck: document.getElementById("session-effort-check"),
    missed: document.getElementById("session-missed-btn"),
    gotIt: document.getElementById("session-gotit-btn"),
    navRow: document.getElementById("session-nav-row"),
    navPos: document.getElementById("session-nav-pos"),
    prev: document.getElementById("session-prev-btn"),
    next: document.getElementById("session-next-btn"),
    summaryTitle: document.getElementById("summary-title"),
    summarySubtitle: document.getElementById("summary-subtitle"),
    summaryReviewed: document.getElementById("summary-reviewed"),
    summaryGotIt: document.getElementById("summary-gotit"),
    summaryMissed: document.getElementById("summary-missed"),
    summaryNextDue: document.getElementById("summary-next-due"),
    summaryExitReview: document.getElementById("summary-exit-review-btn"),
    summaryBack: document.getElementById("summary-back-btn"),
  },
  about: {
    overlay: document.getElementById("about-overlay"),
    version: document.getElementById("about-version"),
    description: document.getElementById("about-description"),
    close: document.getElementById("about-close-btn"),
    dismiss: document.getElementById("about-dismiss-btn"),
  },
  breadcrumbs: {
    home: document.getElementById("bc-home"),
    sep1: document.getElementById("bc-sep1"),
    deck: document.getElementById("bc-deck"),
    sep2: document.getElementById("bc-sep2"),
    tail: document.getElementById("bc-tail"),
  },
  actions: {
    create: document.getElementById("create-action"),
    import: document.getElementById("import-action"),
    export: document.getElementById("export-action"),
    delete: document.getElementById("delete-action"),
    study: document.getElementById("study-action"),
    addNamespace: document.getElementById("add-namespace-btn"),
    clearDeck: document.getElementById("d-clear"),
    clearCard: document.getElementById("c-clear"),
    deleteCard: document.getElementById("c-delete"),
    exportDeck: document.getElementById("export-btn"),
    about: document.getElementById("about-action"),
  },
  importFile: document.getElementById("import-file"),
};

let toastTimer = 0;

/** @description Escape HTML entities in a string. */
function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

/** @description Set the status message text in the UI. */
function setStatus(message) {
  els.status.textContent = message;
}

/** @description Compute the base status depending on current view and state. */
function currentBaseStatus() {
  if (state.view === "session" && state.session) {
    if (state.session.mode === "review") {
      return state.session.phase === "summary"
        ? `${els.session.summarySubtitle.textContent} · Summary ready · Enter to return to decks`
        : els.status.textContent;
    }
    return els.status.textContent;
  }

  if (state.view === "study") {
    return "Select a deck or press / to search";
  }

  if (state.activeDeckId !== null) {
    const deck = currentDeck();
    if (!deck) {
      return DECK_CREATE_STATUS;
    }
    if (state.editingCardId) {
      const card = selectedCard();
      return card
        ? `Editing card "${card.Prompt}"`
        : `Editing deck ${deck.Name} · ${currentCards().length} card${currentCards().length === 1 ? "" : "s"}`;
    }
    return `Editing deck ${deck.Name} · ${currentCards().length} card${currentCards().length === 1 ? "" : "s"}`;
  }

  if (state.editingDeckId) {
    const deck = state.decks.find((item) => item.ID === state.editingDeckId);
    if (deck) {
      return `Editing ${deck.Name} - update fields then click Save changes`;
    }
  }

  return DECK_CREATE_STATUS;
}

/** @description Restore the status message to the current base status. */
function restoreBaseStatus() {
  setStatus(currentBaseStatus());
}

/** @description Open the About dialog and set focus appropriately. */
function openAboutDialog() {
  state.aboutReturnFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  state.aboutOpen = true;
  els.about.version.textContent = `Version ${APP_INFO.version || "dev"}`;
  els.about.description.textContent = APP_INFO.description || "";
  els.about.overlay.classList.remove("hidden");
  els.about.overlay.setAttribute("aria-hidden", "false");
  setStatus("About shortcutdeck · Esc to close");
  els.about.close.focus();
}

/** @description Close the About dialog and restore prior focus. */
function closeAboutDialog() {
  if (!state.aboutOpen) {
    return;
  }

  state.aboutOpen = false;
  els.about.overlay.classList.add("hidden");
  els.about.overlay.setAttribute("aria-hidden", "true");
  restoreBaseStatus();
  state.aboutReturnFocus?.focus?.();
  state.aboutReturnFocus = null;
}

/** @description Capture minimal context useful for import flow restoration. */
function captureImportContext() {
  return {
    view: state.view,
    activeDeckId: state.activeDeckId,
    selectedDeckId: state.selectedDeckId,
  };
}

/** @description Clear any saved import context and prompt state. */
function clearImportContext() {
  state.importPromptActive = false;
  state.importContext = null;
}

/** @description Determine whether import context should be normalized to study view. */
function shouldNormalizeImportContext(context) {
  if (!context) {
    return false;
  }
  return context.view === "session" || context.activeDeckId !== null;
}

/** @description Apply selection state after importing a deck based on context. */
function applyImportedDeckSelection(importedDeckID, context) {
  if (!importedDeckID) {
    return;
  }

  if (state.view === "study") {
    state.selectedDeckId = importedDeckID;
    renderStudyHome();
    focusStudyDeckRowByID(importedDeckID);
    return;
  }

  if (context?.view === "study") {
    state.selectedDeckId = importedDeckID;
    renderStudyHome();
  }
}

/** @description Show a transient toast notification with optional type. */
function showToast(message, type = "") {
  window.clearTimeout(toastTimer);
  els.toast.textContent = message;
  els.toast.className = `toast${type ? ` ${type}` : ""}`;
  toastTimer = window.setTimeout(() => {
    els.toast.className = "toast hidden";
  }, 2800);
}

/** @description Compute a deterministic dot color for a deck. */
function getDeckColor(deck) {
  let hash = 0;
  for (const char of deck.ID || deck.Name || "") {
    hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
  }
  return DOT_COLORS[hash % DOT_COLORS.length];
}

/** @description Return the currently active deck object or null. */
function currentDeck() {
  return state.decks.find((deck) => deck.ID === state.activeDeckId) || null;
}

/** @description Return the list of cards for the active deck. */
function currentCards() {
  return state.cardsByDeck.get(state.activeDeckId) || [];
}

/** @description Return the currently selected card or null. */
function selectedCard() {
  return currentCards().find((card) => card.ID === state.selectedCardId) || null;
}

/** @description Update deck and card counts displayed in the UI. */
function updateStats() {
  const deckCount = state.decks.length;
  const cardCount = Array.from(state.cardsByDeck.values()).reduce((total, cards) => total + cards.length, 0);
  els.statsDecks.innerHTML = `<span>${deckCount}</span> deck${deckCount === 1 ? "" : "s"}`;
  els.statsCards.innerHTML = `<span>${cardCount}</span> shortcut${cardCount === 1 ? "" : "s"}`;
}

/** @description Update breadcrumb UI according to current view and selection. */
function updateBreadcrumbs() {
  const activeDeck = currentDeck();
  if (state.view === "study") {
    els.breadcrumbs.home.classList.add("active");
    els.breadcrumbs.sep1.classList.add("hidden");
    els.breadcrumbs.deck.classList.add("hidden");
    els.breadcrumbs.sep2.classList.add("hidden");
    els.breadcrumbs.tail.classList.add("hidden");
  } else if (state.view === "session" && state.session) {
    els.breadcrumbs.home.classList.remove("active");
    const deck = state.decks.find((item) => item.ID === state.session.deckID);
    els.breadcrumbs.sep1.classList.remove("hidden");
    els.breadcrumbs.deck.classList.remove("hidden");
    els.breadcrumbs.sep2.classList.remove("hidden");
    els.breadcrumbs.tail.classList.remove("hidden");
    els.breadcrumbs.deck.textContent = deck?.Name || "Study";
    els.breadcrumbs.tail.textContent = state.session.phase === "summary"
      ? "Session summary"
      : state.session.mode === "refresher" ? "Quick Refresher" : "Ongoing Review";
  } else if (activeDeck) {
    els.breadcrumbs.home.classList.remove("active");
    els.breadcrumbs.sep1.classList.remove("hidden");
    els.breadcrumbs.deck.classList.remove("hidden");
    els.breadcrumbs.sep2.classList.add("hidden");
    els.breadcrumbs.tail.classList.add("hidden");
    els.breadcrumbs.deck.textContent = `Edit deck - ${activeDeck.Name}`;
  } else {
    els.breadcrumbs.home.classList.remove("active");
    els.breadcrumbs.sep1.classList.add("hidden");
    els.breadcrumbs.deck.classList.add("hidden");
    els.breadcrumbs.sep2.classList.add("hidden");
    els.breadcrumbs.tail.classList.remove("hidden");
    els.breadcrumbs.tail.textContent = "Create deck";
  }
}

/** @description Compare two decks by name for sorting. */
function compareDecksByName(left, right) {
  return left.Name.localeCompare(right.Name);
}

/** @description Return decks sorted by name. */
function sortedDecks() {
  return [...state.decks].sort(compareDecksByName);
}

/** @description Return the default deck ID to select in study view. */
function defaultStudyDeckID() {
  return sortedDecks()[0]?.ID || null;
}

/** @description Ensure a deck is selected for study, optionally constrained to visible decks. */
function ensureStudyDeckSelection(visibleDecks = null) {
  const decks = visibleDecks || sortedDecks();
  if (!decks.length) {
    state.selectedDeckId = null;
    return null;
  }

  const selectedVisible = decks.find((deck) => deck.ID === state.selectedDeckId);
  if (selectedVisible) {
    return selectedVisible.ID;
  }

  state.selectedDeckId = decks[0].ID;
  return state.selectedDeckId;
}

/** @description Render the namespace rows in the deck form. */
function renderDeckNamespaceRows(namespaces = {}) {
  els.deckForm.namespaces.innerHTML = "";
  const entries = Object.entries(namespaces);
  if (entries.length === 0) {
    addDeckNamespaceRow();
    return;
  }
  for (const [name, values] of entries) {
    addDeckNamespaceRow(name, values.join(", "));
  }
}

/** @description Add a namespace input row to the deck form. */
function addDeckNamespaceRow(name = "", values = "") {
  const row = document.createElement("div");
  row.className = "ns-row";
  row.innerHTML = `
    <input class="ns-name" type="text" placeholder="name" value="${escapeHTML(name)}">
    <span class="ns-sep">→</span>
    <input class="ns-values" type="text" placeholder="val1, val2, ..." value="${escapeHTML(values)}">
    <button class="ns-remove" type="button" aria-label="Remove namespace">×</button>
  `;
  row.querySelector(".ns-remove").addEventListener("click", () => {
    row.remove();
    if (!els.deckForm.namespaces.children.length) {
      addDeckNamespaceRow();
    }
  });
  els.deckForm.namespaces.appendChild(row);
}

/** @description Collect and return current values from the deck form. */
function collectDeckFormValues() {
  const tagNamespaces = {};
  for (const row of els.deckForm.namespaces.querySelectorAll(".ns-row")) {
    const name = row.querySelector(".ns-name").value.trim();
    const values = row
      .querySelector(".ns-values")
      .value
      .split(",")
      .map((value) => value.trim())
      .filter(Boolean);
    if (name) {
      tagNamespaces[name] = values;
    }
  }

  return {
    Name: els.deckForm.name.value.trim(),
    Description: els.deckForm.desc.value.trim(),
    DefaultReverseMode: els.deckForm.mode.value,
    TagNamespaces: tagNamespaces,
  };
}

/** @description Normalize a namespace token by trimming whitespace. */
function normalizeNamespaceToken(value) {
  // Namespace values are edited in a comma-separated field, so separator
  // whitespace should not count as a dirty change.
  return String(value || "").trim();
}

/** @description Get the current deck form state as a serializable object. */
function currentDeckFormState() {
  const tagNamespaces = {};
  for (const row of els.deckForm.namespaces.querySelectorAll(".ns-row")) {
    const name = normalizeNamespaceToken(row.querySelector(".ns-name").value);
    const values = row
      .querySelector(".ns-values")
      .value
      .split(",")
      .map((value) => normalizeNamespaceToken(value))
      .filter((value) => value.length > 0);
    if (name) {
      tagNamespaces[name] = values;
    }
  }

  return {
    Name: trimTrailingWhitespace(els.deckForm.name.value),
    Description: trimTrailingWhitespace(els.deckForm.desc.value),
    DefaultReverseMode: els.deckForm.mode.value,
    TagNamespaces: tagNamespaces,
  };
}

/** @description Return the baseline state for the deck form. */
function deckFormBaselineState() {
  return state.deckFormBaseline || {
    Name: "",
    Description: "",
    DefaultReverseMode: "prompt_first",
    TagNamespaces: {},
  };
}

/** @description Compare two tag-namespace objects for deep equality. */
function sameTagNamespaces(left, right) {
  const leftKeys = Object.keys(left);
  const rightKeys = Object.keys(right);
  if (leftKeys.length !== rightKeys.length) {
    return false;
  }

  for (const key of leftKeys) {
    if (!(key in right)) {
      return false;
    }
    const leftValues = left[key];
    const rightValues = right[key];
    if (leftValues.length !== rightValues.length) {
      return false;
    }
    if (leftValues.some((value, index) => value !== rightValues[index])) {
      return false;
    }
  }

  return true;
}

/** @description Return whether the deck form contains unsaved changes. */
function hasUnsavedDeckChanges() {
  const current = currentDeckFormState();
  const baseline = deckFormBaselineState();

  return current.Name !== baseline.Name
    || current.Description !== baseline.Description
    || current.DefaultReverseMode !== baseline.DefaultReverseMode
    || !sameTagNamespaces(current.TagNamespaces, baseline.TagNamespaces);
}

/** @description Prompt to confirm discarding unsaved deck changes. */
function confirmDiscardDeckChanges() {
  return window.confirm("Discard unsaved deck changes?");
}

/** @description Reset the deck form to its default empty state. */
function resetDeckForm() {
  state.editingDeckId = null;
  state.deckFormBaseline = {
    Name: "",
    Description: "",
    DefaultReverseMode: "prompt_first",
    TagNamespaces: {},
  };
  if (state.activeDeckId === null) {
    state.selectedDeckId = null;
  }
  els.deckForm.name.value = "";
  els.deckForm.desc.value = "";
  els.deckForm.mode.value = "prompt_first";
  els.deckForm.submit.textContent = "Add deck";
  renderDeckNamespaceRows({});
}

/** @description Populate the deck form with values from a deck object. */
function populateDeckForm(deck) {
  state.selectedDeckId = deck.ID;
  state.editingDeckId = deck.ID;
  state.deckFormBaseline = {
    Name: trimTrailingWhitespace(deck.Name),
    Description: trimTrailingWhitespace(deck.Description),
    DefaultReverseMode: deck.DefaultReverseMode || "prompt_first",
    TagNamespaces: Object.fromEntries(
      Object.entries(deck.TagNamespaces || {}).map(([name, values]) => [
        normalizeNamespaceToken(name),
        values.map((value) => normalizeNamespaceToken(value)),
      ]),
    ),
  };
  els.deckForm.name.value = deck.Name || "";
  els.deckForm.desc.value = deck.Description || "";
  els.deckForm.mode.value = deck.DefaultReverseMode || "prompt_first";
  els.deckForm.submit.textContent = "Save changes";
  renderDeckNamespaceRows(deck.TagNamespaces || {});
  setStatus(`Editing ${deck.Name} - update fields then click Save changes`);
}

/** @description Render the list of decks in the sidebar. */
function renderDecks() {
  updateStats();
  if (!state.decks.length) {
    els.deckList.innerHTML = `
      <div class="empty-state">
        <div class="empty-title">No decks yet</div>
        <div class="empty-sub">Fill in the form and click Add deck</div>
      </div>
    `;
    return;
  }

  els.deckList.innerHTML = "";
  for (const deck of state.decks) {
    const cardCount = (state.cardsByDeck.get(deck.ID) || []).length;
    const namespaceKeys = Object.keys(deck.TagNamespaces || {});
    const metaParts = [];
    if (deck.Description) {
      metaParts.push(deck.Description);
    } else {
      metaParts.push("No description");
    }
    if (namespaceKeys.length) {
      const namespaceLabel = namespaceKeys.slice(0, 2).join(", ");
      metaParts.push(namespaceKeys.length > 2 ? `${namespaceLabel}, ...` : namespaceLabel);
    }

    const row = document.createElement("div");
    row.className = `deck-row${state.selectedDeckId === deck.ID ? " selected" : ""}`;
    row.dataset.deckId = deck.ID;
    row.tabIndex = 0;
    row.innerHTML = `
      <div class="deck-row-top">
        <span class="de-dot" style="background:${getDeckColor(deck)}"></span>
        <div class="deck-row-info">
          <div class="deck-row-name">${escapeHTML(deck.Name)}</div>
          <div class="deck-row-meta">${escapeHTML(metaParts.join(" · "))}</div>
        </div>
        <span class="deck-row-mode">${escapeHTML(formatReverseMode(deck.DefaultReverseMode))}</span>
      </div>
      <div class="deck-row-footer">
        <button class="btn-edit-cards" type="button">Edit deck</button>
        <span class="deck-row-count">${cardCount} card${cardCount === 1 ? "" : "s"}</span>
      </div>
    `;

    row.addEventListener("click", (event) => {
      if (event.target.closest(".btn-edit-cards")) {
        return;
      }
      focusDeckByID(deck.ID);
    });
    row.addEventListener("focus", () => {
      if (!selectDeckByID(deck.ID, { rerender: false })) {
        restoreDeckFocusAfterBlockedSelection();
      }
    });
    row.addEventListener("dblclick", () => openCardEditor(deck.ID));
    row.addEventListener("keydown", (event) => {
      if (event.key === "Enter") {
        event.preventDefault();
        openCardEditor(deck.ID);
      } else if (event.key === "ArrowDown" || event.key === "ArrowUp") {
        event.preventDefault();
        focusAdjacentDeck(deck.ID, event.key === "ArrowDown" ? 1 : -1);
      }
    });

    row.querySelector(".btn-edit-cards").addEventListener("click", () => {
      openCardEditor(deck.ID);
    });

    els.deckList.appendChild(row);
  }
}

/** @description Render tag hint controls for a deck in the card editor. */
function renderTagHints(deck) {
  els.cardForm.tagHints.innerHTML = "";
  const entries = getDeckTagNamespaceEntries(deck);
  if (!entries.length) {
    els.cardForm.tagInput.classList.remove("hidden");
    els.cardForm.tagInput.disabled = false;
    els.cardForm.tagInput.placeholder = "Type tag, press Enter";
    return;
  }

  els.cardForm.tagInput.classList.add("hidden");
  els.cardForm.tagInput.disabled = true;
  els.cardForm.tagInput.value = "";

  for (const [name, values] of entries) {
    const row = document.createElement("div");
    row.className = "tag-select-row";
    const availableValues = values.filter((value) => !state.cardTags.includes(value));
    const options = availableValues
      .map((value) => `<option value="${escapeHTML(value)}">${escapeHTML(value)}</option>`)
      .join("");
    row.innerHTML = `
      <label class="tag-select-label" for="tag-select-${escapeHTML(name)}">${escapeHTML(name)}</label>
      <select
        class="tag-namespace-select"
        id="tag-select-${escapeHTML(name)}"
        data-tag-namespace="${escapeHTML(name)}"
        ${availableValues.length ? "" : "disabled"}
      >
        <option value="">${availableValues.length ? `Add ${escapeHTML(name)} tag...` : `All ${escapeHTML(name)} tags selected`}</option>
        ${options}
      </select>
    `;
    row.querySelector(".tag-namespace-select")?.addEventListener("change", (event) => {
      const { value } = event.target;
      if (addTag(value)) {
        event.target.value = "";
      }
      event.target.focus();
    });
    els.cardForm.tagHints.appendChild(row);
  }
}

/** @description Render the current card tags as tag chips in the UI. */
function renderCardTags() {
  for (const chip of els.cardForm.tagWrap.querySelectorAll(".tag-chip")) {
    chip.remove();
  }

  for (const tag of state.cardTags) {
    const chip = document.createElement("span");
    chip.className = "tag-chip";
    chip.innerHTML = `
      <span>${escapeHTML(tag)}</span>
      <button class="tag-chip-remove" type="button" aria-label="Remove tag ${escapeHTML(tag)}">×</button>
    `;
    chip.querySelector(".tag-chip-remove").addEventListener("click", () => {
      state.cardTags = state.cardTags.filter((value) => value !== tag);
      renderCardTags();
    });
    els.cardForm.tagWrap.insertBefore(chip, els.cardForm.tagInput);
  }

  const deck = currentDeck();
  if (deck) {
    renderTagHints(deck);
  }
}

/** @description Reset the card form to its default empty state. */
function resetCardForm() {
  state.editingCardId = null;
  state.selectedCardId = null;
  state.cardTags = [];
  state.cardFormBaseline = {
    Prompt: "",
    Answer: "",
    Notes: "",
    Tags: [],
  };
  els.cardForm.prompt.value = "";
  els.cardForm.answer.value = "";
  els.cardForm.notes.value = "";
  els.cardForm.tagInput.value = "";
  els.cardForm.tagInput.classList.remove("hidden");
  els.cardForm.tagInput.disabled = false;
  els.cardForm.tagInput.placeholder = "Type tag, press Enter";
  els.cardForm.tagHints.innerHTML = "";
  els.cardForm.source.textContent = "manual";
  els.cardForm.submit.textContent = "Add card";
  els.actions.deleteCard.classList.add("hidden");
  renderCardTags();
}

/** @description Populate the card form with values from a card object. */
function populateCardForm(card) {
  state.selectedCardId = card.ID;
  state.editingCardId = card.ID;
  state.cardTags = [...(card.Tags || [])];
  state.cardFormBaseline = {
    Prompt: trimTrailingWhitespace(card.Prompt),
    Answer: trimTrailingWhitespace(card.Answer),
    Notes: trimTrailingWhitespace(card.Notes),
    Tags: [...(card.Tags || [])],
  };
  els.cardForm.prompt.value = card.Prompt || "";
  els.cardForm.answer.value = card.Answer || "";
  els.cardForm.notes.value = card.Notes || "";
  els.cardForm.tagInput.value = "";
  els.cardForm.source.textContent = card.Source || "manual";
  els.cardForm.submit.textContent = "Save card";
  els.actions.deleteCard.classList.remove("hidden");
  renderCardTags();
  setStatus(`Editing card "${card.Prompt}"`);
}

/** @description Render the list of cards for the active deck. */
function renderCards() {
  const deck = currentDeck();
  if (!deck) {
    els.cardList.innerHTML = "";
    els.cardListHeader.textContent = "Cards";
    return;
  }

  const cards = currentCards();
  els.cardListHeader.textContent = `Cards (${cards.length})`;
  if (!cards.length) {
    els.cardList.innerHTML = `
      <div class="empty-state">
        <div class="empty-title">No cards yet</div>
        <div class="empty-sub">Use the card form to add the first shortcut in ${escapeHTML(deck.Name)}</div>
      </div>
    `;
    return;
  }

  els.cardList.innerHTML = "";
  for (const card of cards) {
    const row = document.createElement("div");
    row.className = `card-row${state.selectedCardId === card.ID ? " selected" : ""}`;
    row.dataset.cardId = card.ID;
    row.tabIndex = 0;

    const tags = (card.Tags || [])
      .map((tag) => `<span class="card-tag">${escapeHTML(tag)}</span>`)
      .join("");

    row.innerHTML = `
      <div class="card-row-body">
        <div class="card-row-prompt">${escapeHTML(card.Prompt)}</div>
        <div class="card-row-answer">${escapeHTML(card.Answer)}</div>
        ${tags ? `<div class="card-row-tags">${tags}</div>` : ""}
        ${card.Notes ? `<div class="card-row-notes">${escapeHTML(card.Notes)}</div>` : ""}
      </div>
    `;

    const editCard = () => {
      focusCardByID(card.ID, { populateForm: true });
      els.cardForm.prompt.focus();
    };

    row.addEventListener("click", () => {
      focusCardByID(card.ID, { populateForm: true });
    });
    row.addEventListener("focus", () => {
      if (!selectCardByID(card.ID, { rerender: false, populateForm: true })) {
        restoreCardFocusAfterBlockedSelection();
      }
    });
    row.addEventListener("dblclick", editCard);
    row.addEventListener("keydown", (event) => {
      if (event.key === "Enter") {
        event.preventDefault();
        editCard();
      } else if (event.key === "ArrowDown" || event.key === "ArrowUp") {
        event.preventDefault();
        focusAdjacentCard(card.ID, event.key === "ArrowDown" ? 1 : -1);
      }
    });

    els.cardList.appendChild(row);
  }
}

/** @description Render the deck meta bar shown above the card editor. */
function renderCardEditorMeta(deck) {
  const pills = Object.keys(deck.TagNamespaces || {})
    .map((name) => `<span class="ns-pill">${escapeHTML(name)}</span>`)
    .join("");

  els.deckMetaBar.innerHTML = `
    <button class="dmi-back" type="button" id="card-back-btn">← Back</button>
    <span class="dmi-back-sep"></span>
    <span class="dmi-dot" style="background:${getDeckColor(deck)}"></span>
    <span class="dmi-name">${escapeHTML(deck.Name)}</span>
    ${deck.Description ? `<span class="dmi-sep">·</span><span class="dmi-desc">${escapeHTML(deck.Description)}</span>` : ""}
    ${pills ? `<div class="dmi-ns">${pills}</div>` : ""}
  `;

  els.deckMetaBar.querySelector("#card-back-btn").addEventListener("click", () => {
    maybeLeaveCardEditor();
  });
}

/** @description Switch the UI to the deck creation/editing view. */
function showDeckView() {
  state.view = "decks";
  state.activeDeckId = null;
  state.session = null;
  resetCardForm();
  els.views.study.classList.add("hidden");
  els.views.session.classList.add("hidden");
  els.views.decks.classList.remove("hidden");
  els.views.cards.classList.add("hidden");
  syncSidebarActionState("create");
  updateBreadcrumbs();
  setStatus(DECK_CREATE_STATUS);
  renderDecks();
}

/** @description Enter the deck-creation mode and show the deck view. */
function startCreateDeckMode() {
  resetDeckForm();
  state.selectedDeckId = null;
  showDeckView();
}

/** @description Prepare the UI for adding a new card to a deck. */
function setAddCardMode(deck) {
  resetCardForm();
  renderCards();
  setStatus(`Add a card to ${deck.Name} · fill in Prompt and Answer to get started`);
  els.cardForm.prompt.focus();
}

/** @description Focus a deck row element by deck ID. */
function focusDeckRowByID(deckID) {
  if (!deckID) {
    return;
  }
  window.requestAnimationFrame(() => {
    els.deckList.querySelector(`[data-deck-id="${deckID}"]`)?.focus();
  });
}

/** @description Return to the deck list view and restore focus to selected deck. */
function returnToDeckList() {
  const deckID = state.selectedDeckId;
  showDeckView();
  if (deckID) {
    focusDeckRowByID(deckID);
  }
}

/** @description Open the card editor for a given deck ID (async). */
async function openCardEditor(deckID) {
  state.view = "cards";
  state.activeDeckId = deckID;
  state.session = null;
  const deck = currentDeck();
  els.views.study.classList.add("hidden");
  els.views.session.classList.add("hidden");
  els.views.decks.classList.add("hidden");
  els.views.cards.classList.remove("hidden");
  syncSidebarActionState("create");
  updateBreadcrumbs();
  renderCardEditorMeta(deck);
  renderTagHints(deck);
  await ensureCardsLoaded(deckID);
  setAddCardMode(deck);
}

/** @description Human-readable label for deck reverse mode. */
function formatReverseMode(mode) {
  return {
    prompt_first: "Prompt first",
    answer_first: "Answer first",
    both: "Both",
  }[mode] || mode || "Prompt first";
}

/** @description Return tag-namespace entries that have values for a deck. */
function getDeckTagNamespaceEntries(deck) {
  return Object.entries(deck?.TagNamespaces || {}).filter(([, values]) => values.length > 0);
}

/** @description Return whether a deck constrains tags to namespaces. */
function deckUsesConstrainedTags(deck) {
  return getDeckTagNamespaceEntries(deck).length > 0;
}

/** @description Return a Set of allowed tag values for a deck. */
function getDeckAllowedTags(deck) {
  return new Set(getDeckTagNamespaceEntries(deck).flatMap(([, values]) => values));
}

/** @description Normalize a tag value by trimming whitespace. */
function normalizeTagValue(tag) {
  return String(tag || "").trim();
}

/** @description Validate a tag against deck constraints, returning result. */
function validateDeckTag(tag, deck = currentDeck()) {
  const normalizedTag = normalizeTagValue(tag);
  if (!normalizedTag) {
    return { ok: false, reason: "empty", tag: normalizedTag };
  }

  if (!deckUsesConstrainedTags(deck)) {
    return { ok: true, tag: normalizedTag };
  }

  const allowedTags = getDeckAllowedTags(deck);
  if (allowedTags.has(normalizedTag)) {
    return { ok: true, tag: normalizedTag };
  }

  return { ok: false, reason: "not_allowed", tag: normalizedTag };
}

/** @description Add a validated tag to the current card tag list. */
function addTag(tag) {
  const validation = validateDeckTag(tag);
  if (!validation.ok) {
    if (validation.reason === "not_allowed") {
      const deck = currentDeck();
      showToast(`"${validation.tag}" is not in this deck's allowed tags`, "error");
      setStatus(deck ? `Pick a tag defined on ${deck.Name} before saving.` : "Pick a valid tag before saving.");
    }
    return false;
  }

  if (state.cardTags.includes(validation.tag)) {
    return false;
  }

  state.cardTags = [...state.cardTags, validation.tag];
  renderCardTags();
  return true;
}

/** @description Commit a pending tag typed into the tag input control. */
function commitPendingCardTag() {
  if (els.cardForm.tagInput.disabled) {
    return true;
  }

  const pendingTag = els.cardForm.tagInput.value.trim();
  if (!pendingTag) {
    return true;
  }

  const added = addTag(pendingTag);
  if (added) {
    els.cardForm.tagInput.value = "";
  }
  return added;
}

/** @description Focus the appropriate tag control (input or select) for card tags. */
function focusCardTagControl() {
  if (els.cardForm.tagInput.disabled) {
    window.requestAnimationFrame(() => {
      els.cardForm.tagHints.querySelector(".tag-namespace-select:not(:disabled)")?.focus();
    });
    return;
  }

  els.cardForm.tagInput.focus();
}

/** @description Validate current card tags against a deck's allowed tags. */
function validateCardTagsForDeck(deck) {
  if (!deckUsesConstrainedTags(deck)) {
    return true;
  }

  const allowedTags = getDeckAllowedTags(deck);
  const invalidTag = state.cardTags.find((tag) => !allowedTags.has(tag));
  if (!invalidTag) {
    return true;
  }

  showToast(`"${invalidTag}" is not in this deck's allowed tags`, "error");
  setStatus("Pick tags from the deck's defined values before saving.");
  return false;
}

/** @description Perform a fetch to API_BASE and throw on non-OK responses. */
async function apiFetch(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, options);
  if (!response.ok) {
    let message = `Request failed (${response.status})`;
    try {
      const body = await response.json();
      if (body.error || body.Error) {
        message = body.error || body.Error;
      }
    } catch {
      // Keep generic message.
    }
    throw new Error(message);
  }
  return response;
}

/** @description Fetch JSON from the API and return parsed body. */
async function fetchJSON(path, options = {}) {
  const response = await apiFetch(path, options);
  return response.json();
}

/** @description Load decks and their cards from the server into state. */
async function loadDecks() {
  const decks = (await fetchJSON("/decks")) || [];
  state.decks = decks;

  const cardsByDeck = new Map();
  await Promise.all(
    decks.map(async (deck) => {
      const cards = (await fetchJSON(`/decks/${deck.ID}/cards`)) || [];
      cardsByDeck.set(deck.ID, cards);
    }),
  );

  state.cardsByDeck = cardsByDeck;
  renderDecks();
  renderStudyHome();
  updateStats();
}

/** @description Ensure cards for a given deck ID are loaded into state. */
async function ensureCardsLoaded(deckID) {
  const cards = (await fetchJSON(`/decks/${deckID}/cards`)) || [];
  state.cardsByDeck.set(deckID, cards);
  updateStats();
}

/** @description Handle creating or updating a deck from the deck form (async). */
async function handleDeckSubmit() {
  const payload = collectDeckFormValues();
  if (!payload.Name) {
    showToast("Name is required", "error");
    els.deckForm.name.focus();
    return;
  }

  try {
    if (state.editingDeckId) {
      const deck = await fetchJSON(`/decks/${state.editingDeckId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      state.decks = state.decks.map((current) => (current.ID === deck.ID ? deck : current));
      state.selectedDeckId = deck.ID;
      showToast(`${deck.Name} saved`);
    } else {
      const deck = await fetchJSON("/decks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      state.decks = [...state.decks, deck];
      state.cardsByDeck.set(deck.ID, []);
      state.selectedDeckId = deck.ID;
      showToast(`${deck.Name} added`);
    }

    state.decks.sort((a, b) => a.Name.localeCompare(b.Name));
    resetDeckForm();
    renderDecks();
    updateStats();
    setStatus(DECK_CREATE_STATUS);
  } catch (error) {
    setStatus(error.message);
    showToast(error.message, "error");
  }
}

/** @description Handle deletion of the currently selected deck (async). */
async function handleDeckDelete() {
  if (!state.selectedDeckId) {
    showToast("Select a deck first", "error");
    return;
  }

  const deck = state.decks.find((item) => item.ID === state.selectedDeckId);
  if (!deck) {
    return;
  }

  if (!window.confirm(`Delete "${deck.Name}" and all its cards? This cannot be undone.`)) {
    return;
  }

  try {
    await apiFetch(`/decks/${deck.ID}`, { method: "DELETE" });
    state.decks = state.decks.filter((item) => item.ID !== deck.ID);
    state.cardsByDeck.delete(deck.ID);
    state.studyScopeByDeck.delete(deck.ID);
    state.dueCountsByDeck.delete(deck.ID);

    const fallbackDeckID = sortedDecks()[0]?.ID || null;
    if (state.selectedDeckId === deck.ID) {
      state.selectedDeckId = fallbackDeckID;
    }

    if (state.activeDeckId === deck.ID) {
      showDeckView();
    }
    resetDeckForm();
    renderDecks();
    renderStudyHome();
    updateStats();
    if (state.view === "study") {
      if (state.selectedDeckId) {
        focusStudyDeckRowByID(state.selectedDeckId);
      } else {
        focusCreateAction();
      }
    }
    showToast(`${deck.Name} deleted`);
  } catch (error) {
    setStatus(error.message);
    showToast(error.message, "error");
  }
}

/** @description Build the card payload object from the card form (and optionally an existing card). */
function buildCardPayload(existingCard) {
  return {
    Prompt: els.cardForm.prompt.value.trim(),
    Answer: els.cardForm.answer.value.trim(),
    Notes: els.cardForm.notes.value.trim(),
    Tags: [...state.cardTags],
    Source: existingCard?.Source || "manual",
    Interval: existingCard?.Interval ?? 0,
    EaseFactor: existingCard?.EaseFactor ?? 2.5,
    Repetitions: existingCard?.Repetitions ?? 0,
    DueDate: existingCard?.DueDate || new Date().toISOString(),
    LastReviewedAt: existingCard?.LastReviewedAt || "0001-01-01T00:00:00Z",
  };
}

/** @description Trim trailing whitespace from a string. */
function trimTrailingWhitespace(value) {
  return String(value || "").replace(/\s+$/, "");
}

/** @description Toggle active/muted state for sidebar actions. */
function syncSidebarActionState(activeAction) {
  for (const [name, element] of Object.entries({
    study: els.actions.study,
    create: els.actions.create,
  })) {
    element.classList.toggle("active", name === activeAction);
    element.classList.toggle("muted", name !== activeAction);
  }
}

/** @description Build study scope options from a deck's tag namespaces. */
function getStudyScopeOptions(deck) {
  const options = [{ value: "", label: "All shortcuts" }];
  for (const [namespace, values] of Object.entries(deck.TagNamespaces || {})) {
    for (const value of values) {
      options.push({
        value: `${namespace}::${value}`,
        label: `${namespace}: ${value}`,
      });
    }
  }
  return options;
}

/** @description Get the selected study scope value for a deck ID. */
function getSelectedStudyScope(deckID) {
  return state.studyScopeByDeck.get(deckID) || "";
}

/** @description Extract the tag portion from a study scope value. */
function getStudyScopeTag(scopeValue) {
  if (!scopeValue) {
    return "";
  }
  const separatorIndex = scopeValue.indexOf("::");
  return separatorIndex === -1 ? scopeValue : scopeValue.slice(separatorIndex + 2);
}

/** @description Return a human-readable label for the selected study scope. */
function getStudyScopeLabel(deckID) {
  const scopeValue = getSelectedStudyScope(deckID);
  if (!scopeValue) {
    return "All shortcuts";
  }
  const [namespace = "", value = ""] = scopeValue.split("::");
  return namespace && value ? `${namespace}: ${value}` : scopeValue;
}

/** @description Determine session reverse mode, honoring an override. */
function getSessionReverseMode(deckID, overrideMode) {
  if (overrideMode) {
    return overrideMode;
  }
  const deck = state.decks.find((item) => item.ID === deckID);
  return deck?.DefaultReverseMode || "prompt_first";
}

/** @description Open or toggle the study options panel for a deck. */
function openStudyOptions(deckID, action) {
  const reverseMode = getSessionReverseMode(deckID);
  const isSamePanel = state.studyOptions?.deckID === deckID && state.studyOptions?.action === action;
  state.selectedDeckId = deckID;
  state.studyOptions = isSamePanel ? null : { deckID, action, reverseMode };
  renderStudyHome();
  if (!isSamePanel) {
    window.requestAnimationFrame(() => {
      els.study.list.querySelector(`[data-session-mode-for="${deckID}-${action}"]`)?.focus();
    });
  }
}

/** @description Close the study options panel if open. */
function closeStudyOptions() {
  if (!state.studyOptions) {
    return;
  }
  state.studyOptions = null;
  renderStudyHome();
}

/** @description Return the deck object for the active session, or null. */
function getSessionDeck() {
  return state.session ? state.decks.find((deck) => deck.ID === state.session.deckID) || null : null;
}

/** @description Return the currently presented session card, or null. */
function currentSessionCard() {
  return state.session?.cards[state.session.index] || null;
}

/** @description Build the card presentation list for a quick refresher session. */
function buildQuickRefresherCards(cards, reverseMode) {
  if (reverseMode === "both") {
    const promptFirstPass = cards.map((card, cardIndex) => ({
      card,
      side: "prompt_first",
      cardIndex,
      sideIndex: 1,
      sideCount: 2,
      cardCount: cards.length,
    }));
    const answerFirstPass = cards.map((card, cardIndex) => ({
      card,
      side: "answer_first",
      cardIndex,
      sideIndex: 2,
      sideCount: 2,
      cardCount: cards.length,
    }));
    return [...promptFirstPass, ...answerFirstPass];
  }
  const side = reverseMode === "answer_first" ? "answer_first" : "prompt_first";
  return cards.map((card, cardIndex) => ({ card, side, cardIndex, sideIndex: 1, sideCount: 1, cardCount: cards.length }));
}

/** @description Build card presentation list for review sessions (alias). */
function buildReviewCards(cards, reverseMode) {
  return buildQuickRefresherCards(cards, reverseMode);
}

/** @description Focus the selected study action button for given action. */
function focusSelectedStudyButton(action) {
  const deckID = state.selectedDeckId;
  if (!deckID) {
    focusStudySearch();
    return;
  }
  window.requestAnimationFrame(() => {
    els.study.list.querySelector(`[data-deck-id="${deckID}"] [data-study-action="${action}"]`)?.focus();
  });
}

/** @description Render the currently active session card or session summary. */
function renderSessionCard() {
  if (!state.session) {
    return;
  }

  if (state.session.phase === "summary") {
    renderSessionSummary();
    return;
  }

  els.session.body.classList.remove("hidden");
  els.session.summary.classList.add("hidden");

  const presentation = currentSessionCard();
  const deck = getSessionDeck();
  if (!presentation || !deck) {
    leaveSessionToStudy();
    return;
  }

  const { card, side } = presentation;
  const isAnswerFirst = side === "answer_first";
  const total = state.session.cards.length;
  const position = state.session.index + 1;
  const percent = Math.round((position / total) * 100);
  const scopeLabel = state.session.scopeLabel;
  const primaryValue = isAnswerFirst ? card.Answer : card.Prompt;
  const answerValue = isAnswerFirst ? card.Prompt : card.Answer;
  const isRefresher = state.session.mode === "refresher";
  const sessionLabel = isRefresher ? "Quick Refresher" : "Ongoing Review";
  const titleLabel = deck.Name || "Deck";
  const subtitleLabel = scopeLabel !== "All shortcuts" ? `${sessionLabel} · ${scopeLabel}` : sessionLabel;

  els.session.title.textContent = titleLabel;
  els.session.subtitle.textContent = subtitleLabel;
  els.session.progressFill.style.width = `${percent}%`;
  const cardPositionLabel = `card ${presentation.cardIndex + 1} of ${presentation.cardCount}`;
  const sideLabel = presentation.sideCount > 1 ? ` · side ${presentation.sideIndex} of ${presentation.sideCount}` : "";
  els.session.progressText.textContent = `${cardPositionLabel}${sideLabel}`;
  els.session.navPos.textContent = `${cardPositionLabel}${sideLabel}`;
  els.session.cardLabel.textContent = isAnswerFirst ? "What does this shortcut do…" : "What is the shortcut for…";
  els.session.cardValue.textContent = primaryValue || "";
  els.session.cardAnswer.textContent = answerValue || "";
  els.session.cardNotes.textContent = card.Notes || "";
  els.session.cardTag.textContent = scopeLabel;
  els.session.cardTag.classList.toggle("hidden", scopeLabel === "All shortcuts");
  els.session.cardAnswer.classList.toggle("hidden", !state.session.revealed);
  els.session.cardNotes.classList.toggle("hidden", !state.session.revealed || !card.Notes);
  els.session.revealHint.classList.toggle("hidden", state.session.revealed);
  els.session.answerActions.classList.toggle("hidden", isRefresher || !state.session.revealed);
  els.session.navRow.classList.toggle("hidden", !isRefresher || !state.session.revealed);
  els.session.prev.disabled = state.session.index === 0;
  els.session.next.disabled = state.session.index >= total - 1;
  if (!state.session.revealed) {
    els.session.effortCheck.checked = false;
  }

  updateBreadcrumbs();
  if (isRefresher) {
    setStatus(
      state.session.revealed
        ? `${titleLabel} · ${subtitleLabel} · j/k or arrows to move · Esc to return`
        : `${titleLabel} · ${subtitleLabel} · Space to reveal · Esc to return`,
    );
    return;
  }

  setStatus(
    state.session.revealed
      ? `${titleLabel} · ${subtitleLabel} · Rate with buttons · Esc to return`
      : `${titleLabel} · ${subtitleLabel} · Space to reveal · Esc to return`,
  );
}

/** @description Format a human label for a due bucket day offset. */
function formatDueBucketLabel(dayOffset) {
  if (dayOffset === 0) {
    return "Still due today";
  }
  if (dayOffset === 1) {
    return "Tomorrow";
  }
  return `In ${dayOffset} days`;
}

/** @description Summarize upcoming due counts for a deck as day-offset buckets. */
function summarizeUpcomingDue(deckID) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const cards = state.cardsByDeck.get(deckID) || [];
  const buckets = new Map();

  for (const card of cards) {
    const due = new Date(card.DueDate);
    due.setHours(0, 0, 0, 0);
    const dayOffset = Math.max(0, Math.round((due.getTime() - today.getTime()) / 86400000));
    buckets.set(dayOffset, (buckets.get(dayOffset) || 0) + 1);
  }

  return Array.from(buckets.entries())
    .sort((left, right) => left[0] - right[0])
    .slice(0, 4);
}

/** @description Render the session summary UI for review sessions. */
function renderSessionSummary() {
  if (!state.session) {
    return;
  }

  const deck = getSessionDeck();
  const subtitle = `${deck?.Name || "Deck"} · Ongoing Review`;
  const reviewed = state.session.reviewStats?.reviewed || 0;
  const gotIt = state.session.reviewStats?.gotIt || 0;
  const missed = state.session.reviewStats?.missed || 0;
  const rows = summarizeUpcomingDue(state.session.deckID);

  els.session.body.classList.add("hidden");
  els.session.summary.classList.remove("hidden");
  let summaryTitle = "Review paused";
  if (state.session.noDueCards) {
    summaryTitle = "Nothing due right now";
  } else if (state.session.completed) {
    summaryTitle = "Done for now";
  }

  els.session.summaryTitle.textContent = summaryTitle;
  els.session.summarySubtitle.textContent = subtitle;
  els.session.summaryReviewed.textContent = String(reviewed);
  els.session.summaryGotIt.textContent = String(gotIt);
  els.session.summaryMissed.textContent = String(missed);
  els.session.summaryNextDue.innerHTML = rows.length
    ? rows.map(([dayOffset, count]) => `
      <div class="next-due-row">
        <span>${escapeHTML(formatDueBucketLabel(dayOffset))}</span>
        <span>${count} card${count === 1 ? "" : "s"}</span>
      </div>
    `).join("")
    : `<div class="next-due-row"><span>No upcoming cards</span><span>All clear</span></div>`;

  updateBreadcrumbs();
  setStatus(`${subtitle} · Summary ready · Enter to return to decks`);
}

/** @description Reveal the answer for the current session card. */
function revealSessionCard() {
  if (!state.session || state.session.revealed) {
    return;
  }
  state.session.revealed = true;
  renderSessionCard();
}

/** @description Navigate within the session by index direction (-1 or +1). */
function navigateSession(direction) {
  if (!state.session) {
    return;
  }

  const nextIndex = state.session.index + direction;
  if (nextIndex < 0 || nextIndex >= state.session.cards.length) {
    return;
  }

  state.session.index = nextIndex;
  state.session.revealed = false;
  renderSessionCard();
}

/** @description Exit the session and return to the study view. */
function leaveSessionToStudy(action = "refresher") {
  state.session = null;
  void showStudyView({ focusTarget: action });
}

/** @description Start a quick refresher session for a deck (async). */
async function startQuickRefresher(deckID, reverseModeOverride = "") {
  const deck = state.decks.find((item) => item.ID === deckID);
  if (!deck) {
    return;
  }

  const scopeTag = getStudyScopeTag(getSelectedStudyScope(deckID));
  const scopeLabel = getStudyScopeLabel(deckID);
  const query = scopeTag ? `?tags=${encodeURIComponent(scopeTag)}` : "";

  setStatus(`Loading Quick Refresher for ${deck.Name}...`);
  try {
    const cards = (await fetchJSON(`/decks/${deckID}/cards${query}`)) || [];
    if (!cards.length) {
      showToast("No cards found for that deck/scope", "error");
      setStatus(`No cards available for ${deck.Name}${scopeLabel !== "All shortcuts" ? ` · ${scopeLabel}` : ""}`);
      return;
    }

    state.selectedDeckId = deck.ID;
    state.studyOptions = null;
    state.session = {
      mode: "refresher",
      phase: "active",
      completed: false,
      deckID: deck.ID,
      scopeLabel,
      cards: buildQuickRefresherCards(cards, getSessionReverseMode(deck.ID, reverseModeOverride)),
      index: 0,
      revealed: false,
    };
    state.view = "session";
    state.activeDeckId = null;
    els.views.study.classList.add("hidden");
    els.views.decks.classList.add("hidden");
    els.views.cards.classList.add("hidden");
    els.views.session.classList.remove("hidden");
    syncSidebarActionState("study");
    renderSessionCard();
    showToast(`Quick Refresher loaded for ${deck.Name}`);
    window.requestAnimationFrame(() => {
      els.session.cardFace.focus();
    });
  } catch (error) {
    setStatus(error.message);
    showToast(error.message, "error");
  }
}

/** @description Start an ongoing review session for a deck (async). */
async function startReviewSession(deckID, reverseModeOverride = "") {
  let deck = state.decks.find((item) => item.ID === deckID);
  if (!deck) {
    return;
  }

  const scopeTag = getStudyScopeTag(getSelectedStudyScope(deckID));
  const scopeLabel = getStudyScopeLabel(deckID);
  const query = scopeTag ? `?tags=${encodeURIComponent(scopeTag)}` : "";

  setStatus(`Loading Ongoing Review for ${deck.Name}...`);
  try {
    if (!deck.ReviewActive) {
      const startedDeck = await fetchJSON(`/decks/${deckID}/start-review`, { method: "POST" });
      state.decks = state.decks.map((item) => (item.ID === startedDeck.ID ? startedDeck : item));
      deck = startedDeck;
    }

    const cards = (await fetchJSON(`/decks/${deckID}/due${query}`)) || [];
    state.dueCountsByDeck.set(deckID, cards.length);
    if (!cards.length) {
      state.selectedDeckId = deck.ID;
      state.studyOptions = null;
      state.session = {
        mode: "review",
        phase: "summary",
        completed: false,
        noDueCards: true,
        deckID: deck.ID,
        scopeLabel,
        cards: [],
        index: 0,
        revealed: false,
        reviewStats: {
          reviewed: 0,
          gotIt: 0,
          missed: 0,
        },
      };
      state.view = "session";
      state.activeDeckId = null;
      els.views.study.classList.add("hidden");
      els.views.decks.classList.add("hidden");
      els.views.cards.classList.add("hidden");
      els.views.session.classList.remove("hidden");
      syncSidebarActionState("study");
      renderSessionSummary();
      showToast(`No due cards right now for ${deck.Name}`);
      window.requestAnimationFrame(() => {
        els.session.summaryBack.focus();
      });
      return;
    }

    state.selectedDeckId = deck.ID;
    state.studyOptions = null;
    state.session = {
      mode: "review",
      phase: "active",
      completed: false,
      noDueCards: false,
      deckID: deck.ID,
      scopeLabel,
      cards: buildReviewCards(cards, getSessionReverseMode(deck.ID, reverseModeOverride)),
      index: 0,
      revealed: false,
      reviewStats: {
        reviewed: 0,
        gotIt: 0,
        missed: 0,
      },
    };
    state.view = "session";
    state.activeDeckId = null;
    els.views.study.classList.add("hidden");
    els.views.decks.classList.add("hidden");
    els.views.cards.classList.add("hidden");
    els.views.session.classList.remove("hidden");
    syncSidebarActionState("study");
    renderSessionCard();
    showToast(`Ongoing Review loaded for ${deck.Name}`);
    window.requestAnimationFrame(() => {
      els.session.cardFace.focus();
    });
  } catch (error) {
    setStatus(error.message);
    showToast(error.message, "error");
  }
}

/** @description Map gesture inputs to a numeric review grade. */
function reviewGradeFromGesture(gotIt, neededEffort) {
  if (!gotIt && !neededEffort) {
    return 0;
  }
  if (!gotIt && neededEffort) {
    return 1;
  }
  if (gotIt && neededEffort) {
    return 3;
  }
  return 5;
}

/** @description Transition a review session into its summary phase. */
function showReviewSummary(completed) {
  if (!state.session || state.session.mode !== "review") {
    return;
  }
  state.session.phase = "summary";
  state.session.completed = completed;
  renderSessionSummary();
  window.requestAnimationFrame(() => {
    els.session.summaryBack.focus();
  });
}

/** @description Submit a review grade for the current session card (async). */
async function submitReview(gotIt) {
  if (!state.session || state.session.mode !== "review" || !state.session.revealed) {
    return;
  }

  const presentation = currentSessionCard();
  if (!presentation?.card?.ID) {
    return;
  }

  const neededEffort = els.session.effortCheck.checked;
  const grade = reviewGradeFromGesture(gotIt, neededEffort);

  try {
    const updatedCard = await fetchJSON(`/cards/${presentation.card.ID}/review`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ grade }),
    });

    const cards = currentCards().map((card) => (card.ID === updatedCard.ID ? updatedCard : card));
    state.cardsByDeck.set(state.session.deckID, cards);
    state.session.reviewStats.reviewed += 1;
    if (gotIt) {
      state.session.reviewStats.gotIt += 1;
    } else {
      state.session.reviewStats.missed += 1;
    }
    state.session.index += 1;
    state.session.revealed = false;

    if (state.session.index >= state.session.cards.length) {
      showToast("Review round complete");
      showReviewSummary(true);
      return;
    }

    renderSessionCard();
  } catch (error) {
    setStatus(error.message);
    showToast(error.message, "error");
  }
}

/** @description Select a deck in the study list by ID and rerender. */
function selectStudyDeckByID(deckID) {
  const deck = state.decks.find((item) => item.ID === deckID);
  if (!deck) {
    return false;
  }
  state.selectedDeckId = deck.ID;
  renderStudyHome();
  return true;
}

/** @description Focus and select text in the study search input. */
function focusStudySearch() {
  window.requestAnimationFrame(() => {
    els.study.search?.focus();
    els.study.search?.select();
  });
}

/** @description Focus the study scope select for a given deck ID. */
function focusStudyScopeByID(deckID) {
  window.requestAnimationFrame(() => {
    els.study.list.querySelector(`[data-study-scope-for="${deckID}"]`)?.focus();
  });
}

/** @description Focus the study deck row element for a given deck ID. */
function focusStudyDeckRowByID(deckID) {
  window.requestAnimationFrame(() => {
    els.study.list.querySelector(`[data-deck-id="${deckID}"]`)?.focus();
  });
}

/** @description Focus the create action button in the sidebar. */
function focusCreateAction() {
  window.requestAnimationFrame(() => {
    els.actions.create.focus();
  });
}

/** @description Sync the selected study deck state when focus moves. */
function syncStudySelectionFromFocus(deckID) {
  if (state.selectedDeckId === deckID) {
    return;
  }
  state.selectedDeckId = deckID;
  syncSelectedStudyEntryByID(deckID);
}

/** @description Update study list DOM to mark a deck row as selected. */
function syncSelectedStudyEntryByID(deckID) {
  for (const row of els.study.list.querySelectorAll(".study-entry.selected")) {
    row.classList.remove("selected");
  }
  els.study.list.querySelector(`[data-deck-id="${deckID}"]`)?.classList.add("selected");
}

/** @description Format the study meta text for a deck (card counts, due). */
function formatStudyMeta(deck, cardCount) {
  const parts = [`${cardCount} shortcut${cardCount === 1 ? "" : "s"}`];
  if (deck.ReviewActive) {
    const dueCount = state.dueCountsByDeck.get(deck.ID) ?? 0;
    parts.push(`${dueCount} due today`);
  }
  return parts.join(" · ");
}

/** @description Render the study home list with decks, scopes, and actions. */
function renderStudyHome() {
  if (!els.study.list) {
    return;
  }

  const query = state.studySearch.trim().toLowerCase();
  const decks = sortedDecks()
    .filter((deck) => {
      if (!query) {
        return true;
      }
      const haystack = `${deck.Name} ${deck.Description || ""}`.toLowerCase();
      return haystack.includes(query);
    });

  if (!decks.length) {
    state.selectedDeckId = null;
    els.study.list.innerHTML = `
      <div class="study-empty">
        ${state.decks.length ? "No decks match the current filter." : "No decks yet. Create a deck first, then come back to Study."}
      </div>
    `;
    return;
  }

  ensureStudyDeckSelection(decks);

  els.study.list.innerHTML = "";
  for (const deck of decks) {
    const cardCount = (state.cardsByDeck.get(deck.ID) || []).length;
    const selectedScope = getSelectedStudyScope(deck.ID);
    const scopeOptions = getStudyScopeOptions(deck)
      .map((option) => `
        <option value="${escapeHTML(option.value)}"${option.value === selectedScope ? " selected" : ""}>
          ${escapeHTML(option.label)}
        </option>
      `)
      .join("");
    const reviewButtonLabel = deck.ReviewActive ? "Ongoing Review →" : "Start Review";
    const dueCount = state.dueCountsByDeck.get(deck.ID) ?? 0;
    const reviewIndicator = deck.ReviewActive
      ? dueCount > 0
        ? `<span class="study-due">${dueCount} due today</span>`
        : `<span class="study-all-good">All caught up</span>`
      : "";
    const optionsState = state.studyOptions?.deckID === deck.ID ? state.studyOptions : null;
    const optionsMarkup = optionsState ? `
      <div class="study-session-options">
        <div class="study-session-options-title">${optionsState.action === "review" ? "Review options" : "Quick Refresher options"}</div>
        <div class="study-session-options-note">Applies to this session only. Deck default stays unchanged.</div>
        <label class="study-session-field">
          <span>Card presentation mode</span>
          <select class="scope-select" data-session-mode-for="${escapeHTML(`${deck.ID}-${optionsState.action}`)}">
            <option value="prompt_first"${optionsState.reverseMode === "prompt_first" ? " selected" : ""}>Prompt first</option>
            <option value="answer_first"${optionsState.reverseMode === "answer_first" ? " selected" : ""}>Answer first</option>
            <option value="both"${optionsState.reverseMode === "both" ? " selected" : ""}>Both</option>
          </select>
        </label>
        <div class="study-session-option-actions">
          <button class="btn" type="button" data-session-start="${optionsState.action}">Start now</button>
          <button class="btn btn-ghost" type="button" data-session-cancel="${optionsState.action}">Cancel</button>
        </div>
      </div>
    ` : "";

    const row = document.createElement("div");
    row.className = `study-entry${state.selectedDeckId === deck.ID ? " selected" : ""}`;
    row.dataset.deckId = deck.ID;
    row.tabIndex = 0;
    row.innerHTML = `
      <div class="study-entry-top">
        <div class="study-name-wrap">
          <div class="study-name">
            <span class="study-dot" style="background:${getDeckColor(deck)}"></span>
            <span>${escapeHTML(deck.Name)}</span>
          </div>
          ${deck.Description ? `<div class="study-description">${escapeHTML(deck.Description)}</div>` : ""}
        </div>
        <div class="study-meta">${escapeHTML(formatStudyMeta(deck, cardCount))}</div>
      </div>
      <div class="study-entry-footer">
        <label class="study-scope">
          <select class="scope-select" data-study-scope-for="${escapeHTML(deck.ID)}" aria-label="Scope for ${escapeHTML(deck.Name)}">
            ${scopeOptions}
          </select>
        </label>
        <div class="study-actions">
          <div class="study-split-action">
            <button class="btn" type="button" data-study-action="refresher">Quick Refresher</button>
            <button class="study-action-chevron" type="button" aria-label="Quick Refresher session options" data-study-options="refresher">▾</button>
          </div>
          <div class="study-split-action">
            <button class="btn btn-review" type="button" data-study-action="review">${reviewButtonLabel}</button>
            <button class="study-action-chevron" type="button" aria-label="Review session options" data-study-options="review">▾</button>
          </div>
          ${reviewIndicator}
        </div>
        ${optionsMarkup}
      </div>
    `;

    row.addEventListener("click", (event) => {
      if (event.target.closest("button") || event.target.closest("select")) {
        return;
      }
      if (selectStudyDeckByID(deck.ID)) {
        focusStudyDeckRowByID(deck.ID);
      }
    });
    row.addEventListener("focus", () => {
      syncStudySelectionFromFocus(deck.ID);
    });
    row.addEventListener("keydown", (event) => {
      const visibleDecks = decks;
      const currentIndex = visibleDecks.findIndex((item) => item.ID === deck.ID);
      if (event.key === "ArrowDown" || event.key === "ArrowUp") {
        event.preventDefault();
        const nextIndex = currentIndex + (event.key === "ArrowDown" ? 1 : -1);
        if (nextIndex < 0 || nextIndex >= visibleDecks.length) {
          return;
        }
        const nextDeckID = visibleDecks[nextIndex].ID;
        if (selectStudyDeckByID(nextDeckID)) {
          focusStudyDeckRowByID(nextDeckID);
        }
      } else if (event.key === "Enter") {
        event.preventDefault();
        focusStudyScopeByID(deck.ID);
      }
    });

    row.querySelector(".scope-select").addEventListener("focus", () => {
      syncStudySelectionFromFocus(deck.ID);
    });
    row.querySelector(".scope-select").addEventListener("change", async (event) => {
      state.selectedDeckId = deck.ID;
      state.studyScopeByDeck.set(deck.ID, event.target.value);
      if (deck.ReviewActive) {
        await loadDueCountForDeck(deck.ID);
      }
      renderStudyHome();
      focusStudyScopeByID(deck.ID);
    });
    const refresherButton = row.querySelector('[data-study-action="refresher"]');
    const reviewButton = row.querySelector('[data-study-action="review"]');
    const refresherOptionsButton = row.querySelector('[data-study-options="refresher"]');
    const reviewOptionsButton = row.querySelector('[data-study-options="review"]');

    refresherButton.addEventListener("focus", () => {
      syncStudySelectionFromFocus(deck.ID);
      window.requestAnimationFrame(() => {
        els.study.list.querySelector(`[data-deck-id="${deck.ID}"] [data-study-action="refresher"]`)?.focus();
      });
    });
    reviewButton.addEventListener("focus", () => {
      syncStudySelectionFromFocus(deck.ID);
      window.requestAnimationFrame(() => {
        els.study.list.querySelector(`[data-deck-id="${deck.ID}"] [data-study-action="review"]`)?.focus();
      });
    });
    refresherOptionsButton.addEventListener("focus", () => {
      syncStudySelectionFromFocus(deck.ID);
      window.requestAnimationFrame(() => {
        els.study.list.querySelector(`[data-deck-id="${deck.ID}"] [data-study-options="refresher"]`)?.focus();
      });
    });
    reviewOptionsButton.addEventListener("focus", () => {
      syncStudySelectionFromFocus(deck.ID);
      window.requestAnimationFrame(() => {
        els.study.list.querySelector(`[data-deck-id="${deck.ID}"] [data-study-options="review"]`)?.focus();
      });
    });

    refresherButton.addEventListener("click", () => {
      state.selectedDeckId = deck.ID;
      void startQuickRefresher(deck.ID);
    });
    reviewButton.addEventListener("click", () => {
      state.selectedDeckId = deck.ID;
      void startReviewSession(deck.ID);
    });
    refresherOptionsButton.addEventListener("click", () => openStudyOptions(deck.ID, "refresher"));
    reviewOptionsButton.addEventListener("click", () => openStudyOptions(deck.ID, "review"));

    if (optionsState) {
      row.querySelector(`[data-session-mode-for="${deck.ID}-${optionsState.action}"]`)?.addEventListener("change", (event) => {
        state.studyOptions = {
          ...optionsState,
          reverseMode: event.target.value,
        };
      });
      row.querySelector(`[data-session-start="${optionsState.action}"]`)?.addEventListener("click", () => {
        if (optionsState.action === "review") {
          void startReviewSession(deck.ID, state.studyOptions?.reverseMode || "");
        } else {
          void startQuickRefresher(deck.ID, state.studyOptions?.reverseMode || "");
        }
      });
      row.querySelector(`[data-session-cancel="${optionsState.action}"]`)?.addEventListener("click", () => {
        closeStudyOptions();
        focusSelectedStudyButton(optionsState.action);
      });
    }

    els.study.list.appendChild(row);
  }
}

async function loadDueCountForDeck(deckID) {
  const deck = state.decks.find((item) => item.ID === deckID);
  if (!deck || !deck.ReviewActive) {
    state.dueCountsByDeck.set(deckID, 0);
    return 0;
  }

  const selectedScope = getSelectedStudyScope(deckID);
  const selectedTag = getStudyScopeTag(selectedScope);
  const query = selectedTag ? `?tags=${encodeURIComponent(selectedTag)}` : "";
  const dueCards = (await fetchJSON(`/decks/${deckID}/due${query}`)) || [];
  const dueCount = dueCards.length;
  state.dueCountsByDeck.set(deckID, dueCount);
  return dueCount;
}

/** @description Refresh due counts for all review-active decks. */
async function refreshStudyDueCounts() {
  await Promise.all(
    state.decks
      .filter((deck) => deck.ReviewActive)
      .map((deck) => loadDueCountForDeck(deck.ID)),
  );
}

/** @description Show the study view and optionally focus a target. */
async function showStudyView(options = {}) {
  const { focusTarget = "selected" } = options;
  state.view = "study";
  state.activeDeckId = null;
  state.session = null;
  state.studyOptions = null;
  resetCardForm();
  els.views.study.classList.remove("hidden");
  els.views.session.classList.add("hidden");
  els.views.decks.classList.add("hidden");
  els.views.cards.classList.add("hidden");
  syncSidebarActionState("study");
  updateBreadcrumbs();
  setStatus("Loading study home...");
  renderStudyHome();

  try {
    await refreshStudyDueCounts();
    renderStudyHome();
    const selectedDeckID = ensureStudyDeckSelection();
    setStatus("Select a deck or press / to search");
    if (focusTarget === "refresher" || focusTarget === "review") {
      focusSelectedStudyButton(focusTarget);
    } else if (focusTarget === "search") {
      focusStudySearch();
    } else if (focusTarget === "none") {
      // Keep the first deck selected without auto-moving live focus.
    } else if (selectedDeckID) {
      focusStudyDeckRowByID(selectedDeckID);
    } else {
      focusCreateAction();
    }
  } catch (error) {
    setStatus(error.message);
    showToast(error.message, "error");
  }
}

/** @description Handle Escape key behavior while in study view. */
function handleStudyEscape() {
  if (state.studyOptions) {
    const { action } = state.studyOptions;
    closeStudyOptions();
    setStatus("Session options closed");
    focusSelectedStudyButton(action);
    return;
  }

  const selectedDeckID = state.selectedDeckId;
  const selectedScope = selectedDeckID ? getSelectedStudyScope(selectedDeckID) : "";

  if (selectedScope) {
    state.studyScopeByDeck.set(selectedDeckID, "");
    renderStudyHome();
    setStatus("Study scope reset to All shortcuts");
    focusStudyScopeByID(selectedDeckID);
    return;
  }

  if (selectedDeckID) {
    const fallbackDeckID = defaultStudyDeckID();
    state.selectedDeckId = fallbackDeckID;
    renderStudyHome();
    setStatus("Study selection reset");
    if (fallbackDeckID) {
      focusStudyDeckRowByID(fallbackDeckID);
    } else {
      focusStudySearch();
    }
    return;
  }

  if (state.studySearch) {
    state.studySearch = "";
    els.study.search.value = "";
    renderStudyHome();
    setStatus("Study filter cleared");
    focusStudySearch();
    return;
  }

  setStatus("Select a deck or press / to search");
  if (defaultStudyDeckID()) {
    focusStudyDeckRowByID(defaultStudyDeckID());
  } else {
    focusStudySearch();
  }
}

/** @description Update deck list DOM to mark a deck row as selected. */
function syncSelectedDeckRowByID(deckID) {
  for (const row of els.deckList.querySelectorAll(".deck-row.selected")) {
    row.classList.remove("selected");
  }
  els.deckList.querySelector(`[data-deck-id="${deckID}"]`)?.classList.add("selected");
}

/** @description Select a deck by ID for editing; optionally rerender. */
function selectDeckByID(deckID, options = {}) {
  const { rerender = true } = options;
  const deck = state.decks.find((item) => item.ID === deckID);
  if (!deck) {
    return false;
  }

  if (!canSwitchDeckSelection(deck.ID)) {
    return false;
  }

  state.selectedDeckId = deck.ID;
  if (state.activeDeckId === null) {
    populateDeckForm(deck);
  }
  if (rerender) {
    renderDecks();
  } else {
    // Focus can move across existing rows without tearing the list down.
    syncSelectedDeckRowByID(deck.ID);
  }
  return true;
}

/** @description Move focus to an adjacent deck row by direction. */
function focusAdjacentDeck(currentDeckID, direction) {
  const currentIndex = state.decks.findIndex((deck) => deck.ID === currentDeckID);
  const nextIndex = currentIndex + direction;
  if (currentIndex === -1 || nextIndex < 0 || nextIndex >= state.decks.length) {
    return;
  }

  const deck = state.decks[nextIndex];
  if (!selectDeckByID(deck.ID)) {
    return;
  }
  els.deckList.querySelector(`[data-deck-id="${deck.ID}"]`)?.focus();
}

/** @description Select and focus a deck row by ID. */
function focusDeckByID(deckID) {
  if (!selectDeckByID(deckID)) {
    return;
  }
  els.deckList.querySelector(`[data-deck-id="${deckID}"]`)?.focus();
}

/** @description Update card list DOM to mark a card row as selected. */
function syncSelectedCardRowByID(cardID) {
  for (const row of els.cardList.querySelectorAll(".card-row.selected")) {
    row.classList.remove("selected");
  }
  els.cardList.querySelector(`[data-card-id="${cardID}"]`)?.classList.add("selected");
}

/** @description Select a card by ID and optionally populate the form. */
function selectCardByID(cardID, options = {}) {
  const { rerender = true, populateForm = false } = options;
  const card = currentCards().find((item) => item.ID === cardID);
  if (!card) {
    return false;
  }

  if (!canSwitchCardSelection(card.ID)) {
    return false;
  }

  state.selectedCardId = card.ID;
  if (populateForm) {
    populateCardForm(card);
  }
  if (rerender) {
    renderCards();
  } else {
    // Keep DOM focus stable when selection changes via tab/focus movement.
    syncSelectedCardRowByID(card.ID);
  }
  return true;
}

/** @description Move focus to an adjacent card by direction. */
function focusAdjacentCard(currentCardID, direction) {
  const cards = currentCards();
  const currentIndex = cards.findIndex((card) => card.ID === currentCardID);
  const nextIndex = currentIndex + direction;
  if (currentIndex === -1 || nextIndex < 0 || nextIndex >= cards.length) {
    return;
  }

  const card = cards[nextIndex];
  if (!selectCardByID(card.ID)) {
    return;
  }
  els.cardList.querySelector(`[data-card-id="${card.ID}"]`)?.focus();
}

/** @description Select and focus a card row by ID. */
function focusCardByID(cardID, options = {}) {
  if (!cardID) {
    return;
  }
  if (!selectCardByID(cardID, options)) {
    return;
  }
  els.cardList.querySelector(`[data-card-id="${cardID}"]`)?.focus();
}

/** @description Restore focus to a deck row or deck form after a blocked selection. */
function restoreDeckFocusAfterBlockedSelection() {
  if (state.selectedDeckId) {
    focusDeckRowByID(state.selectedDeckId);
    return;
  }

  window.requestAnimationFrame(() => {
    els.deckForm.name.focus();
  });
}

/** @description Restore focus to a card row or card form after a blocked selection. */
function restoreCardFocusAfterBlockedSelection() {
  if (state.selectedCardId) {
    focusCardByID(state.selectedCardId, { rerender: false, populateForm: false });
    return;
  }

  window.requestAnimationFrame(() => {
    els.cardForm.prompt.focus();
  });
}

/** @description Restore sensible focus after a blocked leave action. */
function restoreFocusAfterBlockedLeave() {
  if (state.activeDeckId !== null) {
    window.requestAnimationFrame(() => {
      if (state.editingCardId) {
        els.cardForm.prompt.focus();
      } else {
        els.cardForm.prompt.focus();
      }
    });
    return;
  }

  if (state.view === "study") {
    if (state.selectedDeckId) {
      focusStudyDeckRowByID(state.selectedDeckId);
      return;
    }
    window.requestAnimationFrame(() => {
      els.study.search.focus();
    });
    return;
  }

  window.requestAnimationFrame(() => {
    els.deckForm.name.focus();
  });
}

/** @description Return the current card form state as a serializable object. */
function currentCardFormState() {
  return {
    Prompt: trimTrailingWhitespace(els.cardForm.prompt.value),
    Answer: trimTrailingWhitespace(els.cardForm.answer.value),
    Notes: trimTrailingWhitespace(els.cardForm.notes.value),
    Tags: [...state.cardTags],
  };
}

/** @description Return the baseline state for the card form. */
function cardFormBaselineState() {
  return state.cardFormBaseline || {
    Prompt: "",
    Answer: "",
    Notes: "",
    Tags: [],
  };
}

/** @description Determine whether the card form contains unsaved changes. */
function hasUnsavedCardChanges() {
  const current = currentCardFormState();
  const baseline = cardFormBaselineState();

  return current.Prompt !== baseline.Prompt
    || current.Answer !== baseline.Answer
    || current.Notes !== baseline.Notes
    || current.Tags.length !== baseline.Tags.length
    || current.Tags.some((tag, index) => tag !== baseline.Tags[index]);
}

/** @description Prompt to confirm discarding unsaved card changes. */
function confirmDiscardCardChanges() {
  return window.confirm("Discard unsaved card changes?");
}

/** @description Determine if it's safe to switch deck selection, prompting if needed. */
function canSwitchDeckSelection(nextDeckID) {
  if (state.activeDeckId !== null || !hasUnsavedDeckChanges()) {
    return true;
  }

  if (state.selectedDeckId === nextDeckID) {
    return true;
  }

  return confirmDiscardDeckChanges();
}

/** @description Determine if it's safe to switch card selection, prompting if needed. */
function canSwitchCardSelection(nextCardID) {
  if (state.activeDeckId === null || !hasUnsavedCardChanges()) {
    return true;
  }

  if (state.selectedCardId === nextCardID) {
    return true;
  }

  return confirmDiscardCardChanges();
}

/** @description Determine if it's safe to leave the current editing context. */
function canLeaveCurrentEditingContext() {
  if (state.activeDeckId !== null) {
    if (!hasUnsavedCardChanges()) {
      return true;
    }
    return confirmDiscardCardChanges();
  }

  if (hasUnsavedDeckChanges()) {
    return confirmDiscardDeckChanges();
  }

  return true;
}

/** @description Leave the card editor, optionally confirming unsaved changes. */
function maybeLeaveCardEditor() {
  if (!hasUnsavedCardChanges()) {
    returnToDeckList();
    return;
  }

  if (confirmDiscardCardChanges()) {
    returnToDeckList();
  }
}

/** @description Cancel deck editing, optionally confirming unsaved changes. */
function maybeCancelDeckEdit() {
  if (!hasUnsavedDeckChanges()) {
    resetDeckForm();
    renderDecks();
    setStatus(DECK_CREATE_STATUS);
    els.deckForm.name.focus();
    return;
  }

  if (confirmDiscardDeckChanges()) {
    resetDeckForm();
    renderDecks();
    setStatus(DECK_CREATE_STATUS);
    els.deckForm.name.focus();
  }
}

/** @description Cancel card editing, optionally confirming unsaved changes. */
function maybeCancelCardEdit() {
  if (!hasUnsavedCardChanges()) {
    setAddCardMode(currentDeck());
    return;
  }

  if (confirmDiscardCardChanges()) {
    setAddCardMode(currentDeck());
  }
}

/** @description Handle creating or updating a card from the card form (async). */
async function handleCardSubmit() {
  const deck = currentDeck();
  if (!deck) {
    showToast("Select a deck first", "error");
    return;
  }

  if (!commitPendingCardTag()) {
    focusCardTagControl();
    return;
  }

  const prompt = els.cardForm.prompt.value.trim();
  const answer = els.cardForm.answer.value.trim();
  if (!prompt || !answer) {
    showToast("Prompt and answer are required", "error");
    if (!prompt) {
      els.cardForm.prompt.focus();
    } else {
      els.cardForm.answer.focus();
    }
    return;
  }

  if (!validateCardTagsForDeck(deck)) {
    focusCardTagControl();
    return;
  }

  try {
    const wasEditing = Boolean(state.editingCardId);
    if (state.editingCardId) {
      const existingCard = selectedCard();
      const updated = await fetchJSON(`/cards/${state.editingCardId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(buildCardPayload(existingCard)),
      });
      state.cardsByDeck.set(
        deck.ID,
        currentCards().map((card) => (card.ID === updated.ID ? updated : card)),
      );
      state.selectedCardId = updated.ID;
      showToast("Card saved");
    } else {
      const created = await fetchJSON(`/decks/${deck.ID}/cards`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(buildCardPayload()),
      });
      state.cardsByDeck.set(deck.ID, [...currentCards(), created]);
      state.selectedCardId = created.ID;
      showToast("Card added");
    }

    resetCardForm();
    renderCards();
    updateStats();
    setStatus(`Editing deck ${deck.Name} · ${currentCards().length} card${currentCards().length === 1 ? "" : "s"}`);
    if (!wasEditing) {
      els.cardForm.prompt.focus();
    }
  } catch (error) {
    setStatus(error.message);
    showToast(error.message, "error");
  }
}

/** @description Delete the currently selected card after confirmation (async). */
async function handleCardDelete() {
  const card = selectedCard();
  if (!card) {
    showToast("Select a card first", "error");
    return;
  }
  if (!window.confirm(`Delete "${card.Prompt}"? This cannot be undone.`)) {
    return;
  }

  try {
    const cards = currentCards();
    const deletedIndex = cards.findIndex((item) => item.ID === card.ID);
    const remainingCards = cards.filter((item) => item.ID !== card.ID);
    const nextCard = remainingCards[Math.max(0, deletedIndex - 1)] || remainingCards[deletedIndex] || null;

    await apiFetch(`/cards/${card.ID}`, { method: "DELETE" });
    state.cardsByDeck.set(state.activeDeckId, remainingCards);

    if (nextCard) {
      populateCardForm(nextCard);
      renderCards();
      els.cardForm.prompt.focus();
    } else {
      resetCardForm();
      renderCards();
      els.cardForm.prompt.focus();
    }

    updateStats();
    showToast("Card deleted");
  } catch (error) {
    setStatus(error.message);
    showToast(error.message, "error");
  }
}

/** @description Export the selected deck as a downloadable JSON file (async). */
async function handleExport() {
  const deckID = state.activeDeckId || state.selectedDeckId;
  if (!deckID) {
    showToast("Select a deck first", "error");
    return;
  }

  try {
    const response = await apiFetch(`/decks/${deckID}/export`);
    const blob = await response.blob();
    const deck = state.decks.find((item) => item.ID === deckID);
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `${(deck?.Name || "deck").toLowerCase().replace(/\s+/g, "-") || "deck"}.json`;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
    showToast("Deck exported");
  } catch (error) {
    setStatus(error.message);
    showToast(error.message, "error");
  }
}

/** @description Handle a file import input change, upload and process the file (async). */
async function handleImportFile(event) {
  const [file] = event.target.files || [];
  const importContext = state.importContext;
  clearImportContext();
  if (!file) {
    restoreBaseStatus();
    return;
  }

  if (shouldNormalizeImportContext(importContext)) {
    await showStudyView({ focusTarget: "none" });
  }

  let importFailed = false;
  try {
    const text = await file.text();
    const response = await fetchJSON("/import?mode=merge", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: text,
    });
    const importedDeck = response.deck || response.Deck;
    if (!importedDeck?.ID) {
      throw new Error("Import response did not include a deck");
    }
    await loadDecks();
    applyImportedDeckSelection(importedDeck.ID, importContext);
    showToast(`${importedDeck.Name} imported`);
  } catch (error) {
    importFailed = true;
    setStatus(`Import failed: ${error.message}`);
    showToast(error.message, "error");
  } finally {
    event.target.value = "";
    if (!importFailed) {
      restoreBaseStatus();
    }
  }
}

/** @description Heuristic to determine whether an element is a typing target. */
function isTypingTarget(target) {
  if (!target) {
    return false;
  }
  const tagName = target.tagName;
  return target.isContentEditable || tagName === "INPUT" || tagName === "TEXTAREA" || tagName === "SELECT";
}

/** @description Focus the appropriate new-item form (deck or card) depending on context. */
function focusNewForm() {
  if (state.activeDeckId) {
    resetCardForm();
    els.cardForm.prompt.focus();
  } else {
    resetDeckForm();
    els.deckForm.name.focus();
  }
}

/** @description Trigger the edit shortcut: edit selected card or deck. */
function triggerEditShortcut() {
  if (state.view === "session") {
    return;
  }
  if (state.activeDeckId) {
    const card = selectedCard();
    if (card) {
      populateCardForm(card);
      els.cardForm.prompt.focus();
    }
    return;
  }

  const deck = state.decks.find((item) => item.ID === state.selectedDeckId);
  if (deck) {
    populateDeckForm(deck);
    els.deckForm.name.focus();
  }
}

/** @description Trigger the delete shortcut: delete selected card or deck. */
function triggerDeleteShortcut() {
  if (state.view === "session") {
    return;
  }
  if (state.activeDeckId) {
    void handleCardDelete();
  } else {
    void handleDeckDelete();
  }
}

/** @description Global Escape handler for various UI contexts. */
function handleEscape() {
  if (state.aboutOpen) {
    closeAboutDialog();
    return;
  }

  if (state.view === "session") {
    if (state.session?.mode === "review") {
      if (state.session.phase === "summary") {
        leaveSessionToStudy("review");
      } else {
        showReviewSummary(false);
      }
      return;
    }
    leaveSessionToStudy();
    return;
  }

  if (state.activeDeckId) {
    if (state.editingCardId) {
      maybeCancelCardEdit();
      return;
    }
    returnToDeckList();
    return;
  }

  if (state.editingDeckId || hasUnsavedDeckChanges()) {
    maybeCancelDeckEdit();
    return;
  }

  if (state.view === "study") {
    handleStudyEscape();
  }
}

/** @description Exit review mode for the current deck (async). */
async function handleExitReview() {
  if (!state.session || state.session.mode !== "review") {
    return;
  }

  const deck = getSessionDeck();
  if (!deck) {
    leaveSessionToStudy("review");
    return;
  }

  try {
    const updatedDeck = await fetchJSON(`/decks/${deck.ID}/exit-review`, { method: "POST" });
    state.decks = state.decks.map((item) => (item.ID === updatedDeck.ID ? updatedDeck : item));
    await ensureCardsLoaded(deck.ID);
    await loadDueCountForDeck(deck.ID);
    showToast(`Exited review for ${updatedDeck.Name}`);
    leaveSessionToStudy("review");
  } catch (error) {
    setStatus(error.message);
    showToast(error.message, "error");
  }
}

/** @description Helper to defer Escape handling until keyup event. */
function handleEscapeKey(event) {
  event.preventDefault();
  handleEscape();
}

/** @description Bind UI event handlers to DOM elements. */
function bindEvents() {
  els.actions.addNamespace.addEventListener("click", () => addDeckNamespaceRow());
  els.actions.clearDeck.addEventListener("click", () => {
    resetDeckForm();
    renderDecks();
    setStatus(DECK_CREATE_STATUS);
  });
  els.actions.clearCard.addEventListener("click", () => {
    resetCardForm();
    renderCards();
  });
  els.actions.deleteCard.addEventListener("click", () => void handleCardDelete());
  els.deckForm.submit.addEventListener("click", () => void handleDeckSubmit());
  els.cardForm.submit.addEventListener("click", () => void handleCardSubmit());
  els.actions.delete.addEventListener("click", () => void triggerDeleteShortcut());
  els.actions.create.addEventListener("click", () => {
    if (!canLeaveCurrentEditingContext()) {
      restoreFocusAfterBlockedLeave();
      return;
    }
    startCreateDeckMode();
    focusNewForm();
  });
  els.actions.import.addEventListener("click", () => {
    if (!canLeaveCurrentEditingContext()) {
      restoreFocusAfterBlockedLeave();
      return;
    }
    state.importContext = captureImportContext();
    state.importPromptActive = true;
    setStatus("Select a single deck JSON file to import");
    els.importFile.click();
  });
  els.actions.export.addEventListener("click", () => void handleExport());
  els.actions.exportDeck.addEventListener("click", () => void handleExport());
  els.actions.about.addEventListener("click", () => openAboutDialog());
  els.importFile.addEventListener("change", (event) => void handleImportFile(event));
  els.about.close.addEventListener("click", () => closeAboutDialog());
  els.about.dismiss.addEventListener("click", () => closeAboutDialog());
  els.about.overlay.addEventListener("click", (event) => {
    if (event.target === els.about.overlay) {
      closeAboutDialog();
    }
  });
  els.breadcrumbs.home.addEventListener("click", () => {
    if (state.view === "study") {
      return;
    }
    if (!canLeaveCurrentEditingContext()) {
      restoreFocusAfterBlockedLeave();
      return;
    }
    if (state.view === "session") {
      void showStudyView({ focusTarget: "none" });
      return;
    }
    showDeckView();
  });
  els.actions.study.addEventListener("click", () => {
    if (!canLeaveCurrentEditingContext()) {
      restoreFocusAfterBlockedLeave();
      return;
    }
    void showStudyView();
  });
  els.session.back.addEventListener("click", () => leaveSessionToStudy());
  els.session.missed.addEventListener("click", () => void submitReview(false));
  els.session.gotIt.addEventListener("click", () => void submitReview(true));
  els.session.prev.addEventListener("click", () => navigateSession(-1));
  els.session.next.addEventListener("click", () => navigateSession(1));
  els.session.summaryBack.addEventListener("click", () => leaveSessionToStudy("review"));
  els.session.summaryExitReview.addEventListener("click", () => void handleExitReview());
  els.study.search.addEventListener("input", (event) => {
    state.studySearch = event.target.value;
    renderStudyHome();
  });

  els.cardForm.tagWrap.addEventListener("click", () => focusCardTagControl());
  els.cardForm.tagInput.addEventListener("keydown", (event) => {
    if (event.key === "Enter" || event.key === ",") {
      event.preventDefault();
      commitPendingCardTag();
    } else if (event.key === "Tab" && !event.shiftKey) {
      event.preventDefault();
      if (els.cardForm.tagInput.value.trim() && !commitPendingCardTag()) {
        return;
      }
      els.cardForm.submit.focus();
    } else if (event.key === "Backspace" && !els.cardForm.tagInput.value && state.cardTags.length) {
      state.cardTags = state.cardTags.slice(0, -1);
      renderCardTags();
    }
  });
  els.cardForm.tagInput.addEventListener("blur", () => {
    if (els.cardForm.tagInput.value.trim()) {
      commitPendingCardTag();
    }
  });

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      event.preventDefault();
      return;
    }

    if (state.aboutOpen) {
      return;
    }

    if (isTypingTarget(event.target)) {
      return;
    }

    if (state.view === "session" && state.session) {
      if (state.session.phase === "summary") {
        if (event.key === "Enter") {
          event.preventDefault();
          leaveSessionToStudy(state.session.mode === "review" ? "review" : "refresher");
        }
        return;
      }
      if (event.key === " " || event.key === "Enter") {
        event.preventDefault();
        revealSessionCard();
        return;
      }
      if (state.session.mode !== "refresher") {
        return;
      }
      if ((event.key === "j" || event.key === "ArrowDown") && state.session.revealed) {
        event.preventDefault();
        navigateSession(1);
        return;
      }
      if ((event.key === "k" || event.key === "ArrowUp") && state.session.revealed) {
        event.preventDefault();
        navigateSession(-1);
        return;
      }
      return;
    }

    if (state.view === "study" && event.key === "/") {
      event.preventDefault();
      focusStudySearch();
      return;
    }

    if (event.key === "n") {
      event.preventDefault();
      focusNewForm();
      return;
    }
    if (event.key === "e") {
      event.preventDefault();
      triggerEditShortcut();
      return;
    }
    if (event.key === "d" || event.key === "Delete") {
      event.preventDefault();
      triggerDeleteShortcut();
      return;
    }
  });

  document.addEventListener("keyup", (event) => {
    if (event.key === "Escape") {
      // Let the keypress finish before opening browser-native discard prompts.
      handleEscapeKey(event);
    }
  });

  window.addEventListener("beforeunload", (event) => {
    if (!hasUnsavedDeckChanges() && !hasUnsavedCardChanges()) {
      return;
    }
    event.preventDefault();
    event.returnValue = "";
  });

  window.addEventListener("focus", () => {
    if (!state.importPromptActive) {
      return;
    }

    window.requestAnimationFrame(() => {
      if (state.importPromptActive && !els.importFile.value) {
        clearImportContext();
        restoreBaseStatus();
      }
    });
  });
}

/** @description Initialize the app: bind events, reset forms, and load decks (async). */
async function init() {
  bindEvents();
  resetDeckForm();
  resetCardForm();
  setStatus("Loading decks...");
  try {
    await loadDecks();
    if (state.decks.length === 0) {
      showDeckView();
      focusNewForm();
    } else {
      await showStudyView({ focusTarget: "selected" });
    }
  } catch (error) {
    setStatus(error.message);
    showToast(error.message, "error");
  }
}

void init();
