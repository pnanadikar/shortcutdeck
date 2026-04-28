# Scheduler Unit Tests

## Ease factor floor

Apply Again (grade 0) repeatedly — ease factor must never drop below 1.3 regardless of how many consecutive failures
Apply Hard (grade 3) repeatedly — same floor check, slower descent
Ease factor that is already exactly 1.3 receiving another Again — must stay at 1.3, not go below

## Relapse after a good run

Card with repetitions=5, high ease factor, long interval receives Again — interval must reset to 1, repetitions must reset to 0
Card with repetitions=5 receives Again, then recovers through Hard/Good/Easy — verify it rebuilds correctly from the reset state rather than resuming from where it was

## Grade=1 (Partial) distinct behaviour

Grade=1 must count as a miss (interval resets) but ease factor penalty must be less severe than Grade=0
Grade=1 followed by Grade=0 — verify both reset interval but produce different ease factors
Grade=1 must not increment repetitions

## Interval compounding over a long chain

10 consecutive Easy (grade 5) reviews — verify each interval is strictly larger than the previous and ease factor grows correctly
10 consecutive Good (grade 4) reviews — verify intervals grow but more slowly than Easy chain
Alternating Hard/Good — verify interval still grows but slowly

## Boundary conditions on new cards

NewState() produces exactly interval=0, ease=2.5, repetitions=0
First review of a new card with Easy — verify interval is 1 (not 0, not 2)
First review with Again — verify interval resets to 1 and repetitions stay at 0

## DueDate correctness

After a review, DueDate must be exactly now + interval days, not off by one
Timezone neutrality — DueDate calculation must not depend on local timezone

## Cross-grade consistency

For the same starting state, Easy (5) must always produce a longer interval than Hard (3)
Partial (1) and Again (0) both reset the interval, so interval ordering is not the meaningful check between those two grades
The meaningful cross-grade assertions are:
- Easy > Hard for interval growth
- Partial is less punitive than Again for ease-factor penalty severity

## Invalid grade inputs

The grade sequence is intentionally sparse: only 0, 1, 3, and 5 are valid inputs
Because `Grade` is currently an `int`, tests should cover unexpected values such as 2, 4, or 99 once the implementation defines explicit behavior
The design should choose one policy and document it clearly:
- return an error
- panic with a clear programmer-error message
- or make invalid values impossible via a stronger type/API boundary
