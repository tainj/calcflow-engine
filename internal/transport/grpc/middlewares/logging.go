package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tainj/distributed_calculator2/internal/auth"
	"github.com/tainj/distributed_calculator2/pkg/logger"
)

// LoggerProvider adds logger to request context
func LoggerProvider(serviceName string) Middleware {
	l := logger.New(serviceName)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), logger.LoggerKey, l)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// loggingResponseWriter - for tracking status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Logging middleware - logs completed requests
func Logging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := logger.GetLoggerFromCtx(r.Context())
			lrw := newLoggingResponseWriter(w)
			start := time.Now()

			defer func() {
				duration := time.Since(start).Milliseconds()
				userId := auth.UserIDFromCtx(r.Context()) // Get userId from context

				// Form attributes for log
				attrs := []any{
					"method", r.Method,
					"uri", r.RequestURI,
					"status_code", lrw.statusCode,
					"elapsed_ms", duration,
				}

				// Add userId only if present
				if userId != "" {
					attrs = append(attrs, "user_id", userId)
				}

				l.Info(r.Context(),
					fmt.Sprintf("%s request to %s completed", r.Method, r.RequestURI),
					attrs...,
				)
			}()

			next.ServeHTTP(lrw, r)
		})
	}
}
