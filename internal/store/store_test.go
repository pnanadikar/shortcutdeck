package store

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/pnanadikar/shortcutdeck/internal/model"
)

func TestCreateDeckRoundTrip(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	deck, err := s.CreateDeck(model.Deck{
		Name:               "Vim",
		Description:        "Editor shortcuts",
		TagNamespaces:      model.TagNamespaces{"mode": {"normal", "insert"}},
		DefaultReverseMode: model.Both,
		ReviewActive:       true,
	})
	if err != nil {
		t.Fatalf("CreateDeck returned error: %v", err)
	}

	got, err := s.GetDeck(deck.ID)
	if err != nil {
		t.Fatalf("GetDeck returned error: %v", err)
	}

	if got.Name != deck.Name || got.Description != deck.Description {
		t.Fatalf("unexpected deck round-trip: %#v", got)
	}
	if got.DefaultReverseMode != model.Both {
		t.Fatalf("unexpected reverse mode: %q", got.DefaultReverseMode)
	}
	if !got.ReviewActive {
		t.Fatalf("expected review_active to round-trip")
	}
	if len(got.TagNamespaces["mode"]) != 2 {
		t.Fatalf("unexpected tag namespaces: %#v", got.TagNamespaces)
	}
}

func TestCreateCardSetsDefaultSchedulingState(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	deck := mustCreateDeck(t, s, "Vim")

	before := normalizeDate(time.Now().UTC())
	card, err := s.CreateCard(model.Card{
		DeckID: deck.ID,
		Prompt: "Delete to end of line",
		Answer: "d$",
	})
	if err != nil {
		t.Fatalf("CreateCard returned error: %v", err)
	}

	if card.Interval != 0 || card.Repetitions != 0 {
		t.Fatalf("unexpected default scheduling state: %#v", card)
	}
	if card.EaseFactor != 2.5 {
		t.Fatalf("unexpected default ease factor: %v", card.EaseFactor)
	}
	if !card.DueDate.Equal(before) {
		t.Fatalf("unexpected due date: got %s want %s", card.DueDate, before)
	}
	if card.Source != "manual" {
		t.Fatalf("unexpected source default: %q", card.Source)
	}
}

func TestGetDueCardsFiltersCorrectly(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	deck := mustCreateDeck(t, s, "Vim")
	today := normalizeDate(time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC))

	_, _ = mustCreateCard(t, s, model.Card{
		DeckID:  deck.ID,
		Prompt:  "Due yesterday",
		Answer:  "h",
		DueDate: today.AddDate(0, 0, -1),
	})
	_, dueToday := mustCreateCard(t, s, model.Card{
		DeckID:  deck.ID,
		Prompt:  "Due today",
		Answer:  "j",
		DueDate: today,
	})
	_, _ = mustCreateCard(t, s, model.Card{
		DeckID:  deck.ID,
		Prompt:  "Due tomorrow",
		Answer:  "k",
		DueDate: today.AddDate(0, 0, 1),
	})

	dueCards, err := s.GetDueCards(deck.ID, today)
	if err != nil {
		t.Fatalf("GetDueCards returned error: %v", err)
	}

	if len(dueCards) != 2 {
		t.Fatalf("expected 2 due cards, got %d", len(dueCards))
	}
	if dueCards[1].ID != dueToday.ID {
		t.Fatalf("unexpected due card ordering/result: %#v", dueCards)
	}
}

func TestDeleteDeckCascadesToCardsAndReviewLog(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	deck := mustCreateDeck(t, s, "Vim")
	_, card := mustCreateCard(t, s, model.Card{
		DeckID: deck.ID,
		Prompt: "Delete line",
		Answer: "dd",
	})

	err := s.LogReview(model.ReviewEntry{
		CardID:        card.ID,
		ReviewedAt:    time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC),
		Grade:         5,
		IntervalAfter: 3,
		EaseAfter:     2.6,
	})
	if err != nil {
		t.Fatalf("LogReview returned error: %v", err)
	}

	if err := s.DeleteDeck(deck.ID); err != nil {
		t.Fatalf("DeleteDeck returned error: %v", err)
	}

	_, err = s.GetCard(card.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after cascade delete, got %v", err)
	}

	entries, err := s.GetReviewLog(card.ID)
	if err != nil {
		t.Fatalf("GetReviewLog returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty review log after cascade delete, got %#v", entries)
	}
}

func TestLogReviewIsAppendOnly(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	deck := mustCreateDeck(t, s, "Vim")
	_, card := mustCreateCard(t, s, model.Card{
		DeckID: deck.ID,
		Prompt: "Delete line",
		Answer: "dd",
	})

	firstReview := model.ReviewEntry{
		ID:            "review-1",
		CardID:        card.ID,
		ReviewedAt:    time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC),
		Grade:         3,
		IntervalAfter: 1,
		EaseAfter:     2.4,
	}
	secondReview := model.ReviewEntry{
		ID:            "review-2",
		CardID:        card.ID,
		ReviewedAt:    time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC),
		Grade:         5,
		IntervalAfter: 4,
		EaseAfter:     2.5,
	}

	if err := s.LogReview(firstReview); err != nil {
		t.Fatalf("LogReview(first) returned error: %v", err)
	}
	if err := s.LogReview(secondReview); err != nil {
		t.Fatalf("LogReview(second) returned error: %v", err)
	}

	entries, err := s.GetReviewLog(card.ID)
	if err != nil {
		t.Fatalf("GetReviewLog returned error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 review entries, got %d", len(entries))
	}
	if entries[0].ID != firstReview.ID || entries[1].ID != secondReview.ID {
		t.Fatalf("unexpected review log contents: %#v", entries)
	}
}

func TestImportDeckMergeUpsertsCardsWithoutTouchingOthers(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	deck := mustCreateDeck(t, s, "Vim")

	_, existing := mustCreateCard(t, s, model.Card{
		ID:      "card-1",
		DeckID:  deck.ID,
		Prompt:  "Original prompt",
		Answer:  "dd",
		DueDate: time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
	})
	_, untouched := mustCreateCard(t, s, model.Card{
		ID:      "card-2",
		DeckID:  deck.ID,
		Prompt:  "Untouched card",
		Answer:  "yy",
		DueDate: time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
	})

	err := s.ImportDeck(ExportPayload{
		Deck: model.Deck{
			ID:                 deck.ID,
			Name:               "Vim Updated",
			Description:        "Updated deck",
			TagNamespaces:      model.TagNamespaces{"mode": {"normal"}},
			DefaultReverseMode: model.AnswerFirst,
			CreatedAt:          deck.CreatedAt,
		},
		Cards: []model.Card{
			{
				ID:         existing.ID,
				DeckID:     deck.ID,
				Prompt:     "Updated prompt",
				Answer:     existing.Answer,
				CreatedAt:  existing.CreatedAt,
				DueDate:    existing.DueDate,
				EaseFactor: existing.EaseFactor,
				Source:     existing.Source,
				Tags:       []string{"editing"},
			},
			{
				ID:         "card-3",
				DeckID:     deck.ID,
				Prompt:     "New card",
				Answer:     "p",
				CreatedAt:  time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
				DueDate:    time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
				EaseFactor: 2.5,
			},
		},
	}, "merge")
	if err != nil {
		t.Fatalf("ImportDeck(merge) returned error: %v", err)
	}

	cards, err := s.ListCards(deck.ID, nil)
	if err != nil {
		t.Fatalf("ListCards returned error: %v", err)
	}
	if len(cards) != 3 {
		t.Fatalf("expected 3 cards after merge import, got %d", len(cards))
	}

	updated, err := s.GetCard(existing.ID)
	if err != nil {
		t.Fatalf("GetCard(updated) returned error: %v", err)
	}
	if updated.Prompt != "Updated prompt" {
		t.Fatalf("expected card to be upserted, got %#v", updated)
	}

	stillThere, err := s.GetCard(untouched.ID)
	if err != nil {
		t.Fatalf("GetCard(untouched) returned error: %v", err)
	}
	if stillThere.Prompt != untouched.Prompt {
		t.Fatalf("expected untouched card to remain unchanged, got %#v", stillThere)
	}
}

func TestImportDeckReplaceDeletesExistingCardsFirst(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	deck := mustCreateDeck(t, s, "Vim")

	_, doomed := mustCreateCard(t, s, model.Card{
		ID:      "card-old",
		DeckID:  deck.ID,
		Prompt:  "Old",
		Answer:  "dd",
		DueDate: time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
	})

	err := s.ImportDeck(ExportPayload{
		Deck: deck,
		Cards: []model.Card{
			{
				ID:         "card-new",
				DeckID:     deck.ID,
				Prompt:     "New",
				Answer:     "yy",
				CreatedAt:  time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
				DueDate:    time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
				EaseFactor: 2.5,
			},
		},
	}, "replace")
	if err != nil {
		t.Fatalf("ImportDeck(replace) returned error: %v", err)
	}

	_, err = s.GetCard(doomed.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected old card to be deleted, got %v", err)
	}

	cards, err := s.ListCards(deck.ID, nil)
	if err != nil {
		t.Fatalf("ListCards returned error: %v", err)
	}
	if len(cards) != 1 || cards[0].ID != "card-new" {
		t.Fatalf("unexpected cards after replace import: %#v", cards)
	}
}

func TestImportDeckUpsertsLearningPathWhenPresent(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	deck := mustCreateDeck(t, s, "Vim")

	path := model.LearningPath{
		ID:          "path-1",
		Name:        "Vim Fundamentals",
		Description: "Core to advanced",
		DeckIDs:     []string{deck.ID},
		CreatedAt:   time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
	}

	err := s.ImportDeck(ExportPayload{
		Deck:         deck,
		Cards:        nil,
		LearningPath: &path,
	}, "merge")
	if err != nil {
		t.Fatalf("ImportDeck(learning path) returned error: %v", err)
	}

	got, err := s.GetLearningPath(path.ID)
	if err != nil {
		t.Fatalf("GetLearningPath returned error: %v", err)
	}
	if got.Name != path.Name || len(got.DeckIDs) != 1 || got.DeckIDs[0] != deck.ID {
		t.Fatalf("unexpected learning path after import: %#v", got)
	}
}

func newTestStore(t *testing.T) Store {
	t.Helper()

	dbPath := "file:" + t.Name() + "?mode=memory&cache=shared"

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	return s
}

func mustCreateDeck(t *testing.T, s Store, name string) model.Deck {
	t.Helper()

	deck, err := s.CreateDeck(model.Deck{
		Name:               name,
		Description:        name + " deck",
		TagNamespaces:      model.TagNamespaces{"topic": {"editing"}},
		DefaultReverseMode: model.PromptFirst,
	})
	if err != nil {
		t.Fatalf("CreateDeck returned error: %v", err)
	}

	return deck
}

func mustCreateCard(t *testing.T, s Store, card model.Card) (model.Deck, model.Card) {
	t.Helper()

	if card.DeckID == "" {
		deck := mustCreateDeck(t, s, "Deck for card")
		card.DeckID = deck.ID
		created, err := s.CreateCard(card)
		if err != nil {
			t.Fatalf("CreateCard returned error: %v", err)
		}

		return deck, created
	}

	deck, err := s.GetDeck(card.DeckID)
	if err != nil {
		t.Fatalf("GetDeck for card returned error: %v", err)
	}

	created, err := s.CreateCard(card)
	if err != nil {
		t.Fatalf("CreateCard returned error: %v", err)
	}

	return deck, created
}
