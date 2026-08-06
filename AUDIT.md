# AUDIT — sudoku-flowN

The captured state of the world this project must live with. ARCHITECTURE.md treats every
finding here as a hard input. Sources: six parallel research passes (2026-08-06) over the
PRD, current Vercel/GitHub/Go documentation, the Go stdlib source, and direct inspection of
`puzzles.txt`. Per the PRD, no existing Sudoku solver codebase was consulted.

## Summary

**The PRD's two hardest-won platform facts are stale, and the safe architecture is the one
that doesn't need to know which version is true.** Current Vercel docs (2026-07) describe a
Go Framework Preset that runs a real `package main` server, and a 300s Hobby duration cap
under Fluid Compute — contradicting the PRD's "serverless functions only" and "10s cap."
Vercel's own docs are internally inconsistent about this, and the PRD correctly warns that
only a real deploy exercises the classic model's `internal/`-import restriction. The
architecture therefore avoids `internal/` packages entirely, keeps one shared
`http.Handler` constructor consumed by both entrypoints, designs time budgets to the
conservative 10s figure, and pins actual behavior with a walking-skeleton deploy spike
before the full build (A1, A2).

**Two PRD success criteria are impossible on a private GitHub Free repo.** Required status
checks (and rulesets) need a public repo on the Free plan, and environment
required-reviewers on private repos need Enterprise. "CI blocks merge" + "production
approval gate" + "free tiers only" jointly force a public repository (C1; DECISIONS.md
D-007).

**The frozen contract has four ambiguities that must be resolved before any code exists,
because byte-identical cross-iteration comparison is the product's purpose.** The
`invalid_input` dual shape, the stdlib ServeMux plain-text 405, the `solveTimeMs`
wall-clock-vs-byte-identical tension, and the `unsolvable` naked-vs-hidden contradiction
scope each admit two readings; all four are frozen by ADR (S/A findings below;
DESIGN_DECISIONS.md ADR-0004…ADR-0008).

**Every upper-ladder technique is legal under the logic-only ban only in a specific
positive formulation.** The textbook narration of simple colouring's "color wrap" is
structurally Nishio (assume-propagate-revert); the legal form combines already-certain
conjugate-pair biconditionals with the one-digit-per-unit rule directly. Wings are
constructive dilemmas over already-certain disjunctions; fish are pure counting arguments.
The event log must justify deductions positively, and finned/sashimi fish must not creep in
(L1).

**The metric quartet and event log are only comparable if scan orders, tie-breaks, and
counting conventions are frozen now.** Row-major cells, ascending digits, lexicographic
combinations, canonical serialization order, one productive event per pass, and one counted
candidate-membership accessor are the frozen conventions (L2, L3; ADR-0007).

**Coverage arithmetic is wrong without `-coverpkg=./...`.** Verified hands-on: cross-package
exercised code silently reports 0% under plain `go test ./...`, corrupting the 80% floor in
either direction. The CI gate must use `go test -race -coverprofile=coverage.out
-coverpkg=./... ./...` and float-safe threshold comparison (C2).

**The seed corpus is clean but narrower than the contract it feeds.** All 55 lines verified
well-formed (LF-only, 25/10/10/10 sections); every blank is `0` — the `.` input path is
never exercised by fixtures and needs synthetic tests; section header text does not match
the catalog's required display names and needs an ordinal mapping; file section labels are
NOT solver-grade assertions (D1–D3).

## Architecture

### A1. Vercel platform facts have moved since the PRD was written

- **Where:** vercel.com/docs/functions/runtimes/go (dated 2026-07-01);
  vercel.com/docs/functions/limitations (2026-07-01); vercel.com/docs/functions/runtimes
  (2026-07-29); PRD.md §Constraints "Deployment".
- **What:** Current docs describe (1) a Go Framework Preset: auto-detects root `go.mod` +
  `cmd/server/main.go`, runs a real `net/http` server listening on `$PORT`, built by a
  normal whole-module `go build`; (2) Hobby max duration 300s (default and max) under Fluid
  Compute; (3) platform request/response body cap 4.5 MB; (4) bundle cap 250 MB. The PRD
  asserts "serverless functions only," "cannot run a `package main` server," and a "10s
  request cap." Vercel's own general runtimes table still describes only the old per-file
  model — the docs are internally inconsistent, so neither version can be trusted without a
  deploy.
- **So what:** The architecture must not bet on either model. Layout works under both:
  shared handler package, `cmd/server/main.go` (also the preset's detected entrypoint) and
  `api/index.go` (classic model), no `internal/` anywhere. Time budgets assume 10s. An
  early deploy spike pins reality.
- **Open questions to resolve in ARCHITECTURE.md:** Which entrypoint Vercel actually
  builds when both exist — resolved by the spike piece; vercel.json is written for the
  classic model (rewrites → /api) since that path is PRD-proven.

### A2. `internal/` import restriction — real for the classic model, unreproducible locally

- **Where:** github.com/vercel/vercel discussions/5725; jhartman.pl (2026-02); PRD.md
  §Constraints.
- **What:** The classic `/api/*.go` builder rewrites the module path for an isolated
  per-file compile, breaking Go's `internal/` visibility. Community-confirmed only; not in
  official docs. A local `go build` cannot reproduce it.
- **So what:** Zero `internal/` directories in this repo, ever. Shared code lives in plain
  packages. The constraint costs nothing (single-module project, no external consumers to
  hide API from) and removes the only local-build-invisible failure mode.
- **Open questions:** None.

### A3. `go:embed` cannot reach parent directories

- **Where:** pkg.go.dev/embed ("Patterns may not contain '.' or '..' … interpreted
  relative to the package directory").
- **What:** The repo-root `puzzles.txt` cannot be embedded directly from another package's
  directory. Same for UI assets.
- **So what:** The catalog package carries its own copy of `puzzles.txt` next to the
  embedding source file, plus a test byte-comparing the embedded copy against the repo
  root file (the PRD's drift guard). UI assets live inside the web package directory.
- **Open questions:** None.

### A4. Go 1.22+ ServeMux: automatic 405 is plain-text and non-overridable

- **Where:** Go stdlib `net/http/server.go` (ServeMux doc + `findHandler`); go.dev/blog/routing-enhancements.
- **What:** Method-specific patterns produce an automatic 405 via `http.Error` (plain
  text, `Allow` header) with no public override hook. `GET` patterns also match `HEAD`;
  nothing else cross-matches. Overlapping-pattern registration panics at startup.
- **So what:** To keep the PRD's single JSON error envelope on every /v1 response, routes
  register path-only patterns and dispatch on method inside the handler, emitting the
  envelope for 405 with `Allow` set (ADR-0005).
- **Open questions:** None.

### A5. `encoding/json` behaviors that touch the frozen contract

- **Where:** pkg.go.dev/encoding/json.
- **What:** Output field order = struct definition order (stable across builds if structs
  don't change). `omitempty` drops `""` — but the contract requires `grade` always
  present. `Marshal` always HTML-escapes (`<`, `>`, `&`); payloads here are digits and
  fixed technique strings, so no effect in practice. Float encoding is
  shortest-round-trip, variable digit count by value.
- **So what:** Response structs are the contract's source of truth; `grade` carries no
  omitempty; `solveTimeMs` is float64 ms and excluded from byte-identity claims
  (ADR-0006).
- **Open questions:** None.

### A6. Middleware ordering for correct logs under panic

- **Where:** pkg.go.dev/log/slog; blog.questionable.services (logging middleware pattern).
- **What:** A status-recording ResponseWriter wrapper (default 200) plus ordering:
  access-log outermost (starts timer, wraps writer), panic-recovery inside it, routes
  innermost — so a panicking handler still yields one access-log line with status 500.
- **So what:** Fixes the middleware chain shape in ARCHITECTURE.md §Components; identical
  chain in both entrypoints.
- **Open questions:** None.

### A7. Body-cap semantics: `http.MaxBytesReader` is lazy; item-count cap needs the parse

- **Where:** pkg.go.dev/net/http#MaxBytesReader; PRD.md §API contract.
- **What:** MaxBytesReader enforces the byte cap during reads (detect `*MaxBytesError`
  after decode fails); "reject before reading" is satisfied because the cap is installed
  before the first read. A Content-Length fast-path can 413 declared-oversized bodies
  before opening the reader. The 256-item batch cap is checkable only after JSON parse —
  the PRD's own wording there is "before any solving," a different and satisfiable
  guarantee.
- **So what:** Two textually distinct checks in the handler spec: byte cap (fast-path +
  lazy enforcement → 413) and item cap (post-parse, pre-solve → 413).
- **Open questions:** None.

## Security

### S1. `'self'`-only CSP shapes the SPA's construction

- **Where:** MDN CSP (`style-src`, `script-src`, `default-src`); PRD.md §In scope
  hardening bullet.
- **What:** With no `unsafe-inline`: inline `<script>`, inline handlers (`onclick=`),
  `<style>` blocks, and `style=""` attributes are all blocked. Direct CSSOM manipulation
  from JS (`el.style.x = y`) is NOT blocked, but class-toggling against external CSS is
  the clean pattern. External `.js`/`.css`, `addEventListener`, `classList`,
  `textContent`, `createElement`, same-origin `fetch` are unaffected.
- **So what:** UI ships as `index.html` + external `app.js` + `app.css` from `embed.FS`;
  all visual states are CSS classes; no inline anything. CSP:
  `default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self';
  img-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'`.
- **Open questions:** None.

### S2. Empty CORS allowlist = no CORS headers at all

- **Where:** OWASP ASVS CORS guidance; CWE-942; PRD.md §In scope.
- **What:** Correct empty-allowlist behavior: never emit `Access-Control-Allow-Origin`
  (no echo, no `*`); browsers then block cross-origin readers while same-origin is
  unaffected. When an origin is enumerated later: exact-match echo + `Vary: Origin`.
  stdlib ServeMux gives OPTIONS no special handling, and no preflight can ever succeed
  with an empty allowlist, so no OPTIONS route ships in v1.
- **So what:** CORS is a middleware whose allowlist constant ships empty; adding an origin
  is a one-line change plus tests. 405-with-envelope covers stray OPTIONS.
- **Open questions:** None.

### S3. HSTS is a no-op over plain HTTP by design

- **Where:** MDN Strict-Transport-Security.
- **What:** Browsers ignore the header over HTTP; Vercel terminates TLS so production
  responses are HTTPS. Sending unconditionally is harmless locally and correct deployed.
- **So what:** One unconditional header set in the shared middleware; no environment
  branching.
- **Open questions:** None.

### S4. Frame denial and sniffing defense

- **Where:** MDN frame-ancestors; content-security-policy.com.
- **What:** `frame-ancestors 'none'` (modern) + `X-Frame-Options: DENY` (legacy
  defense-in-depth) + `X-Content-Type-Options: nosniff`. stdlib's automatic error paths
  already emit nosniff; duplicate Set is idempotent.
- **So what:** Fixed header set in shared middleware, asserted by handler tests on every
  route.
- **Open questions:** None.

### S5. govulncheck is legitimately non-reproducible

- **Where:** go.dev/doc/tutorial/govulncheck; pkg.go.dev/golang.org/x/vuln/cmd/govulncheck.
- **What:** It fetches vuln.go.dev live at run time — the same commit can flip green→red
  overnight as CVEs publish. It checks the stdlib itself (the only dependency surface
  here). Exit code reflects findings ONLY in plain-text mode (`-json`/`-format` always
  exit 0). Installing via `go install …@version` keeps the project's go.mod free of
  third-party entries; Go 1.24's `tool` directive would not.
- **So what:** CI runs plain-text govulncheck via `go install`; a red run must be triaged
  as possible-new-CVE, not assumed flaky (recorded in EVAL.md regression triggers).
- **Open questions:** None.

## Data Quality

### D1. `puzzles.txt` verified clean — and narrower than the input grammar

- **Where:** /Users/nerd/Git/sudoku-flow2/puzzles.txt (direct inspection).
- **What:** ASCII, LF-only, no trailing whitespace, final newline. Headers at lines
  1/28/40/52: `# === ORIGINAL (unlabeled) ===`, `# === MEDIUM ===`, `# === HARD ===`,
  `# === VERY HARD ===`. Exactly 55 lines matching `^[0-9]{81}$`: 25/10/10/10. Zero `.`
  characters — every blank is `0`.
- **So what:** The contract's `.`-blank input path gets synthetic test fixtures (it will
  never be exercised by the corpus). CRLF/whitespace tolerance is a batch-endpoint
  requirement tested with synthetic inputs, not a property the seed file needs.
- **Open questions:** None.

### D2. Catalog display names require an ordinal mapping

- **Where:** puzzles.txt headers vs PRD.md §API contract (`Original / Medium / Hard /
  Very Hard`).
- **What:** Literal header text (`ORIGINAL (unlabeled)`, `MEDIUM`, …) matches none of the
  four required display strings.
- **So what:** The loader treats headers as section boundaries only; ordinal position maps
  to the four canonical names. Exactly four sections are asserted by test.
- **Open questions:** None.

### D3. File section labels are not grade assertions

- **Where:** PRD.md §Data/fixtures ("unique-solution and ladder-solvable, exercising the
  upper ladder") vs §In scope grading definition.
- **What:** The PRD never claims a `HARD`-section puzzle grades `Hard`. Grade is defined
  purely by the shipped solver's forced-hardest-technique.
- **So what:** CI asserts, for all 55: `status == solved`, oracle-unique, final grid ==
  oracle solution, replay proof passes. CI must NOT assert `grade == section name`.
- **Open questions:** None.

### D4. Per-technique fixtures need curation beyond the corpus

- **Where:** PRD.md §Success criteria (per-technique coverage).
- **What:** Nothing guarantees 30 natural seeds isolate all 13 techniques. Fires-and-sound
  needs a curated grid state per technique; necessity/sufficiency need puzzles whose exact
  hardest step is the target technique. The PRD pre-accepts that some upper-ladder
  techniques (jellyfish named) cannot be isolated — fires-and-sound plus recorded evidence
  is the gate there.
- **So what:** Fixture curation is an explicit build piece with its own acceptance
  criteria; EVAL.md rows carry the per-technique bar.
- **Open questions to resolve in ARCHITECTURE.md:** "Jellyfish above all" scope — resolved
  as: full necessity/sufficiency attempted for techniques 3–10 and 11–13 where curation
  succeeds; any technique where curation demonstrably fails records the attempt and falls
  back to fires-and-sound (the PRD's own escape hatch, evidence required).

## Performance

### P1. A 9×9 solve is sub-millisecond; the timing metric must not truncate to zero

- **Where:** PRD.md UC-5 and §API contract metric quartet.
- **What:** Integer milliseconds would report 0 for nearly every solve, destroying a
  primary comparison axis. Wall clock can never be byte-identical run-to-run.
- **So what:** `solveTimeMs` is float64 milliseconds; determinism guarantees are scoped to
  the event log + three counters (ADR-0006).
- **Open questions:** None.

### P2. Generation is the one open-ended endpoint under a serverless clock

- **Where:** PRD.md UC-3, §Constraints (10s figure; 300s per current docs — unproven).
- **What:** Dig-and-grade retry loops for the expert band plausibly cost hundreds of
  milliseconds to seconds; the platform kill (whatever its true value) returns a bare 504
  that bypasses the app's envelope.
- **So what:** Internal 5s deadline, bounded attempts, `grade == requested difficulty` by
  construction, exhaustion → 500 `generation_failed` (ADR-0009). Serious benchmarking is
  localhost per the PRD.
- **Open questions:** None.

### P3. Caps compose safely

- **Where:** PRD.md §API contract; vercel.com/docs/functions/limitations.
- **What:** App caps (1 MiB body, 256 batch items = goroutine bound) sit far under the
  platform's 4.5 MB cap. Batch of 256 sub-millisecond solves is comfortably inside even a
  10s budget.
- **So what:** No platform-cap interaction to design around beyond what the contract
  already fixes.
- **Open questions:** None.

## CI/CD & Repository

### C1. GitHub Free + private repo cannot enforce the PRD's gates

- **Where:** github.com/pricing ("Repository rules": Free = public only);
  docs.github.com about-protected-branches; docs.github.com environments (required
  reviewers: public or Enterprise); community discussions #174400/#174419 with staff
  confirmation.
- **What:** Required status checks, rulesets, and environment required-reviewers are all
  unavailable on private Free repos. Pro/Team unlock branch protection but NOT environment
  reviewers on private.
- **So what:** The repo goes public when CI wiring lands (DECISIONS.md D-007). The
  production environment's sole required reviewer is the operator ("prevent self-review"
  stays off — a deliberate-pause gate, not a second-pair-of-eyes control, recorded as a
  Known Tradeoff).
- **Open questions:** None.

### C2. Coverage must be computed with `-coverpkg=./...` and compared float-safe

- **Where:** Verified hands-on this session with a two-package module (see research log);
  actions/setup-go docs.
- **What:** Without `-coverpkg=./...`, cross-package exercised code reports 0%.
  `-race` + `-coverprofile` auto-selects atomic mode. Total line format:
  `total: (statements) NN.N%`; extraction via `go tool cover -func` + awk; comparison must
  handle decimals.
- **So what:** The exact CI commands are frozen in ARCHITECTURE.md §CI/CD topology.
- **Open questions:** None.

### C3. Deploy pipeline facts

- **Where:** vercel.com/docs/cli/deploying-from-cli; /docs/cli/build; /docs/git/vercel-for-azure-pipelines;
  docs.github.com REST (environments, pending_deployments); GitHub changelog 2023-10-16.
- **What:** Token-based CLI deploy (`vercel pull --yes --environment=production` →
  `vercel build --prod` → `vercel deploy --prebuilt --prod`) with
  VERCEL_TOKEN/VERCEL_ORG_ID/VERCEL_PROJECT_ID works without Vercel git integration; IDs
  come from a one-time local `vercel link` (`.vercel/project.json`, not committed).
  Environment approval: `POST …/actions/runs/{run_id}/pending_deployments` needs a
  reviewer-owned token (`repo` scope) — the run's own GITHUB_TOKEN cannot self-approve.
  Self-review is allowed unless "prevent self-review" is enabled (it must stay off).
- **So what:** Deploy workflow shape is fixed; the operator's own `gh` session performs
  the approval (the manual gate). No Vercel git integration is ever connected (it would
  bypass the gate).
- **Open questions:** None.

### C4. Post-deploy smoke must tolerate cold starts

- **Where:** vercel.com/docs/functions/runtimes (archival note); community cold-start
  reports (0.8–1.5s typical; +1s+ if archived).
- **What:** First hit after deploy/idle can be slow or transiently unready.
- **So what:** Smoke check retries bounded (5 × 3s) before failing; failure fails the
  workflow and the rollback path is redeploying the prior commit through the same gate.
- **Open questions:** None.

## Solver Domain

### L1. Positive-form legality of the 13 techniques

- **Where:** PRD.md §Logic-only rule + §Domain context; HoDoKu/Sudopedia/SudokuWiki
  technique definitions (conceptual only).
- **What:** Legality test satisfied by all 13 in the right formulation: a deduction is
  legal iff it starts from an already-certain bounded disjunction (bivalue cell, conjugate
  pair, trivalue cell), each branch applies one direct rule, and all branches agree on the
  same positive conclusion (no branch discarded by contradiction). Singles and subsets and
  locked candidates and fish are direct counting/covering arguments. Wings are
  constructive dilemmas. Simple colouring: trap = both colors agree the uncolored cell
  loses the digit; wrap = two already-established facts (same-color biconditional chain +
  one-digit-per-unit) directly imply the color class is false — the event log must justify
  it that way, never as assume-propagate-revert. Finned/sashimi fish are out (different
  case-split shape, not in the ladder).
- **So what:** Per-technique implementation notes in ARCHITECTURE.md; replay verifier
  checks structural witness patterns per technique.
- **Open questions:** None.

### L2. Frozen determinism conventions

- **Where:** PRD.md §In scope (deterministic ordering).
- **What/So what:** Cells row-major (r0..8 × c0..8); digits ascending 1–9; units
  enumerated rows 0–8, columns 0–8, boxes 0–8 row-major; subset/fish combinations in
  lexicographic order; coloring components started from the lowest uncolored row-major
  cell, digits ascending; within a pass the first firing instance in canonical order wins;
  `witnessCells[]` and `eliminations[]` serialized row-major-then-digit-ascending
  regardless of discovery order. These conventions are frozen (ADR-0007).
- **Open questions:** None.

### L3. Metric conventions

- **Where:** PRD.md §API contract metric quartet.
- **What/So what:** One productive event per pass, loop restarts from naked_single after
  every event; `iterations` counts passes started, including the final non-productive pass
  for stalled/unsolvable; `candidateChecks` counts every (cell,digit) candidate-membership
  query by detection logic across all attempted techniques through one shared counted
  accessor; detection short-circuits at the first canonical firing instance (ADR-0007).
- **Open questions:** None.

### L4. Replay-proof verifier specification

- **Where:** PRD.md §Success criteria.
- **What/So what:** The verifier maintains its own shadow candidate state (never trusts
  solver internals). Per event: (1) placements — recompute from prior state that the named
  single's condition held, and the placed digit equals the test-only brute-force oracle's
  value at that cell; (2) eliminations — each eliminated candidate existed, is not the
  oracle's solution value at that cell, and the named technique's witness pattern
  structurally holds in the prior state; (3) `gridAfter` equals prior grid plus exactly
  the stated placement (elimination events leave digits unchanged); (4) after the last
  event the grid equals the oracle's unique solution; (5) a second in-process run
  byte-matches the first (events + three counters). Additionally the verifier
  independently re-implements naked/hidden single detection (~30 lines) to assert no
  single was available whenever an elimination technique fired — the cheapest-first
  scheduling check. Full independent re-implementation of all 13 techniques is rejected as
  circular-by-duplication; per-technique soundness is covered by (2)/(3) against the
  oracle plus curated fixtures.
- **Open questions:** None.

### L5. "Forced" is two properties, both tested

- **Where:** PRD.md §Success criteria vs §Domain context.
- **What/So what:** Logical necessity (replay proof, L4 items 1–2) and ladder-priority
  discipline (L4 singles-availability check + loop structure + per-technique unit tests)
  are distinct; both are in EVAL.md.
- **Open questions:** None.
