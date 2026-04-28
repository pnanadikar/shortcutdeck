package server

import (
	"net/http"

	"github.com/pnanadikar/shortcutdeck/internal/model"
)

func (s *Server) handleListLearningPaths(w http.ResponseWriter, _ *http.Request) {
	paths, err := s.store.ListLearningPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list learning paths")
		return
	}

	filtered := make([]model.LearningPath, 0, len(paths))
	for _, path := range paths {
		filtered = append(filtered, s.filterLearningPath(path))
	}

	writeJSON(w, http.StatusOK, filtered)
}

func (s *Server) handleGetLearningPath(w http.ResponseWriter, r *http.Request) {
	path, err := s.store.GetLearningPath(r.PathValue("id"))
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "learning path not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get learning path")
		return
	}

	writeJSON(w, http.StatusOK, s.filterLearningPath(path))
}

func (s *Server) filterLearningPath(path model.LearningPath) model.LearningPath {
	filtered := make([]string, 0, len(path.DeckIDs))
	for _, deckID := range path.DeckIDs {
		if _, err := s.store.GetDeck(deckID); err == nil {
			filtered = append(filtered, deckID)
		}
	}

	path.DeckIDs = filtered

	return path
}
