package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pnanadikar/shortcutdeck/internal/model"
	"github.com/pnanadikar/shortcutdeck/internal/scheduler"
	"github.com/pnanadikar/shortcutdeck/internal/store"
)

func TestStaticAssetsAreServed(t *testing.T) {
	t.Parallel()

	srv, _ := newTestServer(t)

	cases := []struct {
		name        string
		path        string
		contentType string
		bodyMarker  string
	}{
		{name: "root", path: "/", contentType: "text/html", bodyMarker: "<!DOCTYPE html>"},
		{name: "css", path: "/app.css", contentType: "text/css", bodyMarker: ":root"},
		{name: "js", path: "/app.js", contentType: "text/javascript", bodyMarker: "const API_BASE"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := performRequest(t, srv, http.MethodGet, tc.path, "")

			if rec.Code != http.StatusOK {
				t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
			}
			if !strings.HasPrefix(rec.Header().Get("Content-Type"), tc.contentType) {
				t.Fatalf("unexpected content type: got %q want prefix %q", rec.Header().Get("Content-Type"), tc.contentType)
			}
			if strings.TrimSpace(rec.Body.String()) == "" {
				t.Fatalf("expected non-empty response body")
			}
			if !strings.Contains(rec.Body.String(), tc.bodyMarker) {
				t.Fatalf("expected body to contain %q", tc.bodyMarker)
			}
		})
	}
}

func TestDeckEndpoints(t *testing.T) {
	t.Parallel()

	srv, _ := newTestServer(t)

	create := performJSONRequest(t, srv, http.MethodPost, "/api/v1/decks", model.Deck{
		Name:               "Vim",
		Description:        "Editor shortcuts",
		TagNamespaces:      model.TagNamespaces{"mode": {"normal"}},
		DefaultReverseMode: model.AnswerFirst,
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("create deck status: got %d want %d", create.Code, http.StatusCreated)
	}

	var created model.Deck
	decodeResponse(t, create, &created)
	if created.ID == "" || created.Name != "Vim" || created.DefaultReverseMode != model.AnswerFirst {
		t.Fatalf("unexpected created deck: %#v", created)
	}

	list := performRequest(t, srv, http.MethodGet, "/api/v1/decks", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list decks status: got %d want %d", list.Code, http.StatusOK)
	}

	var decks []model.Deck
	decodeResponse(t, list, &decks)
	if len(decks) != 1 || decks[0].ID != created.ID {
		t.Fatalf("unexpected decks response: %#v", decks)
	}

	get := performRequest(t, srv, http.MethodGet, "/api/v1/decks/"+created.ID, "")
	if get.Code != http.StatusOK {
		t.Fatalf("get deck status: got %d want %d", get.Code, http.StatusOK)
	}

	update := performJSONRequest(t, srv, http.MethodPut, "/api/v1/decks/"+created.ID, model.Deck{
		Name:               "Vim Updated",
		Description:        "Updated description",
		TagNamespaces:      model.TagNamespaces{"topic": {"editing"}},
		DefaultReverseMode: model.Both,
	})
	if update.Code != http.StatusOK {
		t.Fatalf("update deck status: got %d want %d", update.Code, http.StatusOK)
	}

	var updated model.Deck
	decodeResponse(t, update, &updated)
	if updated.Name != "Vim Updated" || updated.DefaultReverseMode != model.Both {
		t.Fatalf("unexpected updated deck: %#v", updated)
	}

	del := performRequest(t, srv, http.MethodDelete, "/api/v1/decks/"+created.ID, "")
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete deck status: got %d want %d", del.Code, http.StatusNoContent)
	}
}

func TestCardEndpoints(t *testing.T) {
	t.Parallel()

	srv, st := newTestServer(t)
	deck := mustCreateDeck(t, st, "Vim")

	create := performJSONRequest(t, srv, http.MethodPost, "/api/v1/decks/"+deck.ID+"/cards", model.Card{
		Prompt: "Delete line",
		Answer: "dd",
		Notes:  "normal mode",
		Tags:   []string{"normal", "editing"},
		Source: "manual",
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("create card status: got %d want %d", create.Code, http.StatusCreated)
	}

	var created model.Card
	decodeResponse(t, create, &created)
	if created.ID == "" || created.DeckID != deck.ID || created.Prompt != "Delete line" {
		t.Fatalf("unexpected created card: %#v", created)
	}

	list := performRequest(t, srv, http.MethodGet, "/api/v1/decks/"+deck.ID+"/cards?tags=normal,editing", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list cards status: got %d want %d", list.Code, http.StatusOK)
	}

	var cards []model.Card
	decodeResponse(t, list, &cards)
	if len(cards) != 1 || cards[0].ID != created.ID {
		t.Fatalf("unexpected cards response: %#v", cards)
	}

	get := performRequest(t, srv, http.MethodGet, "/api/v1/cards/"+created.ID, "")
	if get.Code != http.StatusOK {
		t.Fatalf("get card status: got %d want %d", get.Code, http.StatusOK)
	}

	update := performJSONRequest(t, srv, http.MethodPut, "/api/v1/cards/"+created.ID, model.Card{
		Prompt:         "Delete current line",
		Answer:         "dd",
		Notes:          "updated",
		Tags:           []string{"editing"},
		Source:         created.Source,
		Interval:       created.Interval,
		EaseFactor:     created.EaseFactor,
		Repetitions:    created.Repetitions,
		DueDate:        created.DueDate,
		LastReviewedAt: created.LastReviewedAt,
	})
	if update.Code != http.StatusOK {
		t.Fatalf("update card status: got %d want %d", update.Code, http.StatusOK)
	}

	var updated model.Card
	decodeResponse(t, update, &updated)
	if updated.Prompt != "Delete current line" || len(updated.Tags) != 1 || updated.Tags[0] != "editing" {
		t.Fatalf("unexpected updated card: %#v", updated)
	}

	del := performRequest(t, srv, http.MethodDelete, "/api/v1/cards/"+created.ID, "")
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete card status: got %d want %d", del.Code, http.StatusNoContent)
	}
}

func TestReviewEndpoints(t *testing.T) {
	t.Parallel()

	srv, st := newTestServer(t)
	deck := mustCreateDeck(t, st, "Vim")
	card := mustCreateCard(t, st, deck.ID, model.Card{
		Prompt:  "Delete line",
		Answer:  "dd",
		Tags:    []string{"editing"},
		DueDate: time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
	})

	dueInactive := performRequest(t, srv, http.MethodGet, "/api/v1/decks/"+deck.ID+"/due", "")
	if dueInactive.Code != http.StatusOK {
		t.Fatalf("inactive due status: got %d want %d", dueInactive.Code, http.StatusOK)
	}

	var inactive []model.Card
	decodeResponse(t, dueInactive, &inactive)
	if len(inactive) != 0 {
		t.Fatalf("expected no due cards while inactive, got %#v", inactive)
	}

	start := performRequest(t, srv, http.MethodPost, "/api/v1/decks/"+deck.ID+"/start-review", "")
	if start.Code != http.StatusOK {
		t.Fatalf("start review status: got %d want %d", start.Code, http.StatusOK)
	}

	var started model.Deck
	decodeResponse(t, start, &started)
	if !started.ReviewActive {
		t.Fatalf("expected review_active=true after start review")
	}

	dueActive := performRequest(t, srv, http.MethodGet, "/api/v1/decks/"+deck.ID+"/due?tags=editing", "")
	if dueActive.Code != http.StatusOK {
		t.Fatalf("active due status: got %d want %d", dueActive.Code, http.StatusOK)
	}

	var dueCards []model.Card
	decodeResponse(t, dueActive, &dueCards)
	if len(dueCards) != 1 || dueCards[0].ID != card.ID {
		t.Fatalf("unexpected due cards: %#v", dueCards)
	}

	review := performJSONRequest(t, srv, http.MethodPost, "/api/v1/cards/"+card.ID+"/review", reviewRequest{Grade: int(scheduler.GradeHard)})
	if review.Code != http.StatusOK {
		t.Fatalf("review status: got %d want %d", review.Code, http.StatusOK)
	}

	var reviewed model.Card
	decodeResponse(t, review, &reviewed)
	if reviewed.Interval != 1 || reviewed.Repetitions != 1 || reviewed.LastReviewedAt.IsZero() {
		t.Fatalf("unexpected reviewed card: %#v", reviewed)
	}

	logEntries, err := st.GetReviewLog(card.ID)
	if err != nil {
		t.Fatalf("GetReviewLog returned error: %v", err)
	}
	if len(logEntries) != 1 || logEntries[0].Grade != int(scheduler.GradeHard) {
		t.Fatalf("unexpected review log: %#v", logEntries)
	}

	exit := performRequest(t, srv, http.MethodPost, "/api/v1/decks/"+deck.ID+"/exit-review", "")
	if exit.Code != http.StatusOK {
		t.Fatalf("exit review status: got %d want %d", exit.Code, http.StatusOK)
	}

	var exited model.Deck
	decodeResponse(t, exit, &exited)
	if exited.ReviewActive {
		t.Fatalf("expected review_active=false after exit review")
	}

	resetCard, err := st.GetCard(card.ID)
	if err != nil {
		t.Fatalf("GetCard returned error: %v", err)
	}
	if resetCard.Interval != 0 || resetCard.Repetitions != 0 || resetCard.EaseFactor != 2.5 || !resetCard.LastReviewedAt.IsZero() {
		t.Fatalf("unexpected reset card: %#v", resetCard)
	}
}

func TestLearningPathEndpointsFilterMissingDeckIDs(t *testing.T) {
	t.Parallel()

	srv, st := newTestServer(t)
	deck := mustCreateDeck(t, st, "Vim")

	path, err := st.CreateLearningPath(model.LearningPath{
		Name:        "Vim Path",
		Description: "Fundamentals",
		DeckIDs:     []string{deck.ID, "missing-deck"},
	})
	if err != nil {
		t.Fatalf("CreateLearningPath returned error: %v", err)
	}

	list := performRequest(t, srv, http.MethodGet, "/api/v1/paths", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list paths status: got %d want %d", list.Code, http.StatusOK)
	}

	var paths []model.LearningPath
	decodeResponse(t, list, &paths)
	if len(paths) != 1 || len(paths[0].DeckIDs) != 1 || paths[0].DeckIDs[0] != deck.ID {
		t.Fatalf("unexpected paths response: %#v", paths)
	}

	get := performRequest(t, srv, http.MethodGet, "/api/v1/paths/"+path.ID, "")
	if get.Code != http.StatusOK {
		t.Fatalf("get path status: got %d want %d", get.Code, http.StatusOK)
	}

	var got model.LearningPath
	decodeResponse(t, get, &got)
	if len(got.DeckIDs) != 1 || got.DeckIDs[0] != deck.ID {
		t.Fatalf("unexpected filtered path: %#v", got)
	}
}

func TestImportExportEndpoints(t *testing.T) {
	t.Parallel()

	srv, st := newTestServer(t)
	deck := mustCreateDeck(t, st, "Vim")
	card := mustCreateCard(t, st, deck.ID, model.Card{
		Prompt: "Delete line",
		Answer: "dd",
		Tags:   []string{"editing"},
	})
	path, err := st.CreateLearningPath(model.LearningPath{
		ID:          "path-1",
		Name:        "Vim Path",
		Description: "Fundamentals",
		DeckIDs:     []string{deck.ID},
		CreatedAt:   time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("CreateLearningPath returned error: %v", err)
	}

	exportRec := performRequest(t, srv, http.MethodGet, "/api/v1/decks/"+deck.ID+"/export", "")
	if exportRec.Code != http.StatusOK {
		t.Fatalf("export status: got %d want %d", exportRec.Code, http.StatusOK)
	}

	var exported store.ExportPayload
	decodeResponse(t, exportRec, &exported)
	if exported.Deck.ID != deck.ID || len(exported.Cards) != 1 || exported.Cards[0].ID != card.ID {
		t.Fatalf("unexpected export payload: %#v", exported)
	}

	importPayload := store.ExportPayload{
		Deck: model.Deck{
			ID:                 "deck-imported",
			Name:               "Imported Deck",
			Description:        "Imported",
			TagNamespaces:      model.TagNamespaces{"topic": {"editing"}},
			DefaultReverseMode: model.PromptFirst,
			CreatedAt:          time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
		},
		Cards: []model.Card{
			{
				ID:         "card-imported",
				DeckID:     "deck-imported",
				Prompt:     "Paste",
				Answer:     "p",
				Tags:       []string{"editing"},
				Source:     "imported",
				CreatedAt:  time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
				EaseFactor: 2.5,
				DueDate:    time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
			},
		},
		LearningPath: &model.LearningPath{
			ID:          path.ID,
			Name:        "Imported Path",
			Description: "Updated",
			DeckIDs:     []string{"deck-imported"},
			CreatedAt:   path.CreatedAt,
		},
	}

	importRec := performJSONRequest(t, srv, http.MethodPost, "/api/v1/import?mode=merge", importPayload)
	if importRec.Code != http.StatusOK {
		t.Fatalf("import status: got %d want %d", importRec.Code, http.StatusOK)
	}

	var imported store.ExportPayload
	decodeResponse(t, importRec, &imported)
	if imported.Deck.ID != "deck-imported" || len(imported.Cards) != 1 || imported.Cards[0].ID != "card-imported" {
		t.Fatalf("unexpected import response: %#v", imported)
	}

	importedPath, err := st.GetLearningPath(path.ID)
	if err != nil {
		t.Fatalf("GetLearningPath returned error: %v", err)
	}
	if importedPath.Name != "Imported Path" || len(importedPath.DeckIDs) != 1 || importedPath.DeckIDs[0] != "deck-imported" {
		t.Fatalf("unexpected imported learning path: %#v", importedPath)
	}
}

func TestImportLogsFailuresAndSuccess(t *testing.T) {
	srv, _ := newTestServer(t)
	var logs bytes.Buffer
	srv.logger = bufferLogger{buf: &logs}

	invalid := performRequest(t, srv, http.MethodPost, "/api/v1/import?mode=merge", "{")
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid import status: got %d want %d", invalid.Code, http.StatusBadRequest)
	}
	if got := logs.String(); !strings.Contains(got, "warning: import failed: mode=merge stage=decode") {
		t.Fatalf("expected decode failure log, got %q", got)
	}

	logs.Reset()
	importPayload := store.ExportPayload{
		Deck: model.Deck{
			ID:                 "deck-logged",
			Name:               "Logged Deck",
			TagNamespaces:      model.TagNamespaces{"topic": {"editing"}},
			DefaultReverseMode: model.PromptFirst,
			CreatedAt:          time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
		},
		Cards: []model.Card{
			{
				ID:         "card-logged",
				Prompt:     "Private prompt should not appear in logs",
				Answer:     "Private answer should not appear in logs",
				CreatedAt:  time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
				EaseFactor: 2.5,
				DueDate:    time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	imported := performJSONRequest(t, srv, http.MethodPost, "/api/v1/import?mode=merge", importPayload)
	if imported.Code != http.StatusOK {
		t.Fatalf("import status: got %d want %d", imported.Code, http.StatusOK)
	}

	got := logs.String()
	if !strings.Contains(got, `import complete: mode=merge deck_id="deck-logged" card_count=1`) {
		t.Fatalf("expected import complete log, got %q", got)
	}
	if strings.Contains(got, "Private prompt") || strings.Contains(got, "Private answer") {
		t.Fatalf("import log leaked card content: %q", got)
	}
}

func TestUnknownIDsReturn404(t *testing.T) {
	t.Parallel()

	srv, _ := newTestServer(t)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "get deck", method: http.MethodGet, path: "/api/v1/decks/missing"},
		{name: "list cards", method: http.MethodGet, path: "/api/v1/decks/missing/cards"},
		{name: "due cards", method: http.MethodGet, path: "/api/v1/decks/missing/due"},
		{name: "start review", method: http.MethodPost, path: "/api/v1/decks/missing/start-review"},
		{name: "exit review", method: http.MethodPost, path: "/api/v1/decks/missing/exit-review"},
		{name: "export deck", method: http.MethodGet, path: "/api/v1/decks/missing/export"},
		{name: "get card", method: http.MethodGet, path: "/api/v1/cards/missing"},
		{name: "review card", method: http.MethodPost, path: "/api/v1/cards/missing/review", body: `{"grade":3}`},
		{name: "get path", method: http.MethodGet, path: "/api/v1/paths/missing"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := performRequest(t, srv, tc.method, tc.path, tc.body)
			if rec.Code != http.StatusNotFound {
				t.Fatalf("unexpected status: got %d want %d body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
			}
		})
	}
}

func TestMalformedJSONReturns400(t *testing.T) {
	t.Parallel()

	srv, st := newTestServer(t)
	deck := mustCreateDeck(t, st, "Vim")
	card := mustCreateCard(t, st, deck.ID, model.Card{Prompt: "Delete", Answer: "dd"})

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "create deck", method: http.MethodPost, path: "/api/v1/decks"},
		{name: "update deck", method: http.MethodPut, path: "/api/v1/decks/" + deck.ID},
		{name: "create card", method: http.MethodPost, path: "/api/v1/decks/" + deck.ID + "/cards"},
		{name: "update card", method: http.MethodPut, path: "/api/v1/cards/" + card.ID},
		{name: "review card", method: http.MethodPost, path: "/api/v1/cards/" + card.ID + "/review"},
		{name: "import deck", method: http.MethodPost, path: "/api/v1/import?mode=merge"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := performRequest(t, srv, tc.method, tc.path, "{")
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestInvalidGradeAndImportModeReturn400(t *testing.T) {
	t.Parallel()

	srv, st := newTestServer(t)
	deck := mustCreateDeck(t, st, "Vim")
	card := mustCreateCard(t, st, deck.ID, model.Card{Prompt: "Delete", Answer: "dd"})

	invalidGrade := performJSONRequest(t, srv, http.MethodPost, "/api/v1/cards/"+card.ID+"/review", reviewRequest{Grade: 2})
	if invalidGrade.Code != http.StatusBadRequest {
		t.Fatalf("invalid grade status: got %d want %d", invalidGrade.Code, http.StatusBadRequest)
	}

	invalidMode := performJSONRequest(t, srv, http.MethodPost, "/api/v1/import?mode=nope", store.ExportPayload{})
	if invalidMode.Code != http.StatusBadRequest {
		t.Fatalf("invalid import mode status: got %d want %d", invalidMode.Code, http.StatusBadRequest)
	}
}

func newTestServer(t *testing.T) (*Server, store.Store) {
	t.Helper()

	dbPath := "file:" + t.Name() + "?mode=memory&cache=shared"
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("store.New returned error: %v", err)
	}

	srv := New(st, scheduler.SM2{})
	srv.now = func() time.Time {
		return time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	}

	return srv, st
}

type bufferLogger struct {
	buf *bytes.Buffer
}

func (l bufferLogger) Printf(format string, v ...any) {
	_, _ = fmt.Fprintf(l.buf, format+"\n", v...)
}

func performJSONRequest(t *testing.T, srv *Server, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		t.Fatalf("json.NewEncoder returned error: %v", err)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	return rec
}

func performRequest(t *testing.T, srv *Server, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	return rec
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder, dst any) {
	t.Helper()

	if err := json.Unmarshal(rec.Body.Bytes(), dst); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v; body=%s", err, rec.Body.String())
	}
}

func mustCreateDeck(t *testing.T, st store.Store, name string) model.Deck {
	t.Helper()

	deck, err := st.CreateDeck(model.Deck{
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

func mustCreateCard(t *testing.T, st store.Store, deckID string, card model.Card) model.Card {
	t.Helper()

	card.DeckID = deckID
	created, err := st.CreateCard(card)
	if err != nil {
		t.Fatalf("CreateCard returned error: %v", err)
	}

	return created
}
