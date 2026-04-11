// @title           AutoParts Hub API
// @version         1.0
// @description     API Gateway for AutoParts Hub
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/api-gateway/config"
	_ "github.com/f4ke-n0name/autoparts-hub/services/api-gateway/docs"
	"github.com/f4ke-n0name/autoparts-hub/services/api-gateway/internal/clients"
	"github.com/f4ke-n0name/autoparts-hub/services/api-gateway/internal/handler"
	"github.com/f4ke-n0name/autoparts-hub/services/api-gateway/internal/middleware"
)

func main() {
	_ = godotenv.Load("./.env")
	cfg := config.MustLoad()
	log := logger.SetupLogger(os.Getenv("ENV"))
	log.Info("starting api-gateway", slog.String("port", cfg.HTTP.Port))

	grpcClients, err := clients.New(
		cfg.Services.AuthAddr,
		cfg.Services.CatalogAddr,
		cfg.Services.OrderAddr,
		cfg.Services.InventoryAddr,
	)
	if err != nil {
		log.Error("failed to connect to services", logger.Err(err))
		os.Exit(1)
	}
	defer grpcClients.Close()

	authHandler := handler.NewAuthHandler(grpcClients.Auth)
	catalogHandler := handler.NewCatalogHandler(grpcClients.Catalog)
	orderHandler := handler.NewOrderHandler(grpcClients.Order)
	inventoryHandler := handler.NewInventoryHandler(grpcClients.Inventory)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/swagger/doc.json")))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(grpcClients.Auth))

			r.Route("/catalog", func(r chi.Router) {
				r.Get("/parts", catalogHandler.SearchParts)
				r.Post("/parts", catalogHandler.CreatePart)
				r.Get("/parts/{id}", catalogHandler.GetPart)
				r.Get("/categories", catalogHandler.ListCategories)
			})

			r.Route("/orders", func(r chi.Router) {
				r.Post("/", orderHandler.CreateOrder)
				r.Get("/", orderHandler.ListOrders)
				r.Get("/{id}", orderHandler.GetOrder)
				r.Patch("/{id}/status", orderHandler.UpdateStatus)
			})

			r.Route("/inventory", func(r chi.Router) {
				r.Get("/stock/{part_id}", inventoryHandler.GetStock)
				r.Put("/stock", inventoryHandler.UpdateStock)
			})
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info("swagger UI available", slog.String("url", fmt.Sprintf("http://localhost:%s/swagger/index.html", cfg.HTTP.Port)))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server error", logger.Err(err))
			os.Exit(1)
		}
	}()

	<-quit
	log.Info("shutting down api-gateway")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", logger.Err(err))
	}

	log.Info("api-gateway stopped")
}
