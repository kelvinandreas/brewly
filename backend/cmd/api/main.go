package main

import (
	"net/http"
	"os"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/your-handle/brewly/internal/handler"
	appMiddleware "github.com/your-handle/brewly/internal/middleware"
	"github.com/your-handle/brewly/internal/repository"
	"github.com/your-handle/brewly/internal/usecase"
	"github.com/your-handle/brewly/pkg/db"
	"github.com/your-handle/brewly/internal/domain"
)

func main() {
	_ = godotenv.Load()

	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	// ── Database ───────────────────────────────────────────────────────────────
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal().Msg("DATABASE_URL is required")
	}
	gormDB, err := db.Open(dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	log.Info().Msg("database connected")

	// ── Wiring ─────────────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepo(gormDB)

	authCfg := usecase.AuthConfig{
		AccessSecret:  mustEnv("JWT_SECRET"),
		RefreshSecret: mustEnv("REFRESH_SECRET"),
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
	}
	authUC := usecase.NewAuthUsecase(userRepo, authCfg)
	userUC := usecase.NewUserUsecase(userRepo)

	authH := handler.NewAuthHandler(authUC)
	userH := handler.NewUserHandler(userUC)

	// ── Router ─────────────────────────────────────────────────────────────────
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth routes (unauthenticated)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register-owner", authH.RegisterOwner)
		r.Post("/login", authH.Login)
		r.Post("/refresh", authH.Refresh)
		// logout and me require a valid access token
		r.Group(func(r chi.Router) {
			r.Use(appMiddleware.RequireAuth(authCfg.AccessSecret))
			r.Post("/logout", authH.Logout)
			r.Get("/me", authH.Me)
		})
	})

	// User management — owner only
	r.Route("/api/users", func(r chi.Router) {
		r.Use(appMiddleware.RequireAuth(authCfg.AccessSecret, domain.RoleOwner))
		r.Get("/", userH.List)
		r.Post("/", userH.Create)
		r.Patch("/{id}", userH.Update)
		r.Delete("/{id}", userH.Delete)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Info().Str("port", port).Msg("starting brewly backend")
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatal().Str("key", key).Msg("required environment variable not set")
	}
	return v
}
