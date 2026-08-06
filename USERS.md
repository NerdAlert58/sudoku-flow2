# USERS — sudoku-flowN

> Source of truth for who uses this system and how. Every capability in ARCHITECTURE.md
> must trace back to a use case here; a capability that cannot cite a UC does not exist.

## The User

**The NerdFlow operator** — the single human who runs the nerdflow slash commands in
Claude Code CLI, directing arch → impl → build on a fresh repo with the Golanger agent
writing all Go, then benchmarks the result against the v1 baseline build.

- **Setting:** a terminal + Claude Code CLI + GitHub + a browser, on macOS; deployments on
  Vercel free tier for demos, localhost for all serious timing runs.
- **Volume:** one operator; bursts of activity at build boundaries — a full-corpus batch
  validation (55 puzzles), repeated identical solves for timing-consistency measurement,
  occasional generation calls, UI walk-throughs of individual solves.
- **Tools today:** the v1 baseline build (the product this iteration is compared against),
  curl/jq, the nerdflow commands.
- **Constraints:** cannot tolerate a wrong or cheating solver — a rule-violating solution
  or anything reached by trial-and-error is a total failure regardless of speed; will
  abandon an iteration whose `/v1` contract drifts even slightly, because byte-comparable
  contracts across iterations are the entire point.

**Why this user, narrowly defined:** picking the benchmark operator (not "developers" or
"Sudoku players") forecloses auth, multi-tenancy, persistence, and player-pleasing features
(hints, difficulty balancing for fun), and forces the properties a benchmark needs:
determinism, replayable proof, frozen contracts, honest failure statuses. The embedded UI
exists for the operator's inspection workflow, not for players.

## The Workflow

### The 90 seconds before the system enters the day

The operator has just finished (or resumed) a NerdFlow iteration and wants to know: is this
build correct, honest, and how does it compare? Without this system they have three
options: read the diff (no ground truth about runtime behavior), trust CI logs alone (no
feel for the event log or UI), or run the v1 baseline (which tells them nothing about the
new iteration). Each costs the thing they care about — verified, comparable evidence.

### The system enters the day

The operator opens a terminal: `curl -s -X POST <host>/v1/solve -H 'content-type:
application/json' -d '{"puzzle":"7006050000000000305093000240020000004019070520005010000040
50000310492000007003000"}' | jq` — and reads back status, the solved grid, the metric
quartet, grade, and the full event log, in well under a second on localhost. They repeat
the identical call and expect byte-identical events and counters. They open `<host>/` in a
browser, pick a seed from the dropdown, hit Solve, and step through every deduction with
the technique explanation panel open. Session boundary: closing the tab or ending the curl
session ends everything — the API is stateless, no puzzle, result, or preference persists
server-side across requests, ever.

## Use Cases

### UC-1. Solve a posted puzzle

**Prompt shape:** `POST /v1/solve` `{puzzle: "<81 chars, 1-9 givens, 0 or . blanks>"}`.
**What the system does:** parses and validates; runs the 13-technique constructive ladder
cheapest-first; returns `{apiVersion, input, status, solved, solution, iterations,
eventCount, candidateChecks, solveTimeMs, grade, events[]}` with honest status semantics
(solved / invalid_input / unsolvable / stalled).
**Why an API call is right:** the operator scripts comparisons across deployments; a JSON
contract is the only shape a dashboard or diff tool can consume uniformly.

### UC-2. Replay and verify the solve

**Prompt shape:** the same POST — the event log is part of every solve response.
**What the system does:** emits an ordered event log where each event names the technique,
witness cells, effect (one placement or one pattern's eliminations), and the full grid
after the step — sufficient for the automated replay verifier (and the UI step-viewer) to
mechanically re-derive and check every deduction.
**Why an event log (not a summary) is right:** the no-backtracking guarantee is only
provable from a complete, replayable deduction trace; a summary would require trusting the
solver about the very property under test.

### UC-3. Generate a graded puzzle

**Prompt shape:** `POST /v1/generate` `{difficulty: easy|medium|hard|expert}`.
**What the system does:** generates a unique-solution, ladder-solvable puzzle whose
solver-computed grade equals the requested band (by construction), returns `{puzzle,
difficulty, grade}`; unknown difficulty → 400 envelope; budget exhaustion → honest 500
`generation_failed`, never a mislabeled puzzle.
**Why an API call is right:** generated fixtures feed scripted testing across iterations;
the internal backtracking uniqueness counter stays sealed inside the generator.

### UC-4. Batch-validate a puzzle list

**Prompt shape:** `POST /v1/validate-batch` `{puzzles: [...]}` (≤256 items, ≤1 MiB,
whitespace/CRLF-tolerant lines).
**What the system does:** solves each puzzle in its own goroutine (grid copies, zero
shared mutable state), returns per-puzzle `{puzzle, solved, solveTimeMs, iterations,
hardestTechnique}` in input order plus solvedCount/total; a malformed line is a per-item
failure, never a batch failure.
**Why a batch endpoint is right:** the operator validates whole corpora (the 55-seed file,
future independent test lists) in one call; input-order results make diffing trivial.

### UC-5. Honest parallelism evidence

**Prompt shape:** not an endpoint — `go test -bench` on the committed benchmark, plus the
batch endpoint's goroutine-per-puzzle design.
**What the system does:** batch solving fans out one goroutine per puzzle (real, linearly
scaling parallelism, race-clean under `-race`); separately, an intra-puzzle scan-parallel
solver variant exists behind a flag with a committed benchmark showing it loses to the
sequential solver (a measured negative result — 9×9 solves are sub-millisecond and
goroutine overhead dominates).
**Why a committed benchmark is right:** the claim is about measured reality, not a feature;
committing the numbers makes the negative result reviewable across iterations.

### UC-6. Seed-puzzle catalog for the UI

**Prompt shape:** `GET /v1/puzzles`.
**What the system does:** returns `{sections: [{name, puzzles: [...]}]}` — the 55-puzzle
corpus grouped as Original / Medium / Hard / Very Hard, served from a copy embedded in the
binary with a drift-guard test against the repo-root file.
**Why an endpoint (not a bundled JS asset) is right:** the catalog is API surface — the
future dashboard can enumerate seeds the same way the UI does, and the embedded copy keeps
serverless runtimes filesystem-free.

## What the System Will NOT Do

- **Will NOT backtrack or guess on the solve path** — no forcing chains, Nishio, AIC,
  contradiction-based reasoning, or uniqueness-assuming techniques (UR/BUG); puzzles above
  the ladder return `stalled`, never a fabricated solution. This is the product's #1
  constraint.
- **Will NOT persist anything across requests** — stateless API; no accounts, sessions,
  saved puzzles, or server-side history.
- **Will NOT implement auth or multi-tenancy** — single-operator benchmark tool; adding
  auth would change the timing surface being compared.
- **Will NOT serve cross-origin callers in v1** — the CORS allowlist ships empty; no
  Origin echo, no wildcard, ever. A future dashboard origin is added as an enumerated
  entry.
- **Will NOT ship the React comparison dashboard** — future, separate repo.
- **Will NOT expose the generator's uniqueness counter** — not in the solve path, not in
  any API field.
- **Will NOT reference the v1 baseline repo or any other Sudoku app** — the build stands
  alone so cross-iteration comparison stays clean.
- **Will NOT extend the ladder** beyond the 13 defined techniques (no chains, ALS,
  finned/sashimi fish).

## Capability → Use Case Traceability

| Capability | Justified by | Required dependencies |
|---|---|---|
| 13-technique constructive ladder, deterministic ordering | UC-1, UC-2, UC-3 (grading), UC-4 | Grid/candidate model; frozen scan-order conventions |
| Event log with per-step gridAfter | UC-2, UI step-viewer | Ladder emits one event per productive deduction |
| Metric quartet + grade | UC-1, UC-4, cross-iteration comparison | Counted candidate accessor; pass counter; handler wall clock |
| Difficulty-graded unique generation | UC-3 | Backtracking uniqueness counter (sealed in generator); ladder solver for grading |
| Batch solving, goroutine-per-puzzle | UC-4, UC-5 | Per-goroutine grid copies; 256-item cap as concurrency bound |
| Flagged scan-parallel variant + benchmark | UC-5 | Benchmark harness; committed results file |
| Embedded seed catalog endpoint | UC-6, UI dropdown | Embedded puzzles.txt copy + drift test |
| Embedded SPA (grid, dropdown, solve, stats, step-viewer, explanations) | Intervention moment, UC-2, UC-6 | GET /v1/puzzles, POST /v1/solve; embed.FS; strict-CSP-compatible assets |
| /v1/health self-identification | Deployment smoke + future dashboard segmentation | runtime.Version() |
| HTTP hardening (CSP, HSTS, frame-denial, nosniff, empty CORS, caps, 415, panic recovery, access log) | Production boundary (PRD §In scope); every UC crosses this edge | Shared middleware chain in both entrypoints |
| CI gates + gated Vercel deploy | PRD success criteria (block merge; manual-gate deploy) | GitHub Actions, public repo, Vercel CLI token flow |

**Capabilities the architecture should NOT include** (no UC justifies them): auth/user
management; persistence of any kind; rate limiting beyond body/batch caps (single-operator
tool, no public write surface); WebSockets/streaming; i18n; a puzzle-difficulty tuning UI;
OPTIONS/preflight handling while the CORS allowlist is empty; finned fish or any ladder
extension; client-side solving.

## Demo-Data Reality Check

`puzzles.txt` (55 puzzles: 25 singles-only ORIGINAL + 10/10/10 MEDIUM/HARD/VERY-HARD) is
committed and verified well-formed; UC-1/2/4/6 are fully demo-able against it today, and
the UI dropdown serves it. UC-3 needs no fixtures (it generates). Gaps to curate during
build: per-technique fixture states for all 13 techniques (fires-and-sound), plus
necessity/sufficiency puzzles where isolable (D4 in AUDIT.md); synthetic invalid inputs
(bad length, bad chars, duplicate givens, `.`-blank variants — the corpus never uses `.`);
synthetic batch inputs with CRLF/whitespace and malformed lines. Additional independent
puzzle lists arrive only after implementation, as external testing.
