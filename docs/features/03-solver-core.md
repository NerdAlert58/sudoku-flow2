# Feature: Solver core — grid, parse, singles, loop, events, metrics

**ID:** F-03 · **Roadmap piece:** P-03 · **Status:** Done (2026-08-06) — repo-wide -race green; 95.8% coverage, solver blocks 100%; test-verifier PASS (tuned, 8/8 ACs, mutation-verified); readability PASS; leanness clean

## Description
The deterministic heart: the grid/candidate model, input parsing and validation, the
counted candidate accessor, naked and hidden singles, the one-event-per-pass solve loop
with the priority-zero unsolvable check and pre-pass completion check, event emission,
the metric counters, and grading for the singles tier. Exit bar: every ORIGINAL seed
solves by singles alone, byte-deterministically.

## How it fits the roadmap
W1 (parallel with F-02, F-09), first piece of the solver spine, on the critical path.
Everything ladder-shaped extends the loop this piece freezes.

## Dependencies (must exist before this starts)
- F-01 walking-skeleton — go.mod (module root)

## Unblocks (what waits on this)
- F-04 ladder-mid — the technique interface, loop, and event/metric machinery

## Allow-list (source)
- solver/*.go (non-test files)

## Allow-list (tests)
- solver/*_test.go
- solver/testdata/core/**

## Read-only context
- ARCHITECTURE.md §Components (solver), §Contracts C2
- AUDIT.md L2, L3, L5, D1
- DESIGN_DECISIONS.md ADR-0002, ADR-0006, ADR-0007, ADR-0008, ADR-0014
- USERS.md UC-1, UC-2
- EVAL.md rows "UC-1 Solve", "UC-2 Determinism", "Solve-path containment"
- PRD.md §Domain context (ladder table rows 1–2), §Logic-only rule

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
None — no CI/CD role.

## Acceptance criteria
- **AC-1:** `solver.Parse` accepts exactly the frozen grammar (81 chars, `1-9` givens,
  `0` or `.` blanks — `.` fixtures included since the corpus never uses it), rejects bad
  lengths, bad characters, and duplicate givens in any row/column/box with a
  distinguishable error.
- **AC-2:** All 25 ORIGINAL seeds return `status:"solved"` using only `naked_single` and
  `hidden_single` events, with correct solutions (validated against full-grid rule
  conformance at this stage; oracle cross-check arrives in F-06). Eval row: "UC-1 Solve"
  (ORIGINAL slice).
- **AC-3:** Two consecutive in-process solves of every ORIGINAL seed produce
  byte-identical `events[]`, `iterations`, `eventCount`, `candidateChecks`. Eval row:
  "UC-2 Determinism".
- **AC-4:** Loop semantics observably match ADR-0007: one productive event per pass;
  `iterations == eventCount` for every solved seed; a hand-built stalled fixture yields
  `iterations == eventCount + 1`; events carry canonical serialization order.
- **AC-5:** The ADR-0014 complete-grid edge holds: an 81-given valid grid returns
  solved/`"Easy"`/0/0/0/`events:[]`; a complete-but-duplicate grid fails parse.
- **AC-6:** A hand-built zero-candidate fixture returns `unsolvable` (checked before
  `stalled` can be concluded); a valid-but-above-singles grid returns `stalled` with
  grade `""`.
- **AC-7:** Metric counters are per-solve state: concurrent solves of different puzzles
  under `-race` produce the same per-puzzle counters as serial runs.
- **AC-8:** An import-guard test asserts the solver package imports nothing beyond the
  Go standard library. Eval row: "Solve-path containment" (base assertion; extended in
  F-06/F-12).

## Testing requirements
Unit tests for parse grammar/table-driven invalid inputs; golden-corpus test over the
ORIGINAL section; determinism double-run; loop-invariant and edge fixtures; race test
for AC-7. Every AC maps to at least one named test.

## Test command
(inherit from CONTEXT.md §Test discipline)

## Coverage command
(inherit)

## Coverage report
(inherit)

## Test-exempt lines
None.

## Health check
N/A (library piece; no deploy)

## Rollback command
N/A (library piece; no deploy)

## Env vars required
None.

## Readability budget
(inherit from CONTEXT.md §Rigor)

## Complexity exemptions
None.

## Manual setup required
None.

## Implementation notes (filled in by the building agent)
> The agent implementing this feature records its decisions and rationale here as it
> builds. Cross-cutting discoveries propagate to ROADMAP.md or ARCHITECTURE.md.

- **Layout:** `grid.go` (Grid, Parse, sentinels, canonical 27-unit table), `event.go`
  (wire-frozen event structs), `candidates.go` (bitmask candidate state, counted
  accessor, placement propagation, zero-candidate scan), `ladder.go` (ordered technique
  registry + `runPass`), `singles.go` (the two fire funcs), `solver.go` (Solve loop,
  per-solve state, result/grade assembly). All files well inside the readability budget.
- **Candidate state is incremental**, not per-pass recomputed: `[81]uint16` bitmasks
  initialized once from the givens; a placement zeroes the cell's mask and strips the
  digit from all three houses. Chosen because F-04/F-05 eliminations must persist across
  passes — recompute-from-grid would silently undo them.
- **Counting convention:** `solveState.hasCandidate` is the single counted accessor
  (ADR-0007). Naked-single detection queries all 9 digits per empty cell, then tests
  count==1; hidden-single queries every empty cell of a unit per digit; both
  short-circuit only at the first canonical firing instance. Candidate initialization
  and the top-of-pass zero-candidate scan read the bitmasks directly — setup and
  set-emptiness tests, not (cell,digit) detection queries — so they add zero checks,
  which also keeps the ADR-0014 complete-grid edge at exactly 0/0/0.
- **Hidden-single scan nesting:** units outer (rows 0-8 → cols 0-8 → boxes 0-8
  row-major), digits ascending inner, first sole place wins.
- **Ladder extension point (F-04/F-05):** `ladder.go` holds an ordered
  `[]technique{band, fire}`; `runPass` walks it cheapest-first and records the
  highest-fired index. Grade = band of the hardest technique fired; `"Easy"` when
  nothing fired on a solved grid (ADR-0014); `""` for stalled/unsolvable. Later pieces
  append entries without touching `Solve`.
- **Technique name strings** live in the fire funcs (passed to `place`); ladder entries
  carry only the band. A `name` field is deferred until a real caller needs it
  (F-04's `hardestTechnique` is the likely one).
- **Status strings** are literals at the three `result` call sites; exported constants
  deferred until httpapi consumes them (F-04+).
