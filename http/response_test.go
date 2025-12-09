package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONOK(t *testing.T) {
	ass := assert.New(t)

	handler := JSON(func(w http.ResponseWriter, r *http.Request) *Response {
		return OK(map[string]string{"hello": "world"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	ass.Equal(http.StatusOK, rec.Code)
	ass.Equal("application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var body map[string]string
	ass.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	ass.Equal("world", body["hello"])
}

func TestJSONError(t *testing.T) {
	ass := assert.New(t)

	handler := JSON(func(w http.ResponseWriter, r *http.Request) *Response {
		return BadRequest("missing field")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	ass.Equal(http.StatusBadRequest, rec.Code)

	var body map[string]string
	ass.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	ass.Equal("bad request", body["message"])
	ass.Equal("missing field", body["details"])
}

func TestStreamSuccess(t *testing.T) {
	ass := assert.New(t)

	data := "streamed data"
	handler := Stream(func(w http.ResponseWriter, r *http.Request) *Response {
		return OK(io.NopCloser(bytes.NewBufferString(data))).SetContentType("text/plain")
	})

	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	ass.Equal(http.StatusOK, rec.Code)
	ass.Equal("text/plain", rec.Header().Get("Content-Type"))
	ass.Equal(data, rec.Body.String())
}

func TestStreamFallbackToJSON(t *testing.T) {
	ass := assert.New(t)

	handler := Stream(func(w http.ResponseWriter, r *http.Request) *Response {
		return OK("not-a-reader")
	})

	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	ass.Equal(http.StatusOK, rec.Code)
	ass.Equal("application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var body string
	ass.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	ass.Equal("not-a-reader", body)
}

func TestHeadersAreIncluded(t *testing.T) {
	ass := assert.New(t)

	handler := JSON(func(w http.ResponseWriter, r *http.Request) *Response {
		return OK("ok").
			AddHeader("X-Test", "123").
			AddHeader("X-Another", "456")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	ass.Equal("123", rec.Header().Get("X-Test"))
	ass.Equal("456", rec.Header().Get("X-Another"))
}

func TestRespondWithErrorHelper(t *testing.T) {
	ass := assert.New(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondWithError(w, http.StatusForbidden, "forbidden", "token expired")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	ass.Equal(http.StatusForbidden, rec.Code)

	var body map[string]string
	ass.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	ass.Equal("forbidden", body["message"])
	ass.Equal("token expired", body["details"])
}

func TestNoContent(t *testing.T) {
	ass := assert.New(t)

	handler := JSON(func(w http.ResponseWriter, r *http.Request) *Response {
		return NoContent()
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	ass.Equal(http.StatusNoContent, rec.Code)
	ass.Equal(0, rec.Body.Len()) // No body expected
}

func TestEncodingFailureReturns500(t *testing.T) {
	ass := assert.New(t)

	// Create a type that cannot be JSON-marshaled
	type Bad struct {
		Ch chan int `json:"ch"`
	}

	handler := JSON(func(w http.ResponseWriter, r *http.Request) *Response {
		return OK(Bad{Ch: make(chan int)})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	ass.Equal(http.StatusInternalServerError, rec.Code)
	ass.Contains(rec.Body.String(), "encoding error")
}
