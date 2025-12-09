package http

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mama165/sdk-go/logs"
	"github.com/stretchr/testify/assert"
)

func TestLogObfuscatesPassword(t *testing.T) {
	ass := assert.New(t)

	var buf bytes.Buffer
	logger := logs.GetLoggerFromBufferWithLogger(&buf, slog.LevelDebug)

	body := `{"email":"user@example.com","password":"secret"}`
	r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	mw := LogJSONBodyMiddleware(logger)
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

	// slog can produce multiple lines
	// Parsing the last one
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	last := lines[len(lines)-1]

	var logEntry map[string]any
	err := json.Unmarshal(last, &logEntry)
	ass.NoError(err)

	bodyField := logEntry["body"].(map[string]any)
	ass.Equal("user@example.com", bodyField["email"])
	ass.Equal("*****", bodyField["password"])
}

func TestNoLogOnGetWithoutBody(t *testing.T) {
	ass := assert.New(t)

	var buf bytes.Buffer
	logger := logs.GetLoggerFromBufferWithLogger(&buf, slog.LevelDebug)

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	LogJSONBodyMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

	// Then no log are expected
	ass.Equal(http.StatusOK, w.Code)
	ass.Len(buf.Bytes(), 0)
}

func TestLogWithBodyError(t *testing.T) {
	ass := assert.New(t)

	var buf bytes.Buffer
	logger := logs.GetLoggerFromBufferWithLogger(&buf, slog.LevelDebug)

	body := `{"email":`
	r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LogJSONBodyMiddleware(logger)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	).ServeHTTP(w, r)

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	ass.NotEmpty(lines)

	foundError := false

	for _, line := range lines {
		var entry map[string]any
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		if entry["level"] == "ERROR" && entry["error"] != nil {
			foundError = true
			break
		}
	}

	ass.True(foundError)
	ass.Equal(http.StatusOK, w.Code)
}

func TestMultipartIsIgnored(t *testing.T) {
	ass := assert.New(t)

	var buf bytes.Buffer
	logger := logs.GetLoggerFromBufferWithLogger(&buf, slog.LevelDebug)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()
	_, err := writer.CreateFormField("file")
	ass.NoError(err)

	r := httptest.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()

	LogJSONBodyMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

	// Then no log are expected
	ass.Equal(http.StatusOK, w.Code)
	ass.Len(buf.Bytes(), 0)
}
