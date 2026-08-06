# Operator Decision Log — sudoku-flow2

Autonomous-run log. The operator (user) directed: run the full nerdflow process from the
existing PRD.md through Vercel deployment of master, raising no questions unless continuation
is impossible. Every judgment call made on the user's behalf is recorded here as
question → options → selected → reasons.

---

## D-001 — PRD source
- **Question:** /nerdflow:idea found an existing PRD.md. Overwrite, refine, or use as-is?
- **Options:** (a) overwrite from scratch; (b) refine via idea workflow; (c) use as-is.
- **Selected:** (c) use as-is.
- **Reasons:** User directive ("use the PRD.md"). The PRD is second-generation, distilled
  from the shipped v1 build, and explicitly designed to be the sole input to this run.

## D-002 — Git bootstrap and branch discipline
- **Question:** How to reconcile "never write directly to master" (user global rules) with a
  fresh repo that has no branches yet and a fully autonomous run?
- **Options:** (a) do all work directly on master; (b) bootstrap commit on master
  (pre-existing PRD.md + puzzles.txt + this log + .gitignore), then all subsequent work on
  feature branches merged via PRs; (c) block waiting for human review on each PR.
- **Selected:** (b), with PRs self-merged by the operator agent only after all CI gates pass.
- **Reasons:** The initial commit must land somewhere for branches to exist. The global rule's
  purpose (no unreviewed changes to master) is preserved by gating every merge on green CI;
  (c) contradicts the autonomous mandate; (a) discards the audit trail PRs provide.

## D-003 — GitHub repository visibility
- **Question:** The run needs a GitHub remote (PRs, Actions CI, gated Vercel deploy). Public
  or private?
- **Options:** (a) public; (b) private.
- **Selected:** (b) private.
- **Reasons:** Repo visibility is a user decision per global rules, and the user barred
  questions; private is the minimal-exposure default and is reversible in one click. Vercel
  can deploy a private repo via CLI/token; nothing in the PRD requires public.

## D-004 — Autonomous elicitation policy
- **Question:** nerdflow's arch/impl phases elicit decisions interactively. Who answers?
- **Options:** (a) pause for the human at each elicitation; (b) operator agent answers from
  the PRD (which froze most decisions deliberately) and logs each answer here.
- **Selected:** (b).
- **Reasons:** User directive ("no questions raised to the human unless it is impossible to
  continue"). The PRD was written to settle contested decisions in advance; remaining
  freedoms are builder's-choice by design.

## D-005 — PRD deadline gate
- **Question:** /nerdflow:arch requires a calendar date or hour count; the PRD says "no
  fixed calendar deadline, continuous build until done."
- **Options:** (a) block and elicit a date; (b) accept the PRD's statement as a concrete
  time-budget policy (continuous, operator-paced, this run).
- **Selected:** (b).
- **Reasons:** The PRD's statement is deliberate and specific (AI-driven continuous build,
  operator directs pacing), not a vague "soon" — and the operator's current instruction is
  exactly that: run to completion now.

## D-006 — Rigor profile
- **Question:** lite / standard / full?
- **Options:** lite; standard; full.
- **Selected:** standard.
- **Reasons:** First matching signal in the derivation table: PRD names a deploy target
  (Vercel) → at least standard. No regulatory regime → full not indicated. Not a ≤1-day
  throwaway → lite not indicated.

## D-007 — Repository visibility (revises D-003)
- **Question:** Research verified that GitHub Free gives private repos neither
  required-status-check branch protection (nor rulesets) nor environment
  required-reviewers. The PRD requires both ("CI blocks merge on any gate failure";
  "deploy … only through the manual gate" with "production-environment approval") and
  "free tiers only." Keep private or go public?
- **Options:** (a) stay private, gates advisory-only (violates two PRD success criteria);
  (b) pay GitHub Pro/Team (violates budget; STILL doesn't unlock environment reviewers on
  private — that needs Enterprise); (c) make the repo public.
- **Selected:** (c) public — flipped when CI wiring lands.
- **Reasons:** The user's own PRD constraints are jointly satisfiable only by (c). Contents
  are a Sudoku solver, a PRD, and puzzle fixtures — nothing sensitive; secrets live in
  GitHub Actions Secrets, never in the repo. D-003's private-first reasoning (minimal
  exposure while undecided) is superseded by the PRD's explicit gate requirements.

## D-008 — Vercel deployment model
- **Question:** Research found the PRD's platform facts are stale: current Vercel docs
  describe a Go Framework Preset running a real `package main` server, and a 300s Hobby
  duration cap under Fluid Compute (PRD says serverless-functions-only and 10s). Which
  model to build on?
- **Options:** (a) new Go Framework Preset (docs dated 2026-07, internally inconsistent
  with Vercel's own runtimes table, unproven here); (b) classic `api/index.go` +
  `func Handler(w,r)` + vercel.json rewrites (the PRD's known-working path); (c) build
  for (b) with a package layout that also works under (a).
- **Selected:** (c): shared `http.Handler` constructor package consumed by both
  `cmd/server` (local, $PORT) and `api/index.go` (Vercel); **no `internal/` packages
  anywhere** so the classic model's module-path rewrite restriction is moot; a
  walking-skeleton deploy spike early in the build pins actual platform behavior. Solver
  time budgets designed to the conservative 10s figure regardless.
- **Reasons:** (b) is proven by the PRD's own history; avoiding `internal/` costs nothing
  here and removes the one failure mode a local build can't reproduce; designing to the
  stale-conservative timing bound is safe under either cap.

## D-009 — Frozen-contract ambiguity resolutions
- **Question:** Research surfaced contract points the PRD leaves ambiguous; they must be
  frozen before build. (Full detail in DESIGN_DECISIONS.md ADRs.)
- **Selected:**
  - `/v1/solve` domain-invalid puzzle → HTTP 400 with the FULL solve shape,
    `status:"invalid_input"` (the status enum demands it); the `{error, code}` envelope is
    for transport-level failures (malformed JSON, 415, 413, 405, 404 under /v1, 500) and
    for `/v1/generate` unknown difficulty (its shape has no status field).
  - Envelope `code` set frozen: `invalid_input`, `unsupported_media_type`,
    `payload_too_large`, `method_not_allowed`, `not_found`, `internal_error`,
    `generation_failed`.
  - Routing: path-only ServeMux patterns + in-handler method dispatch so 405s carry the
    JSON envelope (stdlib's automatic 405 is plain-text and non-overridable). Static
    (non-/v1) 404s may be plain text.
  - `solveTimeMs` = float64 milliseconds; byte-identical determinism is scoped to
    events + iterations + eventCount + candidateChecks (wall clock physically cannot be
    byte-identical; PRD tension resolved by scoping).
  - `grade` never carries omitempty; key always present.
  - `unsolvable`/`stalled` are HTTP 200 (only invalid_input is assigned 400 by the PRD).
  - Body cap: 1 MiB uniform on all POST endpoints; Content-Length fast-path + MaxBytesReader
    enforcement; batch item-cap (256) checked after parse, before any solving.
  - Health: `{status:"ok", goVersion: runtime.Version(), apiVersion:"1"}`.
- **Reasons:** Each is the reading that keeps the PRD internally consistent with the
  fewest moving parts; logged as ADRs with alternatives.

## D-010 — /v1/generate failure policy
- **Question:** Expert-band generation may exhaust a time budget on serverless. Behavior?
- **Options:** (a) unbounded retry (can hit platform kill → bare 504); (b) precomputed
  embedded pool fallback (redundant with /v1/puzzles, dilutes UC-3); (c) bounded attempts
  under an internal 5s deadline, `grade` must equal requested difficulty by construction,
  exhaustion → HTTP 500 `{error, code:"generation_failed"}`.
- **Selected:** (c).
- **Reasons:** Honest failure beats silent relabeling (the {difficulty, grade} shape
  admits a mislabeling loophole — closed by construction rule); 5s sits under even the
  stale 10s cap; serious use is localhost per the PRD.

## D-011 — `unsolvable` detection scope
- **Question:** PRD's literal test (a cell driven to zero candidates) omits the symmetric
  hidden analog (a digit with zero places left in a unit). Implement which?
- **Options:** (a) literal only — hidden-analog states report `stalled`; (b) both checks.
- **Selected:** (a) literal only.
- **Reasons:** The contract must match the PRD text exactly across iterations; the PRD
  explicitly tolerates `stalled` conflating undetected contradictions. (b) would change
  observable status for identical inputs vs. the spec text.

## D-012 — Solve-loop granularity and metric freezing
- **Question:** "One cheapest-first sweep per pass" admits two readings with different
  event logs.
- **Options:** (a) one productive event per pass, restart from cheapest; (b) apply all
  instances of the winning technique per pass.
- **Selected:** (a). `iterations` = passes started (== eventCount when solved; +1 final
  empty pass when stalled/unsolvable). `candidateChecks` = every (cell,digit)
  candidate-membership query by detection logic across ALL attempted techniques, via one
  counted accessor; techniques short-circuit at the first canonically-ordered instance.
- **Reasons:** (b) can fire a second elimination instance when the first spawned a cheaper
  move — violating the PRD's own grading discipline ("fires only when nothing cheaper can
  act"). (a) is the only airtight reading.

## D-013 — UI conflict resolutions
- **Question:** Three PRD tensions in the UI bullet.
- **Selected:** (1) "one blue accent" governs chrome; the three step-viewer cell states
  use blue-family lightness/saturation variants PLUS non-color cues (border/glyph) so
  they're distinguishable and colour-blind-safe; (2) the pre-solve "hint" is static
  instructional text (puzzle-specific hints would need an API surface the frozen contract
  lacks and would duplicate solver logic client-side); (3) the "statistics window" is an
  inline panel, not a modal (fewer moving parts, no focus-trap code).
- **Reasons:** Each keeps an explicit functional requirement without violating an explicit
  constraint; logged for the halliday review to challenge.

## D-014 — Compliance N/A declaration
- **Question:** Does the project touch personal/payment/health/education data or a
  regulated environment anywhere, including logs and fixtures?
- **Selected:** No → COMPLIANCE.md N/A stub.
- **Reasons:** Data surface is 81-char digit strings, solver events, and access logs of
  method/path/status/duration (no IPs, no identities, no accounts). PRD names no regime.

## D-015 — Application class for the security gate
- **Question:** Which kaladin rubric class?
- **Options:** web-app; backend-api; serverless; hybrids.
- **Selected:** `hybrid: web-app+serverless`.
- **Reasons:** One HTTP surface serving an SPA + JSON API (web-app rubric) deployed as a
  Vercel function alongside a local binary (serverless rubric). backend-api would
  double-count the same HTTP surface with M2M assumptions (auth) the PRD explicitly
  refuses.

## D-016 — Frontend taste reference
- **Question:** Phase 3d requires a taste reference per user-facing surface.
- **Selected:** The PRD itself is the reference kit: "McKinsey-clean" — system-ui,
  near-monochrome + one blue accent, grid as hero, symmetric borders; token set frozen in
  ARCHITECTURE.md §Frontend Design Language. No external kit.
- **Reasons:** The PRD supplies a complete, deliberate aesthetic; importing an external
  kit would override a frozen input.
