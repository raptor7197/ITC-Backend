package router

import (
	"backend-ITC/internal/config"
	"backend-ITC/internal/firebase"
	"backend-ITC/internal/handlers"
	"backend-ITC/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Setup initializes and returns the Gin router with all routes
func Setup(cfg *config.Config, fc *firebase.Client) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Configure CORS
	corsConfig := cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}

	// In development, allow all origins
	if cfg.IsDevelopment() {
		corsConfig.AllowOrigins = []string{"*"}
		corsConfig.AllowCredentials = false
	}

	r.Use(cors.New(corsConfig))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(fc)
	registrationHandler := handlers.NewRegistrationHandler(fc)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(fc)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "Conference API is running",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/google", authHandler.GoogleLogin)
			auth.POST("/verify", authHandler.VerifyToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			// User routes
			protected.GET("/me", authHandler.GetCurrentUser)

			// Registration routes
			registrations := protected.Group("/registrations")
			{
				registrations.POST("", registrationHandler.CreateRegistration)
				registrations.GET("/me", registrationHandler.GetMyRegistration)
				registrations.PUT("/me", registrationHandler.UpdateRegistration)
				registrations.DELETE("/me", registrationHandler.DeleteRegistration)
			}
		}

		// Admin routes (add admin middleware as needed)
		admin := v1.Group("/admin")
		admin.Use(authMiddleware.RequireAuth())
		{
			admin.GET("/registrations", registrationHandler.GetAllRegistrations)
		}
	}

	return r
}
