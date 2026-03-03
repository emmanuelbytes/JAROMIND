// models/enrollment.go
package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Enrollment represents a student's enrollment in a course
type Enrollment struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EnrollmentID     string             `json:"enrollmentId" bson:"enrollmentId"` // UUID for easy reference
	UserID           primitive.ObjectID `json:"userId" bson:"userId"`
	CourseID         string             `json:"courseId" bson:"courseId"` // Can be UUID or ObjectID hex
	
	// Personal Information
	FullName         string             `json:"fullName" bson:"fullName"`
	Email            string             `json:"email" bson:"email"`
	Phone            string             `json:"phone" bson:"phone"`
	
	// Educational Background
	Education        string             `json:"education" bson:"education"`
	Experience       string             `json:"experience" bson:"experience"`
	
	// Learning Details
	LearningGoal     string             `json:"learningGoal" bson:"learningGoal"`
	Schedule         string             `json:"schedule" bson:"schedule"`         // e.g., "flexible", "weekdays", "weekends"
	StudyTime        string             `json:"studyTime" bson:"studyTime"`       // e.g., "1-2 hours/day"
	
	// Marketing/Referral
	ReferralSource   string             `json:"referralSource" bson:"referralSource"`
	
	// Payment Information
	PaymentMethod    string             `json:"paymentMethod,omitempty" bson:"paymentMethod,omitempty"`
	PaymentStatus    string             `json:"paymentStatus" bson:"paymentStatus"` // "pending", "completed", "failed", "free"
	Amount           float64            `json:"amount" bson:"amount"`
	TransactionID    string             `json:"transactionId,omitempty" bson:"transactionId,omitempty"`
	
	// Terms & Conditions
	TermsAccepted    bool               `json:"termsAccepted" bson:"termsAccepted"`
	
	// Enrollment Status
	Status           string             `json:"status" bson:"status"` // "active", "completed", "cancelled", "pending"
	EnrolledAt       time.Time          `json:"enrolledAt" bson:"enrolledAt"`
	
	// Progress Tracking
	Progress         int                `json:"progress" bson:"progress"`
	CompletedLessons []string           `json:"completedLessons" bson:"completedLessons"`
	LastAccessedAt   time.Time          `json:"lastAccessedAt" bson:"lastAccessedAt"`
	CompletedAt      *time.Time         `json:"completedAt,omitempty" bson:"completedAt,omitempty"`
	
	// Certificate
	CertificateURL   string             `json:"certificateUrl,omitempty" bson:"certificateUrl,omitempty"`
	CertificateIssued bool              `json:"certificateIssued" bson:"certificateIssued"`
	
	// Metadata
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// EnrollmentRequest represents the incoming enrollment request from frontend
type EnrollmentRequest struct {
	CourseID       string `json:"courseId" binding:"required"`
	FullName       string `json:"fullName" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Phone          string `json:"phone" binding:"required"`
	Education      string `json:"education" binding:"required"`
	Experience     string `json:"experience"`
	LearningGoal   string `json:"learningGoal" binding:"required,min=20"`
	Schedule       string `json:"schedule"`
	StudyTime      string `json:"studyTime"`
	ReferralSource string `json:"referralSource"`
	PaymentMethod  string `json:"paymentMethod"`
	TermsAccepted  bool   `json:"termsAccepted" binding:"required"`
}

// EnrollmentResponse represents the response sent back to frontend
type EnrollmentResponse struct {
	Success      bool       `json:"success"`
	Message      string     `json:"message"`
	Enrollment   *Enrollment `json:"enrollment,omitempty"`
	NextSteps    []string   `json:"nextSteps,omitempty"`
	PaymentURL   string     `json:"paymentUrl,omitempty"` // For paid courses
}

// EnrollmentSummary for listing enrollments
type EnrollmentSummary struct {
	Enrollment Enrollment      `json:"enrollment"`
	Course     interface{}     `json:"course"` // Can be Course or bson.M
	User       interface{}     `json:"user,omitempty"`
}