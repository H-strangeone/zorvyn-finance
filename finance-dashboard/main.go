package main

import (
	"context"
	"finance-dashboard/config"
	"finance-dashboard/handlers"
	"finance-dashboard/middleware"
	"finance-dashboard/services"
	"finance-dashboard/store"
	"finance-dashboard/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Step 1: load config
	config.Load()

	// Step 2: initialize stores
	userStore := store.NewInMemoryUserStore()
	transactionStore := store.NewInMemoryTransactionStore()
	roleRequestStore := store.NewInMemoryRoleRequestStore()

	// Step 3: initialize services
	authService := services.NewAuthService(userStore)
	userService := services.NewUserService(userStore)
	transactionService := services.NewTransactionService(transactionStore)
	dashboardService := services.NewDashboardService(transactionStore)
	roleRequestService := services.NewRoleRequestService(roleRequestStore, userStore)

	// Step 4: seed admin
	seeded, err := userService.SeedAdmin(
		config.App.SeedAdminName,
		config.App.SeedAdminEmail,
		config.App.SeedAdminPassword,
	)
	if err != nil {
		log.Printf("⚠️  seed admin warning: %v", err)
	} else if seeded {
		log.Println("✅ Seed admin created")
	} else {
		log.Println("✅ Active admin exists, seed skipped")
	}

	// Step 5: initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	roleRequestHandler := handlers.NewRoleRequestHandler(roleRequestService)

	// Step 6: initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(60, time.Minute)

	// Step 7: initialize router
	// gin.New() not gin.Default() — full control over middleware
	router := gin.New()
	router.Use(gin.Recovery()) // recover from panics
	router.Use(gin.Logger())   // request logging
	router.Use(rateLimiter.RateLimit())

	// Step 8: register routes
	api := router.Group("/api/v1")

	// Public routes — no auth required
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes — auth middleware applied to group
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(userStore))
	{
		// User routes
		// /me registered before /:id — gin matches in order
		// static path must come before parameterized path
		users := protected.Group("/users")
		{
			users.GET("/me", middleware.RequireAnyRole(), userHandler.GetMe)
			users.GET("", middleware.RequireAdmin(), userHandler.GetAll)
			users.POST("", middleware.RequireAdmin(), userHandler.Create)
			users.GET("/:id", middleware.RequireAdmin(), userHandler.GetByID)
			users.PUT("/:id", middleware.RequireAdmin(), userHandler.Update)
			users.PATCH("/:id/status", middleware.RequireAdmin(), userHandler.UpdateStatus)
		}

		// Transaction routes
		transactions := protected.Group("/transactions")
		{
			transactions.GET("", middleware.RequireAnalystOrAdmin(), transactionHandler.GetAll)
			transactions.POST("", middleware.RequireAdmin(), transactionHandler.Create)
			transactions.GET("/:id", middleware.RequireAnalystOrAdmin(), transactionHandler.GetByID)
			transactions.PUT("/:id", middleware.RequireAdmin(), transactionHandler.Update)
			transactions.DELETE("/:id", middleware.RequireAdmin(), transactionHandler.Delete)
		}

		// Dashboard routes — all authenticated roles
		dashboard := protected.Group("/dashboard")
		{
			dashboard.GET("/summary", middleware.RequireAnyRole(), dashboardHandler.GetSummary)
			dashboard.GET("/by-category", middleware.RequireAnyRole(), dashboardHandler.GetByCategory)
			dashboard.GET("/trends", middleware.RequireAnyRole(), dashboardHandler.GetTrends)
			dashboard.GET("/recent", middleware.RequireAnyRole(), dashboardHandler.GetRecent)
		}

		// Role request routes
		// /mine registered before /:id — same reason as /me
		roleRequests := protected.Group("/role-requests")
		{
			roleRequests.GET("/mine", middleware.RequireAnyRole(), roleRequestHandler.GetMine)
			roleRequests.POST("", middleware.RequireAnyRole(), roleRequestHandler.Create)
			roleRequests.GET("", middleware.RequireAdmin(), roleRequestHandler.GetAll)
			roleRequests.GET("/:id", middleware.RequireAdmin(), roleRequestHandler.GetByID)
			roleRequests.PATCH("/:id", middleware.RequireAdmin(), roleRequestHandler.Process)
		}
	}

	// 404 handler — unknown routes return JSON not Gin's default HTML
	// keeps response format consistent across all endpoints
	router.NoRoute(func(c *gin.Context) {
		utils.SendError(c, utils.NewNotFoundError("route"))
	})

	// Step 9: configure http.Server for graceful shutdown
	// Why http.Server directly instead of router.Run()?
	// router.Run() blocks and has no shutdown hook
	// http.Server gives us Shutdown() method — waits for in-flight requests
	srv := &http.Server{
		Addr:    ":" + config.App.AppPort,
		Handler: router,
	}

	// Step 10: start server in goroutine so it doesn't block
	// main goroutine waits for signal below
	go func() {
		log.Printf("🚀 Server starting on port %s", config.App.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// http.ErrServerClosed is expected on clean shutdown — not a fatal error
			// any other error IS fatal
			log.Fatalf("FATAL: server error: %v", err)
		}
	}()

	// Step 11: block until OS signal received
	// SIGINT = Ctrl+C
	// SIGTERM = kill command / container orchestrator shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // blocks here until signal received

	log.Println("⏳ Shutting down server...")

	// Step 12: graceful shutdown with 5 second timeout
	// Shutdown() stops accepting new requests
	// waits for in-flight requests to complete
	// returns after all connections are closed or timeout exceeded
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("FATAL: server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited cleanly")
}