package solver

const allDigits uint16 = 0b1111111110

func initialCandidates(g *Grid) [81]uint16 {
	var cands [81]uint16
	for i := range g {
		if g[i] == 0 {
			cands[i] = allDigits &^ peerDigits(g, i)
		}
	}
	return cands
}

func peerDigits(g *Grid, i int) uint16 {
	var m uint16
	for _, u := range cellUnitIndexes(i) {
		for _, j := range units[u] {
			m |= 1 << g[j]
		}
	}
	return m
}

// hasCandidate is the single counted accessor of ADR-0007: every
// (cell,digit) membership query made by detection logic goes through it.
func (s *solveState) hasCandidate(i, d int) bool {
	s.checks++
	return s.cands[i]&(1<<d) != 0
}

func (s *solveState) applyPlacement(i, d int) {
	s.grid[i] = byte(d)
	s.cands[i] = 0
	for _, u := range cellUnitIndexes(i) {
		for _, j := range units[u] {
			s.cands[j] &^= 1 << d
		}
	}
}

func (s *solveState) zeroCandidateCell() bool {
	for i := range s.grid {
		if s.grid[i] == 0 && s.cands[i] == 0 {
			return true
		}
	}
	return false
}
