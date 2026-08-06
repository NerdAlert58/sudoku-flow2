# Feature: HTTP contract — full /v1 surface + edge matrix

**ID:** F-10 · **Roadmap piece:** P-10 · **Status:** Not started

## Description
Completes the frozen wire contract: `/v1/solve`, `/v1/generate`, `/v1/validate-batch`
(goroutine-per-puzzle fan-out), `/v1/puzzles`, handler-measured `solveTimeMs`, and the
exhaustive contract-edge test matrix (every route × every transport edge, headers
verbatim, envelope codes, solve-shape 400, ADR-0014 batch values, panic hygiene).

## How it fits the roadmap
W6 — the first convergence point (ladder + generator + catalog meet the shell). On the
critical path.

## Dependencies (must exist before this starts)
- F-05 ladder-upper — solver.Solve over the full ladder
- F-08 generator — generate.Generate for the /v1/generate handler
- F-09 catalog — catalog.Sections for the /v1/puzzles handler

## Unblocks (what waits on this)
- F-11 web-ui — live endpoints for the SPA's solve flow and visual smoke
- F-13 ship — the complete API it deploys

## Allow-list (source)
- httpapi/** (non-test files)

## Allow-list (tests)
- httpapi/*_test.go

## Read-only context
- ARCHITECTURE.md §Contracts C1, §Components (httpapi), §Summary (middleware chain)
- AUDIT.md A4, A5, A6, A7, S1, S2, S3, S4, P3, D1
- DESIGN_DECISIONS.md ADR-0004, ADR-0005, ADR-0006, ADR-0014
- USERS.md UC-1, UC-2, UC-3, UC-4, UC-6
- EVAL.md rows "UC-1 Solve", "UC-2 Determinism", "UC-3 Generate", "UC-4 Batch",
  "UC-6 Catalog", "Contract edge"
- SECURITY.md F-9

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
None.

## Acceptance criteria
- **AC-1:** POST /v1/solve returns the exact frozen shape for all 55 corpus puzzles
  (solved, oracle-equal via the F-06 suite, field-for-field shape assertion including
  field order via raw-body comparison on a golden response). Eval row: "UC-1 Solve"
  (integration).
- **AC-2:** Handler-level determinism: double-POST of identical puzzles yields identical
  bodies except `solveTimeMs` (which is a positive float). Eval row: "UC-2 Determinism".
- **AC-3:** POST /v1/generate returns `{puzzle, difficulty, grade}` with grade ==
  requested difficulty (seeded); unknown difficulty → 400 `{error, code:"invalid_input"}`;
  deadline exhaustion (test-injected) → 500 `{error, code:"generation_failed"}`. Eval
  row: "UC-3 Generate" (integration).
- **AC-4:** POST /v1/validate-batch over the full corpus returns 55 in-order results
  with `solvedCount:55` under `-race`; ADR-0014 values hold exactly for malformed lines,
  CRLF/whitespace lines, and stalled items; 257 items → 413 before any solving; >1 MiB →
  413. Eval row: "UC-4 Batch".
- **AC-5:** GET /v1/puzzles serves the four canonical sections (25/10/10/10) matching
  the embedded catalog. Eval row: "UC-6 Catalog" (integration).
- **AC-6:** The full contract-edge matrix passes: per route — 415 wrong content type,
  413 both caps, 400 malformed JSON, 405 wrong method with envelope + Allow, 404 unknown
  /v1 path, solve-shape 400 for domain-invalid puzzles (including `.`-blank variants and
  the ADR-0014 complete-grid fixture), frozen header set verbatim on every response, no
  ACAO header on any response including cross-origin requests. Eval row: "Contract
  edge".
- **AC-7:** A handler forced to panic returns 500 `{error, code:"internal_error"}` whose
  message is a fixed generic string: no panic value, stack frame, or file path appears
  in any error body (asserted over all error paths). **Source:** SECURITY.md §F-9
- **AC-8:** `solveTimeMs` is measured in the handler around solver.Solve only and is
  excluded from all byte-identity assertions (ADR-0006 observable).

## Testing requirements
Table-driven route×edge matrix; full-corpus integration tests; race-enabled batch tests;
golden-body shape test; panic-path tests. Every AC maps to named tests.

## Test command
(inherit from CONTEXT.md §Test discipline)

## Coverage command
(inherit)

## Coverage report
(inherit)

## Test-exempt lines
None.

## Health check
N/A (library/handler piece; deploy surface is F-13's)

## Rollback command
N/A (no deploy in this piece)

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
