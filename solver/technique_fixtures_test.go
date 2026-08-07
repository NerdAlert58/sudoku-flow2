package solver_test

import (
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/oracle"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// F-07 per-technique evidence (EVAL row "Per-technique"; AUDIT.md D4).
// Isolation grids were curated by the ladder-cap search recorded in
// testdata/techniques/EVIDENCE.md; every property asserted here re-verifies
// them from scratch on each run.
const (
	// Corpus ORIGINAL seed s0p0: naked singles alone solve it.
	gridNakedSingleIso = "700605000000000030509300024002000000401907052000501000004050000310492000007003000"
	// State after 24 events of corpus MEDIUM seed s1p2 (unique-solution):
	// no naked single exists, a hidden single opens, singles finish it.
	gridHiddenSingleIso = "589710000364850107721340085950127340430508071070403050600905700007600000890270060"
	// Corpus MEDIUM seed s1p0.
	gridPointingIso = "000000600520040100070062003092000000000000001437029000803094002000000068000030000"
	// Corpus MEDIUM seed s1p5 with r2c7=9 filled from its solution
	// (beam-search curated; the bare seed needs upper-ladder techniques).
	gridClaimingIso = "000050008709000000046003090004020000900600800600004010000100000502000300100090670"
	// Corpus MEDIUM seed s1p1.
	gridNakedSubsetIso = "010370000084901030000000000640000900002100000000000205026509100500000042701000000"
	// Corpus HARD seed s2p0.
	gridXWingIso = "809600000004000061010400700403080000080506002001020070000000030150000900000807005"
	// Corpus HARD seed s2p1.
	gridXYWingIso = "000002003004800050002010000000060000000940760507003000000270004001000006000085072"
	// Corpus VERY-HARD seed s3p0.
	gridWWingIso = "007000008400062001091040000009000402800200000000016000010000706000090053000350800"
	// Corpus VERY-HARD seed s3p1.
	gridSimpleColouringIso = "070090200000006004000002010021079000450080006000000100017900850004050000000000069"
)

// fires: committed state whose full solve opens with the technique (the
// F-04/F-05 fixture constants for positions 3-13). iso: committed puzzle
// proving necessity (capped below stalls) and sufficiency (capped at
// solves); empty = isolation unattainable after the recorded attempt
// (AC-4, testdata/techniques/EVIDENCE.md).
var perTechnique = []struct {
	pos   int
	name  string
	fires string
	iso   string
}{
	{1, "naked_single", gridNakedSingleIso, gridNakedSingleIso},
	{2, "hidden_single", gridHiddenSingleIso, gridHiddenSingleIso},
	{3, "locked_candidates_pointing", gridPointing, gridPointingIso},
	{4, "locked_candidates_claiming", gridClaiming, gridClaimingIso},
	{5, "naked_subset", gridNaked, gridNakedSubsetIso},
	{6, "hidden_subset", gridHidden, ""},
	{7, "x_wing", gridXWing, gridXWingIso},
	{8, "swordfish", gridSwordfish, ""},
	{9, "jellyfish", gridJellyfish, ""},
	{10, "xy_wing", gridXYWing, gridXYWingIso},
	{11, "xyz_wing", gridXYZWing, ""},
	{12, "w_wing", gridWWing, gridWWingIso},
	{13, "simple_colouring", gridSCTrap, gridSimpleColouringIso},
}

// AC-1: each technique fires on its committed fixture and the whole solve
// passes the F-06 verifier. ReplayVerify skips oracle-truth checks when the
// state is not unique-solution (six fixture states admit two completions),
// so the opening event's eliminations are additionally proven sound by
// exact brute force: placing an eliminated digit must admit zero
// completions.
func TestPerTechnique_FiresAndVerifierSound(t *testing.T) {
	for _, tc := range perTechnique {
		t.Run(tc.name, func(t *testing.T) {
			g := mustParse(t, tc.fires)
			res := solver.Solve(g)
			if len(res.Events) == 0 {
				t.Fatalf("no events (status %q): %s must fire", res.Status, tc.name)
			}
			if got := res.Events[0].Technique; got != tc.name {
				t.Fatalf("Events[0].Technique = %q, want %q", got, tc.name)
			}
			if err := oracle.ReplayVerify(g, res); err != nil {
				t.Fatalf("ReplayVerify: %v", err)
			}
			if tc.pos <= 2 {
				if res.Events[0].Placement == nil {
					t.Fatal("singles fixture must open with a placement")
				}
				if _, count := oracle.Solve(g); count != 1 {
					t.Fatalf("oracle count = %d, want 1: the singles evidence relies on the verifier's oracle anchoring", count)
				}
			}
			for _, e := range res.Events[0].Eliminations {
				i := e.Row*9 + e.Col
				mutated := tc.fires[:i] + string('0'+byte(e.Digit)) + tc.fires[i+1:]
				if _, count := oracle.Solve(mustParse(t, mutated)); count != 0 {
					t.Errorf("elimination %+v unsound: %d completions place that digit", e, count)
				}
			}
		})
	}
}

// AC-3: for every technique where curation succeeded, one committed puzzle
// carries both halves - capped below the technique the solve stalls
// (necessity), capped at it the whole puzzle solves (sufficiency) - and the
// capped solve demonstrably fires the technique.
func TestPerTechnique_NecessityAndSufficiency(t *testing.T) {
	proven := 0
	for _, tc := range perTechnique {
		if tc.iso == "" {
			continue
		}
		proven++
		t.Run(tc.name, func(t *testing.T) {
			g := mustParse(t, tc.iso)
			if st := solver.SolveCapped(g, tc.pos-1).Status; st != "stalled" {
				t.Errorf("necessity: capped at %d Status = %q, want stalled", tc.pos-1, st)
			}
			at := solver.SolveCapped(g, tc.pos)
			if at.Status != "solved" {
				t.Fatalf("sufficiency: capped at %d Status = %q, want solved", tc.pos, at.Status)
			}
			fired := false
			for _, ev := range at.Events {
				if ev.Technique == tc.name {
					fired = true
					break
				}
			}
			if !fired {
				t.Errorf("capped-at-%d solve never fired %s", tc.pos, tc.name)
			}
			assertRuleConformantSolution(t, tc.iso, at.Solution.String())
		})
	}
	if proven != 9 {
		t.Fatalf("proven techniques = %d, want 9 per the EVIDENCE.md status table", proven)
	}
}

// AC-4: the four fallback techniques hold necessity at their fires fixture
// (capped below, nothing acts) while sufficiency demonstrably fails (capped
// at, the solve still stalls). If a solver change ever makes one of these
// solve, this test flags it: upgrade the technique to proven and retire its
// EVIDENCE.md fallback entry.
func TestPerTechnique_FallbackFixtureStatus(t *testing.T) {
	fallback := 0
	for _, tc := range perTechnique {
		if tc.iso != "" {
			continue
		}
		fallback++
		t.Run(tc.name, func(t *testing.T) {
			g := mustParse(t, tc.fires)
			if st := solver.SolveCapped(g, tc.pos-1).Status; st != "stalled" {
				t.Errorf("capped at %d Status = %q, want stalled: a cheaper technique acts on the fixture state", tc.pos-1, st)
			}
			if st := solver.SolveCapped(g, tc.pos).Status; st != "stalled" {
				t.Errorf("capped at %d Status = %q, want stalled: sufficiency now holds - upgrade %s to proven in testdata/techniques/EVIDENCE.md", tc.pos, st, tc.name)
			}
		})
	}
	if fallback != 4 {
		t.Fatalf("fallback techniques = %d, want 4 per the EVIDENCE.md status table", fallback)
	}
}
