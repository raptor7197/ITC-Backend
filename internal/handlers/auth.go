package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"backend-ITC/internal/firebase"
	"backend-ITC/internal/models"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	firebaseClient *firebase.Client
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(fc *firebase.Client) *AuthHandler {
	return &AuthHandler{
		firebaseClient: fc,
	}
}

// GoogleAuthRequest represents the request body for Google auth
type GoogleAuthRequest struct {
	IDToken string `json:"idToken" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	User    *models.User `json:"user,omitempty"`
	Token   string       `json:"token,omitempty"`
}

// GoogleLogin handles Google OAuth login
// The frontend should send the Firebase ID token after Google sign-in
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req GoogleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request: idToken is required",
		})
		return
	}

	ctx := context.Background()

	// Verify the Firebase ID token
	token, err := h.firebaseClient.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid or expired token",
		})
		return
	}

	// Get user info from Firebase Auth
	userRecord, err := h.firebaseClient.GetUser(ctx, token.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to retrieve user information",
		})
		return
	}

	// Create or update user in Firestore
	user := &models.User{
		UID:         userRecord.UID,
		Email:       userRecord.Email,
		DisplayName: userRecord.DisplayName,
		PhotoURL:    userRecord.PhotoURL,
		Provider:    "google.com",
		LastLoginAt: time.Now(),
	}

	// Save user to Firestore
	err = h.saveUserToFirestore(ctx, user)
	if err != nil {
		// Log error but don't fail the login
		// The user is authenticated, we just couldn't save their profile
		c.JSON(http.StatusOK, AuthResponse{
			Success: true,
			Message: "Logged in successfully (profile save pending)",
			User:    user,
			Token:   req.IDToken,
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Logged in successfully",
		User:    user,
		Token:   req.IDToken,
	})
}

// VerifyToken verifies a Firebase ID token from the Authorization header
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Authorization header is required",
		})
		return
	}

	// Extract token from "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid authorization header format",
		})
		return
	}

	idToken := tokenParts[1]
	ctx := context.Background()

	token, err := h.firebaseClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid or expired token",
		})
		return
	}

	// Get user from Firestore
	user, err := h.getUserFromFirestore(ctx, token.UID)
	if err != nil {
		// User not in Firestore, get from Auth
		userRecord, err := h.firebaseClient.GetUser(ctx, token.UID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "Failed to retrieve user information",
			})
			return
		}
		user = &models.User{
			UID:         userRecord.UID,
			Email:       userRecord.Email,
			DisplayName: userRecord.DisplayName,
			PhotoURL:    userRecord.PhotoURL,
		}
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Token is valid",
		User:    user,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Firebase handles token invalidation on the client side
	// This endpoint is mainly for any server-side cleanup if needed
	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// GetCurrentUser returns the current authenticated user's profile
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	user, ok := userVal.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to retrieve user information",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "User retrieved successfully",
		User:    user,
	})
}

// saveUserToFirestore saves or updates a user in Firestore
func (h *AuthHandler) saveUserToFirestore(ctx context.Context, user *models.User) error {
	// Check if user exists
	docRef := h.firebaseClient.Firestore.Collection("users").Doc(user.UID)
	doc, err := docRef.Get(ctx)

	if err != nil || !doc.Exists() {
		// New user - set created timestamp
		user.CreatedAt = time.Now()
	} else {
		// Existing user - preserve created timestamp
		var existingUser models.User
		if err := doc.DataTo(&existingUser); err == nil {
			user.CreatedAt = existingUser.CreatedAt
		}
	}

	user.UpdatedAt = time.Now()

	_, err = docRef.Set(ctx, user)
	return err
}

// getUserFromFirestore retrieves a user from Firestore by UID
func (h *AuthHandler) getUserFromFirestore(ctx context.Context, uid string) (*models.User, error) {
	doc, err := h.firebaseClient.Firestore.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
