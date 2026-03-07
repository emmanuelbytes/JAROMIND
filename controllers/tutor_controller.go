// controllers/tutor_controller.go
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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/AbaraEmmanuel/jaromind-backend/database"
	"github.com/AbaraEmmanuel/jaromind-backend/models"
)

// ── Collection helpers ────────────────────────────────────────────────────────

func getTutorsCollection() *mongo.Collection {
	return database.DB.Collection("tutors")
}

func getTutorBookingsCollection() *mongo.Collection {
	return database.DB.Collection("tutor_bookings")
}

// ── Pricing constants (must stay in sync with booking.js) ────────────────────

var sessionMultipliers = map[string]float64{
	"1:1":      1.0,
	"Group":    0.6,
	"Workshop": 0.4,
}

const (
	platformFee   float64 = 2.0
	promoDiscount float64 = 5.0
)

// ================================================
// PUBLIC TUTOR ENDPOINTS
// ================================================

// GetAllTutors - GET /tutors
// Query params: subject, search, isOnline
func GetAllTutors(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	filter := bson.M{"isActive": true}

	if subject := c.Query("subject"); subject != "" && subject != "All" {
		filter["subjects"] = bson.M{"$in": []string{subject}}
	}

	if search := c.Query("search"); search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"subjects": bson.M{"$regex": search, "$options": "i"}},
			{"tags": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	if isOnline := c.Query("isOnline"); isOnline == "true" {
		filter["isOnline"] = true
	}

	opts := options.Find().SetSort(bson.D{
		{Key: "rating", Value: -1},
		{Key: "sessionCount", Value: -1},
	})

	cursor, err := getTutorsCollection().Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch tutors",
		})
		return
	}
	defer cursor.Close(ctx)

	var tutors []models.TutorProfile
	if err = cursor.All(ctx, &tutors); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to parse tutors",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tutors":  tutors,
		"count":   len(tutors),
	})
}

// GetTutorByID - GET /tutors/:id
func GetTutorByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tutorID := c.Param("id")

	var tutor models.TutorProfile
	err := getTutorsCollection().FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"tutorId": tutorID},
			{"_id": convertToObjectID(tutorID)},
		},
		"isActive": true,
	}).Decode(&tutor)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Tutor not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tutor":   tutor,
	})
}

// GetTutorAvailability - GET /tutors/:id/availability?date=YYYY-MM-DD
// Returns every time slot flagged as available or booked.
func GetTutorAvailability(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tutorID := c.Param("id")
	dateStr := c.Query("date")

	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "date query param is required (YYYY-MM-DD)",
		})
		return
	}

	sessionDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid date format, expected YYYY-MM-DD",
		})
		return
	}

	allSlots := []string{
		"8:00 AM", "9:00 AM", "10:00 AM", "11:00 AM",
		"12:00 PM", "1:00 PM", "2:00 PM", "3:00 PM",
		"4:00 PM", "5:00 PM", "6:00 PM", "7:00 PM",
	}

	startOfDay := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	cursor, err := getTutorBookingsCollection().Find(ctx, bson.M{
		"tutorId":     tutorID,
		"sessionDate": bson.M{"$gte": startOfDay, "$lt": endOfDay},
		"status":      bson.M{"$in": []string{"pending", "confirmed"}},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to check availability",
		})
		return
	}
	defer cursor.Close(ctx)

	bookedSlots := make(map[string]bool)
	for cursor.Next(ctx) {
		var booking models.TutorBooking
		if err := cursor.Decode(&booking); err == nil {
			bookedSlots[booking.TimeSlot] = true
		}
	}

	type SlotInfo struct {
		Slot      string `json:"slot"`
		Available bool   `json:"available"`
	}

	slots := make([]SlotInfo, 0, len(allSlots))
	for _, s := range allSlots {
		slots = append(slots, SlotInfo{
			Slot:      s,
			Available: !bookedSlots[s],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"date":    dateStr,
		"slots":   slots,
	})
}

// ================================================
// PROTECTED BOOKING ENDPOINTS
// ================================================

// CreateBooking - POST /bookings
func CreateBooking(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authentication required",
		})
		return
	}

	var req models.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid request: %v", err.Error()),
		})
		return
	}

	sessionDate, err := time.Parse("2006-01-02", req.SessionDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid sessionDate format, expected YYYY-MM-DD",
		})
		return
	}

	if sessionDate.Before(time.Now().Truncate(24 * time.Hour)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Session date cannot be in the past",
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

	// Verify tutor exists and is active
	var tutor models.TutorProfile
	err = getTutorsCollection().FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"tutorId": req.TutorID},
			{"_id": convertToObjectID(req.TutorID)},
		},
		"isActive": true,
	}).Decode(&tutor)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Tutor not found or unavailable",
		})
		return
	}

	startOfDay := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Check slot is not already taken by another user
	slotTaken := getTutorBookingsCollection().FindOne(ctx, bson.M{
		"tutorId":     req.TutorID,
		"timeSlot":    req.TimeSlot,
		"sessionDate": bson.M{"$gte": startOfDay, "$lt": endOfDay},
		"status":      bson.M{"$in": []string{"pending", "confirmed"}},
	})
	if slotTaken.Err() == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "This time slot is already booked",
		})
		return
	}

	// Prevent same user booking the same tutor + slot twice
	alreadyBooked := getTutorBookingsCollection().FindOne(ctx, bson.M{
		"userId":      userObjectID,
		"tutorId":     req.TutorID,
		"timeSlot":    req.TimeSlot,
		"sessionDate": bson.M{"$gte": startOfDay, "$lt": endOfDay},
		"status":      bson.M{"$in": []string{"pending", "confirmed"}},
	})
	if alreadyBooked.Err() == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "You already have a booking with this tutor at the selected time",
		})
		return
	}

	// Calculate amount
	multiplier, ok := sessionMultipliers[req.SessionType]
	if !ok {
		multiplier = 1.0
	}
	baseAmount := tutor.HourlyRate * multiplier
	totalAmount := baseAmount + platformFee - promoDiscount
	if totalAmount < 0 {
		totalAmount = 0
	}

	isFree := totalAmount == 0
	paymentStatus := "free"
	bookingStatus := "confirmed"
	if !isFree {
		paymentStatus = "pending"
		bookingStatus = "pending"
	}

	booking := models.TutorBooking{
		ID:            primitive.NewObjectID(),
		BookingID:     uuid.New().String(),
		UserID:        userObjectID,
		TutorID:       req.TutorID,
		SessionType:   req.SessionType,
		SessionDate:   sessionDate,
		TimeSlot:      req.TimeSlot,
		DurationMins:  60,
		StudentName:   req.StudentName,
		StudentEmail:  req.StudentEmail,
		StudentPhone:  req.StudentPhone,
		Notes:         req.Notes,
		Amount:        totalAmount,
		PaymentStatus: paymentStatus,
		PaymentMethod: req.PaymentMethod,
		Status:        bookingStatus,
		BookedAt:      time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	result, err := getTutorBookingsCollection().InsertOne(ctx, booking)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create booking",
		})
		return
	}

	booking.ID = result.InsertedID.(primitive.ObjectID)
	updateTutorSessionCount(req.TutorID, 1)

	response := models.BookingResponse{
		Success:   true,
		Message:   "Booking created successfully",
		Booking:   &booking,
		NextSteps: []string{},
	}

	if isFree {
		response.NextSteps = []string{
			"Your session is confirmed",
			"Check your email for session details",
			"Join the session at the scheduled time",
		}
	} else {
		response.NextSteps = []string{
			"Complete payment to confirm your booking",
			"Check your email for payment instructions",
			"Session will be confirmed after payment",
		}
	}

	c.JSON(http.StatusCreated, response)
}

// GetUserBookings - GET /bookings
func GetUserBookings(c *gin.Context) {
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

	filter := bson.M{"userId": userObjectID}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	cursor, err := getTutorBookingsCollection().Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "sessionDate", Value: -1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch bookings",
		})
		return
	}
	defer cursor.Close(ctx)

	var bookings []models.TutorBooking
	if err = cursor.All(ctx, &bookings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to parse bookings",
		})
		return
	}

	// Batch-fetch all tutors in one query (same optimisation as enrollment controller)
	tutorIDs := make([]string, 0, len(bookings))
	for _, b := range bookings {
		tutorIDs = append(tutorIDs, b.TutorID)
	}

	tutorCursor, err := getTutorsCollection().Find(ctx, bson.M{
		"tutorId": bson.M{"$in": tutorIDs},
	})

	tutorsMap := make(map[string]bson.M)
	if err == nil {
		defer tutorCursor.Close(ctx)
		for tutorCursor.Next(ctx) {
			var t bson.M
			if err := tutorCursor.Decode(&t); err == nil {
				if tid, ok := t["tutorId"].(string); ok {
					delete(t, "_id")
					tutorsMap[tid] = t
				}
			}
		}
	}

	summaries := make([]models.BookingSummary, 0, len(bookings))
	for _, b := range bookings {
		summaries = append(summaries, models.BookingSummary{
			Booking: b,
			Tutor:   tutorsMap[b.TutorID],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"bookings": summaries,
		"count":    len(summaries),
	})
}

// GetBookingByID - GET /bookings/:id
func GetBookingByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bookingID := c.Param("id")
	userID, _ := c.Get("userID")
	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))

	var booking models.TutorBooking
	err := getTutorBookingsCollection().FindOne(ctx, bson.M{
		"userId": userObjectID,
		"$or": []bson.M{
			{"bookingId": bookingID},
			{"_id": convertToObjectID(bookingID)},
		},
	}).Decode(&booking)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Booking not found",
		})
		return
	}

	var tutor bson.M
	getTutorsCollection().FindOne(ctx, bson.M{"tutorId": booking.TutorID}).Decode(&tutor)
	if tutor != nil {
		delete(tutor, "_id")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"booking": booking,
		"tutor":   tutor,
	})
}

// CancelBooking - DELETE /bookings/:id
func CancelBooking(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bookingID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))

	var booking models.TutorBooking
	err := getTutorBookingsCollection().FindOne(ctx, bson.M{
		"userId": userObjectID,
		"$or": []bson.M{
			{"bookingId": bookingID},
			{"_id": convertToObjectID(bookingID)},
		},
	}).Decode(&booking)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Booking not found",
		})
		return
	}

	if booking.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Booking is already cancelled",
		})
		return
	}

	if booking.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Cannot cancel a completed booking",
		})
		return
	}

	now := time.Now()
	_, err = getTutorBookingsCollection().UpdateOne(
		ctx,
		bson.M{"_id": booking.ID},
		bson.M{
			"$set": bson.M{
				"status":      "cancelled",
				"cancelledAt": now,
				"updatedAt":   now,
			},
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to cancel booking",
		})
		return
	}

	if booking.Status == "confirmed" {
		updateTutorSessionCount(booking.TutorID, -1)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Booking cancelled successfully",
	})
}

// ================================================
// ADMIN ENDPOINTS
// ================================================

// AdminGetAllBookings - GET /admin/bookings
func AdminGetAllBookings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	filter := bson.M{}

	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	if tutorID := c.Query("tutorId"); tutorID != "" {
		filter["tutorId"] = tutorID
	}
	if paymentStatus := c.Query("paymentStatus"); paymentStatus != "" {
		filter["paymentStatus"] = paymentStatus
	}

	sortBy := c.DefaultQuery("sortBy", "bookedAt")
	order := c.DefaultQuery("order", "desc")
	sortOrder := -1
	if order == "asc" {
		sortOrder = 1
	}

	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	cursor, err := getTutorBookingsCollection().Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: sortBy, Value: sortOrder}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch bookings",
		})
		return
	}
	defer cursor.Close(ctx)

	var bookings []models.TutorBooking
	if err = cursor.All(ctx, &bookings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to parse bookings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"bookings": bookings,
		"count":    len(bookings),
		"page":     page,
		"limit":    limit,
	})
}

// AdminUpdateBookingStatus - PUT /admin/bookings/:id/status
func AdminUpdateBookingStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bookingID := c.Param("id")

	var req struct {
		Status        string `json:"status" binding:"required"`
		PaymentStatus string `json:"paymentStatus"`
		TransactionID string `json:"transactionId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	validStatuses := map[string]bool{
		"pending": true, "confirmed": true,
		"completed": true, "cancelled": true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid status. Must be one of: pending, confirmed, completed, cancelled",
		})
		return
	}

	setFields := bson.M{
		"status":    req.Status,
		"updatedAt": time.Now(),
	}
	if req.PaymentStatus != "" {
		setFields["paymentStatus"] = req.PaymentStatus
	}
	if req.TransactionID != "" {
		setFields["transactionId"] = req.TransactionID
	}
	if req.Status == "completed" {
		now := time.Now()
		setFields["completedAt"] = now
	}

	result, err := getTutorBookingsCollection().UpdateOne(
		ctx,
		bson.M{
			"$or": []bson.M{
				{"bookingId": bookingID},
				{"_id": convertToObjectID(bookingID)},
			},
		},
		bson.M{"$set": setFields},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update booking",
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Booking not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Booking status updated successfully",
	})
}

// AdminCreateTutor - POST /admin/tutors
func AdminCreateTutor(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req struct {
		Name       string   `json:"name" binding:"required"`
		Email      string   `json:"email" binding:"required,email"`
		Phone      string   `json:"phone"`
		Bio        string   `json:"bio"`
		Subjects   []string `json:"subjects" binding:"required"`
		Tags       []string `json:"tags"`
		HourlyRate float64  `json:"hourlyRate" binding:"required,gt=0"`
		AvatarURL  string   `json:"avatarUrl"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid request: %v", err.Error()),
		})
		return
	}

	// Prevent duplicate email
	existing := getTutorsCollection().FindOne(ctx, bson.M{"email": req.Email})
	if existing.Err() == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "A tutor with this email already exists",
		})
		return
	}

	tutor := models.TutorProfile{
		ID:           primitive.NewObjectID(),
		TutorID:      uuid.New().String(),
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		Bio:          req.Bio,
		Subjects:     req.Subjects,
		Tags:         req.Tags,
		HourlyRate:   req.HourlyRate,
		AvatarURL:    req.AvatarURL,
		IsOnline:     false,
		IsActive:     true,
		Rating:       0,
		ReviewCount:  0,
		SessionCount: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := getTutorsCollection().InsertOne(ctx, tutor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create tutor",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Tutor created successfully",
		"tutor":   tutor,
	})
}

// AdminUpdateTutor - PUT /admin/tutors/:id
func AdminUpdateTutor(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tutorID := c.Param("id")

	var req struct {
		Name       string   `json:"name"`
		Phone      string   `json:"phone"`
		Bio        string   `json:"bio"`
		Subjects   []string `json:"subjects"`
		Tags       []string `json:"tags"`
		HourlyRate float64  `json:"hourlyRate"`
		AvatarURL  string   `json:"avatarUrl"`
		IsOnline   *bool    `json:"isOnline"`
		IsActive   *bool    `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	setFields := bson.M{"updatedAt": time.Now()}

	if req.Name != ""         { setFields["name"] = req.Name }
	if req.Phone != ""        { setFields["phone"] = req.Phone }
	if req.Bio != ""          { setFields["bio"] = req.Bio }
	if req.AvatarURL != ""    { setFields["avatarUrl"] = req.AvatarURL }
	if len(req.Subjects) > 0  { setFields["subjects"] = req.Subjects }
	if len(req.Tags) > 0      { setFields["tags"] = req.Tags }
	if req.HourlyRate > 0     { setFields["hourlyRate"] = req.HourlyRate }
	if req.IsOnline != nil    { setFields["isOnline"] = *req.IsOnline }
	if req.IsActive != nil    { setFields["isActive"] = *req.IsActive }

	result, err := getTutorsCollection().UpdateOne(
		ctx,
		bson.M{
			"$or": []bson.M{
				{"tutorId": tutorID},
				{"_id": convertToObjectID(tutorID)},
			},
		},
		bson.M{"$set": setFields},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update tutor",
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Tutor not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tutor updated successfully",
	})
}

// AdminDeleteTutor - DELETE /admin/tutors/:id  (soft delete — sets isActive: false)
func AdminDeleteTutor(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tutorID := c.Param("id")

	result, err := getTutorsCollection().UpdateOne(
		ctx,
		bson.M{
			"$or": []bson.M{
				{"tutorId": tutorID},
				{"_id": convertToObjectID(tutorID)},
			},
		},
		bson.M{
			"$set": bson.M{
				"isActive":  false,
				"updatedAt": time.Now(),
			},
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete tutor",
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Tutor not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tutor deactivated successfully",
	})
}

// ================================================
// HELPERS
// ================================================

func updateTutorSessionCount(tutorID string, delta int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	getTutorsCollection().UpdateOne(
		ctx,
		bson.M{"tutorId": tutorID},
		bson.M{"$inc": bson.M{"sessionCount": delta}},
	)
}