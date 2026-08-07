# Feature: Catalog — embedded seed corpus

**ID:** F-09 · **Roadmap piece:** P-09 · **Status:** Not started

## Description
The `catalog` package: an embedded copy of `puzzles.txt`, section parsing (headers as
boundaries, skip `#`/blank lines), ordinal mapping to the four canonical display names,
and the drift-guard test byte-comparing the embedded copy against the repo-root source
of truth.

## How it fits the roadmap
W1, parallel with F-02 and F-03 (disjoint surfaces). Feeds F-10's /v1/puzzles handler.

## Dependencies (must exist before this starts)
- F-01 walking-skeleton — go.mod (module root)

## Unblocks (what waits on this)
- F-10 http-contract — the /v1/puzzles handler consumes catalog.Sections

## Allow-list (source)
- catalog/*.go (non-test files)
- catalog/puzzles.txt

## Allow-list (tests)
- catalog/*_test.go

## Read-only context
- ARCHITECTURE.md §Contracts C4, §Components (catalog)
- AUDIT.md A3, D1, D2, D3
- USERS.md UC-6
- EVAL.md row "UC-6 Catalog"

## Compliance requirements
None — COMPLIANCE.md declares `Applicable hats: N/A`.

## CI/CD requirements
None.

## Acceptance criteria
- **AC-1:** `catalog.Sections()` returns exactly four sections named `Original`,
  `Medium`, `Hard`, `Very Hard` (ordinal mapping — never the literal header text) with
  25/10/10/10 well-formed 81-char puzzles in file order. Eval row: "UC-6 Catalog"
  (package half; the handler half lands in F-10).
- **AC-2:** The drift-guard test byte-compares `catalog/puzzles.txt` against the
  repo-root `puzzles.txt` and fails on any difference.
- **AC-3:** No section name is asserted to equal a solver grade anywhere (AUDIT D3 —
  provenance labels, not grade claims).
- **AC-4:** A malformed embedded file (fixture-injected via test seam) panics at
  init/first use — startup defect, never a silent runtime condition.

## Testing requirements
Parsing tests, drift test, count/name assertions, malformed-input behavior.

## Test command
(inherit from CONTEXT.md §Test discipline)

## Coverage command
(inherit)

## Coverage report
(inherit)

## Test-exempt lines
None.

## Health check
N/A (library piece)

## Rollback command
N/A (library piece)

## Env vars required
None.

## Readability budget
(inherit from CONTEXT.md §Rigor)

## Complexity exemptions
None.

## Manual setup required
None.

## Implementation notes (filled in by the building agent)
> Decisions and rationale land here as the piece builds.

- Two files by concept: `catalog/catalog.go` (embed, `Section`, `Sections()`, `Raw()`,
  once-parse) and `catalog/parse.go` (`parseCatalog`, `isPuzzleLine`), pairing the
  pre-existing `catalog_test.go` / `parse_test.go` split.
- Parse-once via `sync.OnceValue` (Go 1.21+ stdlib) rather than `init()`: malformed
  embedded data panics at first `Sections()` call (AC-4's "init/first use"), and
  `parseCatalog` stays independently testable against fixtures without tripping the
  package on import.
- Any `#`-prefixed line is a section boundary (AUDIT D2: names from ordinal position
  `Original/Medium/Hard/Very Hard`, never header text). Blank lines skipped. Errors:
  section count != 4 (also covers empty input — zero sections), any non-blank/non-#
  line failing the 81-digit check, or a puzzle line before the first header.
- 81-digit check is a hand-rolled byte loop (`len==81`, `'0'..'9'`) instead of
  `regexp` — same contract as `^[0-9]{81}$`, no package-level regexp state.
- `catalog/puzzles.txt` copied byte-for-byte from repo-root `puzzles.txt`
  (`cp -p`, `cmp`-verified) per AUDIT A3 (`go:embed` cannot reach parent dirs);
  drift-guard test enforces it stays identical.
- Sections/puzzle-count validation (25/10/10/10) is intentionally NOT in
  `parseCatalog` — the pinned error contract lists section count, line grammar, and
  empty input only; counts are asserted by tests against the embedded corpus (AC-1).
