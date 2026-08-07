package solver

type SolveResult struct {
	Status          string
	Solution        Grid
	Events          []Event
	Iterations      int
	EventCount      int
	CandidateChecks int
	Grade           string
}

// solveState carries all per-solve state (ADR-0007: counters are
// per-solve-instance, never package-level).
type solveState struct {
	grid       Grid
	cands      [81]uint16
	events     []Event
	iterations int
	checks     int
	hardest    int
}

func Solve(g Grid) SolveResult {
	s := &solveState{grid: g, cands: initialCandidates(&g), hardest: -1}
	for {
		// Completion is checked before starting a pass, so a complete input
		// runs zero passes (ADR-0014).
		if s.grid.complete() {
			return s.result("solved")
		}
		s.iterations++
		// Priority zero: a zero-candidate cell is unsolvable before any
		// technique — and before stalled can be concluded (ADR-0008).
		if s.zeroCandidateCell() {
			return s.result("unsolvable")
		}
		if !s.runPass() {
			return s.result("stalled")
		}
	}
}

func (s *solveState) place(technique string, i, d int) {
	s.applyPlacement(i, d)
	s.events = append(s.events, Event{
		Seq:          len(s.events) + 1,
		Technique:    technique,
		WitnessCells: []Cell{{Row: i / 9, Col: i % 9}},
		Placement:    &Placement{Row: i / 9, Col: i % 9, Digit: d},
		GridAfter:    s.grid.String(),
	})
}

func (s *solveState) eliminate(technique string, witnesses []Cell, elims []Elimination) {
	for _, e := range elims {
		s.cands[e.Row*9+e.Col] &^= 1 << e.Digit
	}
	s.events = append(s.events, Event{
		Seq:          len(s.events) + 1,
		Technique:    technique,
		WitnessCells: witnesses,
		Eliminations: elims,
		GridAfter:    s.grid.String(),
	})
}

func (s *solveState) result(status string) SolveResult {
	return SolveResult{
		Status:          status,
		Solution:        s.grid,
		Events:          s.events,
		Iterations:      s.iterations,
		EventCount:      len(s.events),
		CandidateChecks: s.checks,
		Grade:           s.grade(status),
	}
}

func (s *solveState) grade(status string) string {
	if status != "solved" {
		return ""
	}
	if s.hardest < 0 {
		// Zero-work complete-grid solve grades at the floor band (ADR-0014).
		return "Easy"
	}
	return ladder[s.hardest].band
}
