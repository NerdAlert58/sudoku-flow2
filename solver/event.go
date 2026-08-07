package solver

type Event struct {
	Seq          int           `json:"seq"`
	Technique    string        `json:"technique"`
	WitnessCells []Cell        `json:"witnessCells"`
	Placement    *Placement    `json:"placement,omitempty"`
	Eliminations []Elimination `json:"eliminations,omitempty"`
	GridAfter    string        `json:"gridAfter"`
}

type Cell struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type Placement struct {
	Row   int `json:"row"`
	Col   int `json:"col"`
	Digit int `json:"digit"`
}

type Elimination struct {
	Row   int `json:"row"`
	Col   int `json:"col"`
	Digit int `json:"digit"`
}
