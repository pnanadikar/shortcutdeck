package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/pnanadikar/shortcutdeck/internal/model"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store is the persistence interface for decks, cards, review history,
// and learning paths.
type Store interface {
	CreateDeck(d model.Deck) (model.Deck, error)
	GetDeck(id string) (model.Deck, error)
	ListDecks() ([]model.Deck, error)
	UpdateDeck(d model.Deck) (model.Deck, error)
	DeleteDeck(id string) error

	CreateCard(c model.Card) (model.Card, error)
	GetCard(id string) (model.Card, error)
	ListCards(deckID string, tags []string) ([]model.Card, error)
	GetDueCards(deckID string, today time.Time) ([]model.Card, error)
	UpdateCard(c model.Card) (model.Card, error)
	DeleteCard(id string) error

	LogReview(entry model.ReviewEntry) error
	GetReviewLog(cardID string) ([]model.ReviewEntry, error)

	CreateLearningPath(p model.LearningPath) (model.LearningPath, error)
	GetLearningPath(id string) (model.LearningPath, error)
	ListLearningPaths() ([]model.LearningPath, error)

	ExportDeck(deckID string) (ExportPayload, error)
	ImportDeck(payload ExportPayload, mode string) error
}

// ExportPayload is the JSON structure used for deck export/import.
type ExportPayload struct {
	Deck         model.Deck          `json:"deck"`
	Cards        []model.Card        `json:"cards"`
	LearningPath *model.LearningPath `json:"learning_path,omitempty"`
}

// SQLiteStore is the SQLite-backed implementation of Store.
type SQLiteStore struct {
	db *sql.DB
}

// New opens the SQLite database, runs migrations, and returns a ready store.
func New(dbPath string) (Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := runMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) CreateDeck(d model.Deck) (model.Deck, error) {
	if d.ID == "" {
		d.ID = newID()
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}
	if d.TagNamespaces == nil {
		d.TagNamespaces = model.TagNamespaces{}
	}
	if d.DefaultReverseMode == "" {
		d.DefaultReverseMode = model.PromptFirst
	}

	tagNamespacesJSON, err := marshalJSON(d.TagNamespaces)
	if err != nil {
		return model.Deck{}, fmt.Errorf("marshal tag namespaces: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO decks (
			id, name, description, tag_namespaces, default_reverse_mode, review_active, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		d.ID,
		d.Name,
		d.Description,
		tagNamespacesJSON,
		string(d.DefaultReverseMode),
		boolToInt(d.ReviewActive),
		formatTime(d.CreatedAt),
	)
	if err != nil {
		return model.Deck{}, fmt.Errorf("create deck: %w", err)
	}

	return d, nil
}

func (s *SQLiteStore) GetDeck(id string) (model.Deck, error) {
	row := s.db.QueryRow(`
		SELECT id, name, description, tag_namespaces, default_reverse_mode, review_active, created_at
		FROM decks
		WHERE id = ?`, id)

	deck, err := scanDeck(row)
	if err != nil {
		return model.Deck{}, fmt.Errorf("get deck: %w", err)
	}

	return deck, nil
}

func (s *SQLiteStore) ListDecks() ([]model.Deck, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, tag_namespaces, default_reverse_mode, review_active, created_at
		FROM decks
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list decks: %w", err)
	}
	defer closeRows(rows)

	var decks []model.Deck
	for rows.Next() {
		deck, err := scanDeck(rows)
		if err != nil {
			return nil, fmt.Errorf("scan deck: %w", err)
		}
		decks = append(decks, deck)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate decks: %w", err)
	}

	return decks, nil
}

func (s *SQLiteStore) UpdateDeck(d model.Deck) (model.Deck, error) {
	tagNamespacesJSON, err := marshalJSON(d.TagNamespaces)
	if err != nil {
		return model.Deck{}, fmt.Errorf("marshal tag namespaces: %w", err)
	}
	if d.DefaultReverseMode == "" {
		d.DefaultReverseMode = model.PromptFirst
	}

	result, err := s.db.Exec(`
		UPDATE decks
		SET name = ?, description = ?, tag_namespaces = ?, default_reverse_mode = ?, review_active = ?
		WHERE id = ?`,
		d.Name,
		d.Description,
		tagNamespacesJSON,
		string(d.DefaultReverseMode),
		boolToInt(d.ReviewActive),
		d.ID,
	)
	if err != nil {
		return model.Deck{}, fmt.Errorf("update deck: %w", err)
	}

	if err := requireRowsAffected(result); err != nil {
		return model.Deck{}, fmt.Errorf("update deck: %w", err)
	}

	return s.GetDeck(d.ID)
}

func (s *SQLiteStore) DeleteDeck(id string) error {
	result, err := s.db.Exec(`DELETE FROM decks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete deck: %w", err)
	}

	return requireRowsAffected(result)
}

func (s *SQLiteStore) CreateCard(c model.Card) (model.Card, error) {
	if c.ID == "" {
		c.ID = newID()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	if c.Source == "" {
		c.Source = "manual"
	}
	if c.EaseFactor == 0 {
		c.EaseFactor = 2.5
	}
	if c.DueDate.IsZero() {
		c.DueDate = normalizeDate(time.Now().UTC())
	}
	if c.Tags == nil {
		c.Tags = []string{}
	}

	tagsJSON, err := marshalJSON(c.Tags)
	if err != nil {
		return model.Card{}, fmt.Errorf("marshal tags: %w", err)
	}

	lastReviewedAt, err := nullableTime(c.LastReviewedAt)
	if err != nil {
		return model.Card{}, fmt.Errorf("format last reviewed at: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO cards (
			id, deck_id, prompt, answer, notes, tags, source, created_at,
			interval, ease_factor, repetitions, due_date, last_reviewed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID,
		c.DeckID,
		c.Prompt,
		c.Answer,
		c.Notes,
		tagsJSON,
		c.Source,
		formatTime(c.CreatedAt),
		c.Interval,
		c.EaseFactor,
		c.Repetitions,
		formatTime(normalizeDate(c.DueDate)),
		lastReviewedAt,
	)
	if err != nil {
		return model.Card{}, fmt.Errorf("create card: %w", err)
	}

	return s.GetCard(c.ID)
}

func (s *SQLiteStore) GetCard(id string) (model.Card, error) {
	row := s.db.QueryRow(`
		SELECT id, deck_id, prompt, answer, notes, tags, source, created_at,
		       interval, ease_factor, repetitions, due_date, last_reviewed_at
		FROM cards
		WHERE id = ?`, id)

	card, err := scanCard(row)
	if err != nil {
		return model.Card{}, fmt.Errorf("get card: %w", err)
	}

	return card, nil
}

func (s *SQLiteStore) ListCards(deckID string, tags []string) ([]model.Card, error) {
	rows, err := s.db.Query(`
		SELECT id, deck_id, prompt, answer, notes, tags, source, created_at,
		       interval, ease_factor, repetitions, due_date, last_reviewed_at
		FROM cards
		WHERE deck_id = ?
		ORDER BY tags, id`, deckID)
	if err != nil {
		return nil, fmt.Errorf("list cards: %w", err)
	}
	defer closeRows(rows)

	var cards []model.Card
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}
		if hasAllTags(card.Tags, tags) {
			cards = append(cards, card)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cards: %w", err)
	}

	return cards, nil
}

func (s *SQLiteStore) GetDueCards(deckID string, today time.Time) ([]model.Card, error) {
	rows, err := s.db.Query(`
		SELECT id, deck_id, prompt, answer, notes, tags, source, created_at,
		       interval, ease_factor, repetitions, due_date, last_reviewed_at
		FROM cards
		WHERE deck_id = ? AND due_date <= ?
		ORDER BY due_date, id`,
		deckID,
		formatTime(normalizeDate(today)),
	)
	if err != nil {
		return nil, fmt.Errorf("get due cards: %w", err)
	}
	defer closeRows(rows)

	var cards []model.Card
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, fmt.Errorf("scan due card: %w", err)
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate due cards: %w", err)
	}

	return cards, nil
}

func (s *SQLiteStore) UpdateCard(c model.Card) (model.Card, error) {
	tagsJSON, err := marshalJSON(c.Tags)
	if err != nil {
		return model.Card{}, fmt.Errorf("marshal tags: %w", err)
	}
	lastReviewedAt, err := nullableTime(c.LastReviewedAt)
	if err != nil {
		return model.Card{}, fmt.Errorf("format last reviewed at: %w", err)
	}

	result, err := s.db.Exec(`
		UPDATE cards
		SET prompt = ?, answer = ?, notes = ?, tags = ?,
		    interval = ?, ease_factor = ?, repetitions = ?, due_date = ?, last_reviewed_at = ?
		WHERE id = ?`,
		c.Prompt,
		c.Answer,
		c.Notes,
		tagsJSON,
		c.Interval,
		c.EaseFactor,
		c.Repetitions,
		formatTime(normalizeDate(c.DueDate)),
		lastReviewedAt,
		c.ID,
	)
	if err != nil {
		return model.Card{}, fmt.Errorf("update card: %w", err)
	}

	if err := requireRowsAffected(result); err != nil {
		return model.Card{}, fmt.Errorf("update card: %w", err)
	}

	return s.GetCard(c.ID)
}

func (s *SQLiteStore) DeleteCard(id string) error {
	result, err := s.db.Exec(`DELETE FROM cards WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete card: %w", err)
	}

	return requireRowsAffected(result)
}

func (s *SQLiteStore) LogReview(entry model.ReviewEntry) error {
	if entry.ID == "" {
		entry.ID = newID()
	}
	if entry.ReviewedAt.IsZero() {
		entry.ReviewedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO review_log (
			id, card_id, reviewed_at, grade, interval_after, ease_after
		) VALUES (?, ?, ?, ?, ?, ?)`,
		entry.ID,
		entry.CardID,
		formatTime(entry.ReviewedAt),
		entry.Grade,
		entry.IntervalAfter,
		entry.EaseAfter,
	)
	if err != nil {
		return fmt.Errorf("log review: %w", err)
	}

	return nil
}

func (s *SQLiteStore) GetReviewLog(cardID string) ([]model.ReviewEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, card_id, reviewed_at, grade, interval_after, ease_after
		FROM review_log
		WHERE card_id = ?
		ORDER BY reviewed_at, id`, cardID)
	if err != nil {
		return nil, fmt.Errorf("get review log: %w", err)
	}
	defer closeRows(rows)

	var entries []model.ReviewEntry
	for rows.Next() {
		entry, err := scanReviewEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("scan review entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate review log: %w", err)
	}

	return entries, nil
}

func (s *SQLiteStore) CreateLearningPath(p model.LearningPath) (model.LearningPath, error) {
	if p.ID == "" {
		p.ID = newID()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}

	deckIDsJSON, err := marshalJSON(p.DeckIDs)
	if err != nil {
		return model.LearningPath{}, fmt.Errorf("marshal deck ids: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO learning_paths (id, name, description, deck_ids, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		p.ID,
		p.Name,
		p.Description,
		deckIDsJSON,
		formatTime(p.CreatedAt),
	)
	if err != nil {
		return model.LearningPath{}, fmt.Errorf("create learning path: %w", err)
	}

	return s.GetLearningPath(p.ID)
}

func (s *SQLiteStore) GetLearningPath(id string) (model.LearningPath, error) {
	row := s.db.QueryRow(`
		SELECT id, name, description, deck_ids, created_at
		FROM learning_paths
		WHERE id = ?`, id)

	path, err := scanLearningPath(row)
	if err != nil {
		return model.LearningPath{}, fmt.Errorf("get learning path: %w", err)
	}

	return path, nil
}

func (s *SQLiteStore) ListLearningPaths() ([]model.LearningPath, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, deck_ids, created_at
		FROM learning_paths
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list learning paths: %w", err)
	}
	defer closeRows(rows)

	var paths []model.LearningPath
	for rows.Next() {
		path, err := scanLearningPath(rows)
		if err != nil {
			return nil, fmt.Errorf("scan learning path: %w", err)
		}
		paths = append(paths, path)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate learning paths: %w", err)
	}

	return paths, nil
}

func (s *SQLiteStore) ExportDeck(deckID string) (ExportPayload, error) {
	deck, err := s.GetDeck(deckID)
	if err != nil {
		return ExportPayload{}, fmt.Errorf("export deck: %w", err)
	}

	cards, err := s.ListCards(deckID, nil)
	if err != nil {
		return ExportPayload{}, fmt.Errorf("export deck cards: %w", err)
	}

	return ExportPayload{
		Deck:  deck,
		Cards: cards,
	}, nil
}

func (s *SQLiteStore) ImportDeck(payload ExportPayload, mode string) error {
	switch mode {
	case "merge", "replace":
	default:
		return fmt.Errorf("import deck: unsupported mode %q", mode)
	}

	return withTx(s.db, func(tx *sql.Tx) error {
		if err := upsertDeckTx(tx, payload.Deck); err != nil {
			return err
		}

		if mode == "replace" {
			if _, err := tx.Exec(`DELETE FROM cards WHERE deck_id = ?`, payload.Deck.ID); err != nil {
				return fmt.Errorf("replace deck cards: %w", err)
			}
		}

		for _, card := range payload.Cards {
			card.DeckID = payload.Deck.ID
			if err := upsertCardTx(tx, card); err != nil {
				return err
			}
		}

		if payload.LearningPath != nil {
			if err := upsertLearningPathTx(tx, *payload.LearningPath); err != nil {
				return err
			}
		}

		return nil
	})
}

func runMigrations(db *sql.DB) error {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		sqlBytes, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("run migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func upsertDeckTx(tx *sql.Tx, d model.Deck) error {
	if d.ID == "" {
		d.ID = newID()
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}
	if d.TagNamespaces == nil {
		d.TagNamespaces = model.TagNamespaces{}
	}
	if d.DefaultReverseMode == "" {
		d.DefaultReverseMode = model.PromptFirst
	}

	tagNamespacesJSON, err := marshalJSON(d.TagNamespaces)
	if err != nil {
		return fmt.Errorf("marshal tag namespaces: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO decks (id, name, description, tag_namespaces, default_reverse_mode, review_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			tag_namespaces = excluded.tag_namespaces,
			default_reverse_mode = excluded.default_reverse_mode,
			review_active = excluded.review_active`,
		d.ID,
		d.Name,
		d.Description,
		tagNamespacesJSON,
		string(d.DefaultReverseMode),
		boolToInt(d.ReviewActive),
		formatTime(d.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert deck: %w", err)
	}

	return nil
}

func upsertCardTx(tx *sql.Tx, c model.Card) error {
	if c.ID == "" {
		c.ID = newID()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	if c.Source == "" {
		c.Source = "manual"
	}
	if c.EaseFactor == 0 {
		c.EaseFactor = 2.5
	}
	if c.Tags == nil {
		c.Tags = []string{}
	}
	if c.DueDate.IsZero() {
		c.DueDate = normalizeDate(time.Now().UTC())
	}

	tagsJSON, err := marshalJSON(c.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}
	lastReviewedAt, err := nullableTime(c.LastReviewedAt)
	if err != nil {
		return fmt.Errorf("format last reviewed at: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO cards (
			id, deck_id, prompt, answer, notes, tags, source, created_at,
			interval, ease_factor, repetitions, due_date, last_reviewed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			deck_id = excluded.deck_id,
			prompt = excluded.prompt,
			answer = excluded.answer,
			notes = excluded.notes,
			tags = excluded.tags,
			source = excluded.source,
			interval = excluded.interval,
			ease_factor = excluded.ease_factor,
			repetitions = excluded.repetitions,
			due_date = excluded.due_date,
			last_reviewed_at = excluded.last_reviewed_at`,
		c.ID,
		c.DeckID,
		c.Prompt,
		c.Answer,
		c.Notes,
		tagsJSON,
		c.Source,
		formatTime(c.CreatedAt),
		c.Interval,
		c.EaseFactor,
		c.Repetitions,
		formatTime(normalizeDate(c.DueDate)),
		lastReviewedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert card: %w", err)
	}

	return nil
}

func upsertLearningPathTx(tx *sql.Tx, p model.LearningPath) error {
	if p.ID == "" {
		p.ID = newID()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}

	deckIDsJSON, err := marshalJSON(p.DeckIDs)
	if err != nil {
		return fmt.Errorf("marshal deck ids: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO learning_paths (id, name, description, deck_ids, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			deck_ids = excluded.deck_ids`,
		p.ID,
		p.Name,
		p.Description,
		deckIDsJSON,
		formatTime(p.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert learning path: %w", err)
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanDeck(s scanner) (model.Deck, error) {
	var (
		deck              model.Deck
		tagNamespacesJSON string
		defaultMode       string
		reviewActive      int
		createdAt         string
	)

	if err := s.Scan(
		&deck.ID,
		&deck.Name,
		&deck.Description,
		&tagNamespacesJSON,
		&defaultMode,
		&reviewActive,
		&createdAt,
	); err != nil {
		return model.Deck{}, err
	}

	if err := unmarshalJSON(tagNamespacesJSON, &deck.TagNamespaces); err != nil {
		return model.Deck{}, err
	}

	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return model.Deck{}, err
	}

	deck.DefaultReverseMode = model.ReverseMode(defaultMode)
	deck.ReviewActive = reviewActive != 0
	deck.CreatedAt = parsedCreatedAt

	return deck, nil
}

func scanCard(s scanner) (model.Card, error) {
	var (
		card           model.Card
		tagsJSON       string
		createdAt      string
		dueDate        string
		lastReviewedAt sql.NullString
	)

	if err := s.Scan(
		&card.ID,
		&card.DeckID,
		&card.Prompt,
		&card.Answer,
		&card.Notes,
		&tagsJSON,
		&card.Source,
		&createdAt,
		&card.Interval,
		&card.EaseFactor,
		&card.Repetitions,
		&dueDate,
		&lastReviewedAt,
	); err != nil {
		return model.Card{}, err
	}

	if err := unmarshalJSON(tagsJSON, &card.Tags); err != nil {
		return model.Card{}, err
	}

	var err error
	card.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return model.Card{}, err
	}
	card.DueDate, err = parseTime(dueDate)
	if err != nil {
		return model.Card{}, err
	}
	if lastReviewedAt.Valid {
		card.LastReviewedAt, err = parseTime(lastReviewedAt.String)
		if err != nil {
			return model.Card{}, err
		}
	}

	return card, nil
}

func scanReviewEntry(s scanner) (model.ReviewEntry, error) {
	var (
		entry      model.ReviewEntry
		reviewedAt string
	)

	if err := s.Scan(
		&entry.ID,
		&entry.CardID,
		&reviewedAt,
		&entry.Grade,
		&entry.IntervalAfter,
		&entry.EaseAfter,
	); err != nil {
		return model.ReviewEntry{}, err
	}

	parsedReviewedAt, err := parseTime(reviewedAt)
	if err != nil {
		return model.ReviewEntry{}, err
	}
	entry.ReviewedAt = parsedReviewedAt

	return entry, nil
}

func scanLearningPath(s scanner) (model.LearningPath, error) {
	var (
		path        model.LearningPath
		deckIDsJSON string
		createdAt   string
	)

	if err := s.Scan(
		&path.ID,
		&path.Name,
		&path.Description,
		&deckIDsJSON,
		&createdAt,
	); err != nil {
		return model.LearningPath{}, err
	}

	if err := unmarshalJSON(deckIDsJSON, &path.DeckIDs); err != nil {
		return model.LearningPath{}, err
	}

	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return model.LearningPath{}, err
	}
	path.CreatedAt = parsedCreatedAt

	return path, nil
}

func withTx(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func marshalJSON(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func unmarshalJSON(data string, v any) error {
	if data == "" {
		data = "{}"
		switch v.(type) {
		case *[]string:
			data = "[]"
		}
	}

	return json.Unmarshal([]byte(data), v)
}

func newID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic("generate id: " + err.Error())
	}

	return hex.EncodeToString(buf)
}

func normalizeDate(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(raw string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, raw)
}

func nullableTime(t time.Time) (any, error) {
	if t.IsZero() {
		return nil, nil
	}

	return formatTime(t), nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}

	return 0
}

func hasAllTags(cardTags []string, required []string) bool {
	if len(required) == 0 {
		return true
	}

	set := make(map[string]struct{}, len(cardTags))
	for _, tag := range cardTags {
		set[tag] = struct{}{}
	}

	for _, tag := range required {
		if _, ok := set[tag]; !ok {
			return false
		}
	}

	return true
}

func requireRowsAffected(result sql.Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func closeRows(rows *sql.Rows) {
	_ = rows.Close()
}

func init() {
	if strings.TrimSpace(strings.Join(mustReadMigrationNames(), "")) == "" {
		panic("store migrations embed is empty")
	}
}

func mustReadMigrationNames() []string {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			panic(err)
		}
		return nil
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names
}
