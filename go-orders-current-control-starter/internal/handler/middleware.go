package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

type requestIDKey struct{}

var requestSequence atomic.Uint64

func requestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func requestLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: capture status code and duration, generate request_id,
			// and write method, path, status, duration and request_id to slog.
			start := time.Now()

			id := fmt.Sprintf("req-%d", requestSequence.Add(1))
			w.Header().Set("X-Request-ID", id)
			ctx := context.WithValue(r.Context(), requestIDKey{}, id)

			sw := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(sw, r.WithContext(ctx))
			logger.Info("http_request",
				"request_id", id,
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

func recoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: recover from panic, log the incident and return a safe JSON 500 response.
			defer func() {
				if value := recover(); value != nil {
					logger.Error("panic recovery", "request_id", requestIDFromContext(r.Context()))
					writeJSON(w, 500, ErrorResponse{Error: "Internal server error", Code: ecInternal})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func (w *statusWriter) WriteHeader(code int) {
	if code >= 100 && code < 200 {
		w.ResponseWriter.WriteHeader(code)
		return
	}
	if w.status != 0 {
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}
