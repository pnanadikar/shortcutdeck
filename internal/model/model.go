package model

import "time"

// Card represents a single keyboard shortcut flashcard.
type Card struct {
	ID      string // UUID
	DeckID  string // parent deck UUID
	Prompt  string // what the shortcut does
	Answer  string // the key combination
	Notes   string // optional context
	Tags    []string
	Source  string // "manual" | "ai-generated" | "imported"
	CreatedAt time.Time

	// Scheduling state (SM-2)
	Interval       int     // days until next review
	EaseFactor     float64 // SM-2 ease factor (min 1.3)
	Repetitions    int     // consecutive successful reviews
	DueDate        time.Time
	LastReviewedAt time.Time
}

// Deck represents a set of cards for one application or context.
type Deck struct {
	ID                 string // UUID
	Name               string // e.g. "Vim", "Finder"
	Description        string // optional
	TagNamespaces      TagNamespaces
	DefaultReverseMode ReverseMode // pre-selected option when starting a session; overridable per session
	ReviewActive       bool        // true when Ongoing Review has been started and not yet exited
	CreatedAt          time.Time
}

// TagNamespaces maps namespace names to their known tag values.
type TagNamespaces map[string][]string

// ReverseMode controls which side of a card is shown first during a session.
type ReverseMode string

const (
	PromptFirst ReverseMode = "prompt_first"
	AnswerFirst ReverseMode = "answer_first"
	Both        ReverseMode = "both"
)

// ReviewEntry is a single review event, stored in review_log.
type ReviewEntry struct {
	ID            string // UUID
	CardID        string
	ReviewedAt    time.Time
	Grade         int     // 0=Again, 1=Partial, 3=Hard, 5=Easy
	IntervalAfter int     // days, after this review
	EaseAfter     float64 // ease factor after this review
}

// LearningPath is an ordered sequence of decks representing a suggested study progression.
type LearningPath struct {
	ID          string // UUID
	Name        string // e.g. "Vim — Zero to Fluent"
	Description string // optional narrative
	DeckIDs     []string
	CreatedAt   time.Time
}
