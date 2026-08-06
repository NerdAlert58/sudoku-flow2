// Package httpapi owns the entire HTTP surface: routes, the error envelope,
// and the hardening middleware chain (ARCHITECTURE.md §Components).
package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func New() http.Handler {
	return NewWithLogWriter(os.Stdout)
}

func NewWithLogWriter(w io.Writer) http.Handler {
	return chain(w, routes())
}

func chain(logW io.Writer, next http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(logW, nil))
	return accessLog(logger, recoverPanic(securityHeaders(cors(next))))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func accessLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.LogAttrs(r.Context(), slog.LevelInfo, "request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				writeEnvelope(w, http.StatusInternalServerError, "internal_error", "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

const contentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'"

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		h.Set("Strict-Transport-Security", "max-age=63072000")
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// corsAllowlist ships empty by design: no Access-Control-* header is ever
// emitted and no Origin is echoed (AUDIT.md S2). Enumerating an origin here
// requires exact-match echo plus Vary: Origin, with tests.
var corsAllowlist = map[string]bool{}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); corsAllowlist[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		next.ServeHTTP(w, r)
	})
}
