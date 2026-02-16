// Package main is the entry point for the Inventory API server.
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

	"inventario/server/internal/config"
	"inventario/server/internal/database"
	"inventario/server/internal/handler"
	"inventario/server/internal/repository"
	"inventario/server/internal/router"
	"inventario/server/internal/service"
	"inventario/server/migrations"
)

func main() {
	// Sub-command: create-user
	if len(os.Args) > 1 && os.Args[1] == "create-user" {
		runCreateUser()
		return
	}

	runServer()
}

func runServer() {
	// ── Configuration ────────────────────────────────────────────────
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	slog.Info("starting inventory API server", "port", cfg.ServerPort)

	// ── Database ─────────────────────────────────────────────────────
	db := database.Connect(cfg.DatabaseURL)
	defer db.Close()

	database.RunMigrations(cfg.DatabaseURL, migrations.FS)

	// ── Repositories ─────────────────────────────────────────────────
	tokenRepo := repository.NewTokenRepository(db)
	userRepo := repository.NewUserRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)

	// ── Services ─────────────────────────────────────────────────────
	authSvc := service.NewAuthService(db, userRepo, tokenRepo, cfg.JWTSecret)
	inventorySvc := service.NewInventoryService(inventoryRepo)
	deviceSvc := service.NewDeviceService(deviceRepo)
	dashboardSvc := service.NewDashboardService(dashboardRepo)

	// ── Handlers ─────────────────────────────────────────────────────
	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(authSvc, cfg.EnrollmentKey)
	inventoryHandler := handler.NewInventoryHandler(inventorySvc)
	deviceHandler := handler.NewDeviceHandler(deviceSvc)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)

	// ── Router ───────────────────────────────────────────────────────
	r := router.Setup(cfg, healthHandler, inventoryHandler, authHandler, deviceHandler, dashboardHandler, tokenRepo)

	// ── HTTP Server ──────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// ── Graceful Shutdown ────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}

func runCreateUser() {
	var username, password string
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--username":
			if i+1 < len(os.Args) {
				username = os.Args[i+1]
				i++
			}
		case "--password":
			if i+1 < len(os.Args) {
				password = os.Args[i+1]
				i++
			}
		}
	}

	if username == "" || password == "" {
		fmt.Println("Usage: server create-user --username <user> --password <pass>")
		os.Exit(1)
	}

	if len(password) < 8 {
		fmt.Println("Error: password must be at least 8 characters")
		os.Exit(1)
	}

	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	db := database.Connect(cfg.DatabaseURL)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	authSvc := service.NewAuthService(db, userRepo, tokenRepo, cfg.JWTSecret)

	if err := authSvc.CreateUser(context.Background(), username, password); err != nil {
		slog.Error("failed to create user", "error", err)
		os.Exit(1)
	}

	fmt.Printf("User '%s' created successfully\n", username)
}
