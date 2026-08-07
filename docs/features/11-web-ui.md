# Feature: Embedded web UI

**ID:** F-11 · **Roadmap piece:** P-11 · **Status:** Done (2026-08-07) — DOM contract 15/15 green; visual smoke 11/11 (pre+post refactor); verifier PASS + re-check PASS (tuned); readability PASS r2 with recorded file-length exemption; leanness applied

## Description
The full embedded SPA replacing F-01's placeholder: the 81-cell grid hero, seed dropdown
grouped by tier, manual entry + paste + Clear, solve with status/grade/quartet display,
the ladder-ordered statistics panel, the step-through viewer with three highlight states
and full transport controls, the clickable event log, and the technique-explanation
panel — all dependency-free, strict-CSP-compatible, built to the frozen design language.

## How it fits the roadmap
W7, after the API is complete (its acceptance needs live solves). On the critical path's
tail.

## Dependencies (must exist before this starts)
- F-10 http-contract — live /v1/puzzles and /v1/solve for the solve flow and smoke

## Unblocks (what waits on this)
- F-13 ship — the UI it deploys

## Allow-list (source)
- web/web.go
- web/index.html
- web/app.css
- web/app.js

## Allow-list (tests)
- web/*_test.go

## Read-only context
- ARCHITECTURE.md §Frontend Design Language (the surface bullet — kit = the PRD itself;
  tokens, grid gap-as-gridlines recipe, highlight classes `.placement`/`.witness`/
  `.elimination`, a11y floor — copy recipe carries verbatim)
- ARCHITECTURE.md §Contracts C1, C5
- AUDIT.md S1 (CSP constraints on asset construction)
- DESIGN_DECISIONS.md ADR-0012
- USERS.md §The Workflow, UC-2, UC-6
- EVAL.md row "UI (SPA)"
- PRD.md §In scope (UI bullet — the feature list of record)

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
None.

## Acceptance criteria
- **AC-1 (DOM-contract layer):** A Go test over the embedded assets asserts: zero inline
  `<script>`/`<style>`/`style=`/`on*=` attributes; no `innerHTML` or
  `insertAdjacentHTML` or `document.write` in app.js; external `app.js`/`app.css`
  referenced from index.html; required element hooks exist for all 11 checklist
  features (grid container, 81 cell inputs pattern, dropdown with optgroups, Clear/
  Solve/transport buttons, status region with aria-live, histogram container,
  event-log list, explanation panel, band chip, step position indicator). Eval row:
  "UI (SPA)" layer (a).
- **AC-2 (visual smoke layer):** The operator (or designated stand-in) loads `/` against
  a locally running server in a real browser and walks the 11-item checklist from
  EVAL.md's "UI (SPA)" row on a Medium seed — every item observably works, including:
  paste of an 81-char string populates the grid; Play auto-advances and stops at the
  last step; clicking any event-log row jumps the grid; the explanation panel covers
  every ladder technique with the band chip updating; highlight states are visually
  distinguishable. Result recorded per-item in Implementation notes. Eval row: "UI
  (SPA)" layer (b).
- **AC-3:** The visual language matches the frozen tokens: system-ui, the exact CSS
  custom properties from ARCHITECTURE.md §Frontend Design Language, gap-as-gridlines
  symmetric borders (1px cells, 3px box boundaries), single blue accent for chrome.
- **AC-4:** The a11y floor holds: per-cell `aria-label="Row R, Column C"`, native
  select/optgroup and buttons, `aria-live="polite"` status region, DOM order ==
  visual order, native focus rings not suppressed.
- **AC-5:** The UI performs zero client-side solving: the step viewer renders purely
  from `events[].gridAfter` and event fields (verified by AC-1's static assertions —
  no solver-shaped code in app.js — plus smoke behavior with the network tab).

## Testing requirements
The DOM-contract Go test (asset grep/parse assertions) plus the operator visual smoke.
No JS test framework (dependency-free rule); behavioral verification is the smoke
layer.

## Test command
(inherit from CONTEXT.md §Test discipline)

## Coverage command
(inherit)

## Coverage report
(inherit)

## Test-exempt lines
- web/app.js:L1-L9999 — browser-only vanilla JS; verified by the DOM-contract static
  assertions + the mandatory visual-smoke layer (no JS unit framework by the
  dependency-free rule)

## Health check
N/A (assets ship inside the binary; deploy surface is F-13's)

## Rollback command
N/A (no deploy in this piece)

## Env vars required
None.

## Readability budget
(inherit from CONTEXT.md §Rigor)

## Complexity exemptions
- web/app.js:<file> — max_file_lines:400 exceeded (528 post-decomposition; 513 pre) — the frozen
  Allow-list (source) pins exactly three asset filenames (index.html, app.css, app.js),
  so file-splitting is structurally unavailable; ~76 lines are contract-pinned data
  tables (13-technique NAMES/BANDS/EXPLAIN maps) and ARCHITECTURE.md §Known Tradeoffs
  explicitly concedes the vanilla-JS cost ("several hundred lines… accepted by the PRD").
  Function-level budgets remain fully enforced (no function exemption granted). —
  authorized by coordinator under autonomous mandate (DECISIONS.md D-031), 2026-08-07

## Manual setup required
Visual smoke requires a human-equivalent browser pass — performed by the operator agent
with the Chrome tooling and recorded per checklist item.

## Implementation notes (filled in by the building agent)

### Visual smoke record (coordinator, real Chrome on macOS, 2026-08-07)

Server: PORT=8123 local binary. All 11 checklist items PASS:
1. Manual entry: digit typed into R1C2 renders; focus auto-advances; native focus ring.
2. Paste: typing the 81-char ORIGINAL #1 string live-populates the grid (verified
   cell-by-cell against the corpus).
3. Clear: full reset (grid, dropdown, metrics, step counter, explanation → hint).
4. Seed dropdown: 4 tier optgroups, 55 options from /v1/puzzles; selecting Medium #1
   loaded its givens (verified against corpus string).
5. Solve: Status solved · Grade Medium; quartet 58 iterations / 58 events / 14831
   candidateChecks / 0.08 ms (iterations==events invariant visible live).
6. Statistics: ladder-ordered histogram (Naked single 36, Hidden single 20, Pointing 2),
   difficulty Medium, accent bars.
7. Highlights: all three states simultaneously distinct at step 29 — witness (tint +
   solid accent border), elimination (dashed border), placement (solid accent fill,
   white digit); givens black vs placed blue.
8. Transport: First/Prev/Play/Next/Last; Play flips to Pause and auto-advances; auto-
   stops at 58/58 flipping back to Play; Next/Last disabled at end; step position +
   per-step description correct throughout.
9. Event-log click (#29 row) jumps the grid/viewer to that step.
10. Explanation panel prose swaps per step; band chip flipped Easy→Medium correctly;
    static pre-solve hint present before solving.
11. A11y floor: native select/buttons, aria-live status region, focus rings intact.
Console: zero errors/exceptions during the entire session.

### Post-refactor re-smoke (coordinator, real Chrome, 2026-08-07 — after renderStep decomposition)
Seed load (Medium #1), Solve (solved · Medium · 58/58/14831/0.05ms), step navigation
(Prev → 57/58 with placement highlight moving R9C9→R9C8, panels/log/transport correct),
all through the decomposed paintGrid/applyHighlights/renderPanels/syncLog path —
behavior identical to the pre-refactor smoke. Closes the verifier re-check's
RUBRIC_GAP-1 staleness condition.

> Decisions, the per-item smoke record, and any deviations land here.

**Build (2026-08-07, feature/f-11).** Replaced the F-01 placeholders in `web/index.html`,
`web/app.css`, `web/app.js`; `web/web.go` unchanged (already exports `FS` over the three
assets).

Decisions:
- **Gridline recipe:** container `gap: 1px` + `background: var(--border)` + literal
  `border: 3px` outer frame; internal box boundaries are 1px gap + 2px same-color
  `border-left`/`border-top` on the 4th/7th column/row cells via `:nth-child` — reads as
  a contiguous 3px line, keeps row-major DOM order and auto-placement (no spacer nodes).
- **Solution field never read.** After any solve the viewer jumps to
  `renderStep(events.length)`, which renders exclusively from `input` (step 0) and
  `events[i-1].gridAfter` — so stalled/unsolvable partial `solution` grids can never be
  auto-filled by construction (AC-5). For solved puzzles this shows the final grid
  immediately with the transport ready to rewind.
- **invalid_input:** a 400 whose body carries `status:"invalid_input"` renders the full
  solve shape honestly (status line, zero-valued quartet); other non-200s render the
  error envelope. Step viewer stays reset for invalid input.
- **Bands/prose:** technique→band map copied from `solver/ladder.go` (Easy: singles;
  Medium: locked candidates + subsets; Hard: fish + XY-wing; Expert: XYZ/W-wing +
  simple colouring); 13-entry plain-English `EXPLAIN` map, 1–2 sentences each.
- **Givens vs placed** (optional per brief): implemented — `.given` bold, `.filled` in
  accent, derived per step from the input grid.
- Editing any cell resets solve state (stale status/steps never shown against an edited
  grid). Paste fills on any 81-char `[0-9.]` input; whitespace stripped.

Layer (a) — DOM contract: all 15 tests in `web/dom_contract_test.go` green
(`go test -race -count=1 ./...` fully green; gofmt/vet/build clean).

Functional self-check (live server, PORT=8123): `/`, `/app.js`, `/app.css` → 200 with
correct content types; `/v1/puzzles` → 4 sections (Original/Medium/Hard/Very Hard);
Medium seed solve → 200 `solved`/`Medium`, 58 events, 81-char `gridAfter`, 0-based
row/col in placements/witnesses/eliminations; `{"puzzle":"bad"}` → 400 full solve shape
with quartet present.

coordinator's browser pass; per-item results to be recorded here.
