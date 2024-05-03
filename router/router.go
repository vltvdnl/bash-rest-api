package router

import (
	middleware "Term-api/middleware/handlers"
	"context"
	"net/http"
	"time"
)

// Router создает роутер сервера
func Router(h *middleware.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/add-command", adaptHandler(h.AddCMD, "POST"))
	mux.HandleFunc("/api/show-commands", adaptHandler(h.AllLog, "GET"))
	mux.HandleFunc("/api/show-commands/", adaptHandler(h.LogCMD, "GET"))
	return mux
}

// adaptHandler проверяет метод хэндлера, создает timeout для любого запроса
func adaptHandler(h http.HandlerFunc, method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		h(w, r.WithContext(ctx))
	}
}
