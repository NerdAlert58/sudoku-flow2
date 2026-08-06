# Design Decisions — sudoku-flowN

ADR log. Ordered by dependency; cross-linked by number. Operator-level provenance for the
autonomous run lives in DECISIONS.md (D-numbers); these ADRs are the frozen architectural
record.

## ADR-0001: Shared-core dual-entrypoint monolith; no `internal/` packages

**Status:** Accepted (2026-08-06)
**Context:** The PRD requires one shared handler serving a local `cmd/server` binary and a
Vercel function byte-identically, and asserts platform facts (serverless-only, `internal/`
import ban, 10s cap) that research shows are stale or unverifiable without a deploy
(AUDIT.md A1/A2). Vercel's docs are internally inconsistent about the new Go server
preset.
**Decision:** One handler graph built by `httpapi.New() http.Handler`, mounted by
`cmd/server/main.go` (`$PORT`) and `api/index.go` (`func Handler(w, r)`); `vercel.json`
rewrites all paths to the classic `api/` function (the PRD-proven model). Zero `internal/`
directories anywhere. Time budgets designed to the conservative 10s figure. A
walking-skeleton deploy spike is the first build piece and pins actual platform behavior.
**Alternatives considered:** Betting on the new Go Framework Preset — rejected: docs
internally inconsistent, unproven here, and the classic model is known to work. Using
`internal/` with the preset — rejected: reintroduces the one failure mode a local build
cannot reproduce, for zero benefit in a single-module repo. Two divergent handlers —
rejected: violates the byte-identical requirement outright.
**Consequences:** Package layout is Vercel-model-agnostic; the spike de-risks everything
downstream; if the spike proves the preset works, switching later is a deletion
(`api/` + rewrites) rather than a redesign. `cmd/server/main.go` doubles as the preset's
detected entrypoint.

## ADR-0002: Package topology seals backtracking out of the solve path

**Status:** Accepted (2026-08-06)
**Context:** The logic-only rule is the product's #1 constraint; the generator is
explicitly exempt but its uniqueness counter "must never leak into the solve path or the
API surface" (PRD). Discipline alone is not verification.
**Decision:** Packages: `solver` (ladder core — imports stdlib only), `generate` (the only
shipped backtracking, imports solver), `oracle` (test-only brute force), `catalog`, `web`,
`httpapi`, `cmd/server`, `api`. Two mechanical guards in CI: an import-graph test
asserting `solver` imports nothing beyond the stdlib and that no shipped (non-test)
package imports `oracle`; and response-shape tests asserting `/v1/solve` and
`/v1/generate` bodies contain exactly the frozen field sets (no counter leakage).
**Alternatives considered:** Convention + code review only — rejected: the PRD demands
verifiable properties; a test that fails on violation is strictly stronger. Putting the
grid model in its own package — rejected: no third caller distinct from solver's users;
would add a boundary with no payoff.
**Consequences:** "Zero backtracking in the solver" is CI-checkable forever; generate and
oracle can use any search internally without threatening the guarantee; the replay proof
(ADR-0013) stays non-circular because oracle is unreachable from shipped code.

## ADR-0003: Public repository

**Status:** Accepted (2026-08-06)
**Context:** PRD success criteria require CI that blocks merge and a
production-environment approval gate, on free tiers only. GitHub Free provides required
status checks/rulesets and environment required-reviewers only for public repos
(private needs Pro/Team for branch protection and Enterprise for environment reviewers)
— AUDIT.md C1.
**Decision:** The repo becomes public when CI wiring lands (DECISIONS.md D-007).
**Alternatives considered:** Private + advisory gates — rejected: fails two explicit
success criteria. Paying GitHub — rejected: violates the budget constraint and still
doesn't unlock environment reviewers below Enterprise.
**Consequences:** Real, server-side-enforced gates; nothing sensitive is in the repo
(secrets live in GitHub Actions Secrets); the future dashboard can read the repo freely.

## ADR-0004: Error surface — solve-shape 400 vs envelope, and the frozen code set

**Status:** Accepted (2026-08-06)
**Context:** The PRD lists `invalid_input` both as a `/v1/solve` `status` value ("decided
at parse, HTTP 400") and as an envelope `code` example — two mutually exclusive bodies for
the same words (AUDIT.md; research conflict R2-d4).
**Decision:** Domain-invalid puzzles POSTed to `/v1/solve` return HTTP 400 with the FULL
solve response shape (`status:"invalid_input"`, `solved:false`, `solution:""`, zeroed
metrics, `grade:""`, `events:[]`, `input` echoing the raw string). The `{error, code}`
envelope covers transport-level failures on every endpoint — malformed JSON body
(`invalid_input`), wrong content type (415 `unsupported_media_type`), oversized body or
batch over-cap (413 `payload_too_large`), wrong method (405 `method_not_allowed`), unknown
`/v1` path (404 `not_found`), panic (500 `internal_error`) — and `/v1/generate`'s domain
errors (unknown difficulty → 400 `invalid_input`; budget exhaustion → 500
`generation_failed`, ADR-0009), since generate's success shape has no status field. The
envelope `code` set is frozen as exactly those seven values.
**Alternatives considered:** Envelope for solve's invalid puzzles — rejected: would orphan
`invalid_input` from the status enum the PRD explicitly ties to HTTP 400, and force
clients to parse two shapes for one endpoint's domain outcomes. Distinct codes for
byte-cap vs item-cap 413 — rejected: one code + distinct human messages carries the same
information with a smaller frozen surface.
**Consequences:** `/v1/solve` clients always parse one shape for domain outcomes; the
envelope is uniform everywhere else; the code list is now contract and additive changes
are forbidden on `/v1`.

## ADR-0005: Path-only routing with in-handler method dispatch

**Status:** Accepted (2026-08-06)
**Context:** Go 1.22+ ServeMux's automatic 405 for method-specific patterns is plain text
with no override hook (AUDIT.md A4); the PRD requires one JSON error envelope from any
endpoint.
**Decision:** Register path-only patterns (`/v1/health`, `/v1/solve`, `/v1/generate`,
`/v1/validate-batch`, `/v1/puzzles`, `/v1/` catch-all → 404 envelope, `/` static);
each /v1 handler checks `r.Method` and returns 405 with the envelope and an `Allow`
header on mismatch. Static-file 404s outside `/v1` remain the file server's plain text.
**Alternatives considered:** Method-specific patterns + accepting plain-text 405 —
rejected: violates the one-envelope rule on the API surface. Wrapping the mux to intercept
405s — rejected: stdlib gives no reliable hook; sniffing response bodies is fragile.
**Consequences:** Five small method checks; full control of the error surface; `GET`
patterns' automatic HEAD handling is replaced by explicit behavior (HEAD on API routes →
405, harmless for this product).

## ADR-0006: `solveTimeMs` is float64 milliseconds; determinism excludes wall clock

**Status:** Accepted (2026-08-06)
**Context:** Solves are sub-millisecond (PRD UC-5); integer ms would report 0 and destroy
a primary comparison axis. The PRD's "byte-identical event log and metric quartet" cannot
physically include wall clock (AUDIT.md P1; research conflict R3-d1).
**Decision:** `solveTimeMs` is a float64 of `time.Since(start)` in milliseconds, measured
in the handler around the `solver.Solve` call only. The byte-identity guarantee (and every
determinism test) covers `events[]`, `iterations`, `eventCount`, `candidateChecks` — not
`solveTimeMs`.
**Alternatives considered:** Integer microseconds under a new name — rejected: the field
name `solveTimeMs` is frozen by the PRD. Integer ms — rejected: zeroes out the signal.
**Consequences:** Determinism tests compare responses with `solveTimeMs` excluded;
encoding/json's shortest-round-trip float formatting is acceptable because the field is
outside byte-identity claims.

## ADR-0007: Frozen determinism conventions and metric semantics

**Status:** Accepted (2026-08-06)
**Context:** Byte-identical event logs require every scan order, tie-break, and counting
convention pinned before any solver code exists (AUDIT.md L2/L3; the "one sweep per pass"
ambiguity, research R2-c1/c2).
**Decision:** (1) Loop: one productive event per pass; after every event the loop restarts
from `naked_single`; the first firing instance in canonical order wins. Canonical order:
cells row-major (r0-8 × c0-8), digits ascending 1-9, units enumerated rows 0-8 then
columns 0-8 then boxes 0-8 row-major, subset/fish combinations lexicographic, coloring
components from the lowest uncolored row-major cell with digits ascending.
(2) `iterations` = passes started, including the final non-productive pass that concludes
`stalled`/`unsolvable` (so `iterations == eventCount` for solved, `eventCount + 1`
otherwise — a test-asserted invariant). (3) `candidateChecks` = every (cell,digit)
candidate-membership query made by detection logic across all attempted techniques,
counted in one shared accessor; detection short-circuits at the first canonical firing
instance. (4) `witnessCells[]` and `eliminations[]` serialize row-major-then-digit
regardless of discovery order. (5) `gridAfter` is the 81-char digit string (`0` = empty).
**Alternatives considered:** Apply-all-instances-per-pass — rejected: a second elimination
instance can fire after the first spawned a cheaper move, violating the PRD's own grading
discipline ("fires only when nothing cheaper can act"). Counting only the firing
technique's checks — rejected: makes `candidateChecks` depend on log narration rather
than work done, gutting its diagnostic value.
**Consequences:** Grading is airtight by construction; two implementations following this
ADR produce identical logs; the invariant gives tests a free structural check.

## ADR-0008: `unsolvable` keeps the PRD's literal scope

**Status:** Accepted (2026-08-06)
**Context:** The PRD defines `unsolvable` as "a cell was driven to zero candidates"; the
symmetric hidden-analog contradiction (a digit with zero places left in a unit) is equally
cheap to detect but not in the text (AUDIT.md L-conflict; research R2-d3).
**Decision:** Implement the literal test only: a zero-candidate cell (checked at the top
of every pass, before concluding `stalled`) → `unsolvable`. Hidden-analog contradictions
surface as `stalled`, which the PRD explicitly tolerates ("deliberately conflates …
unprovably-unsolvable").
**Alternatives considered:** Adding the hidden-analog check — rejected: changes observable
status for identical inputs relative to the spec text, breaking cross-iteration
comparability, the product's purpose.
**Consequences:** Status semantics are exactly reproducible from the PRD text by any
future iteration; one conflation case more in `stalled`, by design.

## ADR-0009: Generation — bounded attempts, grade-equality by construction, honest failure

**Status:** Accepted (2026-08-06)
**Context:** Expert-band dig-and-grade generation is the one open-ended computation behind
an endpoint; the platform kill returns a bare 504 bypassing the envelope; the
`{difficulty, grade}` response shape admits a mislabeling loophole the PRD's spirit
forbids (AUDIT.md P2; research R4).
**Decision:** `generate.Generate` runs randomized fill → clue removal under a ≤2-solution
uniqueness counter → grade via `solver.Solve` → accept iff grade equals the requested
band, retrying under a 5-second context deadline supplied by the handler. On exhaustion:
HTTP 500 `{error, code:"generation_failed"}`. No precomputed pool. Response `grade` can
never differ from `difficulty` on success, by construction.
**Alternatives considered:** Precomputed embedded pool (primary or fallback) — rejected:
functionally collapses UC-3 into UC-6 and dilutes "difficulty-graded generation" into
catalog lookup; live generation with an honest failure is truer to the PRD. Unbounded
retry — rejected: surrenders the error surface to the platform's bare 504.
**Consequences:** A rare 500 on expert generation under serverless is possible and honest;
localhost (the serious environment) has headroom; tests use seeded RNG for
reproducibility; the 5s number sits under even the stale 10s cap (ADR-0001).

## ADR-0010: CI/CD platform — GitHub Actions + Vercel CLI token flow

**Status:** Accepted (2026-08-06)
**Context:** PRD mandates GitHub Actions CI and a Vercel deploy behind a manual
`workflow_dispatch` with production-environment approval. Vercel's git integration
auto-deploys on push, which would bypass the gate. Cite: ARCHITECTURE.md §CI/CD topology;
AUDIT.md C3.
**Decision:** GitHub Actions for everything. Deploys run `vercel pull --yes
--environment=production` → `vercel build --prod` → `vercel deploy --prebuilt --prod`
authenticated by `VERCEL_TOKEN` + `VERCEL_ORG_ID`/`VERCEL_PROJECT_ID` from GitHub Actions
Secrets. Vercel's git integration is never connected. The GitHub `production` environment
carries the operator as required reviewer; approval happens through the operator's own
authenticated `gh` session (`pending_deployments` API or UI) — the run's `GITHUB_TOKEN`
cannot self-approve, which is correct: the gate is outside the workflow.
**Alternatives considered:** Vercel git integration — rejected: auto-deploy defeats the
manual gate. Deploy from the operator's laptop — rejected: fails the PRD criterion that a
deploy reaches Vercel only through the gated workflow.
**Consequences:** The deploy path is reproducible, gated, and secret-scoped; one-time
setup: local `vercel link`, copy IDs to secrets, create the environment with the operator
as reviewer.

## ADR-0011: CI gate set

**Status:** Accepted (2026-08-06)
**Context:** PRD names the gates; research pinned the correct invocations (AUDIT.md C2,
S5). Cite: ARCHITECTURE.md §CI/CD topology.
**Decision:** Five blocking gates on every PR and on master pushes: `go vet ./...`;
`go build ./...`; `go test -race -coverprofile=coverage.out -coverpkg=./... ./...`;
coverage total ≥ 80.0 via `go tool cover -func` with float-safe awk comparison;
`govulncheck ./...` in plain-text mode installed via `go install
golang.org/x/vuln/cmd/govulncheck@latest`. All are required status checks on `master`.
**Alternatives considered:** Coverage without `-coverpkg=./...` — rejected: verified to
understate cross-package coverage (hands-on reproduction in research). govulncheck via Go
1.24 `tool` directive — rejected: writes third-party entries into the go.mod that is the
PRD's own stdlib-only evidence. Marketplace coverage actions — rejected: a one-line awk
does it with zero new dependencies.
**Consequences:** Gates are enforceable, reproducible, and dependency-free; a red
`security-scan` may be a newly published CVE rather than a regression — triage rule in
EVAL.md.

## ADR-0012: UI conflict resolutions and design language

**Status:** Accepted (2026-08-06)
**Context:** Three tensions inside the PRD's UI bullet (research R5-d1/d2; DECISIONS.md
D-013): three distinguishable highlight states vs "one blue accent"; the pre-solve "hint"
vs the frozen API surface; "statistics window" wording.
**Decision:** The blue accent governs chrome (buttons, active states, links); the three
step-viewer cell states use blue-family lightness/saturation variants plus non-color cues
(solid fill / solid border / dashed border), colour-blind-safe. The pre-solve hint is
static instructional text. Statistics render as an inline panel. Full token set and
recipes frozen in ARCHITECTURE.md §Frontend Design Language.
**Alternatives considered:** Three independent hues — rejected: violates the stated
aesthetic. Puzzle-specific hints — rejected: needs an endpoint the frozen contract lacks,
or client-side solver duplication. Modal statistics dialog — rejected: focus-trap
machinery with no offsetting benefit.
**Consequences:** Functional requirements met inside the stated aesthetic; every choice is
a CSS class, testable by the UI smoke pass.

## ADR-0013: Replay-proof verifier design

**Status:** Accepted (2026-08-06)
**Context:** The replay proof is the no-backtracking guarantee (PRD success criteria);
"forced" means both logical necessity and cheapest-first discipline (AUDIT.md L4/L5).
**Decision:** A test-side verifier that maintains its own shadow candidate state (never
reading solver internals). Per event: placements must satisfy the named single's condition
recomputed from prior state AND equal the `oracle` solution at that cell; eliminations
must have existed as candidates, must not equal the oracle's value at that cell, and the
named technique's witness pattern must structurally hold; `gridAfter` must equal prior
grid plus exactly the stated placement; the final grid must equal the oracle's unique
solution; a second run must byte-match (events + three counters). The verifier
additionally re-implements naked/hidden single detection independently (~30 lines) and
asserts no single was available whenever an elimination technique fired — the scheduling
check. It does NOT re-implement all 13 techniques.
**Alternatives considered:** Full independent re-implementation of the ladder — rejected:
duplicates the system under test (circular-by-duplication) and doubles the defect surface;
oracle-anchored soundness + structural witness checks + curated per-technique fixtures
cover the same claims honestly. Trusting solver-reported candidate state — rejected:
circular.
**Consequences:** Every shipped solve is mechanically re-derivable; a guess-shaped step
surfaces as either an unforced placement or a true-candidate elimination; the verifier
runs over all 55 corpus puzzles (and generated samples) in CI.
