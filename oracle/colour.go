package oracle

import "github.com/NerdAlert58/sudoku-flow2/solver"

// colouringHolds: the witnesses form one conjugate-linked component on the
// eliminated digit, two-coloured over its strong links; the eliminations
// are either one whole colour class that sees itself (wrap) or
// non-component cells seeing both colours (trap).
func (v *verifier) colouringHolds(w []int, elims []solver.Elimination) bool {
	d, ok := soleDigit(elims)
	if !ok || len(w) < 2 {
		return false
	}
	colour, ok := v.colourComponent(d, w)
	if !ok {
		return false
	}
	return wrapJustified(w, colour, elims) || trapJustified(w, colour, elims)
}

// colourComponent two-colours the witnesses over conjugate links live in
// the shadow state. ok is false when the witnesses are not one connected
// component or the colouring is inconsistent (an odd conjugate cycle —
// impossible in a state that still has a solution).
func (v *verifier) colourComponent(d int, w []int) ([]int8, bool) {
	colour := make([]int8, len(w))
	for j := range colour {
		colour[j] = -1
	}
	colour[0] = 0
	queue := []int{0}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for j := range w {
			if j == cur || !v.strongLink(w[cur], w[j], d) {
				continue
			}
			if colour[j] == colour[cur] {
				return nil, false
			}
			if colour[j] < 0 {
				colour[j] = 1 - colour[cur]
				queue = append(queue, j)
			}
		}
	}
	for j := range colour {
		if colour[j] < 0 {
			return nil, false
		}
	}
	return colour, true
}

// wrapJustified: the eliminations are exactly one colour class holding two
// cells that see each other — the contradiction that falsifies the class.
func wrapJustified(w []int, colour []int8, elims []solver.Elimination) bool {
	return wrapClass(w, colour, elims, 0) || wrapClass(w, colour, elims, 1)
}

func wrapClass(w []int, colour []int8, elims []solver.Elimination, cls int8) bool {
	var class []int
	for j, i := range w {
		if colour[j] == cls {
			class = append(class, i)
		}
	}
	return selfSeeing(class) && sameCells(elims, class)
}

// selfSeeing: some two cells of the class share a unit.
func selfSeeing(class []int) bool {
	for j, a := range class {
		for _, b := range class[j+1:] {
			if sees(a, b) {
				return true
			}
		}
	}
	return false
}

// sameCells: the elimination cells are exactly the class — no extras, no
// omissions, no duplicates.
func sameCells(elims []solver.Elimination, class []int) bool {
	if len(elims) != len(class) {
		return false
	}
	var pending [81]bool
	for _, i := range class {
		pending[i] = true
	}
	for _, e := range elims {
		i := e.Row*9 + e.Col
		if !pending[i] {
			return false
		}
		pending[i] = false
	}
	return true
}

// trapJustified: every elimination cell is outside the component and sees
// at least one witness of each colour.
func trapJustified(w []int, colour []int8, elims []solver.Elimination) bool {
	for _, e := range elims {
		i := e.Row*9 + e.Col
		if contains(w, i) || !seesBothColours(i, w, colour) {
			return false
		}
	}
	return true
}

func seesBothColours(i int, w []int, colour []int8) bool {
	var seen [2]bool
	for j, c := range w {
		if sees(i, c) {
			seen[colour[j]] = true
		}
	}
	return seen[0] && seen[1]
}
