package main

// @title Book Shop API
// @version 1.0
// @description This is a book shop server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email your-email@domain.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	_ "toptal/docs"
	"toptal/internal/app/auth"
	"toptal/internal/app/config"
	"toptal/internal/app/handler"
	"toptal/internal/app/handler/middleware"
	"toptal/internal/app/health"
	"toptal/internal/app/repository"
	"toptal/internal/app/service"
	"toptal/internal/pkg/pg"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	auth.SetConfig(cfg.Security)

	db, err := pg.Connect(cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func(db *pg.DB) {
		err := db.Close()
		if err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}(db)

	if err := runMigrations(cfg.DB.DSN()); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// repository
	bookRepository := repository.NewBookRepository(db)
	categoryRepository := repository.NewCategoryRepository(db)
	userRepository := repository.NewUserRepository(db)
	cartRepository := repository.NewCartRepository(db, &cfg.Cart)

	// service
	authService := service.NewAuthService(userRepository)
	bookService := service.NewBookService(bookRepository, *authService)
	categoryService := service.NewCategoryService(categoryRepository, *authService)
	cartService := service.NewCartService(cartRepository, &cfg.Cart)
	healthService := health.NewHealthService(db)

	// server
	server := handler.NewServer(bookService, categoryService, authService, cartService, healthService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      middleware.MetricsMiddleware(server.Handler()),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	var metricsServer *http.Server
	if cfg.Metrics.Enabled {
		metricsServer = &http.Server{
			Addr:    fmt.Sprintf(":%s", cfg.Metrics.Port),
			Handler: promhttp.Handler(),
		}

		go func() {
			if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("Metrics server error", "error", err)
			}
		}()
	}

	cartService.StartCartCleanerJob(ctx)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down servers...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if cfg.Metrics.Enabled && metricsServer != nil {
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("Metrics server shutdown error", "error", err)
		}
	}

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", "error", err)
	}

	return nil
}

func runMigrations(psqlInfo string) error {
	slog.Info("Running migrations...")
	m, err := migrate.New("file://migrations", psqlInfo)
	if err != nil {
		return fmt.Errorf("failed to init migrations: %w", err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}
	slog.Info("Migrations applied successfully")
	return nil
}
