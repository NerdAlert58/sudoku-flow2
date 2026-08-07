# Feature: Generator — sealed dig-and-grade generation

**ID:** F-08 · **Roadmap piece:** P-08 · **Status:** Done (2026-08-07) — 100-seed matrix + 20-puzzle replay slice green; grade==band by construction; sealed counter (guard live-fire-proven); verifier PASS r1 (tuned, 4/4 mutations); readability PASS; leanness clean

## Description
The `generate` package: randomized full-grid fill (backtracking, sealed), clue removal
under a ≤2-solution uniqueness counter, grading via the shipped `solver.Solve`, and a
grade-targeted accept/retry loop under the caller's context deadline. Grade equals the
requested band by construction; exhaustion is an error, never a mislabeled puzzle.

## How it fits the roadmap
W5, parallel with F-07 (disjoint surfaces). On the critical path (F-10 needs it).

## Dependencies (must exist before this starts)
- F-05 ladder-upper — the full ladder for grading
- F-06 oracle-replay — the oracle its uniqueness tests verify against

## Unblocks (what waits on this)
- F-10 http-contract — the /v1/generate handler consumes generate.Generate

## Allow-list (source)
- generate/*.go (non-test files)

## Allow-list (tests)
- generate/*_test.go

## Read-only context
- ARCHITECTURE.md §Contracts C3, §Components (generate)
- AUDIT.md P2, L1 (generator exemption boundaries)
- DESIGN_DECISIONS.md ADR-0002, ADR-0009
- USERS.md UC-3
- EVAL.md row "UC-3 Generate"

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
None.

## Acceptance criteria
- **AC-1:** With seeded RNG, 25 generations per band (100 total) each produce a puzzle
  that is oracle-unique (count==1), solves via the ladder (`status:"solved"`), and whose
  solver-computed grade equals the requested band — within the 5-second per-call
  deadline locally. Eval row: "UC-3 Generate".
- **AC-2:** `Generate` returns only grade==band results by construction: no code path
  can return a puzzle whose grade differs from the request (verified by API shape — the
  accept condition — and tests over many seeds).
- **AC-3:** On context deadline/attempt exhaustion, `Generate` returns an error (the
  handler maps it to 500 `generation_failed` in F-10); a test with an
  artificially-short deadline observes the error path.
- **AC-4:** Sealing holds: `generate` imports `solver` and never vice versa (extends the
  F-06 import-guard test); no counter/attempt data appears in any public return value.
- **AC-5:** Generator randomness never touches solve-path determinism: the F-06
  determinism suite still passes unchanged (regression).
- **AC-6:** 20 seeded generated puzzles (5 per band) pass the replay verifier
  end-to-end — imported from the `oracle` package's exported test-support API (F-06's
  home for it; test-scope import, legal under the containment rule) inside
  `generate/*_test.go` — every event of every generated-puzzle solve passes all
  ADR-0013 checks, final grids equal the oracle solutions. Eval row: "UC-2 Replay
  proof" (generated slice).

## Testing requirements
Seeded-RNG matrix tests (fixed seeds committed), deadline error-path test, sealing
assertions, regression rerun.

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

- **Shape:** three files — `generate/generate.go` (C3 entry + dig loop), `generate/fill.go`
  (randomized full-grid backtracking filler, rng-shuffled digit order per cell),
  `generate/count.go` (uniqueness counter, capped at 2, fewest-candidates cell choice).
  Imports across the package: stdlib + `solver` only.
- **Accept condition (AC-2):** `dig` returns only when
  `solver.Solve(g).Status == "solved" && res.Grade == want`; `Generate` then returns the
  band's mapped literal. A grade/band mismatch is unrepresentable — there is no code path
  that returns a puzzle without that check having passed.
- **Grade-on-every-removal:** rather than digging to minimal and grading once, `dig` grades
  after every uniqueness-preserving removal once the grid is at or below `maxGivens = 45`
  givens. This catches each band at whatever depth its grade first appears, which is why
  no per-band depth tuning was needed. The 45-given floor also guarantees every accepted
  puzzle has ≥ 36 blanks (AC-1's EventCount > 0 holds with margin).
- **Retry unit:** a fresh fill per attempt (no re-dig of the same grid, no restart
  heuristics). Empirically unnecessary to do more: across the 120 committed seeds and an
  800-seed robustness sweep (200/band, seeds 70001–70200), zero budget exhaustions; worst
  observed call 79 ms against the 5 s deadline.
- **Determinism/reproducibility:** all rng consumption (fill shuffles, `rng.Perm(81)` dig
  order) is sequential on the caller's `*rand.Rand`; the uniqueness counter uses a fixed
  1..9 digit order, so it consumes no randomness. Same seed → same puzzle.
- **Ctx discipline (AC-3):** band validated first; `ctx.Err()` checked at the top of the
  attempt loop and before every removal inside `dig` — a dead context can never reach the
  accept path. Exhaustion error wraps both `ErrBudgetExhausted` and `ctx.Err()` (two-`%w`
  wrap); error returns are `("", "")`.
- **Sealing (AC-4):** counter/attempt state lives in unexported types; the count reaches
  callers only as dig's keep/restore decision. No oracle import anywhere in shipped code —
  the F-06 source-level import guard passes unchanged, as does the solver determinism
  suite (AC-5).
- **Measured (Apple Silicon, no race):** attempts per call min/med/p90/max over the 30
  committed seeds per band — easy 1/1/1/1 (42µs–196µs), medium 1/3/8/29 (0.4ms–33ms),
  hard 1/9/28/39 (0.4ms–50ms), expert 1/4/13/16 (0.4ms–20ms).
