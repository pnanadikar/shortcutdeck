package server

import (
	"net/http"

	"github.com/pnanadikar/shortcutdeck/internal/model"
)

func (s *Server) handleListDecks(w http.ResponseWriter, _ *http.Request) {
	decks, err := s.store.ListDecks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list decks")
		return
	}

	writeJSON(w, http.StatusOK, decks)
}

func (s *Server) handleCreateDeck(w http.ResponseWriter, r *http.Request) {
	var deck model.Deck
	if err := decodeJSON(r, &deck); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := s.store.CreateDeck(deck)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create deck")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleGetDeck(w http.ResponseWriter, r *http.Request) {
	deck, err := s.store.GetDeck(r.PathValue("id"))
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "deck not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get deck")
		return
	}

	writeJSON(w, http.StatusOK, deck)
}

func (s *Server) handleUpdateDeck(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	current, err := s.store.GetDeck(id)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "deck not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get deck")
		return
	}

	var deck model.Deck
	if err := decodeJSON(r, &deck); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	deck.ID = id
	deck.CreatedAt = current.CreatedAt
	deck.ReviewActive = current.ReviewActive

	updated, err := s.store.UpdateDeck(deck)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "deck not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update deck")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDeleteDeck(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteDeck(r.PathValue("id")); err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "deck not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete deck")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
