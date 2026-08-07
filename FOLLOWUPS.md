# FOLLOWUPS

Deferred advisory findings from build reviews (dated; none block Done).

## 2026-08-06
- F-12 leanness: widen `readCorpusSections` to `testing.TB` and collapse
  `veryHardGrids` in scanparallel_bench_test.go (~-18 lines) — was allow-list-blocked
  during parallel waves.
- F-06 leanness: `oracle/witness.go` — replace hand-rolled `contains` with
  `slices.Contains` (3 call sites, -8 lines) and collapse `slotIn` body to
  `slices.Index` (-5 lines).
- F-05 residual: zero-elimination-event violations surface as deterministic
  hang-timeout rather than a crisp assertion (verifier M4); a Solve()-internal guard
  would close it.
- F-06 verifier RUBRIC_GAP-3: ReplayVerify does not validate Seq monotonicity or
  EventCount/Iterations consistency (ADR-0007 solver invariants own the counters).
- F-06 round-2 note: per-checker mutation falsifiers exist for pointing + fish only;
  the wholesale-skip falsifier covers the arm. F-07 fixtures add positive per-checker
  coverage.
- F-07 verifier RUBRIC_GAP-1: SolveCapped's zero-candidate branch has no self-test
  (mutation survives today because every committed capped state genuinely stalls);
  add a capped-unsolvable byte-match test if iso fixtures ever evolve.
- F-08 verifier RUBRIC_GAP-2: add a two-call seed-determinism self-equality test
  (property, not golden bytes) to catch hidden global-rng/time deps; verified true
  empirically today.
