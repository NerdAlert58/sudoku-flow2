package httpapi

// Shared transport gate for the /v1 handlers, in the frozen precedence:
// method (ADR-0005) → content type → byte cap (AUDIT.md A7). Every rejection
// is the JSON envelope with a frozen code (ADR-0004).

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
)

const maxBodyBytes = 1 << 20

// gateMethod is ADR-0005's in-handler method dispatch: on mismatch it writes
// 405 with the exact Allow header and reports false.
func gateMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	writeEnvelope(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed; only "+method+" is supported")
	return false
}

// gatePostJSON gates the three POST routes. The Content-Length fast path
// 413s a declared-oversized body without a single read; unknown-length
// bodies get the lazy MaxBytesReader cap instead (AUDIT.md A7). 415 is
// decided from the header alone, so the body is equally unread there.
func gatePostJSON(w http.ResponseWriter, r *http.Request) bool {
	if !gateMethod(w, r, http.MethodPost) {
		return false
	}
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		writeEnvelope(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "content type must be application/json")
		return false
	}
	if r.ContentLength > maxBodyBytes {
		writeEnvelope(w, http.StatusRequestEntityTooLarge, "payload_too_large", "request body exceeds 1 MiB")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	return true
}

// decodeBody decodes the gated JSON body into v; on failure it writes the
// envelope (413 when the lazy byte cap tripped, else 400) and reports false.
func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	err := json.NewDecoder(r.Body).Decode(v)
	if err == nil {
		return true
	}
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		writeEnvelope(w, http.StatusRequestEntityTooLarge, "payload_too_large", "request body exceeds 1 MiB")
		return false
	}
	writeEnvelope(w, http.StatusBadRequest, "invalid_input", "request body is not valid JSON for this endpoint")
	return false
}
