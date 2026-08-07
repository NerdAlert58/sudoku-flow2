# Feature: Ladder upper tier — fish, wings, simple colouring

**ID:** F-05 · **Roadmap piece:** P-05 · **Status:** Done (2026-08-06) — ALL 55 corpus puzzles solve logic-only; -race green; 98.8% coverage (solver 100% blocks); test-verifier PASS r2 (tuned, 6+1 mutations); readability PASS; leanness clean

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

- **Layout:** three new files, one concept each — `solver/fish.go` (x_wing /
  swordfish / jellyfish as one parameterized plain-fish core, k=2/3/4),
  `solver/wings.go` (xy_wing, xyz_wing, w_wing + shared geometry helpers:
  `sees`, `peerCells`, `exactCandidates`, `elimsSeeingAll`, `sortedCells`),
  `solver/colouring.go` (simple_colouring). Registry: seven `elimination(...)`
  entries appended to `ladder` in `solver/ladder.go` after `hidden_subset`, in
  the frozen order; bands Hard (7–10) / Expert (11–13) ride the existing
  per-entry `band` field, so grading needed no new code.
- **Plain-fish gate:** one check — `popcount(cover union) == k` — is both the
  fish definition and the finned/sashimi exclusion (AC-4). Base lines are
  filtered to >=2 candidate spots before combination enumeration; combos run
  lexicographic over the ascending eligible-line list via the existing
  `forEachCombo`. Unproductive plain fish (e.g. the {2,5} rectangle in the
  beyond-ladder stall grid) return false from the combo callback and
  enumeration continues, so they never fire and never mask the stall.
- **AUDIT L1 code shape (AC-3):** `detectSimpleColouring` builds the conjugate
  graph from units with exactly two candidate cells (already-true
  biconditionals), 2-colours each component by pinned BFS (FIFO, neighbours
  ascending, seed colour 0, first assignment wins), then concludes wrap by
  direct fact-combination: `classSharesUnit` finds two same-colour cells
  sharing a unit and the whole class is eliminated in one event. There is no
  trial assignment, no propagation, no revert anywhere in the package — source
  reviewed; F-06 replay is the mechanical backstop. Odd cycles need no special
  case: first-assignment-wins forces the colliding pair into one colour class
  and the wrap check catches them as same-colour-same-unit.
- **w_wing ordering detail:** eliminations depend only on (A, B, Y), not on
  which strong-link unit witnesses the pattern, so `wWingLink` computes elims
  first and short-circuits before the unit scan when empty; the canonical 0–26
  unit scan then only picks W1/W2 for the event. Event-identical to scanning
  units first, cheaper on candidateChecks, still deterministic.
- **Convention compliance:** every candidate query flows through the counted
  `hasCandidate` accessor (directly or via `digitSpotsInUnit` /
  `cellCandidates`); witnesses sort row-major, eliminations are born sorted by
  row-major scans; no maps are iterated (arrays/slices only), so determinism
  double-run holds by construction.
- **Readability budget:** all functions <=50 lines, nesting <=3, gocyclo max 9
  (<=10), files 100/206/139 lines (<=400). No `// COMPLEXITY:` markers needed.
- **Verification:** gofmt clean, `go vet` clean, `go test -race -count=1 ./...`
  green repo-wide including all-55 corpus (AC-1), determinism double-run
  (AC-6), iteration-cap invariant, finned-only negative (AC-4), grading bands
  (AC-5), golden ORIGINAL#1 log, and all F-03/F-04 suites. Corpus
  hardest-technique histogram: naked_single 25, locked_candidates_pointing 5,
  naked_subset 5, x_wing 1, xy_wing 9, w_wing 8, simple_colouring 2 (55/55
  solved; swordfish/jellyfish/xyz_wing appear in fixtures but are never the
  corpus-hardest).
