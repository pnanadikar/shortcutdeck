package scheduler

import (
	"testing"
	"time"
)

func TestNewStateDefaults(t *testing.T) {
	t.Parallel()

	state := NewState()
	today := normalizeDate(time.Now().UTC())

	if state.Interval != 0 {
		t.Fatalf("expected interval 0, got %d", state.Interval)
	}
	if state.EaseFactor != 2.5 {
		t.Fatalf("expected ease factor 2.5, got %v", state.EaseFactor)
	}
	if state.Repetitions != 0 {
		t.Fatalf("expected repetitions 0, got %d", state.Repetitions)
	}
	if !state.DueDate.Equal(today) {
		t.Fatalf("expected due date %s, got %s", today, state.DueDate)
	}
	if !state.LastReviewedAt.IsZero() {
		t.Fatalf("expected zero last reviewed at, got %s", state.LastReviewedAt)
	}
}

func TestScheduleAgainResetsRepetitions(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)
	state := SchedulingState{
		Interval:    6,
		EaseFactor:  2.5,
		Repetitions: 3,
	}

	got := SM2{}.Schedule(state, GradeAgain, now)

	if got.Repetitions != 0 {
		t.Fatalf("expected repetitions reset to 0, got %d", got.Repetitions)
	}
	if got.Interval != 1 {
		t.Fatalf("expected interval reset to 1, got %d", got.Interval)
	}
}

func TestSchedulePartialIsLessSevereThanAgain(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)
	state := SchedulingState{
		Interval:    6,
		EaseFactor:  2.5,
		Repetitions: 3,
	}

	again := SM2{}.Schedule(state, GradeAgain, now)
	partial := SM2{}.Schedule(state, GradePartial, now)

	if partial.EaseFactor <= again.EaseFactor {
		t.Fatalf("expected partial to be less severe than again: partial=%v again=%v", partial.EaseFactor, again.EaseFactor)
	}
	if partial.Repetitions != 0 {
		t.Fatalf("expected partial repetitions reset to 0, got %d", partial.Repetitions)
	}
}

func TestScheduleFirstHardReviewSetsIntervalToOne(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)
	got := SM2{}.Schedule(NewState(), GradeHard, now)

	if got.Interval != 1 {
		t.Fatalf("expected first hard interval 1, got %d", got.Interval)
	}
	if got.Repetitions != 1 {
		t.Fatalf("expected repetitions 1, got %d", got.Repetitions)
	}
}

func TestScheduleSecondEasyReviewSetsIntervalToSix(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 19, 15, 4, 5, 0, time.UTC)
	state := SchedulingState{
		Interval:    1,
		EaseFactor:  2.5,
		Repetitions: 1,
	}

	got := SM2{}.Schedule(state, GradeEasy, now)

	if got.Interval != 6 {
		t.Fatalf("expected second successful review interval 6, got %d", got.Interval)
	}
	if got.Repetitions != 2 {
		t.Fatalf("expected repetitions 2, got %d", got.Repetitions)
	}
}

func TestScheduleEasyGrowsFasterThanHardAfterSecondRepetition(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 20, 15, 4, 5, 0, time.UTC)
	state := SchedulingState{
		Interval:    6,
		EaseFactor:  2.5,
		Repetitions: 2,
	}

	hard := SM2{}.Schedule(state, GradeHard, now)
	easy := SM2{}.Schedule(state, GradeEasy, now)

	if easy.Interval <= hard.Interval {
		t.Fatalf("expected easy interval > hard interval: easy=%d hard=%d", easy.Interval, hard.Interval)
	}
}

func TestScheduleSuccessiveCorrectReviewsCompoundInterval(t *testing.T) {
	t.Parallel()

	s := SM2{}
	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)

	first := s.Schedule(NewState(), GradeEasy, now)
	second := s.Schedule(first, GradeEasy, now.AddDate(0, 0, 1))
	third := s.Schedule(second, GradeEasy, now.AddDate(0, 0, 7))

	if first.Interval != 1 {
		t.Fatalf("expected first interval 1, got %d", first.Interval)
	}
	if second.Interval != 6 {
		t.Fatalf("expected second interval 6, got %d", second.Interval)
	}
	if third.Interval <= second.Interval {
		t.Fatalf("expected third interval to grow beyond second: second=%d third=%d", second.Interval, third.Interval)
	}
}

func TestScheduleAppliesEaseFactorFloor(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)
	state := SchedulingState{
		Interval:    6,
		EaseFactor:  1.31,
		Repetitions: 3,
	}

	got := SM2{}.Schedule(state, GradeAgain, now)

	if got.EaseFactor != 1.3 {
		t.Fatalf("expected ease factor floor 1.3, got %v", got.EaseFactor)
	}
}

func TestScheduleRepeatedAgainNeverDropsBelowEaseFloor(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)
	state := SchedulingState{
		Interval:    10,
		EaseFactor:  2.5,
		Repetitions: 5,
	}

	for range 20 {
		state = SM2{}.Schedule(state, GradeAgain, now)
		now = now.AddDate(0, 0, 1)
		if state.EaseFactor < 1.3 {
			t.Fatalf("ease factor dropped below floor: %v", state.EaseFactor)
		}
	}
}

func TestScheduleRepeatedHardNeverDropsBelowEaseFloor(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)
	state := SchedulingState{
		Interval:    10,
		EaseFactor:  2.5,
		Repetitions: 5,
	}

	for range 20 {
		state = SM2{}.Schedule(state, GradeHard, now)
		now = now.AddDate(0, 0, 1)
		if state.EaseFactor < 1.3 {
			t.Fatalf("ease factor dropped below floor: %v", state.EaseFactor)
		}
	}
}

func TestScheduleEaseFactorStaysAtFloorWhenAlreadyAtFloor(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)
	state := SchedulingState{
		Interval:    6,
		EaseFactor:  1.3,
		Repetitions: 3,
	}

	got := SM2{}.Schedule(state, GradeAgain, now)

	if got.EaseFactor != 1.3 {
		t.Fatalf("expected ease factor to stay at floor, got %v", got.EaseFactor)
	}
}

func TestScheduleUpdatesLastReviewedAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 15, 4, 5, 123, time.UTC)
	got := SM2{}.Schedule(NewState(), GradeEasy, now)

	if !got.LastReviewedAt.Equal(now.UTC()) {
		t.Fatalf("expected last reviewed at %s, got %s", now.UTC(), got.LastReviewedAt)
	}
}

func TestScheduleRelapseAfterGoodRunResetsProgress(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)
	state := SchedulingState{
		Interval:    45,
		EaseFactor:  2.8,
		Repetitions: 5,
	}

	got := SM2{}.Schedule(state, GradeAgain, now)

	if got.Interval != 1 {
		t.Fatalf("expected relapse interval reset to 1, got %d", got.Interval)
	}
	if got.Repetitions != 0 {
		t.Fatalf("expected relapse repetitions reset to 0, got %d", got.Repetitions)
	}
}

func TestScheduleRecoveryAfterRelapseBuildsFromResetState(t *testing.T) {
	t.Parallel()

	s := SM2{}
	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)
	state := SchedulingState{
		Interval:    45,
		EaseFactor:  2.8,
		Repetitions: 5,
	}

	relapse := s.Schedule(state, GradeAgain, now)
	recoveredHard := s.Schedule(relapse, GradeHard, now.AddDate(0, 0, 1))
	recoveredEasy := s.Schedule(recoveredHard, GradeEasy, now.AddDate(0, 0, 2))

	if relapse.Repetitions != 0 {
		t.Fatalf("expected relapse repetitions reset to 0, got %d", relapse.Repetitions)
	}
	if recoveredHard.Repetitions != 1 {
		t.Fatalf("expected recovery to rebuild from repetition 1, got %d", recoveredHard.Repetitions)
	}
	if recoveredEasy.Repetitions != 2 {
		t.Fatalf("expected second recovery review to reach repetition 2, got %d", recoveredEasy.Repetitions)
	}
	if recoveredEasy.Interval != 6 {
		t.Fatalf("expected rebuilt second success interval 6, got %d", recoveredEasy.Interval)
	}
}

func TestScheduleDueDateUsesProvidedNow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)
	got := SM2{}.Schedule(NewState(), GradeEasy, now)
	want := normalizeDate(now).AddDate(0, 0, 1)

	if !got.DueDate.Equal(want) {
		t.Fatalf("expected due date %s, got %s", want, got.DueDate)
	}
}

func TestScheduleTimezoneNeutralDueDate(t *testing.T) {
	t.Parallel()

	loc := time.FixedZone("UTC+5:30", 5*60*60+30*60)
	now := time.Date(2026, 3, 18, 23, 30, 0, 0, loc)
	got := SM2{}.Schedule(NewState(), GradeEasy, now)
	want := normalizeDate(now.UTC()).AddDate(0, 0, 1)

	if !got.DueDate.Equal(want) {
		t.Fatalf("expected timezone-neutral due date %s, got %s", want, got.DueDate)
	}
}

func TestScheduleIsDeterministicForFixedInputs(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 15, 4, 5, 0, time.UTC)
	state := SchedulingState{
		Interval:    6,
		EaseFactor:  2.5,
		Repetitions: 2,
		DueDate:     normalizeDate(now),
	}

	first := SM2{}.Schedule(state, GradeEasy, now)
	second := SM2{}.Schedule(state, GradeEasy, now)

	if first != second {
		t.Fatalf("expected deterministic output, got first=%#v second=%#v", first, second)
	}
}

func TestSchedulePartialAndAgainAdvanceDueDateFromProvidedNow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 22, 0, 0, 0, time.UTC)
	state := SchedulingState{
		Interval:    10,
		EaseFactor:  2.0,
		Repetitions: 4,
	}

	again := SM2{}.Schedule(state, GradeAgain, now)
	partial := SM2{}.Schedule(state, GradePartial, now)
	want := normalizeDate(now).AddDate(0, 0, 1)

	if !again.DueDate.Equal(want) || !partial.DueDate.Equal(want) {
		t.Fatalf("expected miss due date %s, got again=%s partial=%s", want, again.DueDate, partial.DueDate)
	}
}

func TestSchedulePartialDoesNotIncrementRepetitions(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	state := SchedulingState{
		Interval:    12,
		EaseFactor:  2.2,
		Repetitions: 4,
	}

	got := SM2{}.Schedule(state, GradePartial, now)

	if got.Repetitions != 0 {
		t.Fatalf("expected partial to reset repetitions to 0, got %d", got.Repetitions)
	}
}

func TestSchedulePartialThenAgainProduceDifferentEaseFactors(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	state := SchedulingState{
		Interval:    12,
		EaseFactor:  2.2,
		Repetitions: 4,
	}

	partial := SM2{}.Schedule(state, GradePartial, now)
	again := SM2{}.Schedule(state, GradeAgain, now)

	if partial.Interval != 1 || again.Interval != 1 {
		t.Fatalf("expected both miss grades to reset interval to 1, got partial=%d again=%d", partial.Interval, again.Interval)
	}
	if partial.EaseFactor <= again.EaseFactor {
		t.Fatalf("expected partial ease factor > again ease factor, got partial=%v again=%v", partial.EaseFactor, again.EaseFactor)
	}
}

func TestScheduleTenConsecutiveEasyReviewsStrictlyIncreaseIntervals(t *testing.T) {
	t.Parallel()

	s := SM2{}
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	state := NewState()
	intervals := make([]int, 0, 10)

	for range 10 {
		state = s.Schedule(state, GradeEasy, now)
		intervals = append(intervals, state.Interval)
		now = state.DueDate
	}

	for i := 1; i < len(intervals); i++ {
		if intervals[i] <= intervals[i-1] {
			t.Fatalf("expected strictly increasing intervals, got %v", intervals)
		}
	}
}

func TestScheduleAlternatingHardAndEasyStillGrows(t *testing.T) {
	t.Parallel()

	s := SM2{}
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	state := NewState()
	grades := []Grade{GradeEasy, GradeHard, GradeEasy, GradeHard, GradeEasy, GradeHard}
	lastInterval := 0

	for _, grade := range grades {
		state = s.Schedule(state, grade, now)
		if state.Interval < 1 {
			t.Fatalf("expected positive interval, got %d", state.Interval)
		}
		if state.Interval < lastInterval {
			t.Fatalf("expected alternating chain not to regress: last=%d current=%d", lastInterval, state.Interval)
		}
		lastInterval = state.Interval
		now = state.DueDate
	}
}

func TestScheduleCrossGradeIntervalOrdering(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	state := SchedulingState{
		Interval:    10,
		EaseFactor:  2.5,
		Repetitions: 3,
	}

	again := SM2{}.Schedule(state, GradeAgain, now)
	partial := SM2{}.Schedule(state, GradePartial, now)
	hard := SM2{}.Schedule(state, GradeHard, now)
	easy := SM2{}.Schedule(state, GradeEasy, now)

	if easy.Interval <= hard.Interval || hard.Interval <= again.Interval {
		t.Fatalf("expected interval ordering easy > hard > again, got easy=%d hard=%d again=%d", easy.Interval, hard.Interval, again.Interval)
	}
	if partial.Interval != again.Interval {
		t.Fatalf("expected partial and again to share miss interval reset, got partial=%d again=%d", partial.Interval, again.Interval)
	}
	if partial.EaseFactor <= again.EaseFactor {
		t.Fatalf("expected partial to be less punitive than again on ease factor, got partial=%v again=%v", partial.EaseFactor, again.EaseFactor)
	}
}
