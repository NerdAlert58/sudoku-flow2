package catalog

import (
	"fmt"
	"strings"
)

// AUDIT D2: headers are section boundaries only; names come from ordinal position.
var sectionNames = [4]string{"Original", "Medium", "Hard", "Very Hard"}

func parseCatalog(data []byte) ([]Section, error) {
	var groups [][]string
	for i, line := range strings.Split(string(data), "\n") {
		switch {
		case line == "":
		case strings.HasPrefix(line, "#"):
			groups = append(groups, nil)
		case !isPuzzleLine(line):
			return nil, fmt.Errorf("line %d: %q does not match ^[0-9]{81}$", i+1, line)
		case len(groups) == 0:
			return nil, fmt.Errorf("line %d: puzzle before first section header", i+1)
		default:
			groups[len(groups)-1] = append(groups[len(groups)-1], line)
		}
	}
	if len(groups) != len(sectionNames) {
		return nil, fmt.Errorf("catalog has %d sections; want %d", len(groups), len(sectionNames))
	}
	secs := make([]Section, len(groups))
	for i, g := range groups {
		secs[i] = Section{Name: sectionNames[i], Puzzles: g}
	}
	return secs, nil
}

func isPuzzleLine(s string) bool {
	if len(s) != 81 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
