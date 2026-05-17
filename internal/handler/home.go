package handler

import (
	"log/slog"
	"net/http"
)

func Greet(w http.ResponseWriter, r *http.Request) {
	slog.Info("Hello there!")
	w.Write([]byte("Hello World"))
}
