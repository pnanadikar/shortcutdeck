package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/pnanadikar/shortcutdeck/internal/appinfo"
	"github.com/pnanadikar/shortcutdeck/internal/model"
	"github.com/pnanadikar/shortcutdeck/internal/scheduler"
	"github.com/pnanadikar/shortcutdeck/internal/store"
	webassets "github.com/pnanadikar/shortcutdeck/web"
)

type Server struct {
	store     store.Store
	scheduler scheduler.Scheduler
	logger    Logger
	now       func() time.Time
	mux       *http.ServeMux
	staticFS  fs.FS
}

type Logger interface {
	Printf(format string, v ...any)
}

type noopLogger struct{}

func (noopLogger) Printf(string, ...any) {}

type errorResponse struct {
	Error string `json:"error"`
}

type reviewRequest struct {
	Grade int `json:"grade"`
}

func New(store store.Store, scheduler scheduler.Scheduler) *Server {
	return NewWithLogger(store, scheduler, nil)
}

func NewWithLogger(store store.Store, scheduler scheduler.Scheduler, logger Logger) *Server {
	if logger == nil {
		logger = noopLogger{}
	}

	s := &Server{
		store:     store,
		scheduler: scheduler,
		logger:    logger,
		now:       func() time.Time { return time.Now().UTC() },
		staticFS:  webassets.FS,
	}

	s.mux = http.NewServeMux()
	s.registerRoutes()

	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /", s.handleRoot)
	s.mux.HandleFunc("GET /app.css", s.handleStaticAsset("app.css"))
	s.mux.HandleFunc("GET /app.js", s.handleStaticAsset("app.js"))

	s.mux.HandleFunc("GET /api/v1/decks", s.handleListDecks)
	s.mux.HandleFunc("POST /api/v1/decks", s.handleCreateDeck)
	s.mux.HandleFunc("GET /api/v1/decks/{id}", s.handleGetDeck)
	s.mux.HandleFunc("PUT /api/v1/decks/{id}", s.handleUpdateDeck)
	s.mux.HandleFunc("DELETE /api/v1/decks/{id}", s.handleDeleteDeck)
	s.mux.HandleFunc("GET /api/v1/decks/{id}/cards", s.handleListCards)
	s.mux.HandleFunc("POST /api/v1/decks/{id}/cards", s.handleCreateCard)
	s.mux.HandleFunc("GET /api/v1/decks/{id}/due", s.handleGetDueCards)
	s.mux.HandleFunc("POST /api/v1/decks/{id}/start-review", s.handleStartReview)
	s.mux.HandleFunc("POST /api/v1/decks/{id}/exit-review", s.handleExitReview)
	s.mux.HandleFunc("GET /api/v1/decks/{id}/export", s.handleExportDeck)

	s.mux.HandleFunc("GET /api/v1/cards/{id}", s.handleGetCard)
	s.mux.HandleFunc("PUT /api/v1/cards/{id}", s.handleUpdateCard)
	s.mux.HandleFunc("DELETE /api/v1/cards/{id}", s.handleDeleteCard)
	s.mux.HandleFunc("POST /api/v1/cards/{id}/review", s.handleReviewCard)

	s.mux.HandleFunc("GET /api/v1/paths", s.handleListLearningPaths)
	s.mux.HandleFunc("GET /api/v1/paths/{id}", s.handleGetLearningPath)

	s.mux.HandleFunc("POST /api/v1/import", s.handleImportDeck)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	body, err := s.renderIndexHTML()
	if err != nil {
		http.Error(w, "failed to render app", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(body))
}

func (s *Server) handleStaticAsset(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.serveStaticFile(w, r, name)
	}
}

func (s *Server) serveStaticFile(w http.ResponseWriter, r *http.Request, name string) {
	file, err := s.staticFS.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer func() {
		_ = file.Close()
	}()

	info, err := file.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.ServeContent(w, r, info.Name(), info.ModTime(), file.(io.ReadSeeker))
}

func (s *Server) renderIndexHTML() ([]byte, error) {
	body, err := fs.ReadFile(s.staticFS, "index.html")
	if err != nil {
		return nil, err
	}

	appInfoJSON, err := json.Marshal(struct {
		Version     string `json:"version"`
		Description string `json:"description"`
	}{
		Version:     appinfo.Version,
		Description: appinfo.Description,
	})
	if err != nil {
		return nil, err
	}

	return bytes.ReplaceAll(body, []byte("__APP_INFO_JSON__"), appInfoJSON), nil
}

func decodeJSON(r *http.Request, dst any) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return io.EOF
	}

	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}

	if dec.More() {
		return errors.New("unexpected extra JSON values")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if value == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func parseTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags
}

func filterCardsByTags(cards []model.Card, tags []string) []model.Card {
	if len(tags) == 0 {
		return cards
	}

	filtered := make([]model.Card, 0, len(cards))
	for _, card := range cards {
		if hasAllTags(card.Tags, tags) {
			filtered = append(filtered, card)
		}
	}

	return filtered
}

func hasAllTags(cardTags []string, want []string) bool {
	if len(want) == 0 {
		return true
	}

	set := make(map[string]struct{}, len(cardTags))
	for _, tag := range cardTags {
		set[tag] = struct{}{}
	}

	for _, tag := range want {
		if _, ok := set[tag]; !ok {
			return false
		}
	}

	return true
}

func normalizeDate(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func defaultSchedulingState(now time.Time) scheduler.SchedulingState {
	return scheduler.SchedulingState{
		Interval:       0,
		EaseFactor:     2.5,
		Repetitions:    0,
		DueDate:        normalizeDate(now),
		LastReviewedAt: time.Time{},
	}
}

func applySchedulingState(card model.Card, state scheduler.SchedulingState) model.Card {
	card.Interval = state.Interval
	card.EaseFactor = state.EaseFactor
	card.Repetitions = state.Repetitions
	card.DueDate = state.DueDate
	card.LastReviewedAt = state.LastReviewedAt

	return card
}

func schedulingStateFromCard(card model.Card) scheduler.SchedulingState {
	return scheduler.SchedulingState{
		Interval:       card.Interval,
		EaseFactor:     card.EaseFactor,
		Repetitions:    card.Repetitions,
		DueDate:        card.DueDate,
		LastReviewedAt: card.LastReviewedAt,
	}
}

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
