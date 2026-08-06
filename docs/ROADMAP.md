# Roadmap — sudoku-flowN

**Status:** draft · **Date:** 2026-08-06 · **Source inputs (frozen):** PRD.md, AUDIT.md, USERS.md, ARCHITECTURE.md, DESIGN_DECISIONS.md, EVAL.md, SECURITY.md, COMPLIANCE.md

## Overview

Build the stateless Go Sudoku benchmark service (13-technique logic-only solver,
replayable event log, generation, batch validation, embedded SPA) as 13 feature pieces
over a DAG. Strategy in one line: deploy-spike and CI gates first, then a serial solver
spine (core → mid ladder → upper ladder → oracle/replay) with catalog and benchmark work
fanned out beside it, converging at the HTTP contract, the UI, and a gated production
ship.

## Project pieces

| ID | Piece | Purpose | Depends on | Unblocks | Parallel with |
|---|---|---|---|---|---|
| F-01 | walking-skeleton | go.mod + shared handler shell + both entrypoints + Vercel preview deploy spike pins platform reality | — | F-02, F-03, F-09 | — |
| F-02 | ci-cd | Workflows, public repo, branch protection, production environment + secrets; gates live from here on | F-01 | F-13 | F-03, F-09 |
| F-03 | solver-core | Grid model, parse/validate, candidates, singles, one-event-per-pass loop, events, metrics, grading; 25 ORIGINAL seeds solve | F-01 | F-04 | F-02, F-09 |
| F-04 | ladder-mid | Locked candidates (pointing/claiming) + naked/hidden subsets | F-03 | F-05 | — |
| F-05 | ladder-upper | X-wing, swordfish, jellyfish, XY/XYZ/W-wing, simple colouring; all 55 seeds solve | F-04 | F-06, F-08, F-12 | — |
| F-06 | oracle-replay | Test-only brute-force oracle, replay-proof verifier, determinism tests, corpus proofs, import-graph containment | F-05 | F-07, F-08 | F-12 |
| F-07 | technique-fixtures | Per-technique fires-and-sound fixtures + ladder-cap harness + necessity/sufficiency where isolable | F-06 | F-13 | F-08 |
| F-08 | generator | Sealed dig-and-grade generation, grade==band by construction, seeded band tests | F-05, F-06 | F-10 | F-07 |
| F-09 | catalog | Embedded puzzles.txt copy, section parsing, canonical names, drift-guard test | F-01 | F-10 | F-02, F-03 |
| F-10 | http-contract | Full /v1 surface (solve/generate/validate-batch/puzzles) + complete contract-edge test matrix + batch fan-out | F-05, F-08, F-09 | F-11, F-13 | F-12 |
| F-11 | web-ui | Embedded SPA per Frontend Design Language, DOM-contract test, operator visual smoke | F-10 | F-13 | F-12 |
| F-12 | scan-parallel-bench | SolveScanParallel behind explicit call, committed negative-result benchmark, containment static-scan test | F-05 | F-13 | F-06, F-10, F-11 |
| F-13 | ship | Production deploy through the manual gate, deployed smoke (health/UI/headers), integration closeout | F-02, F-07, F-10, F-11, F-12 | — | — |

## Dependency graph

```
F-01 ──┬── F-02 ─────────────────────────────────────────┐
       ├── F-03 ── F-04 ── F-05 ──┬── F-06 ──┬── F-07 ───┤
       │                          │          └── F-08 ─┐ │
       │                          └── F-12 ────────────┼─┤
       └── F-09 ───────────────────────────────────────┤ │
                                          F-08, F-09 ──┴─► F-10 ── F-11 ──► F-13
                                                          (F-02, F-07, F-12 also → F-13)
```

Waves: W0 {F-01} → W1 {F-02 ∥ F-03 ∥ F-09} → W2 {F-04} → W3 {F-05} → W4 {F-06 ∥ F-12}
→ W5 {F-07 ∥ F-08} → W6 {F-10} → W7 {F-11} → W8 {F-13}.

## Critical path

F-01 → F-03 → F-04 → F-05 → F-06 → F-08 → F-10 → F-11 → F-13 — the solver spine plus its
converging HTTP/UI tail determines minimum schedule. Convergence points: **F-10** (three
inbound edges — ladder, generator, catalog — any slip serializes here) and **F-13** (five
inbound edges — the ship gate proves the whole system, so every loose end surfaces here).
The riskiest single piece is F-05 (seven technique detectors incl. the positively-encoded
colour wrap; all-55 corpus proof is its exit bar).

## Parallelization plan

- **W1:** F-02 (.github/**, repo settings) ∥ F-03 (solver/**) ∥ F-09 (catalog/**) —
  disjoint write surfaces. A fresh agent needs: its feature file + ARCHITECTURE.md
  §Components/§Contracts sections it cites (+ §CI/CD topology for F-02).
- **W4:** F-06 (oracle/** + solver replay/determinism/corpus test files) ∥ F-12
  (solver/scanparallel.go + docs/bench/** + its bench test file) — same directory,
  disjoint files; allow-lists are exact.
- **W5:** F-07 (solver/testdata/** + technique test files) ∥ F-08 (generate/**).
- Everything else is serial by data dependency. go.mod is written once at F-01 and never
  modified again (stdlib-only forever) — no shared-manifest collisions exist.

## Cross-cutting contracts

### Wire contract C1 and envelope
Owner: ARCHITECTURE.md §Contracts C1 + DESIGN_DECISIONS.md ADR-0004/ADR-0014. Every
piece touching HTTP obeys the frozen shapes, status enums, seven envelope codes, and
raw-echo rule. Never restated in briefs.

### Determinism conventions and metric semantics
Owner: DESIGN_DECISIONS.md ADR-0007 (+ ADR-0006 scope). All solver work obeys the frozen
scan orders, tie-breaks, one-event-per-pass loop, per-solve counter context, and
canonical serialization order.

### Containment rules
Owner: DESIGN_DECISIONS.md ADR-0002/ADR-0015. solver imports stdlib only; nothing
shipped imports oracle; nothing outside benchmarks references SolveScanParallel — all
three mechanically tested (F-06/F-12 own the tests).

### Security header set
Owner: AUDIT.md S1/S3/S4 (CSP verbatim, HSTS max-age=63072000, frame-denial, nosniff).
Asserted per-route by F-10's tests and on the deployed surface by F-13's smoke.

### Frontend design language
Owner: ARCHITECTURE.md §Frontend Design Language (tokens, grid recipe, highlight
classes, a11y floor). F-11 copies from it verbatim.

### CI gate commands
Owner: ARCHITECTURE.md §CI/CD topology (exact commands incl. -coverpkg and pinned
tooling). F-02 implements; every piece runs them locally before PR.

## Integration acceptance

| UC | Eval-matrix row | Integration test | Owning feature | Runs at |
|---|---|---|---|---|
| UC-1 | "UC-1 Solve" | Handler-level golden corpus: POST all 55 to /v1/solve, assert solved + oracle-equal + exact shape | F-10 | W6 (F-10 exit) |
| UC-2 | "UC-2 Replay proof" + "UC-2 Determinism" | Replay verifier over all 55 (F-06) + 20 seeded generated, 5/band (F-08 AC-6); handler double-POST byte-compare minus solveTimeMs (F-10) | F-06 (corpus), F-08 (generated slice), F-10 (handler-level) | W4 corpus; W5 generated slice; full at W6 |
| UC-3 | "UC-3 Generate" | Seeded handler test: 25/band generations → oracle-unique, ladder-solved, grade==difficulty; unknown difficulty → 400 envelope | F-08 (package), F-10 (handler) | W5 first; full at W6 |
| UC-4 | "UC-4 Batch" | Full-corpus batch POST under -race: 55 in-order results, solvedCount 55, caps + malformed-line fixtures | F-10 | W6 |
| UC-5 | "UC-5 Parallelism evidence" | -race green on every CI run; committed benchmark shows sequential wins | F-12 | W4 |
| UC-6 | "UC-6 Catalog" | GET /v1/puzzles → 4 canonical sections, 25/10/10/10; drift test embedded==root | F-09 (package+drift), F-10 (handler) | W1 first; full at W6 |

(The UI and deployed-health eval rows are owned by F-11 and F-13 respectively; they
prove the intervention-moment surface and the deployment criterion rather than a UC.)

## High-level acceptance criteria

- **F-01:** both entrypoints serve /v1/health with the frozen header set; a real Vercel
  preview deploy answers 200 with the pinned Go version.
- **F-02:** the five gates block a real PR; repo public; production environment gated.
- **F-03:** 25 ORIGINAL seeds solve via singles only, byte-deterministic.
- **F-04/F-05:** ladder grows in frozen order; all 55 solve at F-05.
- **F-06:** every event of every corpus solve passes the replay proof; containment tests
  green.
- **F-07:** 13/13 fires-and-sound; necessity/sufficiency or recorded evidence.
- **F-08:** 100 seeded generations valid, unique, correctly graded within deadline.
- **F-09:** catalog serves canonical sections; drift guard green.
- **F-10:** full contract-edge matrix green; batch race-clean.
- **F-11:** 11-item UI checklist passes in both layers.
- **F-12:** committed benchmark records the negative result; containment green.
- **F-13:** production deploy through the gate; deployed smoke green.

## Day-Zero prerequisites

- [x] GitHub repo exists with master default (github.com/NerdAlert58/sudoku-flow2)
- [x] gh CLI authed with repo scope (approval of pending deployments needs it)
- [x] Vercel CLI authed (account nerdalert58); project created at first F-01 deploy
- [x] Go ≥1.22 toolchain locally (1.24.1 present)
- [ ] VERCEL_TOKEN for CI: created fresh via API if possible, else CLI token with logged
      deviation (DECISIONS.md D-022) — provisioned during F-02
- [x] Seed corpus committed (puzzles.txt, 55 verified)

## Risk monitoring

| Gate | Highest risk during | In-plan mitigation |
|---|---|---|
| Vercel platform model (AUDIT A1/A2) | F-01 | Preview deploy spike before any dependent work; layout valid under both models |
| Colour-wrap/technique legality (AUDIT L1) | F-05 | Positive-form implementation notes in brief; replay proof + F-07 fixtures |
| Coverage math (AUDIT C2) | F-02 | -coverpkg command frozen; float-safe compare; LCOV conversion pinned |
| Expert generation latency (AUDIT P2) | F-08 | 5s deadline, seeded tests, honest generation_failed |
| Fixture isolation (AUDIT D4) | F-07 | Cap harness; recorded-evidence fallback is the accepted gate |
| Supply chain on public repo (AUDIT S6) | F-02 onward | Pinned actions/tools, contents:read, fork-PR posture |
| govulncheck nondeterminism (AUDIT S5) | every CI run | EVAL.md triage rule: reproduce, read, patch or record |

## Followups

- Enumerate a dashboard origin in the CORS allowlist when the external React dashboard
  exists (future repo).
- Consider SHA-pinning GitHub Actions beyond major tags (advisory, SECURITY.md).
- Revisit no-rate-limiting tradeoff if a deployment ever becomes long-lived.
- These belong in a future `FOLLOWUPS.md` or `STATE.md`.

## Frozen-input contract

This plan is the input to per-piece work. If a piece's allow-list is wrong or an
acceptance criterion is missing a requirement, stop and amend the relevant brief
explicitly (dated entry in the Amendment record). Do not silently expand allow-lists. Do
not silently soften gates. The plan can change; silent drift cannot.

## How to dispatch

> Each `docs/features/NN-*.md` file is a self-contained brief. To build feature F-NN:
>
> 1. Confirm the feature's `Dependencies` are all `Status: done` in their feature files. If not, pick a feature whose dependencies are satisfied.
> 2. If another feature is being built concurrently and its allow-list overlaps yours, create an isolated worktree: `git worktree add ../<project>-fNN -b feature/F-NN`. If allow-lists are disjoint and you are the only builder on this wave, you may skip the worktree.
> 3. Open a fresh Claude Code session at the worktree (or the repo root). Paste:
>
>    > Read `docs/features/NN-<name>.md` and the sections of ARCHITECTURE.md / AUDIT.md / USERS.md / DESIGN_DECISIONS.md it cites under Read-only context. Build the feature. Follow the acceptance criteria. Record decisions in the Implementation notes section as you go. Cross-cutting discoveries propagate to ROADMAP.md or ARCHITECTURE.md, not just the local notes. Open a PR when every acceptance criterion is observably true.
>
> 4. When the feature lands, update its file: change `Status: Not started` to `Status: Done` and append a one-line completion note.
> 5. For programmatic dispatch — picking the ready set, fanning out subagents, or creating worktrees in bulk — run `/nerdflow:build`.

## Amendment record

### Amendment 2026-08-06 — jasnah round-1 blocking findings + observations
Pre-freeze amendment during impl Phase 5b. (1) UC-2's generated-puzzle replay slice
gained an owning AC — F-08 AC-6 (20 seeded, 5/band, through the F-06 verifier); the
integration map's UC-2 "Runs at" now reads W4 corpus / W5 generated slice. (2) F-01 AC-6
now carries the SECURITY §F-10 signal on both platform-model branches
(configure-or-record, never silent platform maximum). Observations adopted: F-02 AC-8
asserts the frozen header set in the deploy smoke; F-01 AC-9 asserts the 80% floor at
skeleton exit; F-06 AC-5 widened to the full C3 sealing claim (only httpapi imports
generate); F-10 AC-3 pins the handler-level N (3/band).

### Amendment <YYYY-MM-DD> — <reason>
(template — real amendments land here during execution)
