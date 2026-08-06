package solver

func (s *solveState) fireNakedSingle() bool {
	for i := 0; i < 81; i++ {
		if s.grid[i] != 0 {
			continue
		}
		if d, ok := s.soleCandidate(i); ok {
			s.place("naked_single", i, d)
			return true
		}
	}
	return false
}

func (s *solveState) soleCandidate(i int) (int, bool) {
	count, digit := 0, 0
	for d := 1; d <= 9; d++ {
		if s.hasCandidate(i, d) {
			count++
			digit = d
		}
	}
	return digit, count == 1
}

func (s *solveState) fireHiddenSingle() bool {
	for u := range units {
		for d := 1; d <= 9; d++ {
			if i, ok := s.solePlaceInUnit(u, d); ok {
				s.place("hidden_single", i, d)
				return true
			}
		}
	}
	return false
}

func (s *solveState) solePlaceInUnit(u, d int) (int, bool) {
	count, at := 0, -1
	for _, i := range units[u] {
		if s.grid[i] != 0 {
			continue
		}
		if s.hasCandidate(i, d) {
			count++
			at = i
		}
	}
	return at, count == 1
}
