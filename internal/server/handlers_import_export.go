package server

import (
	"net/http"

	"github.com/pnanadikar/shortcutdeck/internal/store"
)

func (s *Server) handleExportDeck(w http.ResponseWriter, r *http.Request) {
	payload, err := s.store.ExportDeck(r.PathValue("id"))
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "deck not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to export deck")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="deck-export.json"`)
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleImportDeck(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if mode != "merge" && mode != "replace" {
		s.logger.Printf("warning: import failed: mode=%q stage=mode err=%q", mode, "invalid import mode")
		writeError(w, http.StatusBadRequest, "invalid import mode")
		return
	}

	var payload store.ExportPayload
	if err := decodeJSON(r, &payload); err != nil {
		s.logger.Printf("warning: import failed: mode=%s stage=decode err=%q", mode, err.Error())
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.store.ImportDeck(payload, mode); err != nil {
		s.logger.Printf(
			"warning: import failed: mode=%s stage=store deck_id=%q card_count=%d err=%q",
			mode,
			payload.Deck.ID,
			len(payload.Cards),
			err.Error(),
		)
		writeError(w, http.StatusInternalServerError, "failed to import deck")
		return
	}

	imported, err := s.store.ExportDeck(payload.Deck.ID)
	if err != nil {
		s.logger.Printf(
			"warning: import failed: mode=%s stage=load deck_id=%q card_count=%d err=%q",
			mode,
			payload.Deck.ID,
			len(payload.Cards),
			err.Error(),
		)
		writeError(w, http.StatusInternalServerError, "failed to load imported deck")
		return
	}

	s.logger.Printf("import complete: mode=%s deck_id=%q card_count=%d", mode, payload.Deck.ID, len(payload.Cards))
	writeJSON(w, http.StatusOK, imported)
}
