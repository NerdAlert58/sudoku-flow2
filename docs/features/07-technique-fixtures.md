# Feature: Per-technique fixture suite + ladder-cap harness

**ID:** F-07 · **Roadmap piece:** P-07 · **Status:** Done (2026-08-07) — 13/13 fires-and-sound via ReplayVerify; 9/13 necessity+sufficiency proven; 4/13 recorded-evidence fallback (EVIDENCE.md); verifier PASS r1 (tuned)

## Description
The per-technique evidence the PRD's success criteria demand: a ladder-cap test harness
(solve with the ladder truncated below/at a position), curated fires-and-sound fixtures
for all 13 techniques run through the F-06 verifier, and necessity/sufficiency puzzles
for every technique where curation succeeds — with a recorded-evidence fallback where it
demonstrably does not (jellyfish pre-flagged by the PRD).

## How it fits the roadmap
W5, parallel with F-08. Off the critical path but gates F-13 (a ship without
per-technique evidence fails the PRD).

## Dependencies (must exist before this starts)
- F-06 oracle-replay — the oracle, the verifier, and the corpus proofs it extends

## Unblocks (what waits on this)
- F-13 ship — per-technique evidence is a ship gate

## Allow-list (source)
(none — test-only piece; the ladder-cap harness is test code)

## Allow-list (tests)
- solver/technique_fixtures_test.go
- solver/laddercap_test.go
- solver/export_test.go
- solver/testdata/techniques/**

## Read-only context
- PRD.md §Success criteria (per-technique coverage), §Domain context
- AUDIT.md D4, L1
- EVAL.md row "Per-technique" + §Datasets and fixtures
- DESIGN_DECISIONS.md ADR-0013

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
None.

## Acceptance criteria
- **AC-1:** For each of the 13 techniques there exists a committed fixture on which the
  technique fires, and the resulting event passes the F-06 verifier's soundness checks
  (13/13). Eval row: "Per-technique".
- **AC-2:** A ladder-cap harness exists that solves with the ladder truncated at any
  position, and its behavior is itself tested (capping below `naked_single`'s puzzle
  stalls it; capping at the full ladder reproduces normal results).
- **AC-3:** For every technique where curation succeeds: a necessity fixture (ladder
  capped below the technique → `stalled` at the state where only it can act) and a
  sufficiency fixture (capped at the technique → whole puzzle `solved`).
- **AC-4:** For every technique where necessity/sufficiency curation fails after a real
  attempt, the fixture file records the attempt evidence (what was searched, why
  isolation is believed unattainable) and the technique carries fires-and-sound
  status — the PRD's accepted fallback, never a silent skip. The final per-technique
  status table (proven vs fires-and-sound-with-evidence) lands in Implementation notes.

## Testing requirements
This is a test-only piece: fixtures + harness + assertions. Red-state discipline is
N/A-by-nature (fixtures against an already-built solver verify claims rather than drive
code); any fixture that FAILS is a solver defect and blocks this piece until resolved
upstream via the amendment path.

## Test command
(inherit from CONTEXT.md §Test discipline)

## Coverage command
(inherit)

## Coverage report
(inherit)

## Test-exempt lines
None.

## Health check
N/A (test-only piece)

## Rollback command
N/A (test-only piece)

## Env vars required
None.

## Readability budget
(inherit from CONTEXT.md §Rigor)

## Complexity exemptions
None.

## Manual setup required
None.

## Implementation notes (filled in by the building agent)
> Decisions, curation evidence, and the final per-technique status table land here.

**F-07 (test-author session, baseline 12e3933).** Ladder-cap harness `SolveCapped(g, maxPos)`
+ `LadderSize()` live in `solver/export_test.go` (in-package test code; no source change).
Harness self-tests in `solver/laddercap_test.go`: cap 0 stalls every incomplete state
(0 events, 1 iteration, grid untouched), cap 0 on a complete grid byte-matches Solve
(ADR-0014), cap 13 byte-matches Solve on 5 corpus seeds spanning all 4 sections.
Per-technique fixtures in `solver/technique_fixtures_test.go` (fires grids reuse the
F-04/F-05 constants; isolation grids committed as consts with provenance). Soundness =
ReplayVerify nil on the full solve from every fires fixture PLUS exact brute-force proof
of the opening event's eliminations (closes the oracle-truth gap on the six two-completion
fixture states, and supplies the committed positive verifier evidence F-06 flagged for
hidden_subset and jellyfish). Curation search and fallback evidence recorded in
`solver/testdata/techniques/EVIDENCE.md`.

Final status table (9 proven, 4 fires-and-sound-with-evidence):

| Pos | Technique | Fires | Sound | Necessity | Sufficiency | Status |
|-----|-----------|-------|-------|-----------|-------------|--------|
| 1 | naked_single | Y | Y | Y | Y | proven |
| 2 | hidden_single | Y | Y | Y | Y | proven |
| 3 | locked_candidates_pointing | Y | Y | Y | Y | proven |
| 4 | locked_candidates_claiming | Y | Y | Y | Y | proven (curated seed s1p5 + r2c7=9) |
| 5 | naked_subset | Y | Y | Y | Y | proven |
| 6 | hidden_subset | Y | Y | Y (at fixture) | N | fires-and-sound + evidence |
| 7 | x_wing | Y | Y | Y | Y | proven |
| 8 | swordfish | Y | Y | Y (at fixture) | N | fires-and-sound + evidence |
| 9 | jellyfish | Y | Y | Y (at fixture) | N | fires-and-sound + evidence (PRD pre-flagged) |
| 10 | xy_wing | Y | Y | Y | Y | proven |
| 11 | xyz_wing | Y | Y | Y (at fixture) | N | fires-and-sound + evidence |
| 12 | w_wing | Y | Y | Y | Y | proven |
| 13 | simple_colouring | Y | Y | Y | Y | proven |
