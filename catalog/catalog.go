package catalog

import (
	_ "embed"
	"sync"
)

//go:embed puzzles.txt
var raw []byte

type Section struct {
	Name    string   `json:"name"`
	Puzzles []string `json:"puzzles"`
}

var sections = sync.OnceValue(func() []Section { return mustParse(raw) })

func mustParse(data []byte) []Section {
	secs, err := parseCatalog(data)
	if err != nil {
		panic("catalog: malformed catalog data: " + err.Error())
	}
	return secs
}

func Sections() []Section { return sections() }

func Raw() []byte { return raw }
