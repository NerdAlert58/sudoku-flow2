package solver

// SolveCapped is the F-07 ladder-cap harness (test-only, EVAL row
// "Per-technique"): Solve's loop with only the first maxPos ladder
// techniques enabled. maxPos = len(ladder) reproduces Solve exactly;
// maxPos = 0 disables every technique.
func SolveCapped(g Grid, maxPos int) SolveResult {
	if maxPos > len(ladder) {
		maxPos = len(ladder)
	}
	s := &solveState{grid: g, cands: initialCandidates(&g), hardest: -1}
	for {
		if s.grid.complete() {
			return s.result("solved")
		}
		s.iterations++
		if s.zeroCandidateCell() {
			return s.result("unsolvable")
		}
		fired := false
		for i := 0; i < maxPos; i++ {
			if ladder[i].fire(s) {
				if i > s.hardest {
					s.hardest = i
				}
				fired = true
				break
			}
		}
		if !fired {
			return s.result("stalled")
		}
	}
}

// LadderSize exposes len(ladder) so cap tests cannot drift from the registry.
func LadderSize() int { return len(ladder) }
