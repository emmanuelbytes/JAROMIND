// router/routes.go - UPDATED WITH ENROLLMENT ENDPOINTS
package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/AbaraEmmanuel/jaromind-backend/controllers"
	"github.com/AbaraEmmanuel/jaromind-backend/middleware"
)

func RegisterRoutes(router *gin.Engine) {
	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:8000",
			"http://localhost:8001",
			"http://localhost:8003",
			"http://127.0.0.1:8003",
			"http://127.0.0.1:5500",
			"http://localhost:3000",
			"http://localhost:5500",
			"https://edu-tech-v1-mu.vercel.app",
			"https://course-management-portal.vercel.app",
			"https://upload.jaromind.com",
			"https://jaromind.com",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"time":    time.Now().Unix(),
			"service": "Jaromind Backend API",
			"version": "1.0.0",
		})
	})

	// ======================
	// PUBLIC ROUTES
	// ======================
	
	// Auth routes
	router.POST("/register", controllers.RegisterUser)
	router.POST("/login", controllers.LoginUser)
	router.POST("/admin/login", controllers.AdminLogin)
	
	// Public course routes
	router.GET("/courses", controllers.GetAllCourses)
	router.GET("/courses/:id", controllers.GetCourseByID)
	router.GET("/courses/:id/stats", controllers.GetCourseStats)
	router.GET("/courses/:id/reviews", controllers.GetCourseReviews)
	router.GET("/courses/:id/rating", controllers.GetCourseRating)

	// ======================
	// PROTECTED USER ROUTES
	// ======================
	userProtected := router.Group("/")
	userProtected.Use(middleware.JWTAuthMiddleware())
	{
		// User profile
		userProtected.GET("/user/profile", controllers.GetProfile)
		
		// ✅ NEW ENROLLMENT ENDPOINTS (matches frontend)
		userProtected.POST("/enrollments", controllers.CreateEnrollment)           // Main enrollment endpoint
		userProtected.GET("/enrollments", controllers.GetUserEnrollmentsNew)        // Get user's enrollments
		userProtected.GET("/enrollments/:id", controllers.GetEnrollmentByID)        // Get single enrollment
		userProtected.PUT("/enrollments/:id/progress", controllers.UpdateEnrollmentProgress) // Update progress
		userProtected.PUT("/enrollments/:id/status", controllers.UpdateEnrollmentStatus)     // Update status
		userProtected.DELETE("/enrollments/:id", controllers.CancelEnrollment)      // Cancel enrollment
		
		// Legacy enrollment endpoint (keep for backward compatibility)
		userProtected.POST("/user/enroll/:id", controllers.EnrollInCourse)
		userProtected.GET("/user/enrollments", controllers.GetUserEnrollments)
		userProtected.PUT("/courses/:id/progress", controllers.UpdateProgress)

		// Reviews
		userProtected.POST("/courses/:id/review", controllers.CreateReview)
	}

	// ======================
	// REVIEW-SPECIFIC PROTECTED ROUTES
	// ======================
	reviewProtected := router.Group("/reviews")
	reviewProtected.Use(middleware.JWTAuthMiddleware())
	{
		reviewProtected.GET("/:reviewId", controllers.GetReview)
		reviewProtected.PUT("/:reviewId", controllers.UpdateReview)
		reviewProtected.DELETE("/:reviewId", controllers.DeleteReview)
	}

	// ======================
	// ADMIN ROUTES
	// ======================
	adminProtected := router.Group("/admin")
	adminProtected.Use(middleware.JWTAuthMiddleware())
	// adminProtected.Use(middleware.AdminAuthMiddleware()) // Uncomment when ready
	{
		// Course management
		adminProtected.POST("/courses", controllers.CreateCourse)
		adminProtected.PUT("/courses/:id", controllers.UpdateCourse)
		adminProtected.DELETE("/courses/:id", controllers.DeleteCourse)
		
		// ✅ ADMIN ENROLLMENT MANAGEMENT
		adminProtected.GET("/enrollments", controllers.GetAllEnrollments)           // Get all enrollments with filters
		adminProtected.GET("/enrollments/:id", controllers.GetEnrollmentByID)       // Get single enrollment (admin view)
		adminProtected.PUT("/enrollments/:id/status", controllers.UpdateEnrollmentStatus) // Update enrollment status
		
		// Future admin endpoints (uncomment when implemented)
		/*
		adminProtected.GET("/dashboard", controllers.GetAdminDashboard)
		adminProtected.GET("/users", controllers.GetAllUsers)
		adminProtected.GET("/analytics", controllers.GetAnalytics)
		adminProtected.GET("/enrollments/stats", controllers.GetEnrollmentStats)
		*/
	}

	// ======================
	// WEBHOOK ROUTES (for payment callbacks)
	// ======================
	webhook := router.Group("/webhook")
	{
		// These don't need authentication as they come from payment providers
		// but should verify signatures/secrets
		webhook.POST("/paystack", func(c *gin.Context) {
			// TODO: Implement Paystack webhook handler
			c.JSON(200, gin.H{"status": "received"})
		})
		
		webhook.POST("/flutterwave", func(c *gin.Context) {
			// TODO: Implement Flutterwave webhook handler
			c.JSON(200, gin.H{"status": "received"})
		})
	}

	// ======================
	// 404 HANDLER
	// ======================
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"success": false,
			"error":   "Route not found",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		})
	})
}