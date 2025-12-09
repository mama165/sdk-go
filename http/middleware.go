package http

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const (
	maxObservedBytes = 8 << 10 // 8 KB
)

// LogJSONBodyMiddleware returns an HTTP middleware that logs JSON request bodies.
//
// Behavior:
//   - Only processes requests with Content-Type starting with "application/json".
//   - Skips empty bodies or non-JSON requests.
//   - In Dev/Test (logger enabled at DEBUG level):
//   - Reads and decodes the JSON body.
//   - Sanitizes sensitive fields (password, token, etc.).
//   - Logs the full JSON payload for debugging.
//   - In Production (logger level < DEBUG):
//   - Reads up to maxObservedBytes (8 KB) of the body.
//   - Computes an SHA-256 hash of the body.
//   - Logs minimal metadata without exposing sensitive data.
//   - Restores the request body so the next handler can consume it.
//     ⚠️ Do not use DEBUG logging in production for sensitive data.
//     ⚠️ This middleware is transparent and does not modify request behavior.
//
// Example usage:
//
//	http.Handle("/api", LogJSONBodyMiddleware(logger)(myHandler))
func LogJSONBodyMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ct := r.Header.Get("Content-Type")
			if !strings.HasPrefix(ct, "application/json") {
				next.ServeHTTP(w, r)
				return
			}

			if r.Body == nil || r.ContentLength == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// 🔀 Only for dev & test
			if logger.Enabled(r.Context(), slog.LevelDebug) {
				logJSONDebug(logger, r)
				next.ServeHTTP(w, r)
				return
			}
			logJSONSafe(logger, r)
			next.ServeHTTP(w, r)
		})
	}
}

func logJSONDebug(logger *slog.Logger, r *http.Request) {
	var buf bytes.Buffer
	tee := io.TeeReader(r.Body, &buf)

	var payload any
	err := json.NewDecoder(tee).Decode(&payload)

	switch {
	case err == io.EOF:
		logger.Debug("incoming request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

	case err != nil:
		logger.Error("invalid JSON body",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)

	default:
		Sanitize(payload)
		logger.Debug("incoming request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("body", payload),
		)
	}

	r.Body = io.NopCloser(&buf)
}

func logJSONSafe(logger *slog.Logger, r *http.Request) {
	limited := io.LimitReader(r.Body, maxObservedBytes)
	hasher := sha256.New()
	hash := hex.EncodeToString(hasher.Sum(nil))

	read, err := io.Copy(hasher, limited)
	if err != nil {
		logger.Warn("request body read error",
			slog.Any("error", err),
		)
	} else {
		logger.Info("incoming request observed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("content_type", r.Header.Get("Content-Type")),
			slog.Int64("content_length", r.ContentLength),
			slog.Int64("observed_bytes", read),
			slog.String("body_sha256", hash),
		)
	}

	// restore body (transparent)
	r.Body = io.NopCloser(io.MultiReader(io.LimitReader(strings.NewReader(""), 0), r.Body))
}

func Sanitize(v any) {
	switch t := v.(type) {

	case map[string]any:
		for k, v := range t {
			if isSensitiveKey(k) {
				t[k] = "*****"
			} else {
				Sanitize(v)
			}
		}

	case []any:
		for _, v := range t {
			Sanitize(v)
		}
	}
}

func isSensitiveKey(k string) bool {
	switch strings.ToLower(k) {
	case "password", "token", "access_token", "refresh_token", "secret", "authorization":
		return true
	default:
		return false
	}
}
