// router/routes.go
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

	// Auth
	router.POST("/register", controllers.RegisterUser)
	router.POST("/login", controllers.LoginUser)
	router.POST("/admin/login", controllers.AdminLogin)

	// Public course routes
	router.GET("/courses", controllers.GetAllCourses)
	router.GET("/courses/:id", controllers.GetCourseByID)
	router.GET("/courses/:id/stats", controllers.GetCourseStats)
	router.GET("/courses/:id/reviews", controllers.GetCourseReviews)
	router.GET("/courses/:id/rating", controllers.GetCourseRating)

	// Public tutor routes
	router.GET("/tutors", controllers.GetAllTutors)
	router.GET("/tutors/:id", controllers.GetTutorByID)
	router.GET("/tutors/:id/availability", controllers.GetTutorAvailability)

	// Public tutor application (no auth needed to apply)
	router.POST("/apply/tutor", controllers.SubmitTutorApplication)

	// ======================
	// PROTECTED USER ROUTES
	// ======================
	userProtected := router.Group("/")
	userProtected.Use(middleware.JWTAuthMiddleware())
	{
		// User profile
		userProtected.GET("/user/profile", controllers.GetProfile)

		// Enrollment endpoints
		userProtected.POST("/enrollments", controllers.CreateEnrollment)
		userProtected.GET("/enrollments", controllers.GetUserEnrollmentsNew)
		userProtected.GET("/enrollments/:id", controllers.GetEnrollmentByID)
		userProtected.PUT("/enrollments/:id/progress", controllers.UpdateEnrollmentProgress)
		userProtected.PUT("/enrollments/:id/status", controllers.UpdateEnrollmentStatus)
		userProtected.DELETE("/enrollments/:id", controllers.CancelEnrollment)

		// Legacy enrollment endpoints (backward compatibility)
		userProtected.POST("/user/enroll/:id", controllers.EnrollInCourse)
		userProtected.GET("/user/enrollments", controllers.GetUserEnrollments)
		userProtected.PUT("/courses/:id/progress", controllers.UpdateProgress)

		// Reviews
		userProtected.POST("/courses/:id/review", controllers.CreateReview)

		// Tutor Bookings
		userProtected.POST("/bookings", controllers.CreateBooking)
		userProtected.GET("/bookings", controllers.GetUserBookings)
		userProtected.GET("/bookings/:id", controllers.GetBookingByID)
		userProtected.DELETE("/bookings/:id", controllers.CancelBooking)
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

		// Enrollment management
		adminProtected.GET("/enrollments", controllers.GetAllEnrollments)
		adminProtected.GET("/enrollments/:id", controllers.GetEnrollmentByID)
		adminProtected.PUT("/enrollments/:id/status", controllers.UpdateEnrollmentStatus)

		// Tutor management
		adminProtected.POST("/tutors", controllers.AdminCreateTutor)
		adminProtected.PUT("/tutors/:id", controllers.AdminUpdateTutor)
		adminProtected.DELETE("/tutors/:id", controllers.AdminDeleteTutor)

		// Booking management
		adminProtected.GET("/bookings", controllers.AdminGetAllBookings)
		adminProtected.PUT("/bookings/:id/status", controllers.AdminUpdateBookingStatus)

		// Tutor application review  ← fixed: moved inside adminProtected group
		adminProtected.GET("/applications", controllers.AdminGetApplications)
		adminProtected.GET("/applications/stats", controllers.AdminApplicationStats)
		adminProtected.GET("/applications/:id", controllers.AdminGetApplication)
		adminProtected.PUT("/applications/:id/review", controllers.AdminReviewApplication)
	}

	// ======================
	// WEBHOOK ROUTES
	// ======================
	webhook := router.Group("/webhook")
	{
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