# F-07 per-technique evidence record (AC-4)

Companion to `solver/technique_fixtures_test.go` and `solver/laddercap_test.go`.
Every claim marked (measured) was produced by the ladder-cap search described
below, run against baseline 12e3933; the committed tests re-verify all fixture
properties on every suite run.

## Status table

| Pos | Technique | Fires | Sound | Necessity | Sufficiency | Status |
|-----|-----------|-------|-------|-----------|-------------|--------|
| 1 | naked_single | Y | Y | Y | Y | proven |
| 2 | hidden_single | Y | Y | Y | Y | proven |
| 3 | locked_candidates_pointing | Y | Y | Y | Y | proven |
| 4 | locked_candidates_claiming | Y | Y | Y | Y | proven (curated: seed s1p5 + r2c7=9) |
| 5 | naked_subset | Y | Y | Y | Y | proven |
| 6 | hidden_subset | Y | Y | Y (at fixture) | N | fires-and-sound + evidence |
| 7 | x_wing | Y | Y | Y | Y | proven |
| 8 | swordfish | Y | Y | Y (at fixture) | N | fires-and-sound + evidence |
| 9 | jellyfish | Y | Y | Y (at fixture) | N | fires-and-sound + evidence (PRD pre-flagged) |
| 10 | xy_wing | Y | Y | Y | Y | proven |
| 11 | xyz_wing | Y | Y | Y (at fixture) | N | fires-and-sound + evidence |
| 12 | w_wing | Y | Y | Y | Y | proven |
| 13 | simple_colouring | Y | Y | Y | Y | proven |

Column meanings:

- **Fires**: the technique is `Events[0]` of a full solve from its committed
  fixture (`TestPerTechnique_FiresAndVerifierSound`).
- **Sound**: the entire solve from the fires fixture passes
  `oracle.ReplayVerify` (nil), which exercises the technique's witness checker
  positively; the opening event's eliminations are additionally proven by
  exact brute force (placing an eliminated digit admits zero completions) —
  this closes the oracle-truth gap for the six fires fixtures that admit two
  completions, where ReplayVerify skips its oracle-anchored checks.
- **Necessity / Sufficiency**: one committed puzzle per proven technique with
  `SolveCapped(g, pos-1)` stalled and `SolveCapped(g, pos)` solved
  (`TestPerTechnique_NecessityAndSufficiency`). "Y (at fixture)" means the
  weaker half only: the fires fixture stalls when capped below the technique
  (nothing cheaper acts), pinned by `TestPerTechnique_FallbackFixtureStatus`
  together with the sufficiency failure.

## What was searched (all measured)

State pool: 55 corpus seeds + 15 committed fixture states + every re-parsed
intermediate state (`gridAfter`) of their full-solve trajectories = 3,190
distinct states. For each state, `SolveCapped` status at every cap 0-13
(44,660 capped solves). Isolation candidates (capped-below stalled AND
capped-at solved) per position:

- p=1: 2,241 · p=2: 156 · p=3: 116 · p=5: 87 · p=7: 24 · p=10: 258 ·
  p=12: 160 · p=13: 49 — proven fixtures picked from these.
- p=4, p=6, p=8, p=9, p=11: **0** natural candidates.

Secondary searches for the five zero-candidate positions:

1. **Distance-1 clear search**: over every pool state that solves at cap p
   AND fires the technique under that cap — such start states number **zero**
   for all five techniques (every corpus firing of these techniques happens
   only after a more expensive technique already acted, so capped trajectories
   stall before reaching them; capped trajectories are prefixes of full
   trajectories by determinism).
2. **Greedy fill hill-climb** from each technique's fixture (fill one empty
   cell per step from an oracle completion, keeping capped-below stalled,
   minimizing empties at the capped-at stall): dead ends at 4-20 remaining
   empties for all five.
3. **Beam search** (widths 12 and 16, depth <= 81, plateau moves allowed,
   both oracle completions where the state admits two, starts = every pool
   state whose cap-p run fires the technique plus clear-1 fixture variants,
   up to 20 starts): **locked_candidates_claiming isolated** at depth 1
   (corpus MEDIUM seed s1p5 with r2c7=9 filled from its solution — now
   `gridClaimingIso`, upgraded to proven). Nodes explored for the rest:
   hidden_subset 75,500 (search space exhausted), swordfish 158,350
   (exhausted), jellyfish 200,000 (budget cap), xyz_wing 200,000 (budget cap).
   No isolation found.

## Why isolation is believed unattainable (reasoning, not measurement)

- **jellyfish** (PRD pre-flagged): a rows-base fish always coexists with a
  cols-base complement covering the same eliminations; whenever the
  complement is smaller it fires earlier as x_wing/swordfish, so jellyfish
  can only act where the complement is also size >= 4 — and such states, when
  otherwise cap-9-solvable, did not occur in 3,190 states plus 200,000
  curated variants. Fired exactly once in all pool full solves (its own
  fixture).
- **swordfish**: the same fish-complement redundancy one size down. All 16
  pool firings occur strictly after a >cap-8 technique acted.
- **hidden_subset**: in any unit a hidden k-subset coexists with a naked
  (free-cells - k)-subset; the naked twin fires first (position 5) whenever
  it is small enough to be enumerated productively. Fired exactly once in all
  pool full solves (its own fixture).
- **xyz_wing**: every state that fires it (57 in full solves, 40 under
  cap-11) also needs w_wing/simple_colouring or beyond-ladder moves to
  finish; sufficiency never materialized in any curated variant.

If a future solver or corpus change makes any of the four solve capped-at,
`TestPerTechnique_FallbackFixtureStatus` fails loudly: upgrade the technique
to proven here and commit its isolation fixture.
