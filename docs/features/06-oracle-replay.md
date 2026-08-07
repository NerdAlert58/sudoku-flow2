# Feature: Oracle + replay proof + containment

**ID:** F-06 · **Roadmap piece:** P-06 · **Status:** Done (2026-08-06) — 55/55 replay proof (3228 events, 0 failures); oracle agreement + uniqueness; tamper quartet + witness falsifier; verifier PASS r2 (tuned); readability PASS; leanness deferred to FOLLOWUPS

## Description
The trust machinery: the test-only brute-force `oracle` (solution + 2-capped uniqueness
count), the ADR-0013 replay verifier with its own shadow candidate state, the
determinism byte-comparisons, the full-corpus proofs (solved + oracle-unique +
oracle-equal), and the import-graph containment tests that make the logic-only guarantee
mechanical. The verifier is implemented as exported test-support API **in the `oracle`
package** (e.g. `oracle.ReplayVerify`) so downstream test code in other packages (F-07
fixtures, F-08's generated-slice AC-6) can import it — legal because the containment
rule bans only non-test importers of oracle — with `solver/replay_test.go` as its
corpus driver.

## How it fits the roadmap
W4, parallel with F-12 (disjoint files). On the critical path. This piece converts
"the solver looks right" into "every shipped solve is mechanically proven."

## Dependencies (must exist before this starts)
- F-05 ladder-upper — a complete ladder whose full corpus behavior can be proven

## Unblocks (what waits on this)
- F-07 technique-fixtures — the oracle + verifier + cap harness it consumes
- F-08 generator — the oracle its uniqueness tests require

## Allow-list (source)
- oracle/*.go (non-test files)

## Allow-list (tests)
- oracle/*_test.go
- solver/replay_test.go
- solver/replayer_test.go
- solver/corpus_test.go
- solver/determinism_test.go
- solver/importguard_test.go

## Read-only context
- ARCHITECTURE.md §Contracts C6, §Components (oracle)
- AUDIT.md L4, L5, D3
- DESIGN_DECISIONS.md ADR-0002, ADR-0013
- EVAL.md rows "UC-2 Replay proof", "UC-2 Determinism", "Solve-path containment"
- USERS.md UC-2

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
None.

## Acceptance criteria
- **AC-1:** The oracle, on hand-checkable tiny fixtures and all 25 ORIGINAL seeds,
  produces solutions agreeing with the ladder solver, and reports count==1 for all 55
  corpus seeds.
- **AC-2:** The replay verifier — exported test-support API in the `oracle` package,
  importable by test code in any package — checks every event of every corpus solve per
  ADR-0013:
  placements satisfy the named single's condition recomputed from the verifier's own
  shadow state AND equal the oracle value; every elimination existed as a candidate, is
  never the oracle's value, and its technique's witness pattern structurally holds;
  `gridAfter` equals prior grid plus exactly the stated placement; the final grid equals
  the oracle solution. 55/55 corpus solves pass with zero exceptions. Eval row: "UC-2
  Replay proof".
- **AC-3:** The verifier independently re-implements naked/hidden single detection and
  asserts no single was available whenever any elimination technique fired
  (cheapest-first scheduling check), over all corpus solves.
- **AC-4:** Determinism: consecutive full-corpus runs byte-match on events + the three
  counters. Eval row: "UC-2 Determinism".
- **AC-5:** Containment tests: `solver` imports stdlib only; no non-test package
  anywhere in the module imports `oracle`; and no non-test package other than `httpapi`
  imports `generate` (the full ARCHITECTURE C3 sealing claim). Eval row: "Solve-path
  containment".

## Testing requirements
This piece IS tests plus the oracle package. The verifier must not read solver
internals — only the public event log and its own shadow state (ADR-0013's
non-circularity requirement).

## Test command
(inherit from CONTEXT.md §Test discipline)

## Coverage command
(inherit)

## Coverage report
(inherit)

## Test-exempt lines
None.

## Health check
N/A (test infrastructure piece)

## Rollback command
N/A (test infrastructure piece)

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

- **Package shape (6 files, one concept each):** `oracle.go` brute force; `shadow.go`
  geometry + shadow candidate state + independent singles detection; `replay.go`
  verifier core (sentinels, event loop, pinned check orders); `witness.go` structural
  checks for locked-candidates/subsets + shared witness helpers; `wing.go` fish + three
  wings; `colour.go` simple colouring. Geometry (units/peers/sees) is re-derived from
  scratch — the verifier imports only solver's exported Grid/SolveResult/Event types
  (ADR-0013 non-circularity; solver's helpers are unexported anyway).
- **Brute force:** bitmask row/col/box masks, most-constrained-cell-first (ties by
  lowest index), digits ascending — deterministic returned solution, count capped at 2
  with early bail. A zero-candidate cell is picked first and prunes naturally.
- **Check order as pinned:** per event: shape/bounds guard → placement: oracle-equality
  (`ErrPlacementNotOracle`) → named single recomputation → witness-is-the-cell →
  gridAfter; elimination: scheduling (`ErrSingleAvailable`, run first as the
  cheapest-first mirror — legal since order-independent per the manifest) → liveness
  (`ErrEliminationNotCandidate`, all stated candidates) → oracle-truth
  (`ErrEliminationIsTruth`) → witness structure (descriptive) → gridAfter-unchanged →
  shadow update. Whole-result: final grid == oracle solution when status==solved and
  count==1. Non-solved / count!=1 results: events still validated; oracle-anchored
  checks (truth, placement-equality, final grid) skipped — no unique truth to compare.
- **Witness checks are per-elimination justification, not completeness:** each check
  asserts the pattern structurally holds in the shadow state AND every *stated*
  elimination is a target the pattern justifies; it does not require the event to state
  *all* justified targets (completeness is the solver's exhaustiveness, not event
  soundness — ADR-0013 asks that the witness pattern "structurally hold").
  Exception: colouring wrap pins "whole colour class" (the falsified class must be
  eliminated exactly), per the F-06 manifest.
- **Witness distinctness enforced centrally** (rejects e.g. a duplicated naked-subset
  cell faking k cells with k-1); hidden-subset requires exactly k confined digits whose
  spots cover the witness set exactly; fish requires ≥2 spots per base line and
  witness-set == base-line spots exactly (exact-k cover doubles as the finned
  exclusion, matching F-05); w_wing tries all A/B-vs-link-pair role assignments;
  colouring 2-colours over strong links recomputed from shadow, rejects
  non-connected/odd-cycle witness sets, wrap requires the eliminated class self-seeing.
- **Corpus evidence beyond the pinned suites** (throwaway probe, not committed): over
  all 55 seeds ReplayVerify passes with 0 failures across 3228 events; 11 of 13
  techniques fire naturally in-corpus (naked_single 2645, hidden_single 415, pointing
  88, claiming 21, naked_subset 25, x_wing 3, swordfish 1, xy_wing 13, xyz_wing 3,
  w_wing 12, simple_colouring 2). hidden_subset and jellyfish (0 corpus events) were
  probed green against the F-04/F-05 fixture grids (stalled solves, genuine single
  event each) — all 13 checkers have positive evidence ahead of F-07.
- **candidateChecks canon (f-05 carry-over):** the verifier reads no counters and
  freezes nothing about absolute `candidateChecks`; the as-built w-wing order remains
  canon, unthreatened by this piece.
