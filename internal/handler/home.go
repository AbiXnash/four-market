package handler

import (
	"log/slog"
	"net/http"

	"github.com/AbiXnash/4-market/internal/response"
)

func Greet(w http.ResponseWriter, r *http.Request) {
	slog.Info("Hello there!")
	response.JSON(w, http.StatusOK, map[string]string{"message": "Hello World"})
}
