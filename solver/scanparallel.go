package solver

import "sync"

// SolveScanParallel is the UC-5 scan-parallel experiment (ADR-0015: reachable
// only by explicit call; the committed benchmark is its sole caller). It
// parallelizes detection scanning within a pass, never semantics: every ladder
// detector probes a private snapshot concurrently, then the lowest-rung firing
// technique is re-fired on the real state, so the committed event stream is
// byte-identical to Solve's. Only CandidateChecks may differ — probe counters
// die with their goroutines and the real state counts only committed fires.
func SolveScanParallel(g Grid) SolveResult {
	s := &solveState{grid: g, cands: initialCandidates(&g), hardest: -1}
	for {
		if s.grid.complete() {
			return s.result("solved")
		}
		s.iterations++
		if s.zeroCandidateCell() {
			return s.result("unsolvable")
		}
		if !s.runPassScanParallel() {
			return s.result("stalled")
		}
	}
}

func (s *solveState) runPassScanParallel() bool {
	fired := make([]bool, len(ladder))
	var wg sync.WaitGroup
	for i := range ladder {
		wg.Add(1)
		go func() {
			defer wg.Done()
			probe := solveState{grid: s.grid, cands: s.cands}
			fired[i] = ladder[i].fire(&probe)
		}()
	}
	wg.Wait()
	for i, ok := range fired {
		if ok {
			ladder[i].fire(s)
			if i > s.hardest {
				s.hardest = i
			}
			return true
		}
	}
	return false
}
