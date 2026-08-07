package httpapi

import (
	"net/http"

	"github.com/NerdAlert58/sudoku-flow2/catalog"
)

func handlePuzzles(w http.ResponseWriter, r *http.Request) {
	if !gateMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, puzzlesResponse{Sections: catalog.Sections()})
}
