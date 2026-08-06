# EVAL — sudoku-flowN

This document defines how we know the system meets the bar. Acceptance is not subjective:
every UC has a measurable success signal, a threshold, and a regression trigger. The
ground truth for solver correctness is the test-only brute-force `oracle` package
(ADR-0013) — never the solver's own output. No LLM-as-judge is used anywhere.

## Eval matrix

| UC | Success signal | How measured | Dataset / fixture | Ship threshold | Regression trigger |
|---|---|---|---|---|---|
| UC-1 Solve | Every corpus puzzle returns `status:"solved"`, solution equals oracle, contract shape exact | Golden corpus test (`solver` + handler level) over all seeds; JSON shape asserted field-for-field | `puzzles.txt` (55) | 55/55 solved, 0 shape deviations | Any corpus puzzle not solved, or any response field added/removed/renamed |
| UC-2 Replay proof | Every event of every solve passes the ADR-0013 verifier (placement forced + oracle-equal; eliminations existed + never oracle-true + witness pattern holds; gridAfter exact; final grid == oracle; no single available when an elimination fired) | Automated replay verifier in CI | All 55 corpus puzzles + 20 seeded generated puzzles (5/band) | 100% of events verified, 0 exceptions | Any single event fails any verifier check |
| UC-2 Determinism | Identical input → identical `events[]`, `iterations`, `eventCount`, `candidateChecks` | Run each corpus solve twice in-process, byte-compare (solveTimeMs excluded per ADR-0006); handler-level double-POST compare | `puzzles.txt` (55) | 55/55 byte-identical | Any diff between consecutive runs of the same input |
| UC-3 Generate | For each band: generated puzzle is oracle-unique, ladder-solved, `grade == difficulty` | Seeded-RNG generator test, N=25 per band; oracle uniqueness ≤2-count check; unknown-difficulty → 400 envelope test | Seeded RNG (fixed seeds committed in test) | 100/100 generations valid & correctly graded within the 5s deadline locally | Any non-unique, non-ladder-solvable, or mis-graded generation; any deadline exhaustion at N=25/band locally |
| UC-4 Batch | Full-corpus batch returns 55 in-order results, `solvedCount:55`; malformed lines are per-item failures; caps enforced 413-before-solve; race-clean | Handler test with all 55 + synthetic batches (CRLF/whitespace lines, malformed lines, 257 items, >1 MiB) under `-race` | `puzzles.txt` + synthetic batch fixtures | All assertions pass under `-race` | Any ordering, aggregation, cap, or race failure |
| UC-5 Parallelism evidence | Goroutine-per-puzzle batch is race-clean; scan-parallel variant benchmarked, committed result shows sequential wins | `go test -race` (every CI run) + `go test -bench` comparing sequential vs flagged scan-parallel; results committed at `docs/bench/scan-parallel.md` | Benchmark corpus: the 10 VERY-HARD seeds | Race: 0 findings. Benchmark: committed file exists with measured numbers and the negative-result statement | Race finding; benchmark file drift from re-measured reality at rebuild |
| UC-6 Catalog | `GET /v1/puzzles` returns exactly 4 sections named Original/Medium/Hard/Very Hard with 25/10/10/10 puzzles; embedded copy byte-equals repo root | Handler test + drift-guard test | `puzzles.txt` + `catalog/puzzles.txt` | Exact match on names, counts, contents; drift test green | Any drift or count/name deviation |
| Contract edge | Every route enforces: 415 wrong content type, 413 both caps, 400 malformed JSON, 405 wrong method w/ envelope + Allow, 404 unknown /v1 path, 500 panic recovery, the frozen header set verbatim on every response (CSP per AUDIT.md S1, `Strict-Transport-Security: max-age=63072000`, frame-denial, nosniff), no ACAO header ever, solve-shape 400 for invalid puzzles | Table-driven handler tests per route × per edge case | Synthetic requests (incl. `.`-blank puzzles, duplicate givens, bad lengths) | Every cell of the route×edge matrix passes | Any envelope code, HTTP status, or header deviation |
| Per-technique | Each of the 13 techniques: fires on its curated fixture and is sound (verifier); where isolable: necessity (ladder capped below → stalls) and sufficiency (capped at → solves) | Per-technique fixture tests with ladder-cap harness | Curated fixture set (built during the fixtures piece; see Datasets) | 13/13 fire-and-sound; necessity+sufficiency proven for every technique where curation succeeds; any fallback carries recorded evidence of the attempt | Any technique failing to fire on its fixture, or an unsound deduction |
| UI (SPA) | Every PRD UI feature present and functional: (1) manual entry, (2) 81-char paste, (3) Clear, (4) tier-grouped seed dropdown fed by /v1/puzzles, (5) solve with status/grade/quartet display, (6) ladder-ordered technique histogram + counts + difficulty, (7) step-viewer with placement/witness/elimination highlighting, (8) transport controls (first/prev/Play/next/last) + step position + per-step description, (9) clickable event-log rows jumping steps, (10) technique-explanation panel w/ band chip covering all 13 techniques + static pre-solve hint, (11) a11y floor (cell aria-labels, native controls, aria-live status, visible focus) | Two layers: (a) automated DOM-contract test — Go test asserting the embedded assets declare the required element hooks, zero inline script/style, zero innerHTML in app.js source; (b) operator visual smoke in a real browser walking the 11-item checklist on a solved seed | Embedded `web/` assets + one Medium seed | 11/11 checklist items pass in BOTH layers | Any checklist item failing; any inline code appearing (CSP violation) |
| Solve-path containment | Backtracking and the parallel variant are unreachable from the serving path | Import-graph test: `solver` imports stdlib only; no shipped package imports `oracle`; plus a static-scan test asserting no non-test code outside benchmarks references `SolveScanParallel` | The package graph itself | All containment assertions pass | Any new import edge or reference breaching containment |
| Coverage | Total statement coverage ≥ 80% | `go test -race -coverprofile -coverpkg=./... ./...` + `go tool cover -func`, float-safe compare | The full test suite | ≥ 80.0% | CI red under 80.0% |
| Deployed health | Production answers `GET /v1/health` 200 `{status:"ok",...}` and `GET /` 200 text/html | Post-deploy smoke in deploy workflow, 5 retries × 3s (cold start) | Live deployment | Both checks green within retry budget | Smoke failure on any deploy |

## Datasets and fixtures

- **`puzzles.txt`** (repo root; 55 puzzles: 25 ORIGINAL singles-only — a frozen invariant
  — plus 10/10/10 MEDIUM/HARD/VERY-HARD). Owner: the operator; the file is frozen seed
  data. Version control: any edit must update `catalog/puzzles.txt` in the same change
  (drift test). Section labels are provenance, not grade assertions — no test may assert
  `grade == section name` (AUDIT.md D3).
- **Curated per-technique fixtures** (`solver/testdata/techniques/`). To be curated during
  the fixtures build piece: for each of the 13 techniques, a grid state where the
  technique fires (fires-and-sound), and where achievable a puzzle whose exact hardest
  step is the technique (necessity/sufficiency via the ladder-cap harness). Coverage:
  happy path per technique + the recorded-evidence fallback for techniques that resist
  isolation (PRD's escape hatch; jellyfish pre-flagged). Owner: the build's test-author;
  curation is a Day-Zero prerequisite for the solver pieces' full acceptance.
- **Synthetic invalid/edge inputs** (table-driven, inline in tests): bad lengths, bad
  characters, duplicate givens, `.`-blank variants (the corpus never uses `.` — AUDIT.md
  D1), an already-complete valid grid (must return `solved`, grade `"Easy"`,
  `iterations:0`, `events:[]` — ADR-0014), a complete-but-invalid grid (duplicate → 400
  invalid_input), CRLF/whitespace batch lines, malformed batch entries, over-cap bodies
  and lists.
- **Seeded generation fixtures**: fixed RNG seeds committed in the generator tests so
  UC-3's 100-generation matrix is reproducible.

## Ground-truth process

Ground truth is computed, not labeled: the `oracle` package's brute-force solver (with a
2-capped solution counter) defines the correct solution and uniqueness for any grid, and
is validated by its own unit tests on hand-checkable tiny cases plus cross-checks against
the 25 singles-only seeds (where the ladder solver and oracle must agree). New fixtures
are added by committing the puzzle string plus a test asserting the oracle finds exactly
one solution; disagreements between solver and oracle always resolve in the oracle's
favor pending investigation. Labels live in the repo as test data; changes go through PR
+ CI like all code.

## How EVAL.md is consumed

Every piece in the delivery plan that ships a UC must cite the relevant Eval matrix row in
its acceptance criteria. A piece is not Done until its eval target hits the ship threshold
— "build complete" alone does not satisfy. The `security-scan` gate's known
non-reproducibility (a red run may be a newly published CVE — AUDIT.md S5) is triaged as:
reproduce locally, read the finding, and either patch the toolchain version or record the
verdict; it is never waved through as flake.
