package router

import (
	"net/http"
	"time"

	"guestbook/backend/handlers"
	"guestbook/backend/middleware"

	"gorm.io/gorm"
)

func New(db *gorm.DB, mysqlPort string, jwtSecret string, tokenExpiration time.Duration) http.Handler {
	mux := http.NewServeMux()

	systemHandler := handlers.NewSystemHandler(db, mysqlPort)
	publicHandler := handlers.NewPublicHandler(db)
	authenticator := middleware.NewAuthenticator(jwtSecret)
	authHandler := handlers.NewAuthHandler(db, authenticator, tokenExpiration)

	mux.HandleFunc("GET /health", systemHandler.Health)
	mux.HandleFunc("GET /migrations", systemHandler.Migrations)
	mux.HandleFunc("GET /api/public/forms/{public_slug}", publicHandler.GetGuestForm)
	mux.HandleFunc("POST /api/public/forms/{public_slug}/visits", publicHandler.SubmitGuestVisit)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.Handle("GET /api/admin/me", authenticator.Middleware(http.HandlerFunc(authHandler.Me)))

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
