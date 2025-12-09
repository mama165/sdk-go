package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mama165/sdk-go/logs"
	"github.com/stretchr/testify/assert"
)

func TestLogJSONBodyMiddlewareObfuscatingPassword(t *testing.T) {
	ass := assert.New(t)
	var buf bytes.Buffer
	logger := logs.GetLoggerFromBufferWithLogger(&buf, slog.LevelDebug)

	body := `{"email":"user@example.com","password":"secret"}`
	r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	middleware := LogJSONBodyMiddleware(logger)
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

	var logEntry map[string]any
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	ass.NoError(err)

	bodyField, ok := logEntry["body"]
	ass.True(ok)

	bodyMap, ok := bodyField.(map[string]any)
	ass.True(ok)
	ass.Equal(http.StatusOK, w.Code)
	ass.Equal("/test", logEntry["url"])
	ass.Equal(bodyMap["email"], "user@example.com")
	ass.Equal(bodyMap["password"], "*****")
}

func TestLogWithGetMethod(t *testing.T) {
	ass := assert.New(t)
	var buf bytes.Buffer
	logger := logs.GetLoggerFromBufferWithLogger(&buf, slog.LevelDebug)

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware := LogJSONBodyMiddleware(logger)
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

	var logEntry map[string]any
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	ass.NoError(err)

	ass.Equal(http.StatusOK, w.Code)
	ass.Equal(http.MethodGet, logEntry["method"])
	ass.Equal("/test", logEntry["url"])
}

func TestLogWithBodyError(t *testing.T) {
	ass := assert.New(t)
	var buf bytes.Buffer
	logger := logs.GetLoggerFromBufferWithLogger(&buf, slog.LevelDebug)

	body := `"email":"user"` // Invalid JSON
	r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	w := httptest.NewRecorder()

	middleware := LogJSONBodyMiddleware(logger)
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

	var logEntry map[string]any
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	ass.NoError(err)

	ass.Equal(http.StatusOK, w.Code)
	ass.Equal("/test", logEntry["url"])
	ass.Equal("failed to decode JSON body", logEntry["msg"])
	ass.Equal("json: cannot unmarshal string into Go value of type map[string]interface {}", logEntry["error"])
}

func TestLogWithFileUploadMethod(t *testing.T) {
	ass := assert.New(t)
	var buf bytes.Buffer
	logger := logs.GetLoggerFromBufferWithLogger(&buf, slog.LevelDebug)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Given a fake file
	part, err := writer.CreateFormFile("file", "test.txt")
	ass.NoError(err)
	_, err = io.Copy(part, strings.NewReader("file content"))
	ass.NoError(err)
	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "/upload", body)
	w := httptest.NewRecorder()
	r.Header.Set("Content-Type", writer.FormDataContentType())

	middleware := LogJSONBodyMiddleware(logger)
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

	var logEntry map[string]any
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	ass.NoError(err)

	ass.Equal(http.MethodPost, logEntry["method"])
	ass.Equal(http.StatusOK, w.Code)
	ass.Equal("/upload", logEntry["url"])
	ass.Equal("[upload] incoming upload request", logEntry["msg"])
}
