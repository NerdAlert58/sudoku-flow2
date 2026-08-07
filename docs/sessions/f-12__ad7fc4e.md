# Session: F-12
Date: 2026-08-06T23:00:00-05:00
Agent: subagent (test-author: general-purpose; builder: agents:golanger)
Piece / Brief: F-12 (docs/features/12-scan-parallel-bench.md)
Baseline SHA -> Head: ad7fc4e -> (merge commit of feature/f-12)

## Accomplished
- SolveScanParallel: goroutine-per-rung probes on value-copied solveState snapshots,
  join, lowest-firing-rung re-fire on real state — byte-identical event streams by
  construction; only CandidateChecks diverges (41/66 grids).
- Committed benchmark docs/bench/scan-parallel.md: sequential ~2.12ms vs scan-parallel
  ~22.85ms per 10-VERY-HARD batch — the PRD's anticipated measured NEGATIVE result
  (~10.8x loss; verifier re-measured ~10.7x independently).
- Containment guard proven non-vacuous by plant checks (author + verifier).

## Decisions made (and why)
- Confinement over synchronization (per-goroutine snapshots; probes mutate state, so
  clones were mandatory and give counter isolation free).
- Solve-loop mirror duplication accepted: threading a strategy parameter through
  production Solve would couple the contained experiment into the serving path against
  ADR-0015's spirit.

## Deviations from the frozen plan
- none. Amendment (3) recorded: AC-1 counter-scope clarified per verifier RUBRIC_GAP.

## Test evidence
- Red-state: compile failure (undefined SolveScanParallel); guard green-vacuous with
  plant-proof of non-vacuity.
- Green-state: repo-wide -race -count=1 green (author, builder, verifier each ran it).
- Coverage: n/a-delta (variant exercised by equivalence suite over 66 grids).
- test-verifier: PASS (tuned; both bite-mutations killed; benchmark re-measured within
  noise; plant-check repeated; tree restored hash-verified).
- Compliance evidence: (no compliance ACs — hats N/A)

## Deployment evidence
- **Target:** manual (cicd_deploy_hook: manual) — deployment: SKIP
- **Rollback command (unrun):** N/A — library/benchmark piece

## Leanness review
- **RIGOR:** basic · **Findings:** 1 shrink (redundant total counter, -2 lines) +
  1 deferred cross-allow-list dedup (readCorpusSections → testing.TB, ~-18 lines,
  blocked by parallel-piece allow-lists) · **Disposition:** advisory-only; dedup
  earmarked for FOLLOWUPS.md after W5/W6 land

## Readability review
- **Mode:** blocking · **RIGOR:** basic · **Findings:** Readable. Ship. (file 50/400;
  functions 15+23 lines; new-source cyclomatic ≤5) · **Disposition:** VERDICT: PASS

## Open / next session
- F-13 ship gate consumes docs/bench/scan-parallel.md as UC-5 evidence.
- FOLLOWUPS candidate: testing.TB widening for readCorpusSections.
