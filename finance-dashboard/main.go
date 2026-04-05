package main

import (
	"finance-dashboard/config"
	"finance-dashboard/handlers"
	"finance-dashboard/middleware"
	"finance-dashboard/services"
	"finance-dashboard/store"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	config.Load()






	userStore := store.NewInMemoryUserStore()
	transactionStore := store.NewInMemoryTransactionStore()
	roleRequestStore := store.NewInMemoryRoleRequestStore()


	authService := services.NewAuthService(userStore)
	userService := services.NewUserService(userStore)
	transactionService := services.NewTransactionService(transactionStore)
	dashboardService := services.NewDashboardService(transactionStore)
	roleRequestService := services.NewRoleRequestService(roleRequestStore, userStore)





	if err := userService.SeedAdmin(
		config.App.SeedAdminName,
		config.App.SeedAdminEmail,
		config.App.SeedAdminPassword,
	); err != nil {
		log.Printf("⚠️  seed admin warning: %v", err)
	} else {
		log.Println("✅ Admin seeded successfully")
	}


	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	roleRequestHandler := handlers.NewRoleRequestHandler(roleRequestService)



	rateLimiter := middleware.NewRateLimiter(60, time.Minute)







	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())


	router.Use(rateLimiter.RateLimit())


	api := router.Group("/api/v1")


	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}


	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(userStore))
	{

		users := protected.Group("/users")
		{



			users.GET("/me", middleware.RequireAnyRole(), userHandler.GetMe)
			users.GET("", middleware.RequireAdmin(), userHandler.GetAll)
			users.POST("", middleware.RequireAdmin(), userHandler.Create)
			users.GET("/:id", middleware.RequireAdmin(), userHandler.GetByID)
			users.PUT("/:id", middleware.RequireAdmin(), userHandler.Update)
			users.PATCH("/:id/status", middleware.RequireAdmin(), userHandler.UpdateStatus)
		}


		transactions := protected.Group("/transactions")
		{
			transactions.GET("", middleware.RequireAnalystOrAdmin(), transactionHandler.GetAll)
			transactions.POST("", middleware.RequireAdmin(), transactionHandler.Create)
			transactions.GET("/:id", middleware.RequireAnalystOrAdmin(), transactionHandler.GetByID)
			transactions.PUT("/:id", middleware.RequireAdmin(), transactionHandler.Update)
			transactions.DELETE("/:id", middleware.RequireAdmin(), transactionHandler.Delete)
		}


		dashboard := protected.Group("/dashboard")
		{
			dashboard.GET("/summary", middleware.RequireAnyRole(), dashboardHandler.GetSummary)
			dashboard.GET("/by-category", middleware.RequireAnyRole(), dashboardHandler.GetByCategory)
			dashboard.GET("/trends", middleware.RequireAnyRole(), dashboardHandler.GetTrends)
			dashboard.GET("/recent", middleware.RequireAnyRole(), dashboardHandler.GetRecent)
		}


		roleRequests := protected.Group("/role-requests")
		{

			roleRequests.GET("/mine", middleware.RequireAnyRole(), roleRequestHandler.GetMine)
			roleRequests.POST("", middleware.RequireAnyRole(), roleRequestHandler.Create)
			roleRequests.GET("", middleware.RequireAdmin(), roleRequestHandler.GetAll)
			roleRequests.GET("/:id", middleware.RequireAdmin(), roleRequestHandler.GetByID)
			roleRequests.PATCH("/:id", middleware.RequireAdmin(), roleRequestHandler.Process)
		}
	}


	log.Printf("🚀 Server starting on port %s", config.App.AppPort)
	if err := router.Run(":" + config.App.AppPort); err != nil {
		log.Fatalf("FATAL: failed to start server: %v", err)
	}
}
