package server

import (
	"net/http"

	"github.com/pnanadikar/shortcutdeck/internal/model"
	"github.com/pnanadikar/shortcutdeck/internal/scheduler"
)

func (s *Server) handleGetDueCards(w http.ResponseWriter, r *http.Request) {
	deckID := r.PathValue("id")
	deck, err := s.store.GetDeck(deckID)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "deck not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get deck")
		return
	}

	if !deck.ReviewActive {
		writeJSON(w, http.StatusOK, []model.Card{})
		return
	}

	cards, err := s.store.GetDueCards(deckID, s.now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get due cards")
		return
	}

	writeJSON(w, http.StatusOK, filterCardsByTags(cards, parseTags(r.URL.Query().Get("tags"))))
}

func (s *Server) handleReviewCard(w http.ResponseWriter, r *http.Request) {
	card, err := s.store.GetCard(r.PathValue("id"))
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "card not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get card")
		return
	}

	var req reviewRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	grade, ok := parseGrade(req.Grade)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid grade value")
		return
	}

	now := s.now()
	next := s.scheduler.Schedule(schedulingStateFromCard(card), grade, now)
	card = applySchedulingState(card, next)

	updated, err := s.store.UpdateCard(card)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update card")
		return
	}

	if err := s.store.LogReview(model.ReviewEntry{
		CardID:        updated.ID,
		ReviewedAt:    now,
		Grade:         int(grade),
		IntervalAfter: updated.Interval,
		EaseAfter:     updated.EaseFactor,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to log review")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleStartReview(w http.ResponseWriter, r *http.Request) {
	deck, err := s.store.GetDeck(r.PathValue("id"))
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "deck not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get deck")
		return
	}

	deck.ReviewActive = true
	updated, err := s.store.UpdateDeck(deck)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start review")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleExitReview(w http.ResponseWriter, r *http.Request) {
	deck, err := s.store.GetDeck(r.PathValue("id"))
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "deck not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get deck")
		return
	}

	cards, err := s.store.ListCards(deck.ID, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list cards")
		return
	}

	state := defaultSchedulingState(s.now())
	for _, card := range cards {
		card = applySchedulingState(card, state)
		if _, err := s.store.UpdateCard(card); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reset card scheduling state")
			return
		}
	}

	deck.ReviewActive = false
	updated, err := s.store.UpdateDeck(deck)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to exit review")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func parseGrade(value int) (scheduler.Grade, bool) {
	switch scheduler.Grade(value) {
	case scheduler.GradeAgain, scheduler.GradePartial, scheduler.GradeHard, scheduler.GradeEasy:
		return scheduler.Grade(value), true
	default:
		return 0, false
	}
}
