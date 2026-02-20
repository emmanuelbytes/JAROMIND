package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Course struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`  // Changed to ObjectID
	Title           string             `json:"title" bson:"title"`
	Description     string             `json:"description" bson:"description"`
	LongDescription string             `json:"longDescription" bson:"longDescription"`
	Type            string             `json:"type" bson:"type"`
	ClassLevel      string             `json:"classLevel,omitempty" bson:"classLevel,omitempty"`
	Subject         string             `json:"subject,omitempty" bson:"subject,omitempty"`
	Subjects        []string           `json:"subjects" bson:"subjects"`
	ImageUrl        string             `json:"imageUrl" bson:"imageUrl"`
	Status          string             `json:"status" bson:"status"`
	Price           float64            `json:"price" bson:"price"`
	LessonCount     int                `json:"lessonCount" bson:"lessonCount"`
	Duration        string             `json:"duration" bson:"duration"`
	Level           string             `json:"level" bson:"level"`
	IsActive        bool               `json:"isActive" bson:"isActive"`
	IsFeatured      bool               `json:"isFeatured" bson:"isFeatured"`
	EnrollmentCount int                `json:"enrollmentCount" bson:"enrollmentCount"`
	Rating          float64            `json:"rating" bson:"rating"`
	ReviewCount     int                `json:"reviewCount" bson:"reviewCount"`
	Features        []string           `json:"features" bson:"features"`
	Prerequisites   []string           `json:"prerequisites" bson:"prerequisites"`
	LearningGoals   []string           `json:"learningGoals" bson:"learningGoals"`
	Tutor           *Tutor             `json:"tutor" bson:"tutor"`
	Curriculum      []CurriculumItem   `json:"curriculum" bson:"curriculum"`
	Certificate     bool               `json:"certificate" bson:"certificate"`
	Language        string             `json:"language" bson:"language"`
	Category        string             `json:"category" bson:"category"`
	Tags            []string           `json:"tags" bson:"tags"`
	Metadata        *CourseMetadata    `json:"metadata" bson:"metadata"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// Other structs remain the same...
type CourseMetadata struct {
	Code                   string              `json:"code" bson:"code"`
	QuizCount              int                 `json:"quizCount" bson:"quizCount"`
	AssignmentCount        int                 `json:"assignmentCount" bson:"assignmentCount"`
	EnrollmentType         string              `json:"enrollmentType" bson:"enrollmentType"`
	MaxCapacity            int                 `json:"maxCapacity" bson:"maxCapacity"`
	AccessLevel            string              `json:"accessLevel" bson:"accessLevel"`
	AccessPrerequisites    string              `json:"accessPrerequisites" bson:"accessPrerequisites"`
	PromoVideoUrl          string              `json:"promoVideoUrl" bson:"promoVideoUrl"`
	AdditionalMaterials    string              `json:"additionalMaterials" bson:"additionalMaterials"`
	TargetAudience         []string            `json:"targetAudience" bson:"targetAudience"`
	TechnicalRequirements  []string            `json:"technicalRequirements" bson:"technicalRequirements"`
	StartDate              string              `json:"startDate" bson:"startDate"`
	EndDate                string              `json:"endDate" bson:"endDate"`
	ScheduleType           string              `json:"scheduleType" bson:"scheduleType"`
	LiveSessionTimes       string              `json:"liveSessionTimes" bson:"liveSessionTimes"`
	PassingGrade           int                 `json:"passingGrade" bson:"passingGrade"`
	AssessmentMethod       string              `json:"assessmentMethod" bson:"assessmentMethod"`
	GradingPolicy          string              `json:"gradingPolicy" bson:"gradingPolicy"`
	AdditionalNotes        string              `json:"additionalNotes" bson:"additionalNotes"`
	Visibility             string              `json:"visibility" bson:"visibility"`
	Restrictions           *AccessRestrictions `json:"restrictions" bson:"restrictions"`
}

type AccessRestrictions struct {
	Geo          bool `json:"geo" bson:"geo"`
	Age          bool `json:"age" bson:"age"`
	Prerequisite bool `json:"prerequisite" bson:"prerequisite"`
	Device       bool `json:"device" bson:"device"`
}

type Tutor struct {
	Name        string `json:"name" bson:"name"`
	Bio         string `json:"bio" bson:"bio"`
	Avatar      string `json:"avatar" bson:"avatar"`
	Email       string `json:"email" bson:"email"`
	Expertise   string `json:"expertise" bson:"expertise"`
	YearsExp    int    `json:"yearsExp" bson:"yearsExp"`
	Credentials string `json:"credentials" bson:"credentials"`
}

type CurriculumItem struct {
	Week        int      `json:"week,omitempty" bson:"week,omitempty"`
	Title       string   `json:"title" bson:"title"`
	Description string   `json:"description" bson:"description"`
	Topics      []string `json:"topics" bson:"topics"`
	Duration    string   `json:"duration" bson:"duration"`
	Resources   []string `json:"resources,omitempty" bson:"resources,omitempty"`
}

type Enrollment struct {
	ID               string     `json:"id" bson:"id"`
	UserID           string     `json:"userId" bson:"userId"`
	CourseID         string     `json:"courseId" bson:"courseId"`
	EnrolledAt       time.Time  `json:"enrolledAt" bson:"enrolledAt"`
	Progress         int        `json:"progress" bson:"progress"`
	CompletedLessons []string   `json:"completedLessons" bson:"completedLessons"`
	LastAccessedAt   time.Time  `json:"lastAccessedAt" bson:"lastAccessedAt"`
	CompletedAt      *time.Time `json:"completedAt,omitempty" bson:"completedAt,omitempty"`
	CertificateURL   string     `json:"certificateUrl,omitempty" bson:"certificateUrl,omitempty"`
	CreatedAt        time.Time  `json:"createdAt" bson:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt" bson:"updatedAt"`
}
