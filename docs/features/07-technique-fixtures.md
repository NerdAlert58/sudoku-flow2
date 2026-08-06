# Feature: Per-technique fixture suite + ladder-cap harness

**ID:** F-07 · **Roadmap piece:** P-07 · **Status:** Not started

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
