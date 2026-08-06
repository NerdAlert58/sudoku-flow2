# Feature: Ladder mid-tier — locked candidates + subsets

**ID:** F-04 · **Roadmap piece:** P-04 · **Status:** Not started

## Description
Techniques 3–6 of the frozen ladder: `locked_candidates_pointing`,
`locked_candidates_claiming`, `naked_subset` (pairs/triples/quads), `hidden_subset`
(pairs/triples/quads) — each a pure elimination technique firing in frozen order with
canonical tie-breaks, each emitting one pattern's eliminations per event.

## How it fits the roadmap
W2, alone (extends solver/**). On the critical path between core and upper ladder.

## Dependencies (must exist before this starts)
- F-03 solver-core — the loop, technique ordering, event/metric machinery

## Unblocks (what waits on this)
- F-05 ladder-upper — the subset/covering helpers and the next ladder positions

## Allow-list (source)
- solver/*.go (non-test files)

## Allow-list (tests)
- solver/*_test.go
- solver/testdata/mid/**

## Read-only context
- PRD.md §Domain context (ladder rows 3–6)
- AUDIT.md L1, L2, L3
- DESIGN_DECISIONS.md ADR-0007
- EVAL.md rows "Per-technique", "UC-2 Determinism"

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
None.

## Acceptance criteria
- **AC-1:** Each of the four techniques fires on a hand-built fixture state, emitting
  the frozen technique string, structurally correct `witnessCells[]`, and eliminations
  in canonical row-major-then-digit order (fires-half of the EVAL "Per-technique" row;
  soundness proof completes in F-06/F-07).
- **AC-2:** Grading reflects the new bands: a puzzle whose hardest fired technique is a
  mid-ladder technique grades `Medium` when solved.
- **AC-3:** A technique fires only when nothing cheaper can act: fixtures verify a
  grid with an available single never fires a mid-ladder technique in that pass.
- **AC-4:** All F-03 acceptance criteria still hold (ORIGINAL corpus, determinism,
  invariants) — regression suite green.
- **AC-5:** Determinism double-run holds over every corpus puzzle that now solves; the
  count of corpus puzzles solving with the ladder capped at technique 6 is recorded in
  Implementation notes (no target — the 30 advanced seeds may need the upper ladder).

## Testing requirements
Per-technique fixture tests (fire + canonical serialization + cheapest-first
discipline), grading tests, regression + determinism reruns.

## Test command
(inherit from CONTEXT.md §Test discipline)

## Coverage command
(inherit)

## Coverage report
(inherit)

## Test-exempt lines
None.

## Health check
N/A (library piece)

## Rollback command
N/A (library piece)

## Env vars required
None.

## Readability budget
(inherit from CONTEXT.md §Rigor)

## Complexity exemptions
None.

## Manual setup required
None.

## Implementation notes (filled in by the building agent)
> Decisions and rationale land here as the piece builds.
