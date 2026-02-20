// router/routes.go - FIXED VERSION
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
			"http://localhost:8003",          // Admin panel
			"http://127.0.0.1:8003",          // Admin alternative
			"http://127.0.0.1:5500",          // VS Code alternative
			"http://localhost:3000",          // React dev
			"http://localhost:5500",          // VS Code
			"https://edu-tech-v1-mu.vercel.app", // Live web app
			"https://course-management-portal.vercel.app",
			"https://upload.jaromind.com",
			"https://jaromind.com",           // Your domain
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
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	// ======================
	// PUBLIC ROUTES
	// ======================
	
	// Auth routes
	router.POST("/register", controllers.RegisterUser)
	router.POST("/login", controllers.LoginUser)
	router.POST("/admin/login", controllers.AdminLogin) // ✅ Admin login
	
	// Public course routes
	router.GET("/courses", controllers.GetAllCourses)
	router.GET("/courses/:id", controllers.GetCourseByID)
	router.GET("/courses/:id/stats", controllers.GetCourseStats)
	router.GET("/courses/:id/reviews", controllers.GetCourseReviews)
    router.GET("/courses/:id/rating", controllers.GetCourseRating)

	// ======================
	// PROTECTED USER ROUTES
	// ======================
	userProtected := router.Group("/user")
	userProtected.Use(middleware.JWTAuthMiddleware())
	{
		userProtected.GET("/profile", controllers.GetProfile)
		userProtected.POST("/enroll/:id", controllers.EnrollInCourse)
		userProtected.GET("/enrollments", controllers.GetUserEnrollments)
		userProtected.PUT("/courses/:id/progress", controllers.UpdateProgress)

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
	// adminProtected.Use(middleware.AdminAuthMiddleware()) // ⚠️ Comment this out for now
	{
		// Course management (these should exist in course_controller.go)
		adminProtected.POST("/courses", controllers.CreateCourse)
		adminProtected.PUT("/courses/:id", controllers.UpdateCourse)
		adminProtected.DELETE("/courses/:id", controllers.DeleteCourse)
		
		// ❌ REMOVE OR COMMENT THESE LINES - they don't exist yet
		// adminProtected.GET("/dashboard", controllers.GetAdminDashboard)
		// adminProtected.GET("/users", controllers.GetAllUsers)
		
		// ✅ OR keep them commented until you create the functions
		/*
		adminProtected.GET("/dashboard", controllers.GetAdminDashboard)
		adminProtected.GET("/users", controllers.GetAllUsers)
		*/
	}

	// ======================
	// 404 HANDLER
	// ======================
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"success": false,
			"error":   "Route not found",
			"path":    c.Request.URL.Path,
		})
	})
}