package httpapi

import (
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/NerdAlert58/sudoku-flow2/web"
)

// routes registers path-only patterns; each /v1 handler dispatches on method
// itself so every /v1 error carries the JSON envelope (ADR-0005).
func routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", handleHealth)
	mux.HandleFunc("/v1/", handleV1NotFound)
	mux.Handle("/", http.FileServerFS(web.FS))
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeEnvelope(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed; only GET is supported")
		return
	}
	writeJSON(w, http.StatusOK, healthResponse{
		Status:     "ok",
		GoVersion:  runtime.Version(),
		APIVersion: "1",
	})
}

func handleV1NotFound(w http.ResponseWriter, r *http.Request) {
	writeEnvelope(w, http.StatusNotFound, "not_found", "no such endpoint")
}

func writeEnvelope(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorEnvelope{Error: message, Code: code})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
