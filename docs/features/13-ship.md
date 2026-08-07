# Feature: Ship — gated production deploy + closeout

**ID:** F-13 · **Roadmap piece:** P-13 · **Status:** In progress (started 2026-08-07, baseline de03e29)

## Description
The terminal gate: a production deployment reached ONLY through the manual
workflow_dispatch + production-environment approval, deployed smoke proving health, UI,
and the frozen header set on the live surface, a final full regression, and the
integration-acceptance closeout confirming every UC's proof ran green on the assembled
system.

## How it fits the roadmap
W8, last. The second convergence point — all five terminal edges meet here.

## Dependencies (must exist before this starts)
- F-02 ci-cd — the deploy workflow, environment, and secrets it triggers
- F-07 technique-fixtures — per-technique evidence complete
- F-10 http-contract — the full API
- F-11 web-ui — the full UI
- F-12 scan-parallel-bench — the committed UC-5 evidence

## Unblocks (what waits on this)
Nothing — terminal.

## Allow-list (source)
(none expected — this piece deploys and verifies; any fix it needs goes through the
amendment path of the owning piece)

## Allow-list (tests)
(none — evidence is live runs and the existing suite)

## Read-only context
- ARCHITECTURE.md §CI/CD topology (deploy topology of record)
- AUDIT.md C3, C4, S3
- DESIGN_DECISIONS.md ADR-0010
- SECURITY.md F-12
- EVAL.md rows "Deployed health", plus the integration-acceptance table in
  docs/ROADMAP.md
- DECISIONS.md D-022, D-023

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
- **AC-1:** A production deployment exists that was created by the deploy workflow via
  workflow_dispatch from master, after a recorded production-environment approval by
  the operator — and by no other path (Vercel git integration unconnected; no local
  `--prod` deploys in the project's deploy history). Evidence: workflow run URL +
  approval record + Vercel deployment metadata in Implementation notes. **Source:**
  ARCHITECTURE.md §CI/CD topology.

## Acceptance criteria
- **AC-2:** The deployed instance answers `GET /v1/health` 200 `{status:"ok", ...}`
  with goVersion matching go.mod, and `GET /` 200 text/html serving the SPA, within the
  bounded cold-start retry budget. Eval row: "Deployed health".
- **AC-3:** The deployed smoke asserts the frozen security header set verbatim on a
  live `/v1` response and on `GET /`. **Source:** SECURITY.md §F-12
- **AC-4:** One live solve round-trips on the deployed instance: POST a corpus puzzle →
  `status:"solved"` with the frozen shape.
- **AC-5:** The full local gate suite is green at the deployed commit (vet, build,
  test -race -coverpkg, coverage ≥80, pinned govulncheck), and the CI run for that
  master commit is green.
- **AC-6:** The integration-acceptance table (docs/ROADMAP.md) is closed out: every UC
  row's integration test ran green at or after the final commit; the closeout table
  (UC → test → run evidence) lands in Implementation notes.
- **AC-7:** The operator performs a final visual smoke of the DEPLOYED UI (load seed,
  solve, step-through) and records it — the deployed surface, not just localhost.

## Testing requirements
Live-deployment smoke (bounded retries), full regression suite, closeout evidence
gathering. No new source.

## Test command
(inherit from CONTEXT.md §Test discipline)

## Coverage command
(inherit)

## Coverage report
(inherit)

## Test-exempt lines
None.

## Health check
(inherit from CONTEXT.md §Deployment discipline)

## Rollback command
(inherit from CONTEXT.md §Deployment discipline)

## Env vars required
None.

## Readability budget
N/A (no source changes; deploy-and-verify piece)

## Complexity exemptions
None.

## Manual setup required
Production-environment approval (the operator's gh session approves the pending
deployment — the manual gate itself, per ARCHITECTURE.md §CI/CD topology).

## Implementation notes (filled in by the building agent)
> Deploy evidence (run URL, approval, deployment URL), smoke results, and the
> integration closeout table land here.
