package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Add this above JWTAuthMiddleware
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// ⚠️ MUST MATCH THE FALLBACK IN utils/jwt.go!
		return "your-jwt-secret-key-change-in-production"
	}
	return secret
}

// And update the jwtSecret variable:
var jwtSecret = []byte(getJWTSecret())
// var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// ✅ Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token: " + err.Error()})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Token expiry check (extra layer)
		if exp, ok := claims["exp"].(float64); ok {
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
				c.Abort()
				return
			}
		}

		// ✅ FIX: Handle BOTH token formats
		// 1. Try "user_id" (new format - lowercase, used by admins)
		// 2. Try "user_Id" (old format - capital I, used by existing students)
		var userID interface{}
		
		if uid, ok := claims["user_id"]; ok {
			// New format (admins and new student tokens)
			userID = uid
		} else if uid, ok := claims["user_Id"]; ok {
			// Old format (existing student tokens - backward compatibility)
			userID = uid
		} else {
			// No user ID found in token
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing user ID"})
			c.Abort()
			return
		}

		// ✅ Get email
		email, _ := claims["email"].(string)
		
		// ✅ Get role with default
		role := "user"
		if r, ok := claims["role"].(string); ok {
			role = r
		}

		// ✅ Save user info in Gin context
		c.Set("userID", userID)  // Change from "user_id" to "userID"
		c.Set("userEmail", email)
		c.Set("userRole", role)
		c.Next()
	}
}

// ✅ AdminAuthMiddleware - Checks if user has admin role
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, check if user is authenticated (call JWTAuthMiddleware logic)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing Authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid Authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token claims",
			})
			c.Abort()
			return
		}

		// ✅ Check if user has admin role
		role, ok := claims["role"].(string)
		if !ok || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Admin access required",
				"hint":    "Your role: " + role,
			})
			c.Abort()
			return
		}

		// ✅ Set user info in context (same as JWTAuthMiddleware)
		c.Set("user_id", claims["user_id"])
		c.Set("email", claims["email"])
		c.Set("role", role)

		c.Next()
	}
}

// ✅ Optional: Combined middleware that does both
func JWTAuthWithAdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First validate JWT
		JWTAuthMiddleware()(c)
		
		// If context was aborted by JWTAuthMiddleware, return
		if c.IsAborted() {
			return
		}
		
		// Then check for admin role
		role, exists := c.Get("role")
		if !exists || role.(string) != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Admin privileges required",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}