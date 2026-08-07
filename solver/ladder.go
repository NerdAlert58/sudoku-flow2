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
	elimination("locked_candidates_pointing", "Medium", (*solveState).detectPointing),
	elimination("locked_candidates_claiming", "Medium", (*solveState).detectClaiming),
	elimination("naked_subset", "Medium", (*solveState).detectNakedSubset),
	elimination("hidden_subset", "Medium", (*solveState).detectHiddenSubset),
	elimination("x_wing", "Hard", (*solveState).detectXWing),
	elimination("swordfish", "Hard", (*solveState).detectSwordfish),
	elimination("jellyfish", "Hard", (*solveState).detectJellyfish),
	elimination("xy_wing", "Hard", (*solveState).detectXYWing),
	elimination("xyz_wing", "Expert", (*solveState).detectXYZWing),
	elimination("w_wing", "Expert", (*solveState).detectWWing),
	elimination("simple_colouring", "Expert", (*solveState).detectSimpleColouring),
}

// elimination wires a detection func's frozen event string once, here at the
// registry. Detections return ok only for instances with non-empty
// eliminations, so an unproductive pattern never fires (ADR-0007).
func elimination(name, band string, detect func(*solveState) ([]Cell, []Elimination, bool)) technique {
	return technique{band: band, fire: func(s *solveState) bool {
		witnesses, elims, ok := detect(s)
		if ok {
			s.eliminate(name, witnesses, elims)
		}
		return ok
	}}
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
