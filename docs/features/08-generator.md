# Feature: Generator — sealed dig-and-grade generation

**ID:** F-08 · **Roadmap piece:** P-08 · **Status:** Not started

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
