package solver

import (
	"math/bits"
	"slices"
)

// Simple colouring (ADR-0007 extension, F-05): digits ascending; conjugate
// edges are units with exactly two candidate cells; components seed from the
// lowest uncoloured row-major linked cell; BFS FIFO with neighbours in
// ascending cell order, seed colour 0, first assignment wins. Per component
// WRAP is checked before TRAP, wrap colour 0 before colour 1.
//
// AUDIT L1 code shape: the component is built from already-true conjugate
// biconditionals and the wrap conclusion is a direct fact-combination — two
// same-colour cells sharing a unit falsify that entire colour class. No
// hypothesis is assigned, propagated, or reverted anywhere.

func (s *solveState) detectSimpleColouring() ([]Cell, []Elimination, bool) {
	for d := 1; d <= 9; d++ {
		adj := s.conjugateAdjacency(d)
		var colour [81]int8
		for i := range colour {
			colour[i] = -1
		}
		for seed := 0; seed < 81; seed++ {
			if colour[seed] >= 0 || len(adj[seed]) == 0 {
				continue
			}
			comp := colourComponent(&adj, seed, &colour)
			if w, e, ok := s.wrapOrTrap(d, comp, &colour); ok {
				return w, e, true
			}
		}
	}
	return nil, nil, false
}

// conjugateAdjacency links the two cells of every unit holding exactly two
// candidates of d; lists are ascending and deduped (two cells can share more
// than one conjugate unit).
func (s *solveState) conjugateAdjacency(d int) [81][]int {
	var adj [81][]int
	for u := 0; u < 27; u++ {
		spots := s.digitSpotsInUnit(u, d)
		if bits.OnesCount16(spots) != 2 {
			continue
		}
		c1 := units[u][bits.TrailingZeros16(spots)]
		c2 := units[u][15-bits.LeadingZeros16(spots)]
		adj[c1] = append(adj[c1], c2)
		adj[c2] = append(adj[c2], c1)
	}
	for i := range adj {
		slices.Sort(adj[i])
		adj[i] = slices.Compact(adj[i])
	}
	return adj
}

func colourComponent(adj *[81][]int, seed int, colour *[81]int8) []int {
	colour[seed] = 0
	comp := []int{seed}
	queue := []int{seed}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, nb := range adj[cur] {
			if colour[nb] >= 0 {
				continue
			}
			colour[nb] = 1 - colour[cur]
			comp = append(comp, nb)
			queue = append(queue, nb)
		}
	}
	slices.Sort(comp)
	return comp
}

func (s *solveState) wrapOrTrap(d int, comp []int, colour *[81]int8) ([]Cell, []Elimination, bool) {
	for _, cls := range [2]int8{0, 1} {
		if classSharesUnit(comp, colour, cls) {
			return sortedCells(comp...), classElims(comp, colour, cls, d), true
		}
	}
	e := s.trapElims(d, comp, colour)
	if len(e) == 0 {
		return nil, nil, false
	}
	return sortedCells(comp...), e, true
}

func classSharesUnit(comp []int, colour *[81]int8, cls int8) bool {
	for i, a := range comp {
		if colour[a] != cls {
			continue
		}
		for _, b := range comp[i+1:] {
			if colour[b] == cls && sees(a, b) {
				return true
			}
		}
	}
	return false
}

func classElims(comp []int, colour *[81]int8, cls int8, d int) []Elimination {
	var es []Elimination
	for _, i := range comp {
		if colour[i] == cls {
			es = append(es, Elimination{Row: i / 9, Col: i % 9, Digit: d})
		}
	}
	return es
}

func (s *solveState) trapElims(d int, comp []int, colour *[81]int8) []Elimination {
	var es []Elimination
	for i := 0; i < 81; i++ {
		if s.grid[i] != 0 || slices.Contains(comp, i) || !seesBothColours(i, comp, colour) {
			continue
		}
		if s.hasCandidate(i, d) {
			es = append(es, Elimination{Row: i / 9, Col: i % 9, Digit: d})
		}
	}
	return es
}

func seesBothColours(i int, comp []int, colour *[81]int8) bool {
	var seen [2]bool
	for _, c := range comp {
		if sees(i, c) {
			seen[colour[c]] = true
		}
	}
	return seen[0] && seen[1]
}
