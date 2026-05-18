package response

import (
	"encoding/json"
	"net/http"
)

type API struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
	Meta    *Meta     `json:"meta,omitempty"`
}

type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Status    int    `json:"status"`
	RequestID string `json:"request_id,omitempty"`
	Path      string `json:"path,omitempty"`
}

type Meta struct {
	RequestID string `json:"request_id,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(API{
		Success: true,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, r *http.Request, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(API{
		Success: false,
		Error: &APIError{
			Code:      code,
			Message:   msg,
			Status:    status,
			RequestID: requestID(r),
			Path:      requestPath(r),
		},
		Meta: &Meta{
			RequestID: requestID(r),
		},
	})
}

func InternalError(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusInternalServerError, "internal_server_error", "internal server error")
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusNotFound, "not_found", "resource not found")
}

func Unauthorized(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusUnauthorized, "unauthorized", "authentication required")
}

func Forbidden(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusForbidden, "forbidden", "access denied")
}

func ValidationError(w http.ResponseWriter, r *http.Request, msg string) {
	Error(w, r, http.StatusUnprocessableEntity, "validation_error", msg)
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

func UnsupportedMediaType(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusUnsupportedMediaType, "unsupported_media_type", "content type must be application/json")
}

func TooManyRequests(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusTooManyRequests, "rate_limit_exceeded", "too many requests")
}

func requestID(r *http.Request) string {
	if r == nil {
		return ""
	}
	return r.Header.Get("X-Request-ID")
}

func requestPath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	return r.URL.Path
}
