package solver

type technique struct {
	band string
	fire func(*solveState) bool
}

// ladder is the ordered technique registry (PRD §Domain context). Each fire
// func detects the first canonical instance, applies it, and emits one event.
// F-04/F-05 append entries here without touching the solve loop.
var ladder = []technique{
	{band: "Easy", fire: (*solveState).fireNakedSingle},
	{band: "Easy", fire: (*solveState).fireHiddenSingle},
}

func (s *solveState) runPass() bool {
	for i := range ladder {
		if ladder[i].fire(s) {
			if i > s.hardest {
				s.hardest = i
			}
			return true
		}
	}
	return false
}
