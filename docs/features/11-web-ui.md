# Feature: Embedded web UI

**ID:** F-11 · **Roadmap piece:** P-11 · **Status:** Not started

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
None.

## Manual setup required
Visual smoke requires a human-equivalent browser pass — performed by the operator agent
with the Chrome tooling and recorded per checklist item.

## Implementation notes (filled in by the building agent)
> Decisions, the per-item smoke record, and any deviations land here.
