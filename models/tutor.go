// models/tutor.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TutorProfile is the standalone bookable tutor document stored in the "tutors" collection.
// Named TutorProfile to avoid conflict with the embedded Tutor struct in course.go.
type TutorProfile struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TutorID      string             `json:"tutorId" bson:"tutorId"` // UUID — used as the public identifier
	UserID       primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"` // Linked user account (optional)

	// Profile
	Name      string   `json:"name" bson:"name"`
	Email     string   `json:"email" bson:"email"`
	Phone     string   `json:"phone,omitempty" bson:"phone,omitempty"`
	AvatarURL string   `json:"avatarUrl,omitempty" bson:"avatarUrl,omitempty"`
	Bio       string   `json:"bio,omitempty" bson:"bio,omitempty"`
	Subjects  []string `json:"subjects" bson:"subjects"`
	Tags      []string `json:"tags" bson:"tags"` // e.g. ["Calculus", "Algebra"]

	// Rates & Availability
	HourlyRate float64 `json:"hourlyRate" bson:"hourlyRate"`
	IsOnline   bool    `json:"isOnline" bson:"isOnline"`
	IsActive   bool    `json:"isActive" bson:"isActive"`

	// Stats (updated on booking / review events)
	Rating       float64 `json:"rating" bson:"rating"`
	ReviewCount  int     `json:"reviewCount" bson:"reviewCount"`
	SessionCount int     `json:"sessionCount" bson:"sessionCount"`

	// Metadata
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

// TutorBooking represents a single booked tutoring session.
type TutorBooking struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BookingID string             `json:"bookingId" bson:"bookingId"` // UUID

	// Parties
	UserID  primitive.ObjectID `json:"userId" bson:"userId"`
	TutorID string             `json:"tutorId" bson:"tutorId"` // TutorProfile.TutorID

	// Session details
	SessionType  string    `json:"sessionType" bson:"sessionType"`   // "1:1" | "Group" | "Workshop"
	SessionDate  time.Time `json:"sessionDate" bson:"sessionDate"`
	TimeSlot     string    `json:"timeSlot" bson:"timeSlot"`         // e.g. "10:00 AM"
	DurationMins int       `json:"durationMins" bson:"durationMins"` // default 60

	// Student snapshot
	StudentName  string `json:"studentName" bson:"studentName"`
	StudentEmail string `json:"studentEmail" bson:"studentEmail"`
	StudentPhone string `json:"studentPhone,omitempty" bson:"studentPhone,omitempty"`
	Notes        string `json:"notes,omitempty" bson:"notes,omitempty"`

	// Payment
	Amount        float64 `json:"amount" bson:"amount"`
	PaymentStatus string  `json:"paymentStatus" bson:"paymentStatus"` // "free" | "pending" | "completed" | "failed"
	PaymentMethod string  `json:"paymentMethod,omitempty" bson:"paymentMethod,omitempty"`
	TransactionID string  `json:"transactionId,omitempty" bson:"transactionId,omitempty"`

	// Lifecycle
	Status      string     `json:"status" bson:"status"` // "pending" | "confirmed" | "completed" | "cancelled"
	BookedAt    time.Time  `json:"bookedAt" bson:"bookedAt"`
	CancelledAt *time.Time `json:"cancelledAt,omitempty" bson:"cancelledAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty" bson:"completedAt,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

// ── Request / Response types ──────────────────────────────────────────────────

// BookingRequest is the JSON payload POSTed from the frontend booking page.
type BookingRequest struct {
	TutorID       string `json:"tutorId" binding:"required"`
	SessionType   string `json:"sessionType" binding:"required,oneof=1:1 Group Workshop"`
	SessionDate   string `json:"sessionDate" binding:"required"`   // "YYYY-MM-DD"
	TimeSlot      string `json:"timeSlot" binding:"required"`      // "10:00 AM"
	StudentName   string `json:"studentName" binding:"required"`
	StudentEmail  string `json:"studentEmail" binding:"required,email"`
	StudentPhone  string `json:"studentPhone"`
	Notes         string `json:"notes"`
	PaymentMethod string `json:"paymentMethod"`
}

// BookingResponse is returned after a booking is created.
type BookingResponse struct {
	Success    bool          `json:"success"`
	Message    string        `json:"message"`
	Booking    *TutorBooking `json:"booking,omitempty"`
	NextSteps  []string      `json:"nextSteps,omitempty"`
	PaymentURL string        `json:"paymentUrl,omitempty"`
}

// BookingSummary enriches a TutorBooking with the tutor's full profile for list responses.
type BookingSummary struct {
	Booking TutorBooking `json:"booking"`
	Tutor   interface{}  `json:"tutor"`
}

type TutorApplication struct {
    ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    AppID     string             `json:"appId" bson:"appId"` // UUID

    // Personal
    FirstName string `json:"firstName" bson:"firstName"`
    LastName  string `json:"lastName"  bson:"lastName"`
    Email     string `json:"email"     bson:"email"`
    Phone     string `json:"phone"     bson:"phone"`
    Country   string `json:"country"   bson:"country"`
    Gender    string `json:"gender"    bson:"gender"`
    PhotoURL  string `json:"photoUrl"  bson:"photoUrl"`

    // Teaching
    PrimarySubject string   `json:"primarySubject" bson:"primarySubject"`
    Subjects       []string `json:"subjects"       bson:"subjects"`
    Levels         []string `json:"levels"         bson:"levels"`
    Languages      []LanguageEntry `json:"languages" bson:"languages"`
    YearsExp       string   `json:"yearsExp"       bson:"yearsExp"`

    // Profile
    Headline string `json:"headline" bson:"headline"`
    Bio      string `json:"bio"      bson:"bio"`
    VideoURL string `json:"videoUrl,omitempty" bson:"videoUrl,omitempty"`

    // Rate & Availability
    HourlyRate    float64  `json:"hourlyRate"    bson:"hourlyRate"`
    Timezone      string   `json:"timezone"      bson:"timezone"`
    AvailSlots    []string `json:"availSlots"    bson:"availSlots"`

    // Review
    Status       string     `json:"status" bson:"status"` // "pending" | "approved" | "rejected" | "revision_requested"
    ReviewerNote string     `json:"reviewerNote,omitempty" bson:"reviewerNote,omitempty"`
    ReviewedAt   *time.Time `json:"reviewedAt,omitempty"   bson:"reviewedAt,omitempty"`
    ReviewedBy   string     `json:"reviewedBy,omitempty"   bson:"reviewedBy,omitempty"` // admin email

    CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

type LanguageEntry struct {
    Language string `json:"language" bson:"language"`
    Level    string `json:"level"    bson:"level"`
}

type ApplicationReviewRequest struct {
    Status       string `json:"status" binding:"required,oneof=approved rejected revision_requested"`
    ReviewerNote string `json:"reviewerNote"`
}