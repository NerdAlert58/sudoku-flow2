# Feature: HTTP contract — full /v1 surface + edge matrix

**ID:** F-10 · **Roadmap piece:** P-10 · **Status:** Done (2026-08-07) — 35+ httpapi tests green; 55-corpus oracle-equal through the handler; verifier PASS (tuned, 5/5 mutations); pin condition discharged; readability PASS; leanness advisory-only

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
  requested difficulty for 3 seeded generations per band at handler level (the 25/band
  matrix lives at package level in F-08 AC-1); unknown difficulty → 400
  `{error, code:"invalid_input"}`; deadline exhaustion (test-injected) → 500
  `{error, code:"generation_failed"}`. Eval row: "UC-3 Generate" (integration).
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

- **Layout:** one file per handler (`solve.go`, `generate.go`, `batch.go`, `puzzles.go`)
  plus `gate.go` for the shared transport gate; wire shapes live in `contracts.go` (the
  C1 declared home); routes wired in `routes.go`. All functions within the readability
  budget (longest: `handleValidateBatch`, 33 lines).
- **Gate precedence** is one function, `gatePostJSON`: method (405 + exact `Allow`,
  ADR-0005) → content type via `mime.ParseMediaType` (415, header-only — body unread) →
  `Content-Length` fast-path 413 (zero reads, AUDIT A7) → lazy `http.MaxBytesReader`.
  `decodeBody` distinguishes `*http.MaxBytesError` (413) from all other decode errors
  (400 `invalid_input`), which makes `{"puzzles":[123]}` and `[]` schema-level 400s for
  free. `gateMethod` is shared by health (refactored from F-01's inline check, same
  message) and puzzles — three callers.
- **solveTimeMs** (ADR-0006): `time.Since` around `solver.Solve` only, converted in one
  helper `msOf` used by both solve and batch items. The invalid-400 path never starts a
  timer, so the golden body's literal `0` is structural, not formatted.
- **Non-solved `solution` field:** not pinned by any test or ADR; it carries
  `res.Solution.String()` (the partial grid) uniformly — the least-code mapping from
  SolveResult and useful to the F-11 UI. Flagging in case the intent was `""` for
  stalled/unsolvable.
- **hardestTechnique:** the solver's ladder registry is unexported by design, so
  `batch.go` carries a 13-entry `ladderRank` map (contract data mirroring PRD §Domain
  context). Drift is caught by `TestBatchContractFullCorpusInOrder`, which recomputes
  expectations from live solver events.
- **Batch fan-out** (AC-4): `wg.Go` (Go 1.25+) goroutine-per-puzzle writing into a
  pre-sized slice — each goroutine owns one index, counters are per-solve (ADR-0007),
  no locks needed; green under `-race` over the 55-puzzle corpus and the 256-item cap
  test.
- **Generate RNG:** per-request `math/rand.Rand` seeded from `crypto/rand` (concurrent
  requests must not share a seed; `crypto/rand.Read` cannot fail on Go 1.24+).
  `mapGenerateError` defaults to 500 `generation_failed` for anything that is not
  `ErrUnknownBand` — the honest-failure posture of ADR-0009. Enum validation is
  delegated to `generate.Generate`'s own `ErrUnknownBand` (checked before its retry
  loop), so the handler has no duplicate enum list.
- **Dependency materialization (coordination flag):** F-08's `generate/` package was a
  declared dependency but is unmerged (`feature/f-08`, commit 9fdf04c, based on this
  branch's baseline 23c93f3). Its committed files were restored verbatim into the
  working tree (`git restore --source=feature/f-08 -- generate/`) — uncommitted,
  untracked — so the repo builds and the F-10 suite runs. Merge order at integration:
  f-08 before (or with) f-10.
