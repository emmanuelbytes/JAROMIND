package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Review represents a course review/rating
type Review struct {
    ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    CourseID   string             `json:"courseId" bson:"course_id"` // Change to string
    UserID     primitive.ObjectID `json:"userId" bson:"user_id"`
    UserName   string             `json:"userName" bson:"user_name"`
    Rating     int                `json:"rating" bson:"rating" binding:"required,min=1,max=5"`
    Comment    string             `json:"comment" bson:"comment" binding:"required,min=10,max=1000"`
    Date       time.Time          `json:"date" bson:"date"`
    CreatedAt  time.Time          `json:"createdAt" bson:"created_at"`
    UpdatedAt  time.Time          `json:"updatedAt" bson:"updated_at"`
}

// ReviewInput represents the input for creating a review
type ReviewInput struct {
	// CourseID string `json:"courseId" binding:"required"`
	Rating   int    `json:"rating" binding:"required,min=1,max=5"`
	Comment  string `json:"comment" binding:"required,min=10,max=1000"`
}

// ReviewResponse represents the response structure
type ReviewResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message,omitempty"`
	Review  *Review   `json:"review,omitempty"`
	Reviews []Review  `json:"reviews,omitempty"`
}