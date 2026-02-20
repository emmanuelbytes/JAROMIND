package controllers

import (
    "net/http"
	"fmt"
    "github.com/AbaraEmmanuel/jaromind-backend/models"
    "github.com/AbaraEmmanuel/jaromind-backend/services_impl"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// ReviewController handles HTTP requests for reviews
type ReviewController struct {
    reviewService *services_impl.ReviewServiceImpl
}

// NewReviewController creates a new review controller
func NewReviewController() *ReviewController {
    return &ReviewController{
        reviewService: services_impl.NewReviewServiceImpl(),
    }
}

// CreateReview handles POST /courses/:courseId/reviews
func CreateReview(ctx *gin.Context) {
    fmt.Println("\n=== CREATE REVIEW - START ===")
    
    // Try multiple key names
    var userID string
    var found bool
    
    keys := []string{"userID", "user_id", "userId", "user"}
    for _, key := range keys {
        if val, exists := ctx.Get(key); exists {
            fmt.Printf("✓ Found key '%s': %v (type: %T)\n", key, val, val)
            if str, ok := val.(string); ok {
                userID = str
                found = true
                break
            }
        }
    }
    
    if !found {
        fmt.Println("✗ ERROR: No user ID found in context")
        ctx.JSON(http.StatusUnauthorized, models.ReviewResponse{
            Success: false,
            Message: "Unauthorized: Please log in to submit a review",
        })
        return
    }
    
    fmt.Printf("✓ Using userID: %s\n", userID)
    
    // Get course ID
    courseID := ctx.Param("id")
    fmt.Printf("✓ Course ID from param: %s\n", courseID)
    
    if courseID == "" {
        fmt.Println("✗ ERROR: Course ID is empty")
        ctx.JSON(http.StatusBadRequest, models.ReviewResponse{
            Success: false,
            Message: "Course ID is required",
        })
        return
    }
    
    // Parse request body
    var input models.ReviewInput
    if err := ctx.ShouldBindJSON(&input); err != nil {
        fmt.Printf("✗ ERROR parsing JSON: %v\n", err)
        ctx.JSON(http.StatusBadRequest, models.ReviewResponse{
            Success: false,
            Message: "Invalid request: " + err.Error(),
        })
        return
    }
    
    fmt.Printf("✓ Parsed input - Rating: %d, Comment: %s\n", input.Rating, input.Comment)
    
    // Validate input
    if input.Rating < 1 || input.Rating > 5 {
        fmt.Printf("✗ ERROR: Invalid rating: %d\n", input.Rating)
        ctx.JSON(http.StatusBadRequest, models.ReviewResponse{
            Success: false,
            Message: "Rating must be between 1 and 5",
        })
        return
    }
    
    // Initialize service
    fmt.Println("✓ Initializing review service...")
    reviewService := services_impl.NewReviewServiceImpl()
    
    // Convert userID string to ObjectID
    fmt.Printf("✓ Converting userID to ObjectID: %s\n", userID)
    userObjectID, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        fmt.Printf("✗ ERROR converting userID to ObjectID: %v\n", err)
        ctx.JSON(http.StatusBadRequest, models.ReviewResponse{
            Success: false,
            Message: "Invalid user ID format",
        })
        return
    }
    
    fmt.Printf("✓ User ObjectID: %v\n", userObjectID)
    
    // Get user name from context (optional)
    userName := ""
    if name, exists := ctx.Get("userName"); exists {
        if str, ok := name.(string); ok {
            userName = str
        }
    }
    if userName == "" {
        userName = "Anonymous"
    }
    fmt.Printf("✓ UserName: %s\n", userName)
    
    // Create review object
    review := &models.Review{
        CourseID:   courseID, // This is a string (UUID)
        UserID:     userObjectID,
        UserName:   userName,
        Rating:     input.Rating,
        Comment:    input.Comment,
    }
    
    fmt.Printf("✓ Created review object: %+v\n", review)
    
    // Create review
    fmt.Println("✓ Calling reviewService.CreateReview...")
    createdReview, err := reviewService.CreateReview(ctx.Request.Context(), review)
    if err != nil {
        fmt.Printf("✗ ERROR from reviewService.CreateReview: %v\n", err)
        ctx.JSON(http.StatusInternalServerError, models.ReviewResponse{
            Success: false,
            Message: "Failed to create review: " + err.Error(),
        })
        return
    }
    
    fmt.Printf("✓ Review created successfully: %+v\n", createdReview)
    fmt.Println("=== CREATE REVIEW - SUCCESS ===")
    
    ctx.JSON(http.StatusCreated, models.ReviewResponse{
        Success: true,
        Message: "Review submitted successfully",
        Review:  createdReview,
    })
}

// GetCourseReviews handles GET /courses/:courseId/reviews
func GetCourseReviews(ctx *gin.Context) {
    reviewService := services_impl.NewReviewServiceImpl()
    
    courseID := ctx.Param("id")
    if courseID == "" {
        ctx.JSON(http.StatusBadRequest, models.ReviewResponse{
            Success: false,
            Message: "Course ID is required",
        })
        return
    }

    reviews, err := reviewService.GetReviewsByCourseID(ctx.Request.Context(), courseID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, models.ReviewResponse{
            Success: false,
            Message: "Failed to retrieve reviews: " + err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, models.ReviewResponse{
        Success: true,
        Reviews: reviews,
    })
}

// GetCourseRating handles GET /courses/:courseId/rating
func GetCourseRating(ctx *gin.Context) {
    reviewService := services_impl.NewReviewServiceImpl()
    
    courseID := ctx.Param("courseId")
    if courseID == "" {
        ctx.JSON(http.StatusBadRequest, models.ReviewResponse{
            Success: false,
            Message: "Course ID is required",
        })
        return
    }

    avgRating, totalReviews, err := reviewService.CalculateCourseRating(ctx.Request.Context(), courseID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, models.ReviewResponse{
            Success: false,
            Message: "Failed to calculate rating: " + err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "success":       true,
        "averageRating": avgRating,
        "totalReviews":  totalReviews,
    })
}

// GetReview handles GET /reviews/:reviewId
func GetReview(ctx *gin.Context) {
    reviewService := services_impl.NewReviewServiceImpl()
    
    reviewID := ctx.Param("reviewId")
    if reviewID == "" {
        ctx.JSON(http.StatusBadRequest, models.ReviewResponse{
            Success: false,
            Message: "Review ID is required",
        })
        return
    }

    review, err := reviewService.GetReviewByID(ctx.Request.Context(), reviewID)
    if err != nil {
        statusCode := http.StatusInternalServerError
        if err.Error() == "review not found" {
            statusCode = http.StatusNotFound
        }
        ctx.JSON(statusCode, models.ReviewResponse{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, models.ReviewResponse{
        Success: true,
        Review:  review,
    })
}

// UpdateReview handles PUT /reviews/:reviewId
func UpdateReview(ctx *gin.Context) {
    reviewService := services_impl.NewReviewServiceImpl()
    
    // Get user from context
    userInterface, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, models.ReviewResponse{
            Success: false,
            Message: "Unauthorized",
        })
        return
    }

    user, ok := userInterface.(map[string]interface{})
    if !ok {
        ctx.JSON(http.StatusUnauthorized, models.ReviewResponse{
            Success: false,
            Message: "Invalid user data",
        })
        return
    }

    // Get review ID
    reviewID := ctx.Param("reviewId")
    if reviewID == "" {
        ctx.JSON(http.StatusBadRequest, models.ReviewResponse{
            Success: false,
            Message: "Review ID is required",
        })
        return
    }

    // Check if review exists and belongs to user
    existingReview, err := reviewService.GetReviewByID(ctx.Request.Context(), reviewID)
    if err != nil {
        statusCode := http.StatusInternalServerError
        if err.Error() == "review not found" {
            statusCode = http.StatusNotFound
        }
        ctx.JSON(statusCode, models.ReviewResponse{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    // Get user ID
    userIDStr := ""
    if id, ok := user["id"].(string); ok {
        userIDStr = id
    } else if id, ok := user["_id"].(string); ok {
        userIDStr = id
    } else if id, ok := user["_id"].(primitive.ObjectID); ok {
        userIDStr = id.Hex()
    }

    // Check if user is the author
    if existingReview.UserID.Hex() != userIDStr {
        ctx.JSON(http.StatusForbidden, models.ReviewResponse{
            Success: false,
            Message: "You can only update your own reviews",
        })
        return
    }

    // Parse request body
    var input models.ReviewInput
    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, models.ReviewResponse{
            Success: false,
            Message: "Invalid request: " + err.Error(),
        })
        return
    }

    // Update review
    existingReview.Rating = input.Rating
    existingReview.Comment = input.Comment

    updatedReview, err := reviewService.UpdateReview(ctx.Request.Context(), reviewID, existingReview)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, models.ReviewResponse{
            Success: false,
            Message: "Failed to update review: " + err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, models.ReviewResponse{
        Success: true,
        Message: "Review updated successfully",
        Review:  updatedReview,
    })
}

// DeleteReview handles DELETE /reviews/:reviewId
func DeleteReview(ctx *gin.Context) {
    reviewService := services_impl.NewReviewServiceImpl()
    
    // Get user from context
    userInterface, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, models.ReviewResponse{
            Success: false,
            Message: "Unauthorized",
        })
        return
    }

    user, ok := userInterface.(map[string]interface{})
    if !ok {
        ctx.JSON(http.StatusUnauthorized, models.ReviewResponse{
            Success: false,
            Message: "Invalid user data",
        })
        return
    }

    // Get review ID
    reviewID := ctx.Param("reviewId")
    if reviewID == "" {
        ctx.JSON(http.StatusBadRequest, models.ReviewResponse{
            Success: false,
            Message: "Review ID is required",
        })
        return
    }

    // Check if review exists and belongs to user
    existingReview, err := reviewService.GetReviewByID(ctx.Request.Context(), reviewID)
    if err != nil {
        statusCode := http.StatusInternalServerError
        if err.Error() == "review not found" {
            statusCode = http.StatusNotFound
        }
        ctx.JSON(statusCode, models.ReviewResponse{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    // Get user ID
    userIDStr := ""
    if id, ok := user["id"].(string); ok {
        userIDStr = id
    } else if id, ok := user["_id"].(string); ok {
        userIDStr = id
    } else if id, ok := user["_id"].(primitive.ObjectID); ok {
        userIDStr = id.Hex()
    }

    // Check if user is the author (or admin - add admin check if needed)
    if existingReview.UserID.Hex() != userIDStr {
        // Add admin check here if you have role-based access
        ctx.JSON(http.StatusForbidden, models.ReviewResponse{
            Success: false,
            Message: "You can only delete your own reviews",
        })
        return
    }

    // Delete review
    if err := reviewService.DeleteReview(ctx.Request.Context(), reviewID); err != nil {
        ctx.JSON(http.StatusInternalServerError, models.ReviewResponse{
            Success: false,
            Message: "Failed to delete review: " + err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, models.ReviewResponse{
        Success: true,
        Message: "Review deleted successfully",
    })
}