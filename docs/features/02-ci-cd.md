# Feature: CI/CD gates + gated deploy pipeline

**ID:** F-02 · **Roadmap piece:** P-02 · **Status:** In progress (worktree, started 2026-08-06, baseline 5559999)

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

**2026-08-06 — F-02 build session (workflow files only; repo settings/secrets are the
coordinator's operator actions).**

- **Job/gate naming (coordinator directive):** four CI jobs — `lint` (go vet), `build`,
  `test`, `security-scan` — are the required-status-check names; the coverage gate is a
  named `coverage` step inside the `test` job (blocks merge via `test`'s status). This
  maps ARCHITECTURE §CI/CD topology's five gates onto four checks: vet→`lint`,
  coverage→step-in-`test`.
- **Action pins (SHA + version comment), verified via GitHub API this session:**
  `actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1` (# v7.0.1 — major tag v7
  points at this commit; matched to refs/tags/v7.0.1) and
  `actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` (# v7.0.0 — major tag v7 =
  refs/tags/v7.0.0). Exceeds SECURITY F-11 / AC-7's "major tag minimum".
- **Tool pins:** `gcov2lcov@v1.1.1` (latest tag per
  `go list -m -versions github.com/jandelgado/gcov2lcov`; closes f-01 session Open item —
  produces `coverage.lcov` in CI, AC-3) and `govulncheck@v1.6.0` (latest golang.org/x/vuln
  tag per `go list -m -versions`; plain-text mode only — JSON always exits 0, AUDIT S5).
  `vercel@58` (npm latest 58.7.1; major pin — same major the F-01 spike validated,
  CLI 58.7.1).
- **Coverage gate:** `go tool cover -func | grep total | awk '{print $3}' | tr -d '%'`
  then float-safe `awk -v total=… 'BEGIN { exit !(total >= 80.0) }'` (AUDIT C2); verified
  locally: passes 85.4 and 80.0, fails 79.9. Test run uses `-race -coverpkg=./...`
  exactly as frozen.
- **Deploy workflow:** workflow_dispatch only; job bound to `environment: production`;
  explicit targets on every vercel command (`--environment=production`, `--prod`,
  `--prebuilt --prod`) per ADR-0016(2); deployment URL captured to step output.
- **Smoke (AC-8):** runs against the production domain https://sudoku-flow2.vercel.app,
  never the per-deployment URL (SSO-protected, ADR-0016(3)); 5 × 3s bounded retries
  (AUDIT C4); asserts /v1/health 200 + `"status":"ok"`, `/` 200 text/html, and the frozen
  header set verbatim on BOTH responses (CSP, HSTS max-age=63072000, X-Frame-Options
  DENY, X-Content-Type-Options nosniff — SECURITY F-12). Header extraction is
  case-insensitive on names, exact on values, CRLF-safe; verified against a synthetic
  response fixture including a negative case.
- **Hardening:** both workflows declare workflow-level `permissions: contents: read`;
  go version single-sourced via `go-version-file: go.mod` (AC-7); secrets referenced only
  in the production-environment job.
- **Validation:** `actionlint` (which also shellchecks run blocks) passes both files.
- **Not done here (operator/coordinator):** repo visibility flip, branch protection with
  the four required checks, `production` environment + reviewer, secrets provisioning
  (D-022 provenance), and the live-run evidence URLs for AC-2/AC-4/AC-5/AC-6.
