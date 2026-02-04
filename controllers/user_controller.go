package controllers

import (
	"context"
	"net/http"
	"time"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/AbaraEmmanuel/jaromind-backend/models"
	"github.com/AbaraEmmanuel/jaromind-backend/database"
	"github.com/AbaraEmmanuel/jaromind-backend/utils"
	servicesimpl "github.com/AbaraEmmanuel/jaromind-backend/services_impl"
	
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"golang.org/x/crypto/bcrypt"
)


// var userService services.UserService

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone" binding:"required"` 
}

func RegisterUser(c *gin.Context) {
	var request RegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	student := models.User{
		Name:     request.Name,
		Email:    request.Email,
		Phone:    request.Phone,
		Password: request.Password,
		// Level will be set to "spark" in the service layer
	}

	err := servicesimpl.NewUserService().Register(student)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration successful! Please check your email to verify your account.",
	})
}

type LoginRequest struct {
	Email    string `bson:"email" binding:"required,email"`
	Password string `bson:"password" binding:"required"`
}

func LoginUser(c *gin.Context) {
	var request LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	token, err := servicesimpl.NewUserService().Login(request.Email, request.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "Login successfully",
		"token":   token,
	})
}

func GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"email":   email,
		"role":    role,
	})
}

func AdminLogin(c *gin.Context) {
	fmt.Println("🔧 AdminLogin called")
	
	var loginData struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		fmt.Println("❌ Bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid input format",
		})
		return
	}

	fmt.Println("📧 Attempting admin login for:", loginData.Email)

	// Get database connection
	client := database.Client
	if client == nil {
		fmt.Println("❌ Database client is nil")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database not connected",
		})
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Use the database name from your .env
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "emmanuelabara265_db_user" // Default fallback
	}
	fmt.Printf("Using database: %s\n", dbName)
	
	db := client.Database(dbName)
	adminsCollection := db.Collection("admins")
	
	// First, check if the admin exists
	var admin bson.M
	err := adminsCollection.FindOne(ctx, bson.M{"email": loginData.Email}).Decode(&admin)
	if err != nil {
		fmt.Printf("❌ Admin not found with email '%s': %v\n", loginData.Email, err)
		
		// Debug: Show all admins that exist
		fmt.Println("\n📋 Checking all admins in database:")
		cursor, err := adminsCollection.Find(ctx, bson.M{})
		if err != nil {
			fmt.Printf("❌ Error finding admins: %v\n", err)
		} else {
			defer cursor.Close(ctx)
			foundAny := false
			for cursor.Next(ctx) {
				foundAny = true
				var result bson.M
				cursor.Decode(&result)
				fmt.Printf("  - Email: %v, Name: %v\n", 
					result["email"], result["name"])
			}
			if !foundAny {
				fmt.Println("  No admins found in collection!")
			}
		}
		
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid email or password",
		})
		return
	}
	
	fmt.Printf("✅ Admin found: %v\n", admin)
	
	// Check if admin is active
	if isActive, ok := admin["isActive"].(bool); ok && !isActive {
		fmt.Println("❌ Admin account is deactivated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Admin account is deactivated",
		})
		return
	}
	
	// Verify password
	passwordHash, ok := admin["password"].(string)
	if !ok {
		fmt.Println("❌ Password field missing or not string")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid admin data",
		})
		return
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(loginData.Password))
	if err != nil {
		fmt.Println("❌ Password mismatch")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid email or password",
		})
		return
	}
	
	fmt.Println("✅ Password verified")
	
	// Get admin ID - FIXED
	var adminID primitive.ObjectID
	if id, ok := admin["_id"].(primitive.ObjectID); ok {
		adminID = id
	} else {
		fmt.Printf("❌ Cannot get admin ID, type: %T\n", admin["_id"])
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid admin ID format",
		})
		return
	}
	
	// Get admin email - FIXED
	adminEmail, ok := admin["email"].(string)
	if !ok {
		fmt.Println("❌ Email field missing or not string")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid admin data",
		})
		return
	}
	
	// Get admin name
	adminName := "Admin"
	if name, ok := admin["name"].(string); ok && name != "" {
		adminName = name
	}
	
	// Generate token - FIXED: Use adminID and adminEmail variables
	token, err := utils.GenerateAdminJWT(adminID.Hex(), adminEmail)
	if err != nil {
		fmt.Println("❌ Token generation failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate token",
		})
		return
	}
	
	fmt.Println("✅ Token generated")
	
	// Return success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user": gin.H{
			"id":    adminID.Hex(),
			"email": adminEmail,
			"name":  adminName,
			"role":  "admin",
		},
	})
	
	fmt.Println("🎉 Admin login successful!")
}


// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}