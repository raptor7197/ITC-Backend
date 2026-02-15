package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend-ITC/internal/firebase"
	"backend-ITC/internal/models"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles authentication middleware
type AuthMiddleware struct {
	firebaseClient *firebase.Client
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(fc *firebase.Client) *AuthMiddleware {
	return &AuthMiddleware{
		firebaseClient: fc,
	}
}

// RequireAuth creates a middleware that validates Firebase ID tokens
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization header format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		idToken := tokenParts[1]
		ctx := context.Background()

		// Verify the Firebase ID token
		token, err := m.firebaseClient.VerifyIDToken(ctx, idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Get user info from Firebase Auth
		userRecord, err := m.firebaseClient.GetUser(ctx, token.UID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Failed to retrieve user information",
			})
			c.Abort()
			return
		}

		// Create user object to pass to handlers
		user := &models.User{
			UID:           userRecord.UID,
			Email:         userRecord.Email,
			DisplayName:   userRecord.DisplayName,
			PhotoURL:      userRecord.PhotoURL,
			EmailVerified: userRecord.EmailVerified,
		}

		// Try to get additional user data from Firestore
		doc, err := m.firebaseClient.Firestore.Collection("users").Doc(token.UID).Get(ctx)
		if err == nil && doc.Exists() {
			var firestoreUser models.User
			if err := doc.DataTo(&firestoreUser); err == nil {
				// Merge Firestore data with Auth data
				user.CreatedAt = firestoreUser.CreatedAt
				user.UpdatedAt = firestoreUser.UpdatedAt
				user.LastLoginAt = firestoreUser.LastLoginAt
				user.Provider = firestoreUser.Provider
			}
		}

		// Set user in context for handlers to use
		c.Set("user", user)
		c.Set("uid", token.UID)
		c.Set("token", token)

		c.Next()
	}
}

// OptionalAuth creates a middleware that validates Firebase ID tokens if present
// but allows requests without authentication to proceed
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without user context
			c.Next()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, continue without user context
			c.Next()
			return
		}

		idToken := tokenParts[1]
		ctx := context.Background()

		// Verify the Firebase ID token
		token, err := m.firebaseClient.VerifyIDToken(ctx, idToken)
		if err != nil {
			// Invalid token, continue without user context
			c.Next()
			return
		}

		// Get user info from Firebase Auth
		userRecord, err := m.firebaseClient.GetUser(ctx, token.UID)
		if err != nil {
			c.Next()
			return
		}

		// Create user object
		user := &models.User{
			UID:           userRecord.UID,
			Email:         userRecord.Email,
			DisplayName:   userRecord.DisplayName,
			PhotoURL:      userRecord.PhotoURL,
			EmailVerified: userRecord.EmailVerified,
		}

		// Set user in context
		c.Set("user", user)
		c.Set("uid", token.UID)
		c.Set("token", token)

		c.Next()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
