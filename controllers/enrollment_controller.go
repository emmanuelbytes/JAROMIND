// controllers/enrollment_controller.go
package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	// "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	
	// "github.com/AbaraEmmanuel/jaromind-backend/database"
	"github.com/AbaraEmmanuel/jaromind-backend/models"
)

// ================================================
// ENROLLMENT ENDPOINTS
// ================================================

// CreateEnrollment - Main endpoint for course enrollment (matches frontend POST /enrollments)
func CreateEnrollment(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get authenticated user ID from JWT middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authentication required",
		})
		return
	}

	// Parse enrollment request
	var req models.EnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid request: %v", err.Error()),
		})
		return
	}

	// Validate terms acceptance
	if !req.TermsAccepted {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "You must accept the terms and conditions",
		})
		return
	}

	// Convert userID to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	// Check if course exists
	var course bson.M
	err = getCoursesCollection().FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"id": req.CourseID},
			{"_id": convertToObjectID(req.CourseID)},
		},
		"isActive": true,
	}).Decode(&course)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Course not found or inactive",
		})
		return
	}

	// Check if user is already enrolled
	existingEnrollment := getEnrollmentsCollection().FindOne(ctx, bson.M{
		"userId":   userObjectID,
		"courseId": req.CourseID,
		"status":   bson.M{"$in": []string{"active", "pending"}},
	})

	if existingEnrollment.Err() == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "You are already enrolled in this course",
		})
		return
	}

	// Get course price
	coursePrice := 0.0
	if price, ok := course["price"].(float64); ok {
		coursePrice = price
	} else if price, ok := course["price"].(int32); ok {
		coursePrice = float64(price)
	}

	isFree := coursePrice == 0

	// Determine payment status
	paymentStatus := "free"
	enrollmentStatus := "active"
	if !isFree {
		paymentStatus = "pending"
		enrollmentStatus = "pending" // Requires payment confirmation
	}

	// Create enrollment
	enrollment := models.Enrollment{
		ID:               primitive.NewObjectID(),
		EnrollmentID:     uuid.New().String(),
		UserID:           userObjectID,
		CourseID:         req.CourseID,
		FullName:         req.FullName,
		Email:            req.Email,
		Phone:            req.Phone,
		Education:        req.Education,
		Experience:       req.Experience,
		LearningGoal:     req.LearningGoal,
		Schedule:         req.Schedule,
		StudyTime:        req.StudyTime,
		ReferralSource:   req.ReferralSource,
		PaymentMethod:    req.PaymentMethod,
		PaymentStatus:    paymentStatus,
		Amount:           coursePrice,
		TermsAccepted:    req.TermsAccepted,
		Status:           enrollmentStatus,
		EnrolledAt:       time.Now(),
		Progress:         0,
		CompletedLessons: []string{},
		LastAccessedAt:   time.Now(),
		CertificateIssued: false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Insert enrollment
	result, err := getEnrollmentsCollection().InsertOne(ctx, enrollment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create enrollment",
		})
		return
	}

	enrollment.ID = result.InsertedID.(primitive.ObjectID)

	// Update course enrollment count (only if active enrollment)
	if enrollmentStatus == "active" {
		updateCourseEnrollmentCount(req.CourseID, 1)
	}

	// Prepare response
	response := models.EnrollmentResponse{
		Success:    true,
		Message:    "Enrollment created successfully",
		Enrollment: &enrollment,
		NextSteps:  []string{},
	}

	if isFree {
		response.NextSteps = []string{
			"Start learning immediately",
			"Access your course dashboard",
			"Complete lessons at your own pace",
		}
	} else {
		response.NextSteps = []string{
			"Complete payment to activate enrollment",
			"Check your email for payment instructions",
			"Access course after payment confirmation",
		}
		// TODO: Generate payment URL based on payment gateway
		// response.PaymentURL = generatePaymentURL(enrollment)
	}

	c.JSON(http.StatusCreated, response)
}

// GetAllEnrollments - Admin: Get all enrollments with filters
func GetAllEnrollments(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Build filter
	filter := bson.M{}

	// Filter by status
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	// Filter by course
	if courseID := c.Query("courseId"); courseID != "" {
		filter["courseId"] = courseID
	}

	// Filter by payment status
	if paymentStatus := c.Query("paymentStatus"); paymentStatus != "" {
		filter["paymentStatus"] = paymentStatus
	}

	// Sorting
	sortBy := c.DefaultQuery("sortBy", "enrolledAt")
	order := c.DefaultQuery("order", "desc")
	sortOrder := 1
	if order == "desc" {
		sortOrder = -1
	}

	opts := options.Find().SetSort(bson.D{{Key: sortBy, Value: sortOrder}})

	// Pagination
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")
	
	// TODO: Implement pagination
	
	cursor, err := getEnrollmentsCollection().Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch enrollments",
		})
		return
	}
	defer cursor.Close(ctx)

	var enrollments []models.Enrollment
	if err = cursor.All(ctx, &enrollments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to parse enrollments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"enrollments": enrollments,
		"count":       len(enrollments),
		"page":        page,
		"limit":       limit,
	})
}

// GetEnrollmentByID - Get single enrollment details
func GetEnrollmentByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	enrollmentID := c.Param("id")
	userID, _ := c.Get("userID")

	// Try to find by enrollmentId (UUID) or _id (ObjectID)
	filter := bson.M{
		"$or": []bson.M{
			{"enrollmentId": enrollmentID},
			{"_id": convertToObjectID(enrollmentID)},
		},
	}

	// Non-admin users can only see their own enrollments
	// TODO: Add role-based access control
	// if userRole != "admin" {
	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))
	filter["userId"] = userObjectID
	// }

	var enrollment models.Enrollment
	err := getEnrollmentsCollection().FindOne(ctx, filter).Decode(&enrollment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Enrollment not found",
		})
		return
	}

	// Get associated course details
	var course bson.M
	getCoursesCollection().FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"id": enrollment.CourseID},
			{"_id": convertToObjectID(enrollment.CourseID)},
		},
	}).Decode(&course)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"enrollment": enrollment,
		"course":     course,
	})
}

// GetUserEnrollmentsNew - Get current user's enrollments (alternative to existing one)
// GetUserEnrollmentsNew - OPTIMIZED VERSION
func GetUserEnrollmentsNew(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	// Filter by status if provided
	filter := bson.M{"userId": userObjectID}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	} else {
		// By default, show active and pending enrollments
		filter["status"] = bson.M{"$in": []string{"active", "pending"}}
	}

	cursor, err := getEnrollmentsCollection().Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "lastAccessedAt", Value: -1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch enrollments",
		})
		return
	}
	defer cursor.Close(ctx)

	var enrollments []models.Enrollment
	if err = cursor.All(ctx, &enrollments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to parse enrollments",
		})
		return
	}

	// ✅ OPTIMIZATION: Collect all course IDs first
	courseIDs := make([]string, 0, len(enrollments))
	courseObjectIDs := make([]primitive.ObjectID, 0, len(enrollments))
	
	for _, enrollment := range enrollments {
		courseIDs = append(courseIDs, enrollment.CourseID)
		if objID := convertToObjectID(enrollment.CourseID); objID != primitive.NilObjectID {
			courseObjectIDs = append(courseObjectIDs, objID)
		}
	}

	// ✅ OPTIMIZATION: Fetch ALL courses in ONE query (instead of N queries)
	courseCursor, err := getCoursesCollection().Find(ctx, bson.M{
		"$or": []bson.M{
			{"id": bson.M{"$in": courseIDs}},
			{"_id": bson.M{"$in": courseObjectIDs}},
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch courses",
		})
		return
	}
	defer courseCursor.Close(ctx)

	// ✅ Build a map for quick lookup
	coursesMap := make(map[string]bson.M)
	for courseCursor.Next(ctx) {
		var course bson.M
		if err := courseCursor.Decode(&course); err != nil {
			continue
		}
		
		// Map by both id and _id for flexibility
		if id, ok := course["id"].(string); ok {
			coursesMap[id] = course
		}
		if objID, ok := course["_id"].(primitive.ObjectID); ok {
			coursesMap[objID.Hex()] = course
		}
	}

	// ✅ Build enrollment summaries using the map (fast lookups!)
	var enrollmentSummaries []models.EnrollmentSummary
	for _, enrollment := range enrollments {
		summary := models.EnrollmentSummary{
			Enrollment: enrollment,
		}

		// Look up course from map (O(1) instead of database query!)
		if course, found := coursesMap[enrollment.CourseID]; found {
			delete(course, "_id") // Remove MongoDB _id
			summary.Course = course
		} else if objID := convertToObjectID(enrollment.CourseID); objID != primitive.NilObjectID {
			if course, found := coursesMap[objID.Hex()]; found {
				delete(course, "_id")
				summary.Course = course
			}
		}

		enrollmentSummaries = append(enrollmentSummaries, summary)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"enrollments": enrollmentSummaries,
		"count":       len(enrollmentSummaries),
	})
}

// UpdateEnrollmentStatus - Update enrollment status (admin or payment callback)
func UpdateEnrollmentStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	enrollmentID := c.Param("id")

	var request struct {
		Status          string `json:"status" binding:"required"`
		PaymentStatus   string `json:"paymentStatus"`
		TransactionID   string `json:"transactionId"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate status
	validStatuses := []string{"active", "pending", "completed", "cancelled"}
	isValid := false
	for _, status := range validStatuses {
		if request.Status == status {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid status",
		})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"status":    request.Status,
			"updatedAt": time.Now(),
		},
	}

	if request.PaymentStatus != "" {
		update["$set"].(bson.M)["paymentStatus"] = request.PaymentStatus
	}

	if request.TransactionID != "" {
		update["$set"].(bson.M)["transactionId"] = request.TransactionID
	}

	// If status is being set to active from pending, update enrollment count
	var oldEnrollment models.Enrollment
	getEnrollmentsCollection().FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"enrollmentId": enrollmentID},
			{"_id": convertToObjectID(enrollmentID)},
		},
	}).Decode(&oldEnrollment)

	result, err := getEnrollmentsCollection().UpdateOne(
		ctx,
		bson.M{
			"$or": []bson.M{
				{"enrollmentId": enrollmentID},
				{"_id": convertToObjectID(enrollmentID)},
			},
		},
		update,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update enrollment",
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Enrollment not found",
		})
		return
	}

	// Update course enrollment count if status changed to active
	if oldEnrollment.Status == "pending" && request.Status == "active" {
		updateCourseEnrollmentCount(oldEnrollment.CourseID, 1)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Enrollment status updated successfully",
	})
}

// UpdateEnrollmentProgress - Update progress (existing functionality)
func UpdateEnrollmentProgress(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	enrollmentID := c.Param("id")
	userID, exists := c.Get("userID")
	
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var request struct {
		Progress         int      `json:"progress" binding:"min=0,max=100"`
		CompletedLessons []string `json:"completedLessons"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))

	update := bson.M{
		"$set": bson.M{
			"progress":         request.Progress,
			"completedLessons": request.CompletedLessons,
			"lastAccessedAt":   time.Now(),
			"updatedAt":        time.Now(),
		},
	}

	// If course is 100% complete
	if request.Progress >= 100 {
		update["$set"].(bson.M)["completedAt"] = time.Now()
		update["$set"].(bson.M)["status"] = "completed"
	}

	result, err := getEnrollmentsCollection().UpdateOne(
		ctx,
		bson.M{
			"userId": userObjectID,
			"$or": []bson.M{
				{"enrollmentId": enrollmentID},
				{"_id": convertToObjectID(enrollmentID)},
			},
		},
		update,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update progress",
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Enrollment not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Progress updated successfully",
	})
}

// CancelEnrollment - Cancel an enrollment
func CancelEnrollment(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	enrollmentID := c.Param("id")
	userID, exists := c.Get("userID")
	
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))

	// Get enrollment to check if it can be cancelled
	var enrollment models.Enrollment
	err := getEnrollmentsCollection().FindOne(ctx, bson.M{
		"userId": userObjectID,
		"$or": []bson.M{
			{"enrollmentId": enrollmentID},
			{"_id": convertToObjectID(enrollmentID)},
		},
	}).Decode(&enrollment)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Enrollment not found",
		})
		return
	}

	// Check if already cancelled or completed
	if enrollment.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Enrollment is already cancelled",
		})
		return
	}

	if enrollment.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Cannot cancel completed enrollment",
		})
		return
	}

	// Update status to cancelled
	_, err = getEnrollmentsCollection().UpdateOne(
		ctx,
		bson.M{"_id": enrollment.ID},
		bson.M{
			"$set": bson.M{
				"status":    "cancelled",
				"updatedAt": time.Now(),
			},
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to cancel enrollment",
		})
		return
	}

	// Decrement enrollment count if was active
	if enrollment.Status == "active" {
		updateCourseEnrollmentCount(enrollment.CourseID, -1)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Enrollment cancelled successfully",
	})
}

// ================================================
// HELPER FUNCTIONS
// ================================================

// convertToObjectID safely converts string to ObjectID
func convertToObjectID(id string) primitive.ObjectID {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID
	}
	return objID
}

// updateCourseEnrollmentCount updates the enrollment count for a course
func updateCourseEnrollmentCount(courseID string, delta int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to update by id field first
	result, err := getCoursesCollection().UpdateOne(
		ctx,
		bson.M{"id": courseID},
		bson.M{"$inc": bson.M{"enrollmentCount": delta}},
	)

	// If not found by id, try by _id
	if err != nil || result.MatchedCount == 0 {
		if objID := convertToObjectID(courseID); objID != primitive.NilObjectID {
			getCoursesCollection().UpdateOne(
				ctx,
				bson.M{"_id": objID},
				bson.M{"$inc": bson.M{"enrollmentCount": delta}},
			)
		}
	}
}

// TODO: Implement payment gateway integration
// func generatePaymentURL(enrollment models.Enrollment) string {
// 	// Integrate with Paystack, Flutterwave, etc.
// 	return ""
// }