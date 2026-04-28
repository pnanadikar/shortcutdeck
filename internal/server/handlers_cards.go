package server

import (
	"net/http"

	"github.com/pnanadikar/shortcutdeck/internal/model"
)

func (s *Server) handleListCards(w http.ResponseWriter, r *http.Request) {
	deckID := r.PathValue("id")
	if _, err := s.store.GetDeck(deckID); err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "deck not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get deck")
		return
	}

	cards, err := s.store.ListCards(deckID, parseTags(r.URL.Query().Get("tags")))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list cards")
		return
	}

	writeJSON(w, http.StatusOK, cards)
}

func (s *Server) handleCreateCard(w http.ResponseWriter, r *http.Request) {
	deckID := r.PathValue("id")
	if _, err := s.store.GetDeck(deckID); err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "deck not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get deck")
		return
	}

	var card model.Card
	if err := decodeJSON(r, &card); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	card.ID = ""
	card.DeckID = deckID

	created, err := s.store.CreateCard(card)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create card")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleGetCard(w http.ResponseWriter, r *http.Request) {
	card, err := s.store.GetCard(r.PathValue("id"))
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "card not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get card")
		return
	}

	writeJSON(w, http.StatusOK, card)
}

func (s *Server) handleUpdateCard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	current, err := s.store.GetCard(id)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "card not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get card")
		return
	}

	var card model.Card
	if err := decodeJSON(r, &card); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	card.ID = id
	card.DeckID = current.DeckID
	card.CreatedAt = current.CreatedAt
	if card.Source == "" {
		card.Source = current.Source
	}

	updated, err := s.store.UpdateCard(card)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "card not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update card")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDeleteCard(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteCard(r.PathValue("id")); err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "card not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete card")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
