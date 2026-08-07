# Feature: Scan-parallel variant + honest negative-result benchmark

**ID:** F-12 · **Roadmap piece:** P-12 · **Status:** Done (2026-08-06) — equivalence 66/66 grids; guard non-vacuous (plant-proven); benchmark committed (~10.8x negative result); test-verifier PASS (tuned); readability PASS; leanness net -2 advisory

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
  solution, events, and the semantic counters Iterations and EventCount;
  CandidateChecks explicitly exempt — it meters scan work, the one thing the variant
  changes) on all 55 corpus seeds — the variant parallelizes scanning, never
  semantics — and is race-clean under `-race`.
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

- **Design (solver/scanparallel.go):** `SolveScanParallel` keeps `Solve`'s exact loop
  (complete → iterations++ → zero-candidate → pass); only the pass differs.
  `runPassScanParallel` launches one goroutine per ladder technique, each probing a
  private `solveState` snapshot (value-copied `grid` + `cands` arrays, own `checks`
  counter — race-clean with no mutex), records fired/not-fired into a per-index slice,
  joins via `sync.WaitGroup`, then re-fires the lowest firing rung on the real state.
  Re-firing on the real state (identical to the probe snapshot; detectors are
  deterministic) makes Status/Solution/Events/Iterations/EventCount/Grade byte-identical
  to sequential by construction; only `CandidateChecks` differs (probe counts are
  discarded, the real state counts only committed fires).
- **Measured result (negative, per PRD UC-5):** sequential ~2.12 ms vs scan-parallel
  ~22.85 ms per 10-VERY-HARD-seed batch (~10.8x slower), 459 KB/7,415 allocs vs
  7.46 MB/~132,865 allocs per op. Environment: go1.26.5 darwin/arm64, Apple M4 Max,
  GOMAXPROCS=14, macOS 26.5.1. Full raw output committed in docs/bench/scan-parallel.md.
- **Verification:** `gofmt -l .` clean; `go vet ./...` clean; `go build ./...` ok;
  `go test -race -count=1 ./...` all green repo-wide (equivalence over 55 corpus seeds +
  fixtures, containment guard, all prior suites).
