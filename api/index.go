// Package handler is the Vercel entrypoint; it mounts the same handler graph
// as cmd/server so both entrypoints behave byte-identically (ADR-0001).
package handler

import (
	"net/http"

	"github.com/NerdAlert58/sudoku-flow2/httpapi"
)

var h = httpapi.New()

func Handler(w http.ResponseWriter, r *http.Request) {
	h.ServeHTTP(w, r)
}
