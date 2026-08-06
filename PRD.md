# PRD — sudoku-flowN

## Origin note
This is the second-generation PRD for sudoku-flow, distilled from the shipped v1 build. Its
purpose is to be the **sole input** to a fresh NerdFlow run (arch → impl → build, with the
Golanger agent writing all Go), so successive NerdFlow iterations can be compared against the
original finished product on identical requirements. Because outputs must be comparable
byte-for-byte at the API boundary, decisions that v1's architecture phase settled — the exact
technique ladder, the four-status outcome model, the frozen metric quartet, and the `/v1` JSON
shapes — are promoted here to requirements. Everything else (package layout, algorithm
internals, generator strategy, UI implementation) remains the builder's to design, and code
size/quality on those freedoms is one of the comparison axes.

## Goal
Let the NerdFlow operator POST any valid Sudoku and get back a provably logic-only solution
with a full, replayable solve-event log — plus generation, batch validation, and a playable
embedded UI — fast enough and contract-stable enough to benchmark NerdFlow iterations against
each other.

## User
**The NerdFlow operator (you).** You run the nerdflow slash commands in Claude Code CLI,
driving arch → impl → build on a fresh repo, with the Golanger agent employed as a subagent to
write all Go code (the agent is itself under evaluation). Your tools are the Claude Code CLI,
the nerdflow commands, a terminal, and GitHub. You direct the work and step away while the AI
builds continuously until done.

You cannot tolerate the solver being wrong or cheating: a returned solution that violates
Sudoku rules, or any solution reached by backtracking / trial-and-error, is a failure
regardless of speed. You judge iterations on solve-timing consistency for identical puzzles
and on code size/quality — but only when correctness and the logic-only constraint are fully
preserved.

## What the user does today
The v1 build of this product exists and is the baseline. This PRD exists so a fresh repo can
be built from scratch against the same requirements and compared to that baseline on
correctness, contract conformance, solve timing, and code leanness. The fresh build must NOT
reference the baseline repo or any other Sudoku app — it stands on its own so the comparison
stays clean.

## Intervention moment
You POST an unsolved puzzle (81-character string, digits `1-9` for givens and `0` or `.` for
blanks) to the running Go API — on localhost or a deployment — and read back JSON: the solved
grid, a status, the metric quartet, a difficulty grade, and a step-by-step solve-event log.
You also open `/` in a browser and use the embedded UI to load a seed puzzle, solve it, and
step through the solution. A separate React comparison dashboard pulling from multiple
deployments is a future, separate repo — out of scope here.

## Use cases
- **UC-1.** POST an 81-char unsolved grid → JSON: solved grid + status + metric quartet + grade.
- **UC-2.** Same POST → an ordered solve-event log; each deduction names the technique used,
  the witness cells, its effect (one placement, or one pattern's eliminations), and the full
  grid after the step — sufficient to mechanically replay and verify the entire solve.
- **UC-3.** POST a difficulty (`easy | medium | hard | expert`) → a generated puzzle with a
  guaranteed-unique solution, solvable by the shipped ladder, returned with its grade.
- **UC-4.** POST a list of puzzle strings → per-puzzle results (solved, time, iterations,
  hardest technique) in input order, plus solved/total counts.
- **UC-5.** Parallelism: batch solving fans out one goroutine per puzzle (real, linear-scaling
  parallelism). Additionally, an intra-puzzle scan-parallel solver variant is built behind a
  flag and benchmarked honestly — published as a **measured negative result** (a 9×9 solve is
  sub-millisecond and sequential; goroutine overhead loses), not a speed claim.
- **UC-6.** GET the seed-puzzle catalog (sectioned by difficulty tier) so the UI can offer a
  puzzle dropdown without filesystem access.

## The logic-only rule (the product's #1 constraint)
- The benchmarked solver uses **constructive, positive deductions only**. Absolutely no
  backtracking, no guess-then-revert, no trial-and-error on the solve path.
- **Banned entirely from the solver:** forcing chains, Nishio, AIC, and all
  contradiction-based reasoning ("assume X, propagate, revert") — logically sound but
  operationally guess-shaped. Also banned: uniqueness-assuming techniques (Unique Rectangles,
  BUG), because arbitrary POSTed grids may be non-unique.
- "All known logic algorithms" is deliberately reframed as a **defined, ordered ladder**
  sufficient for the target puzzle class (see Domain context). Puzzles above the ladder
  honestly return `stalled` — never a guessed or fabricated solution.
- The **generator is exempt**: it may use a standard backtracking uniqueness-counter
  internally (it is an unbenchmarked utility), but that counter must never leak into the
  solve path or the API surface.

## API contract (frozen; identical across iterations by design)
All routes are prefixed `/v1`; every response carries `apiVersion: "1"`. Breaking changes
mint `/v2`. Errors from any endpoint use one envelope: `{error: <human message>, code:
<stable machine code>}` (e.g. `invalid_input`, `unsupported_media_type`, `internal_error`).
POST endpoints require `Content-Type: application/json` (else 415) and are body-capped
(reject oversized bodies with 413 before reading).

- **`GET /v1/health`** → `{status, goVersion, apiVersion}` — a deployment self-identifies so
  a dashboard can segment by host/version.
- **`POST /v1/solve`** `{puzzle}` → `{apiVersion, input, status, solved, solution,
  iterations, eventCount, candidateChecks, solveTimeMs, grade, events[]}`.
  - `status ∈ {solved, invalid_input, unsolvable, stalled}` with honest semantics:
    `invalid_input` = malformed or duplicate-given grid (decided at parse, HTTP 400);
    `unsolvable` = a cell was driven to zero candidates by constructive deduction (an
    in-tier contradiction — NOT a completeness claim); `stalled` = valid grid, no technique
    fires, grid incomplete (deliberately conflates above-tier, unprovably-unsolvable, and
    non-unique — the solver must not run the solution-counting needed to separate them).
  - The **metric quartet** (frozen definitions): `solveTimeMs` = wall clock of the solve
    only, measured in the handler; `eventCount` = productive deductions in the log;
    `iterations` = main-loop scan passes (one cheapest-first sweep per pass);
    `candidateChecks` = total candidate-cell inspections. Primary comparison axes are
    `solveTimeMs` + `eventCount`; the other two are diagnostic.
  - `grade` = difficulty band of a solved puzzle ("" otherwise; key always present).
  - Each event: `{seq, technique, witnessCells[], placement?, eliminations[]?, gridAfter}`
    with 0-based `{row, col}` cells.
- **`POST /v1/generate`** `{difficulty}` → `{puzzle, difficulty, grade}`; unknown difficulty
  is 400/`invalid_input`, never default-and-proceed. The internal uniqueness counter is
  never surfaced.
- **`POST /v1/validate-batch`** `{puzzles: [...]}` → `{apiVersion, results[], solvedCount,
  total}`; each result `{puzzle, solved, solveTimeMs, iterations, hardestTechnique}` in
  input order. List capped at 256 puzzles and body at 1 MiB (over-cap → 413 before any
  solving; the batch cap IS the goroutine-count bound). A single malformed line is a
  per-item not-solved, never a whole-batch failure. Lines are whitespace/CRLF-tolerant.
- **`GET /v1/puzzles`** → `{sections: [{name, puzzles: [...]}]}` — the seed corpus grouped
  by tier (`Original / Medium / Hard / Very Hard`), served from a copy embedded in the
  binary (serverless runtimes have no working directory), with a test guarding against
  drift from the repo-root file.
- **`GET /`** → the embedded single-page UI and its assets.

## In scope
- A Go backend HTTP API implementing the contract above, stateless across requests.
- The full constructive technique ladder (Domain context) with deterministic ordering:
  cheapest technique first; within a technique, row-major scanning; zero randomness on the
  solve path — identical input yields a **byte-identical** event log and metric quartet.
- Difficulty grading: a puzzle's grade is the band of the hardest technique the solver was
  *forced* to use (no cheaper move available at that moment).
- Difficulty-graded generation with a guaranteed-unique solution, graded by the shipped
  solver itself.
- Batch validation with goroutine-per-puzzle parallelism; per-goroutine grid copies, zero
  shared mutable state; `go test -race` mandatory in CI.
- The flagged intra-puzzle parallel experiment with a committed benchmark (UC-5).
- An embedded, dependency-free, "McKinsey-clean" web UI (single binary, `embed.FS`):
  system-ui font, near-monochrome palette with one blue accent, the grid as the hero,
  symmetric grid borders. Features: manual entry + paste of a full 81-char string, a Clear
  button, a seed-puzzle dropdown fed by `GET /v1/puzzles` and grouped by tier section, solve
  with status/grade/metrics display, a statistics window (ladder-ordered technique
  histogram, counts, difficulty), a step-through viewer of the event log with cell
  highlighting (placement / witness / elimination) and full transport controls
  (first / previous / Play auto-advance / next / last, plus a step-position indicator and
  per-step description), a clickable event-log list where selecting any row jumps the grid
  to that step, and a plain-English technique-explanation panel with a difficulty-band chip
  that updates per step, covers every ladder technique, and shows a hint before solving.
  All DOM writes via `textContent`/`createElement` — no `innerHTML`.
- Production-boundary hardening at the HTTP edge: security headers on every response
  (a `'self'`-only CSP with no `unsafe-inline`/`unsafe-eval` — so the SPA's JS and CSS must
  be external files, never inline — plus HSTS, frame-denial, and nosniff); an explicit,
  non-reflecting CORS allowlist that **ships empty** (v1 is same-origin-only; a future
  dashboard origin is added as an enumerated entry — the request Origin is never echoed and
  a wildcard is never emitted); panic recovery to a 500 error envelope; request-body caps;
  content-type validation; and one structured access-log line per request (method, path,
  status, duration).
- CI/CD on GitHub Actions: `go test -race`, `go vet`, `go build`, coverage floor **80%**,
  and `govulncheck`, blocking merge on failure; deploy to Vercel gated behind a manual
  `workflow_dispatch` with a production-environment approval.

## Non-goals (explicit refusals)
- Will NOT use backtracking or any trial-and-error on the solve path (see logic-only rule).
- Will NOT return a guessed or fabricated solution; pure-logic stalls report `stalled`.
- Will NOT persist puzzles or state across requests (stateless API).
- Will NOT implement auth or multi-tenancy.
- Will NOT ship the React comparison dashboard in this repo (future, separate repo).
- Will NOT reference the v1 baseline repo or any other Sudoku app during the build — this
  iteration must stand on its own so cross-iteration comparison stays clean.
- Will NOT implement chain/ALS techniques beyond the defined ladder, nor UR/BUG.

## Success criteria
- Solves every puzzle in the committed `puzzles.txt` (all 55: the 25 singles-tier originals
  and the 30 graded advanced seeds) with zero backtracking.
- Every solution is **mechanically replayable** from its event log: an automated test
  replays each event from the input grid, verifies every placement was forced and every
  elimination sound (validated against an independent brute-force oracle used only in
  tests), and confirms the final grid equals the oracle solution. This replay proof is the
  no-backtracking guarantee — it must hold for every technique that fires.
- Per-technique coverage: every ladder technique ships with fixtures proving it fires and
  is sound; where a technique can be isolated as a puzzle's exact hardest step, fixtures
  additionally prove necessity (stalls with the ladder capped below it) and sufficiency
  (solves capped at it). Some upper-ladder techniques (jellyfish above all) are provably
  near-redundant and cannot be isolated — fires-and-sound is the accepted gate there,
  with the evidence recorded.
- Identical input produces identical output run-to-run (determinism), and the metric
  quartet + timing are returned on every solve.
- Generated puzzles are unique-solution, ladder-solvable, and correctly graded.
- The `/v1` contract matches this PRD exactly, so an external client can call multiple
  deployments uniformly and compare correctness and speed.
- CI blocks merge on any gate failure; a deploy reaches Vercel only through the manual gate;
  the deployed instance answers `/v1/health` with 200 and serves the UI.
- Comparison axes across NerdFlow iterations: consistent solve timing on identical puzzles,
  and reduced code size / improved quality with correctness and logic-only fully preserved.

## Constraints
- **Regulatory:** None.
- **Integration target:** Standalone API; any client meeting the contract may call it.
  Embedded UI included. Future external React dashboard consumes multiple deployments.
- **Tech stack:** Latest Go, stdlib only — `net/http` with Go 1.22+ method+path `ServeMux`
  routing; zero third-party runtime dependencies. The Golanger agent writes all Go —
  source and tests, including the test-driven development work — employed as a subagent;
  the orchestrator directs, runs gates, and writes no Go.
- **Deployment (platform facts, learned the hard way):** Vercel free tier is the zero-cost
  demo target — ephemeral, minimal setup and teardown, since it won't stay up long; all
  serious benchmarking runs on localhost. `@vercel/go` builds **serverless
  functions only** — it requires a file exporting `func Handler(w http.ResponseWriter,
  r *http.Request)` and cannot run a `package main` server; it compiles that entrypoint as
  a module-less build which **cannot import `internal/` packages** (and a local `go build`
  does NOT reproduce this — only a real deploy exercises it). Vercel Hobby also has a 10s
  request cap and variable CPU, so deployed timing is noisy and batch must stay bounded.
  Requirement: one shared handler serves both a local `cmd/server` binary (listening on
  `$PORT`) and the Vercel function with byte-identical behavior — same middleware, routes,
  and logs. All runtime assets (UI, seed catalog) must be embedded in the binary, never
  read from the filesystem.
- **Budget envelope:** Near-zero. Free tiers only; no custom domain.
- **Team:** Solo — you directing, NerdFlow + Golanger building.

## Deadline
No fixed calendar deadline. AI-driven continuous build until done; the operator directs
pacing (including AFK/overnight runs).

## Domain context
The solver implements this exact constructive ladder, cheapest-first (event `technique`
strings in parentheses), with each technique mapped to its difficulty band:

| # | Technique (event string) | Band |
|---|---|---|
| 1 | Naked single (`naked_single`) | Easy |
| 2 | Hidden single (`hidden_single`) | Easy |
| 3 | Locked candidates, pointing (`locked_candidates_pointing`) | Medium |
| 4 | Locked candidates, claiming (`locked_candidates_claiming`) | Medium |
| 5 | Naked pairs/triples/quads (`naked_subset`) | Medium |
| 6 | Hidden pairs/triples/quads (`hidden_subset`) | Medium |
| 7 | X-wing (`x_wing`) | Hard |
| 8 | Swordfish (`swordfish`) | Hard |
| 9 | Jellyfish (`jellyfish`) | Hard |
| 10 | XY-wing (`xy_wing`) | Hard |
| 11 | XYZ-wing (`xyz_wing`) | Expert |
| 12 | W-wing (`w_wing`) | Expert |
| 13 | Simple colouring (`simple_colouring`) | Expert |

Singles place digits; every other technique only eliminates candidates. A technique fires
only when nothing cheaper can act — that discipline is what makes "hardest technique used"
a well-defined grade. Bands follow Sudoku Explainer's tier ordering. Generation difficulty
inputs map easy→Easy, medium→Medium, hard→Hard, expert→Expert grades.

## Data / fixtures available today
- `puzzles.txt`: 55 puzzles, one 81-char grid per line, sectioned by `# === NAME ===`
  headers (loaders must skip `#`/blank lines):
  - **ORIGINAL** (unlabeled first section) — 25 seeds, every one solvable by naked/hidden
    singles alone. This singles-only property is a frozen invariant: it underpins a clean
    singles-tier replay proof and must be preserved as-is.
  - **MEDIUM / HARD / VERY-HARD** — 30 graded advanced puzzles, each unique-solution and
    ladder-solvable, exercising the upper ladder.
  The file is committed as the seed; the finished app must solve all 55 logic-only, and the
  catalog endpoint serves it in these sections.
- Additional puzzle lists arrive only AFTER implementation is complete, as independent
  testing — not during the build.
- Correctness is validated by rule conformance reached via logged, replayable logic steps;
  the test-only brute-force oracle provides ground-truth solutions for soundness checks.

## Reference materials
- None beyond this PRD and the committed `puzzles.txt` — deliberately. No other apps or
  prior solutions may inform the build, to keep cross-iteration comparison clean.
