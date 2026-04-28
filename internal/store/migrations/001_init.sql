CREATE TABLE IF NOT EXISTS decks (
    id                   TEXT PRIMARY KEY,
    name                 TEXT NOT NULL,
    description          TEXT NOT NULL DEFAULT '',
    tag_namespaces       TEXT NOT NULL DEFAULT '{}',
    default_reverse_mode TEXT NOT NULL DEFAULT 'prompt_first',
    review_active        INTEGER NOT NULL DEFAULT 0,
    created_at           DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS cards (
    id               TEXT PRIMARY KEY,
    deck_id          TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    prompt           TEXT NOT NULL,
    answer           TEXT NOT NULL,
    notes            TEXT NOT NULL DEFAULT '',
    tags             TEXT NOT NULL DEFAULT '[]',
    source           TEXT NOT NULL DEFAULT 'manual',
    created_at       DATETIME NOT NULL,
    interval         INTEGER NOT NULL DEFAULT 0,
    ease_factor      REAL    NOT NULL DEFAULT 2.5,
    repetitions      INTEGER NOT NULL DEFAULT 0,
    due_date         DATETIME NOT NULL,
    last_reviewed_at DATETIME
);

CREATE TABLE IF NOT EXISTS review_log (
    id              TEXT PRIMARY KEY,
    card_id         TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    reviewed_at     DATETIME NOT NULL,
    grade           INTEGER NOT NULL,
    interval_after  INTEGER NOT NULL,
    ease_after      REAL    NOT NULL
);

CREATE TABLE IF NOT EXISTS learning_paths (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    deck_ids    TEXT NOT NULL DEFAULT '[]',
    created_at  DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cards_deck_id ON cards(deck_id);
CREATE INDEX IF NOT EXISTS idx_cards_due_date ON cards(due_date);
CREATE INDEX IF NOT EXISTS idx_review_log_card_id ON review_log(card_id);
