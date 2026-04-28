package scheduler

import (
	"math"
	"time"
)

// SchedulingState is the algorithm-agnostic representation of a card's
// current scheduling position.
type SchedulingState struct {
	Interval       int
	EaseFactor     float64
	Repetitions    int
	DueDate        time.Time
	LastReviewedAt time.Time
}

// Grade represents the user's self-assessment of recall quality.
type Grade int

const (
	GradeAgain   Grade = 0
	GradePartial Grade = 1
	GradeHard    Grade = 3
	GradeEasy    Grade = 5
)

// Scheduler is the interface all algorithm implementations must satisfy.
type Scheduler interface {
	Schedule(current SchedulingState, grade Grade, now time.Time) SchedulingState
}

// SM2 is the v1 implementation of Scheduler using the SM-2 algorithm.
type SM2 struct{}

// Schedule implements Scheduler for SM2.
func (SM2) Schedule(current SchedulingState, grade Grade, now time.Time) SchedulingState {
	reviewedAt := now.UTC()
	state := current
	if state.EaseFactor == 0 {
		state.EaseFactor = 2.5
	}

	newEaseFactor := maxEaseFloor(state.EaseFactor + easeDelta(grade))

	var (
		newRepetitions int
		newInterval    int
	)

	if grade < GradeHard {
		newRepetitions = 0
		newInterval = 1
	} else {
		newRepetitions = state.Repetitions + 1
		switch newRepetitions {
		case 1:
			newInterval = 1
		case 2:
			newInterval = 6
		default:
			newInterval = int(math.Round(float64(state.Interval) * newEaseFactor))
			if newInterval < 1 {
				newInterval = 1
			}
		}
	}

	return SchedulingState{
		Interval:       newInterval,
		EaseFactor:     newEaseFactor,
		Repetitions:    newRepetitions,
		DueDate:        normalizeDate(reviewedAt).AddDate(0, 0, newInterval),
		LastReviewedAt: reviewedAt,
	}
}

// NewState returns a SchedulingState with SM-2 defaults for a brand-new card.
func NewState() SchedulingState {
	return SchedulingState{
		Interval:    0,
		EaseFactor:  2.5,
		Repetitions: 0,
		DueDate:     normalizeDate(time.Now().UTC()),
	}
}

func easeDelta(grade Grade) float64 {
	return 0.1 - float64(5-grade)*(0.08+float64(5-grade)*0.02)
}

func maxEaseFloor(value float64) float64 {
	if value < 1.3 {
		return 1.3
	}

	return value
}

func normalizeDate(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
