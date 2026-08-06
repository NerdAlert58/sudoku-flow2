# Feature: CI/CD gates + gated deploy pipeline

**ID:** F-02 · **Roadmap piece:** P-02 · **Status:** Not started

## Description
The enforcement layer: both GitHub Actions workflows, the repo flipped public, required
status checks on master, the `production` environment with the operator as required
reviewer, and deploy credentials provisioned. From this piece on, every merge is blocked
by the five gates and the only path to production is the manual workflow_dispatch gate.

## How it fits the roadmap
W1, parallel with F-03 and F-09 (disjoint surfaces). Off the solver critical path but
gates every subsequent merge — landing it early is the point.

## Dependencies (must exist before this starts)
- F-01 walking-skeleton — a module with passing tests and ≥80% coverage to gate

## Unblocks (what waits on this)
- F-13 ship — the deploy workflow and production environment it triggers

## Allow-list (source)
- .github/workflows/ci.yml
- .github/workflows/deploy.yml

(Repo settings — visibility, branch protection, environment, secrets — are operator
actions via gh/vercel CLI, recorded in Implementation notes; they are not files.)

## Allow-list (tests)
(none — CI wiring piece; its checks are live runs)

## Read-only context
- ARCHITECTURE.md §CI/CD topology (authoritative commands, triggers, gates, hardening block, deploy topology)
- AUDIT.md C1, C2, C3, C4, S5, S6
- DESIGN_DECISIONS.md ADR-0003, ADR-0010, ADR-0011
- SECURITY.md F-11, Advisory notes
- DECISIONS.md D-022 (token acquisition), D-023 (spike vs gate)

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
- **AC-1:** Every path in ARCHITECTURE.md §CI/CD topology's Config file paths
  (`.github/workflows/ci.yml`, `.github/workflows/deploy.yml`, `vercel.json`) exists in
  the repo. **Source:** ARCHITECTURE.md §CI/CD topology.
- **AC-2:** A real `pr:opened`/`pr:updated` event on this piece's PR fires the CI
  workflow and produces pass/fail signals for every gate — vet, build, test (-race,
  -coverpkg), coverage ≥80.0 float-safe, security-scan (pinned govulncheck, plain text)
  — with the CI run URL recorded in Implementation notes. **Source:** ARCHITECTURE.md
  §CI/CD topology.
- **AC-3:** The coverage gate produces an LCOV report (`coverage.lcov` via pinned
  gcov2lcov) so it uses the same LCOV parser + diff-line floor as local /nerdflow:build
  Phase 5b.6, in addition to the 80% total floor. **Source:** ARCHITECTURE.md §CI/CD
  topology.

## Acceptance criteria
- **AC-4:** The repository is public, and `master` has required status checks covering
  all five gates — a PR with a failing gate is observably un-mergeable (evidence: the
  branch-protection API response and a red-check PR state recorded in notes).
- **AC-5:** The GitHub `production` environment exists with the operator as required
  reviewer and "prevent self-review" off; `deploy.yml` triggers only on
  workflow_dispatch, deploys only from master, and its job is bound to the `production`
  environment (a dispatched run observably pauses for approval).
- **AC-6:** `VERCEL_TOKEN`, `VERCEL_ORG_ID`, `VERCEL_PROJECT_ID` exist as
  environment-scoped secrets on `production`; `.vercel/` is git-ignored and absent from
  the repo; the token's provenance (fresh-created vs CLI-token fallback) is recorded per
  DECISIONS.md D-022.
- **AC-7:** Both workflows pin actions (major tag minimum, SHA preferred) and declare
  `permissions: contents: read`; the go version comes from `go-version-file: go.mod`
  (single source). **Source:** SECURITY.md §F-11
- **AC-8:** The deploy workflow's smoke step is present and bounded (5 × 3s retries)
  asserting `/v1/health` 200 + `status:"ok"`, `GET /` 200 text/html, AND the frozen
  security header set verbatim on both responses (so the SECURITY.md §F-12 check
  re-asserts on every future deploy, not only at F-13) — exercised end-to-end at F-13,
  present and syntax-valid here (actionlint or a dry parse).

## Testing requirements
Live CI runs are the tests. Evidence links (run URLs, API responses) land in
Implementation notes.

## Test command
N/A (no executable source added; the piece's checks are the live CI runs on its own PR)

## Coverage command
N/A (same rationale)

## Coverage report
N/A (same rationale)

## Test-exempt lines
None.

## Health check
N/A (no runtime deploy in this piece; deploy.yml is exercised at F-13)

## Rollback command
(inherit from CONTEXT.md §Deployment discipline)

## Env vars required
None at runtime. (CI job env: VERCEL_TOKEN/VERCEL_ORG_ID/VERCEL_PROJECT_ID from the
production environment scope.)

## Readability budget
N/A (YAML wiring; no functions)

## Complexity exemptions
None.

## Manual setup required
- VERCEL_TOKEN provisioning per DECISIONS.md D-022 (operator credential handling — the
  coordinator performs it, logs provenance).
- Repo visibility flip + branch protection + environment creation via gh api (operator
  actions, logged with API responses).

## Implementation notes (filled in by the building agent)
> The agent implementing this feature records its decisions and rationale here as it
> builds. Cross-cutting discoveries propagate to ROADMAP.md or ARCHITECTURE.md.
