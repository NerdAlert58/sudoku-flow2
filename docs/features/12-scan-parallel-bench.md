# Feature: Scan-parallel variant + honest negative-result benchmark

**ID:** F-12 · **Roadmap piece:** P-12 · **Status:** Not started

## Description
UC-5's honesty artifact: `solver.SolveScanParallel` (an intra-puzzle scan-parallel
variant reachable only by explicit call), a Go benchmark comparing it against the
sequential solver on the hardest seeds, committed measured results with the
negative-result statement, and the static-scan containment test that keeps the variant
off every serving path forever.

## How it fits the roadmap
W4+, parallel with F-06/F-10/F-11 (disjoint files). Off the critical path; gates F-13.

## Dependencies (must exist before this starts)
- F-05 ladder-upper — the sequential solver to vary and benchmark against

## Unblocks (what waits on this)
- F-13 ship — the committed benchmark is a ship gate (PRD UC-5)

## Allow-list (source)
- solver/scanparallel.go
- docs/bench/scan-parallel.md

## Allow-list (tests)
- solver/scanparallel_test.go
- solver/scanparallel_bench_test.go
- solver/scanparallel_guard_test.go

## Read-only context
- DESIGN_DECISIONS.md ADR-0015
- USERS.md UC-5
- EVAL.md rows "UC-5 Parallelism evidence", "Solve-path containment"
- PRD.md UC-5 (the negative-result framing of record)

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
None.

## Acceptance criteria
- **AC-1:** `SolveScanParallel` produces results identical to `Solve` (same status,
  solution, events, counters) on all 55 corpus seeds — the variant parallelizes
  scanning, never semantics — and is race-clean under `-race`.
- **AC-2:** A `go test -bench` benchmark compares sequential vs scan-parallel on the 10
  VERY-HARD seeds; the measured numbers are committed to `docs/bench/scan-parallel.md`
  with the honest conclusion (expected: sequential wins on 9×9 — a measured negative
  result, stated as such; if measurement surprises, the file says what was actually
  measured). Eval row: "UC-5 Parallelism evidence".
- **AC-3:** The static-scan guard test walks all non-test Go source outside the solver
  package and fails on any reference to `SolveScanParallel`. Eval row: "Solve-path
  containment".

## Testing requirements
Equivalence tests over the corpus, the benchmark itself, the guard test.

## Test command
(inherit from CONTEXT.md §Test discipline)

## Coverage command
(inherit)

## Coverage report
(inherit)

## Test-exempt lines
None.

## Health check
N/A (library/benchmark piece)

## Rollback command
N/A (library/benchmark piece)

## Env vars required
None.

## Readability budget
(inherit from CONTEXT.md §Rigor)

## Complexity exemptions
None.

## Manual setup required
None.

## Implementation notes (filled in by the building agent)
> Decisions, the measured numbers, and the benchmark environment land here.
