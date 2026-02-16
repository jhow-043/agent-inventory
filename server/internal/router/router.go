// Package router sets up the Gin engine with middleware and route definitions.
package router

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"inventario/server/internal/config"
	"inventario/server/internal/handler"
	"inventario/server/internal/middleware"
	"inventario/server/internal/repository"
)

// Setup creates and configures the Gin engine with all application routes.
func Setup(
	cfg *config.Config,
	healthHandler *handler.HealthHandler,
	inventoryHandler *handler.InventoryHandler,
	authHandler *handler.AuthHandler,
	deviceHandler *handler.DeviceHandler,
	dashboardHandler *handler.DashboardHandler,
	userHandler *handler.UserHandler,
	tokenRepo *repository.TokenRepository,
) *gin.Engine {
	if cfg.LogLevel != slog.LevelDebug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.CORSOrigins))

	// Health probes — no authentication required.
	r.GET("/healthz", healthHandler.Healthz)
	r.HEAD("/healthz", healthHandler.Healthz)
	r.GET("/readyz", healthHandler.Readyz)

	api := r.Group("/api/v1")
	{
		// Agent endpoints.
		api.POST("/enroll", middleware.RateLimit(10, time.Minute), authHandler.Enroll)
		api.POST("/inventory", middleware.DeviceAuth(tokenRepo), inventoryHandler.SubmitInventory)

		// Dashboard authentication.
		api.POST("/auth/login", middleware.RateLimit(5, time.Minute), authHandler.Login)
		api.POST("/auth/logout", authHandler.Logout)

		// Dashboard data endpoints — JWT protected.
		protected := api.Group("")
		protected.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			protected.GET("/auth/me", authHandler.Me)
			protected.GET("/dashboard/stats", dashboardHandler.GetStats)
			protected.GET("/devices", deviceHandler.ListDevices)
			protected.GET("/devices/:id", deviceHandler.GetDevice)
			protected.GET("/users", userHandler.ListUsers)
			protected.POST("/users", userHandler.CreateUser)
			protected.DELETE("/users/:id", userHandler.DeleteUser)
		}
	}

	return r
}
