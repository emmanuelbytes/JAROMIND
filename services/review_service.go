package services

import (
    "context"
    "github.com/AbaraEmmanuel/jaromind-backend/models"
)

// ReviewService defines the interface for review operations
type ReviewService interface {
	// CreateReview creates a new review for a course
	CreateReview(ctx context.Context, review *models.Review) (*models.Review, error)
	
	// GetReviewsByCourseid retrieves all reviews for a specific course
	GetReviewsByCourseID(ctx context.Context, courseID string) ([]models.Review, error)
	
	// GetReviewByID retrieves a single review by ID
	GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error)
	
	// UpdateReview updates an existing review
	UpdateReview(ctx context.Context, reviewID string, review *models.Review) (*models.Review, error)
	
	// DeleteReview deletes a review
	DeleteReview(ctx context.Context, reviewID string) error
	
	// GetReviewByUserAndCourse checks if user already reviewed a course
	GetReviewByUserAndCourse(ctx context.Context, userID, courseID string) (*models.Review, error)
	
	// CalculateCourseRating calculates average rating for a course
	CalculateCourseRating(ctx context.Context, courseID string) (float64, int, error)
}