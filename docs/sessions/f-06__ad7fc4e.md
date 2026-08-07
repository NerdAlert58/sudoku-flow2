# Session: F-06
Date: 2026-08-07T00:15:00-05:00
Agent: subagent (test-author: general-purpose + 1 fix round; builder: agents:golanger)
Piece / Brief: F-06 (docs/features/06-oracle-replay.md)
Baseline SHA -> Head: ad7fc4e -> (merge commit of feature/f-06)

## Accomplished
- oracle package: deterministic MRV brute-force Solve (count capped 2), ReplayVerify
  with own shadow state + independent geometry/singles (non-circularity
  verifier-confirmed by token scan: types-only solver imports), all 13 structural
  witness checks, 4 errors.Is sentinels with pinned intra-event check order.
- Replay proof: 55/55 corpus solves verified event-by-event (3228 events, 0 failures);
  oracle agreement + uniqueness for all 55; tamper quartet + witness-structure
  falsifier (fix round) committed.
- Containment: no shipped import of oracle; generate-only-via-httpapi guard armed
  (vacuous until F-08).

## Decisions made (and why)
- Witness checks assert per-elimination JUSTIFICATION, not completeness — verifier
  RULED COMPLIANT against ADR-0013's text (under-elimination can only stall, never
  unsound; colouring wrap stays exact-class, stricter than floor).
- MRV + ascending digits for oracle determinism.

## Deviations from the frozen plan
- Brief's test allow-list named F-03-era files; new sibling files created instead
  (coordinator-directed, logged in dispatch).

## Test evidence
- Red-state: oracle absent → compile failure; suite dry-run proven against a
  temporary degenerate stub (deleted).
- Green-state: repo-wide -race -count=1 green (author, builder, verifier ×2 each).
- test-verifier: round 1 FAIL (1 blocking: witness-structure arm had no committed
  falsifier — wholesale-skip mutation survived; 3 other mutations killed as predicted;
  non-circularity CONFIRMED airtight) → fix round: TestReplayVerify_WitnessStructureTamper
  (2 subtests, distinct failure modes, premises self-validated, bite proven by author
  AND verifier independently) → round 2 PASS (tuned; sole-and-sufficient falsifier
  confirmed among 95 solver tests).
- RUBRIC_GAPs → FOLLOWUPS.md + F-07 intake (per-checker positive coverage;
  hidden_subset/jellyfish committed evidence is F-07's assigned burden).
- Compliance evidence: (no compliance ACs — hats N/A)

## Deployment evidence
- **Target:** manual (cicd_deploy_hook: manual) — deployment: SKIP
- **Rollback command (unrun):** N/A — test-infrastructure piece

## Leanness review
- **RIGOR:** basic · **Findings:** 2 stdlib substitutions (~-13 lines) →
  **Disposition:** deferred to FOLLOWUPS.md (advisory; no re-gate churn mid-verification)

## Readability review
- **Mode:** blocking · **RIGOR:** basic · **Findings:** Readable. Ship. (max fn ~30/50;
  cyclomatic 10/10 at two sites — at-limit; files ≤239/400) · **Disposition:** PASS

## Open / next session
- F-07 consumes the cap harness need + per-checker positive coverage burden
  (hidden_subset + jellyfish have zero committed positive evidence — F-07 gate enforces).
- F-08's generated-slice replay (AC-6) now unblocked: oracle.ReplayVerify is exported.
