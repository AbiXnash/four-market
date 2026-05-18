package response

import (
	"encoding/json"
	"net/http"
)

type API struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(API{Success: true, Data: data})
}

func Error(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(API{Success: false, Error: msg})
}

func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "internal server error")
}

func NotFound(w http.ResponseWriter) {
	Error(w, http.StatusNotFound, "not found")
}

func Unauthorized(w http.ResponseWriter) {
	Error(w, http.StatusUnauthorized, "unauthorized")
}

func Forbidden(w http.ResponseWriter) {
	Error(w, http.StatusForbidden, "forbidden")
}

func ValidationError(w http.ResponseWriter, msg string) {
	Error(w, http.StatusUnprocessableEntity, msg)
}
