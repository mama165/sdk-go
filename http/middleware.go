package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"log/slog"
)

// responseRecorder intercepts the response body & status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           bytes.NewBuffer(nil),
	}
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// LogRequestResponse logs HTTP request & response bodies using slog
func LogRequestResponse(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()

			// --- Read request body ---
			var reqBody []byte
			if r.Body != nil {
				reqBody, _ = io.ReadAll(r.Body)
			}

			// Restore body to prevent issues in handlers
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

			logger.Info("HTTP request received",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("body", string(reqBody)),
			)

			// --- Wrap response writer ---
			rec := newResponseRecorder(w)

			// Execute next handler
			next.ServeHTTP(rec, r)

			// --- Log the response ---
			logger.Info("HTTP response sent",
				slog.Int("status", rec.statusCode),
				slog.String("body", rec.body.String()),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
