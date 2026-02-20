package services_impl

import (
	"context"
	"errors"
	"time"

	"github.com/AbaraEmmanuel/jaromind-backend/database"
	"github.com/AbaraEmmanuel/jaromind-backend/models"
	"github.com/AbaraEmmanuel/jaromind-backend/services"
	"github.com/AbaraEmmanuel/jaromind-backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type userServiceImpl struct {
	collection *mongo.Collection
}

// Constructor
func NewUserService() services.UserService {
	return &userServiceImpl{
		collection: database.GetCollection("students"),
	}
}

// -------- REGISTER METHOD --------
func (s *userServiceImpl) Register(student models.User) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if email already exists
	count, err := s.collection.CountDocuments(ctx, bson.M{"email": student.Email})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(student.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Generate verification code
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return err
	}

	// Create new student object
	newStudent := models.User{
		ID:        primitive.NewObjectID(),
		Name:      student.Name,
		Email:     student.Email,
		Phone:     student.Phone,         // Added phone
		Password:  string(hashedPassword),
		Level:     "spark",     		  // Default level
		Code:      code,
		Verified:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert into DB
	_, err = s.collection.InsertOne(ctx, newStudent)
	if err != nil {
		return err
	}

	return nil
}

func (s *userServiceImpl) Login(email, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// if !user.Verified {
	// 	return "", errors.New("email is not verified")
	// }

	token, err := utils.GenerateJWT(user.ID.Hex(), user.Email)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil

}
