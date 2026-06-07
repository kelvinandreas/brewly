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
	"github.com/your-handle/brewly/internal/domain"
	"github.com/your-handle/brewly/internal/handler"
	appMiddleware "github.com/your-handle/brewly/internal/middleware"
	"github.com/your-handle/brewly/internal/repository"
	"github.com/your-handle/brewly/internal/usecase"
	"github.com/your-handle/brewly/pkg/db"
	"github.com/your-handle/brewly/pkg/sse"
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

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo     := repository.NewUserRepo(gormDB)
	categoryRepo := repository.NewCategoryRepo(gormDB)
	menuItemRepo := repository.NewMenuItemRepo(gormDB)
	tableRepo    := repository.NewTableRepo(gormDB)
	orderRepo    := repository.NewOrderRepo(gormDB)
	paymentRepo  := repository.NewPaymentRepo(gormDB)

	// ── Config ─────────────────────────────────────────────────────────────────
	authCfg := usecase.AuthConfig{
		AccessSecret:  mustEnv("JWT_SECRET"),
		RefreshSecret: mustEnv("REFRESH_SECRET"),
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
	}
	tableCfg := usecase.TableConfig{
		TableTokenSecret: mustEnv("TABLE_TOKEN_SECRET"),
		FrontendURL:      envOr("FRONTEND_URL", "http://localhost:5173"),
	}

	// ── SSE brokers ────────────────────────────────────────────────────────────
	kitchenBroker := sse.NewBroker()

	// ── Usecases ──────────────────────────────────────────────────────────────
	authUC     := usecase.NewAuthUsecase(userRepo, authCfg)
	userUC     := usecase.NewUserUsecase(userRepo)
	categoryUC := usecase.NewCategoryUsecase(categoryRepo)
	menuItemUC := usecase.NewMenuItemUsecase(menuItemRepo, categoryRepo)
	tableUC    := usecase.NewTableUsecase(tableRepo, tableCfg)
	orderUC    := usecase.NewOrderUsecase(orderRepo, menuItemRepo, kitchenBroker)
	paymentUC  := usecase.NewPaymentUsecase(paymentRepo, orderRepo)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authH     := handler.NewAuthHandler(authUC)
	userH     := handler.NewUserHandler(userUC)
	categoryH := handler.NewCategoryHandler(categoryUC)
	menuItemH := handler.NewMenuItemHandler(menuItemUC)
	tableH    := handler.NewTableHandler(tableUC)
	orderH    := handler.NewOrderHandler(orderUC)
	paymentH  := handler.NewPaymentHandler(paymentUC)
	sseH      := handler.NewSSEHandler(kitchenBroker, authCfg.AccessSecret)
	customerH := handler.NewCustomerHandler(categoryUC, menuItemUC, orderUC)

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

	// Categories — owner only
	r.Route("/api/categories", func(r chi.Router) {
		r.Use(appMiddleware.RequireAuth(authCfg.AccessSecret, domain.RoleOwner))
		r.Get("/", categoryH.List)
		r.Post("/", categoryH.Create)
		r.Patch("/{id}", categoryH.Update)
		r.Delete("/{id}", categoryH.Delete)
	})

	// Menu items — owner only
	r.Route("/api/menu-items", func(r chi.Router) {
		r.Use(appMiddleware.RequireAuth(authCfg.AccessSecret, domain.RoleOwner))
		r.Get("/", menuItemH.List)
		r.Post("/", menuItemH.Create)
		r.Patch("/{id}", menuItemH.Update)
		r.Delete("/{id}", menuItemH.Delete)
	})

	// Tables — owner only
	r.Route("/api/tables", func(r chi.Router) {
		r.Use(appMiddleware.RequireAuth(authCfg.AccessSecret, domain.RoleOwner))
		r.Get("/", tableH.List)
		r.Post("/", tableH.Create)
		r.Patch("/{id}", tableH.Update)
		r.Delete("/{id}", tableH.Delete)
		r.Post("/{id}/regenerate-token", tableH.RegenerateToken)
		r.Get("/{id}/qr.png", tableH.GetQR)
	})

	// Orders — all authenticated staff can read; cashier+owner can create/cancel
	r.Route("/api/orders", func(r chi.Router) {
		r.Use(appMiddleware.RequireAuth(authCfg.AccessSecret))
		r.Get("/", orderH.List)
		r.Group(func(r chi.Router) {
			r.Use(appMiddleware.RequireAuth(authCfg.AccessSecret, domain.RoleCashier, domain.RoleOwner))
			r.Post("/", orderH.Create)
		})
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", orderH.GetByID)
			r.Patch("/status", orderH.AdvanceStatus)
			r.Group(func(r chi.Router) {
				r.Use(appMiddleware.RequireAuth(authCfg.AccessSecret, domain.RoleCashier, domain.RoleOwner))
				r.Post("/cancel", orderH.Cancel)
				r.Post("/payments", paymentH.Create)
				r.Get("/payments", paymentH.List)
			})
		})
	})

	// SSE — token validated inside handler via query param
	r.Get("/api/sse/kitchen", sseH.KitchenStream)

	// Customer endpoints — table token required
	r.Route("/api/customer", func(r chi.Router) {
		r.Use(appMiddleware.RequireTableToken(tableCfg.TableTokenSecret, tableRepo))
		r.Get("/menu", customerH.GetMenu)
		r.Post("/orders", customerH.PlaceOrder)
		r.Get("/orders/mine", customerH.MyOrders)
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

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
