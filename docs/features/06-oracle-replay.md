# Feature: Oracle + replay proof + containment

**ID:** F-06 · **Roadmap piece:** P-06 · **Status:** Not started

## Description
The trust machinery: the test-only brute-force `oracle` (solution + 2-capped uniqueness
count), the ADR-0013 replay verifier with its own shadow candidate state, the
determinism byte-comparisons, the full-corpus proofs (solved + oracle-unique +
oracle-equal), and the import-graph containment tests that make the logic-only guarantee
mechanical.

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
- **AC-2:** The replay verifier checks every event of every corpus solve per ADR-0013:
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
