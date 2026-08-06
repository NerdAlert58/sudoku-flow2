# Feature: Walking skeleton + Vercel deploy spike

**ID:** F-01 · **Roadmap piece:** P-01 · **Status:** Not started

## Description
The smallest end-to-end system: one Go module (stdlib-only, version-pinned), the shared
`httpapi.New()` handler with the complete hardening middleware chain, `/v1/health`, a
placeholder `GET /` page from `embed.FS`, both entrypoints (`cmd/server` on `$PORT`,
`api/index.go` for Vercel), and a real Vercel **preview** deployment that pins which
platform build model actually works. Everything downstream builds inside the shell this
piece proves.

## How it fits the roadmap
Root of the DAG; W0, alone. On the critical path. Its deploy spike retires the project's
single largest platform risk (AUDIT A1/A2) before anything depends on it.

## Dependencies (must exist before this starts)
None — can start immediately.

## Unblocks (what waits on this)
- F-02 ci-cd — a module with passing tests to gate
- F-03 solver-core — the module root (go.mod)
- F-09 catalog — the module root (go.mod)

## Allow-list (source)
- go.mod
- httpapi/** (non-test files)
- cmd/server/main.go
- api/index.go
- vercel.json
- web/web.go
- web/index.html
- web/app.css
- web/app.js
- .gitignore

## Allow-list (tests)
- httpapi/*_test.go

## Read-only context
- ARCHITECTURE.md §Summary (middleware chain, trust posture), §Components (httpapi, cmd/server, api, web), §Contracts C1/C5
- AUDIT.md A1, A2, A4, A5, A6, A7, S1, S2, S3, S4
- DESIGN_DECISIONS.md ADR-0001, ADR-0004, ADR-0005
- SECURITY.md F-10, F-11

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
None — CI/CD lands in F-02.

## Acceptance criteria
- **AC-1:** `go run ./cmd/server` (with `PORT` set) answers `GET /v1/health` 200 with
  exactly `{"status":"ok","goVersion":"<runtime version>","apiVersion":"1"}` and
  `goVersion` equals the go.mod-pinned toolchain.
- **AC-2:** Every response from every route carries the frozen header set verbatim: the
  AUDIT.md S1 CSP, `Strict-Transport-Security: max-age=63072000`,
  `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff` — and never any
  `Access-Control-Allow-Origin` header.
- **AC-3:** Unknown `/v1/*` paths return 404 `{error, code:"not_found"}`; a wrong method
  on `/v1/health` returns 405 envelope with an `Allow` header (ADR-0005).
- **AC-4:** A panicking test handler recovers to 500 `{error, code:"internal_error"}`
  and the request still produces one structured access-log line (method, path, status
  500, duration) — AUDIT A6 ordering observable.
- **AC-5:** `GET /` serves the embedded placeholder page (external .css/.js only, zero
  inline code) with 200 text/html.
- **AC-6:** A real Vercel **preview** deployment (never `--prod`) serves AC-1's health
  response and AC-2's header set from the deployed URL; the build model that worked
  (classic `api/` vs Go server preset) is recorded in Implementation notes. Under
  EITHER model, the platform max duration is explicitly configured (vercel.json) at the
  smallest value covering the 5s generation deadline where the platform allows it, and
  otherwise the effective platform value is observed and recorded in Implementation
  notes with where it was observed — never left silently at platform maximum.
  **Source:** SECURITY.md §F-10
- **AC-7:** go.mod pins a specific currently-supported Go version (≥1.22); the deployed
  `/v1/health` goVersion confirms the platform honored it. **Source:** SECURITY.md §F-11
- **AC-8:** `go.sum` does not exist and go.mod contains zero `require` entries (stdlib
  only), and no `internal/` directory exists anywhere in the repo.
- **AC-9:** Total statement coverage of the module at this piece's exit is ≥ 80.0%
  measured by the frozen coverage command — so F-02's first gated PR inherits a green
  floor, not a shortfall.

## Testing requirements
Handler-level unit tests (httptest) covering AC-1..AC-5 and AC-8's structural
assertions; the deploy spike (AC-6/7) is verified by live curl, evidence pasted into
Implementation notes.

## Test command
(inherit from CONTEXT.md §Test discipline)

## Coverage command
(inherit)

## Coverage report
(inherit)

## Test-exempt lines
- cmd/server/main.go:L1-L999 — thin entrypoint (env read + ListenAndServe); exercised by
  manual smoke, not unit-testable without process orchestration
- api/index.go:L1-L999 — thin Vercel adapter; exercised by the live deploy spike (AC-6)

## Health check
$ curl -fsS "$PREVIEW_URL/v1/health" | grep -q '"status":"ok"'

## Rollback command
N/A (preview deployments are disposable; no production surface exists yet)

## Env vars required
None.

## Readability budget
(inherit from CONTEXT.md §Rigor)

## Complexity exemptions
None.

## Manual setup required
None — Vercel CLI is already authenticated; the first `vercel deploy` creates the
project.

## Implementation notes (filled in by the building agent)
> The agent implementing this feature records its decisions and rationale here as it
> builds — chosen patterns within the architecture's constraints, trade-offs made,
> deviations and why. Cross-cutting discoveries that affect other features must also be
> propagated to ROADMAP.md or ARCHITECTURE.md, not just left here.
