# Feature: Ladder upper tier — fish, wings, simple colouring

**ID:** F-05 · **Roadmap piece:** P-05 · **Status:** Not started

## Description
Techniques 7–13: `x_wing`, `swordfish`, `jellyfish` (plain fish only — no fins, no
sashimi), `xy_wing`, `xyz_wing`, `w_wing`, `simple_colouring` (colour trap + colour
wrap, implemented and event-logged in the positive deductive form AUDIT L1 requires —
never assume-propagate-revert). Exit bar: every one of the 55 corpus seeds solves
logic-only.

## How it fits the roadmap
W3, alone. The riskiest piece on the critical path (seven detectors, the colour-wrap
legality subtlety, and the all-55 exit bar).

## Dependencies (must exist before this starts)
- F-04 ladder-mid — ladder positions 1–6 and shared covering/subset helpers

## Unblocks (what waits on this)
- F-06 oracle-replay — a complete ladder to prove
- F-08 generator — the full ladder for grading
- F-12 scan-parallel-bench — the sequential solver to benchmark against

## Allow-list (source)
- solver/*.go (non-test files)

## Allow-list (tests)
- solver/*_test.go
- solver/testdata/upper/**

## Read-only context
- PRD.md §Domain context (rows 7–13), §Logic-only rule
- AUDIT.md L1 (the positive-form legality test — binding), L2
- DESIGN_DECISIONS.md ADR-0007
- EVAL.md rows "UC-1 Solve", "Per-technique", "UC-2 Determinism"

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
None.

## Acceptance criteria
- **AC-1:** All 55 puzzles in the corpus return `status:"solved"` with rule-conformant
  solutions, zero backtracking. Eval row: "UC-1 Solve".
- **AC-2:** Each of the seven techniques fires on a hand-built fixture with the frozen
  technique string, structurally correct witnesses, canonical serialization, and — for
  fish — base/cover sets enumerated in ascending lexicographic order.
- **AC-3:** Colour-wrap events are justified positively: the event's witness set is the
  conjugate-pair chain plus the shared unit, and no solver code path implements
  hypothesis-propagate-revert (verified by test fixtures whose only solution path is a
  wrap, plus source review recorded in notes; the replay proof in F-06 is the mechanical
  backstop).
- **AC-4:** Fish detection never fires finned/sashimi patterns: fixtures containing only
  a finned pattern do not produce a fish event (they stall or solve via other
  techniques).
- **AC-5:** Grading maps hardest-fired to bands per the frozen table (Hard for 7–10,
  Expert for 11–13) on synthetic fixtures with known ceilings.
- **AC-6:** Determinism double-run holds over all 55; all prior ACs (F-03/F-04) still
  green.

## Testing requirements
Per-technique fixtures, the all-55 golden corpus test, negative fixtures for AC-4,
grading-band tests, full determinism + regression reruns.

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
