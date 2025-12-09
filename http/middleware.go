package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// LogJSONBodyMiddleware is an HTTP middleware that logs the JSON request body
// Reads and logs the body of incoming HTTP requests
// Only if the method is POST, PUT or PATCH
// Only if the Content-Type is "application/json".
// Avoid file upload with multipart content-type
// The body is read once then deserialized into a map for structured logging
// Useful for debugging JSON payloads during development or in controlled environments.
// ⚠️ Note: The request body is logged in plain text.
// ⚠️ Note: Do not use this in production without filtering or masking sensitive fields
// To secure : passwords, tokens, or credentials.
func LogJSONBodyMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle file upload
			if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
				logger.Debug("[upload] incoming upload request",
					slog.String("method", r.Method),
					slog.String("url", r.URL.String()),
				)
			} else {
				// Read and duplicate the body
				var buf bytes.Buffer
				tee := io.TeeReader(r.Body, &buf)

				// Decode the JSON into a generic map
				var jsonBody map[string]any
				err := json.NewDecoder(tee).Decode(&jsonBody)
				switch {
				case err == io.EOF: // Empty body
					logger.Debug("[json] incoming request",
						slog.String("method", r.Method),
						slog.String("url", r.URL.String()),
					)
				case err != nil: // Error decoding JSON
					logger.Error("failed to decode JSON body",
						slog.String("method", r.Method),
						slog.String("url", r.URL.String()),
						slog.Any("error", err),
					)
				default: // Successfully decoded JSON
					if _, ok := jsonBody["password"]; ok {
						jsonBody["password"] = "*****"
					}
					logger.Debug("[json] incoming request",
						slog.String("method", r.Method),
						slog.String("url", r.URL.String()),
						slog.Any("body", jsonBody),
					)
				}

				// Always restore the body for the next handler
				r.Body = io.NopCloser(&buf)
			}
			next.ServeHTTP(w, r)
		})
	}
}
