# Security posture — sudoku-flowN

**Application class:** hybrid: web-app+serverless
**Rubric version:** reference/security-hats/web-app.md @ 98f4290c46465befefa020aa3d5756e62f485d21 · reference/security-hats/serverless.md @ 98f4290c46465befefa020aa3d5756e62f485d21
**Reviewed on:** 2026-08-06
**Reviewer verdict:** FAIL-resolved

Review history: round 1 (kaladin, tuned) — FAIL, 2 blocking / 6 tradeoff / 3 deferred /
29 not-applicable over 58 deduped rubric items. Round 2 after artifact revision — PASS,
all dimensions. Staleness refresh over commit ba8baeb — PASS confirmed. Deferred slugs
below are the superset of both rounds' lists (round-2 wording canonical where they name
the same concern).

## Findings

### F-1 HSTS max-age unpinned
- **Rubric item:** WEB-05 — Security Misconfiguration
- **Severity:** medium
- **Applies to:** AUDIT.md §S3, ARCHITECTURE.md §Summary, EVAL.md contract-edge row
- **Finding:** The header was frozen as emitted but no max-age value appeared in any
  artifact; a trivial or divergent value would pass every stated test while disabling the
  control.
- **Resolution:** resolved-in-arch (AUDIT.md §S3 freezes
  `Strict-Transport-Security: max-age=63072000` with the no-includeSubDomains rationale;
  asserted verbatim by the per-route header tests per EVAL.md).

### F-2 No written threat model
- **Rubric item:** WEB-11 — Threat modeling
- **Severity:** high
- **Applies to:** AUDIT.md §Security, USERS.md §What the System Will NOT Do
- **Finding:** Six tradeoff rationales appealed to a threat model ("no sensitive assets,"
  "single-operator," "ephemeral") that was written nowhere; the accepted-risk posture had
  no recorded derivation.
- **Resolution:** resolved-in-arch (AUDIT.md §S6: four defended assets ranked, three
  attacker classes with mitigations, tradeoffs derived explicitly; folds in the three
  round-1 rubric-gap advisories — deploy-credential blast radius/rotation, CI
  tooling/actions pinning + fork-PR posture, contract fidelity as a defended asset).

### F-3 No per-caller rate limiting
- **Rubric item:** SVR-27 (also WEB-scope abuse surface)
- **Severity:** low
- **Applies to:** ARCHITECTURE.md §Known Tradeoffs
- **Finding:** All endpoints are public and unauthenticated with no rate limit.
- **Resolution:** accepted-tradeoff (ARCHITECTURE.md §Known Tradeoffs; bounded by
  1 MiB/256-item caps, 5s generation deadline, sub-ms solves, platform DDoS protection,
  ephemeral demo lifecycle; rate limiting would alter the benchmarked timing surface;
  revisit if a deployment is ever long-lived).

### F-4 No concurrency/instance ceiling
- **Rubric item:** SVR-14
- **Severity:** low
- **Applies to:** ARCHITECTURE.md §Known Tradeoffs
- **Finding:** No reserved-concurrency bound on the public function; denial-of-wallet is
  possible in principle.
- **Resolution:** accepted-tradeoff (free-tier scope bounds spend; per-invocation work
  bounded by caps/deadline; AUDIT.md §S6 anonymous-caller class).

### F-5 No WAF beyond platform protections
- **Rubric item:** SVR-28
- **Severity:** low
- **Applies to:** ARCHITECTURE.md §Known Tradeoffs
- **Finding:** No WAF layer beyond Vercel's defaults.
- **Resolution:** accepted-tradeoff (the injection surface a WAF filters is structurally
  absent — no DB, no templates, 81-char digit grammar enforced at parse).

### F-6 No code signing on the deploy path
- **Rubric item:** SVR-23
- **Severity:** low
- **Applies to:** DESIGN_DECISIONS.md ADR-0010, ARCHITECTURE.md §Known Tradeoffs
- **Finding:** Deployed artifacts are not signed.
- **Resolution:** accepted-tradeoff (deploy trust pinned to the single gated workflow:
  master-only, environment-scoped secrets, operator approval outside the workflow, Vercel
  git integration never connected — AUDIT.md §S6).

### F-7 No alertable security signals
- **Rubric items:** WEB-24, SVR-25
- **Severity:** low
- **Applies to:** ARCHITECTURE.md §Observability, §Known Tradeoffs
- **Finding:** No alerting/tracing; deploy smoke is the only wired check.
- **Resolution:** accepted-tradeoff (single-operator, stateless, no confidentiality
  assets; smoke failure fails the workflow and the operator acts on it — AUDIT.md §S6).

### F-8 Public repository
- **Rubric item:** threat-model consequence (AUDIT.md §C1)
- **Severity:** low
- **Applies to:** ARCHITECTURE.md §Known Tradeoffs, DESIGN_DECISIONS.md ADR-0003
- **Finding:** Public repo enlarges the supply-chain/CI exposure (fork PRs, visible CI).
- **Resolution:** accepted-tradeoff (required for free-tier required status checks and
  environment approval; no secrets in the repo; fork-PR and pinning posture per AUDIT.md
  §S6 / ADR-0011).

### F-9 Panic path must not leak internals
- **Rubric item:** WEB-14
- **Severity:** medium
- **Applies to:** httpapi panic-recovery path (ADR-0004 envelope)
- **Finding:** The architectural control exists (single envelope, frozen code set); that
  the 500 message carries no panic value, stack frame, or path is verifiable only as a
  runtime test.
- **Resolution:** deferred-to-impl (piece TBD-panic-message-no-internals)
- **Acceptance signal:** a test asserts the 500 `internal_error` envelope carries a fixed
  generic message and no panic value, stack frame, or file path appears in any error body.

### F-10 Platform max duration must be pinned
- **Rubric item:** SVR-13
- **Severity:** medium
- **Applies to:** Vercel function configuration (ADR-0001/0009 budgets)
- **Finding:** Internal budgets (5s generate, 10s design bound) are frozen but the
  platform-configured max duration is stated nowhere and cannot be pinned before the
  deploy spike resolves the build model (AUDIT.md §A1).
- **Resolution:** deferred-to-impl (piece TBD-platform-timeout-pin)
- **Acceptance signal:** the deployed function's max duration is explicitly configured or
  recorded at the smallest value covering the 5s generation deadline, verified during the
  walking-skeleton deploy spike — not left at platform maximum.

### F-11 Go toolchain single-source pin
- **Rubric item:** SVR-21
- **Severity:** medium
- **Applies to:** go.mod, CI workflows, Vercel build
- **Finding:** No artifact pins the Go toolchain; CI and the platform build could drift.
- **Resolution:** deferred-to-impl (piece TBD-go-version-pin)
- **Acceptance signal:** go.mod pins a specific currently-supported Go version (≥1.22
  floor); CI reads the version from go.mod as the single source; `/v1/health` goVersion
  confirms the deployed toolchain at the deploy-spike smoke.

### F-12 Deployed header smoke
- **Rubric items:** WEB-05/WEB-12 (deployed-surface seam)
- **Severity:** low
- **Applies to:** deploy workflow smoke step
- **Finding:** The frozen header set is asserted in-process only; nothing proves the
  platform edge preserves the middleware's headers on the deployed surface.
- **Resolution:** deferred-to-impl (piece TBD-deployed-header-smoke)
- **Acceptance signal:** post-deploy smoke asserts the frozen security header set
  verbatim on a live `/v1` response and on `GET /`.

## Advisory notes (rubric gaps, non-blocking)

- **Operator account security:** the deploy gate's entire authority chain is the
  operator's GitHub and Vercel accounts; neither rubric asks about their 2FA/hardware-key
  posture. Recorded here as the residual root of trust.
- **Mutable action tags:** major-version action tags are mutable; commit-SHA pinning is
  the immutable form and is preferred when the CI piece lands (ARCHITECTURE.md §CI/CD
  workflow-hardening block).
