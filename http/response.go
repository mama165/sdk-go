package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Response struct {
	Payload     interface{}
	StatusCode  int
	contentType string
	header      map[string]string
}

func (r *Response) Error() string {
	return fmt.Sprintf("%d: %v", r.StatusCode, r.Payload)
}

func (r *Response) SetContentType(ct string) *Response {
	r.contentType = ct
	return r
}

func (r *Response) AddHeader(key, value string) *Response {
	if r.header == nil {
		r.header = make(map[string]string)
	}
	r.header[key] = value
	return r
}

func OK(content interface{}) *Response {
	return &Response{Payload: content, StatusCode: http.StatusOK}
}

func Created(content interface{}) *Response {
	return &Response{Payload: content, StatusCode: http.StatusCreated}
}

func BadRequest(details string) *Response {
	return &Response{
		Payload:    map[string]string{"message": "bad request", "details": details},
		StatusCode: http.StatusBadRequest,
	}
}

func NotFound(details string) *Response {
	return &Response{
		Payload:    map[string]string{"message": "not found", "details": details},
		StatusCode: http.StatusNotFound,
	}
}

func Forbidden(details string) *Response {
	return &Response{
		Payload:    map[string]string{"message": "forbidden", "details": details},
		StatusCode: http.StatusForbidden,
	}
}

func Unauthorized(details string) *Response {
	return &Response{
		Payload:    map[string]string{"message": "unauthorized", "details": details},
		StatusCode: http.StatusUnauthorized,
	}
}

func InternalError(details string) *Response {
	return &Response{
		Payload:    map[string]string{"message": "internal server error", "details": details},
		StatusCode: http.StatusInternalServerError,
	}
}

func NoContent() *Response {
	return &Response{StatusCode: http.StatusNoContent}
}

type Handler func(w http.ResponseWriter, r *http.Request) *Response

func JSON(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondWithJSON(w, h(w, r))
	}
}

func Stream(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := h(w, r)

		if rc, ok := resp.Payload.(io.Reader); ok {
			if resp.contentType != "" {
				w.Header().Set("Content-Type", resp.contentType)
			}
			for k, v := range resp.header {
				w.Header().Set(k, v)
			}

			w.WriteHeader(resp.StatusCode)

			if _, err := io.Copy(w, rc); err != nil {
				respondWithError(w, http.StatusInternalServerError, "encoding error", err.Error())
			}
			return
		}

		respondWithJSON(w, resp)
	}
}

func respondWithJSON(w http.ResponseWriter, resp *Response) {
	// 204 → no body allowed
	if resp.StatusCode == http.StatusNoContent {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(resp.Payload); err != nil {
		http.Error(w, "encoding error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	for k, v := range resp.header {
		w.Header().Set(k, v)
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(buf.Bytes())
}

func respondWithError(w http.ResponseWriter, code int, msg string, details string) {
	w.Header().Set("X-Content-Type-Options", "nosniff")

	respondWithJSON(w, &Response{
		Payload: map[string]string{
			"message": msg,
			"details": details,
		},
		StatusCode: code,
	})
}
