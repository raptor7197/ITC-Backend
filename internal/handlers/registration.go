package handlers

import (
	"context"
	"net/http"
	"time"

	"backend-ITC/internal/models"

	fb "backend-ITC/internal/firebase"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

// RegistrationHandler handles registration related requests
type RegistrationHandler struct {
	firebaseClient *fb.Client
}

// NewRegistrationHandler creates a new registration handler
func NewRegistrationHandler(fc *fb.Client) *RegistrationHandler {
	return &RegistrationHandler{
		firebaseClient: fc,
	}
}

// RegistrationResponse represents the response for registration operations
type RegistrationResponse struct {
	Success       bool                  `json:"success"`
	Message       string                `json:"message"`
	Registration  *models.Registration  `json:"registration,omitempty"`
	Registrations []models.Registration `json:"registrations,omitempty"`
}

// CreateRegistration handles creating a new conference registration
func (h *RegistrationHandler) CreateRegistration(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, RegistrationResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	user, ok := userVal.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, RegistrationResponse{
			Success: false,
			Message: "Failed to retrieve user information",
		})
		return
	}

	var input models.RegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, RegistrationResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	ctx := context.Background()

	// Check if user already has a registration
	existingReg, _ := h.getUserRegistration(ctx, user.UID)
	if existingReg != nil {
		c.JSON(http.StatusConflict, RegistrationResponse{
			Success:      false,
			Message:      "User already has a registration. Please update instead.",
			Registration: existingReg,
		})
		return
	}

	// Create registration
	now := time.Now()
	registration := &models.Registration{
		UserID:           user.UID,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		Email:            input.Email,
		Phone:            input.Phone,
		Organization:     input.Organization,
		JobTitle:         input.JobTitle,
		Country:          input.Country,
		City:             input.City,
		DietaryReqs:      input.DietaryReqs,
		SpecialNeeds:     input.SpecialNeeds,
		TicketType:       input.TicketType,
		SessionsOfInt:    input.SessionsOfInt,
		PaymentStatus:    "pending",
		RegistrationDate: now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Save to Firestore
	docRef, _, err := h.firebaseClient.Firestore.Collection("registrations").Add(ctx, registration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RegistrationResponse{
			Success: false,
			Message: "Failed to create registration: " + err.Error(),
		})
		return
	}

	registration.ID = docRef.ID

	c.JSON(http.StatusCreated, RegistrationResponse{
		Success:      true,
		Message:      "Registration created successfully",
		Registration: registration,
	})
}

// GetMyRegistration retrieves the current user's registration
func (h *RegistrationHandler) GetMyRegistration(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, RegistrationResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	user, ok := userVal.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, RegistrationResponse{
			Success: false,
			Message: "Failed to retrieve user information",
		})
		return
	}

	ctx := context.Background()
	registration, err := h.getUserRegistration(ctx, user.UID)
	if err != nil {
		c.JSON(http.StatusNotFound, RegistrationResponse{
			Success: false,
			Message: "Registration not found",
		})
		return
	}

	c.JSON(http.StatusOK, RegistrationResponse{
		Success:      true,
		Message:      "Registration retrieved successfully",
		Registration: registration,
	})
}

// UpdateRegistration updates the current user's registration
func (h *RegistrationHandler) UpdateRegistration(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, RegistrationResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	user, ok := userVal.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, RegistrationResponse{
			Success: false,
			Message: "Failed to retrieve user information",
		})
		return
	}

	var input models.RegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, RegistrationResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	ctx := context.Background()

	// Find existing registration
	existingReg, err := h.getUserRegistration(ctx, user.UID)
	if err != nil {
		c.JSON(http.StatusNotFound, RegistrationResponse{
			Success: false,
			Message: "Registration not found. Please create one first.",
		})
		return
	}

	// Update registration
	updates := map[string]interface{}{
		"firstName":           input.FirstName,
		"lastName":            input.LastName,
		"email":               input.Email,
		"phone":               input.Phone,
		"organization":        input.Organization,
		"jobTitle":            input.JobTitle,
		"country":             input.Country,
		"city":                input.City,
		"dietaryRequirements": input.DietaryReqs,
		"specialNeeds":        input.SpecialNeeds,
		"ticketType":          input.TicketType,
		"sessionsOfInterest":  input.SessionsOfInt,
		"updatedAt":           time.Now(),
	}

	_, err = h.firebaseClient.Firestore.Collection("registrations").Doc(existingReg.ID).Set(ctx, updates, firestore.MergeAll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RegistrationResponse{
			Success: false,
			Message: "Failed to update registration: " + err.Error(),
		})
		return
	}

	// Fetch updated registration
	updatedReg, _ := h.getUserRegistration(ctx, user.UID)

	c.JSON(http.StatusOK, RegistrationResponse{
		Success:      true,
		Message:      "Registration updated successfully",
		Registration: updatedReg,
	})
}

// DeleteRegistration deletes the current user's registration
func (h *RegistrationHandler) DeleteRegistration(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, RegistrationResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	user, ok := userVal.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, RegistrationResponse{
			Success: false,
			Message: "Failed to retrieve user information",
		})
		return
	}

	ctx := context.Background()

	// Find existing registration
	existingReg, err := h.getUserRegistration(ctx, user.UID)
	if err != nil {
		c.JSON(http.StatusNotFound, RegistrationResponse{
			Success: false,
			Message: "Registration not found",
		})
		return
	}

	// Delete registration
	_, err = h.firebaseClient.Firestore.Collection("registrations").Doc(existingReg.ID).Delete(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RegistrationResponse{
			Success: false,
			Message: "Failed to delete registration: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RegistrationResponse{
		Success: true,
		Message: "Registration deleted successfully",
	})
}

// GetAllRegistrations retrieves all registrations (admin only - add admin check as needed)
func (h *RegistrationHandler) GetAllRegistrations(c *gin.Context) {
	ctx := context.Background()

	iter := h.firebaseClient.Firestore.Collection("registrations").Documents(ctx)
	var registrations []models.Registration

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, RegistrationResponse{
				Success: false,
				Message: "Failed to retrieve registrations: " + err.Error(),
			})
			return
		}

		var reg models.Registration
		if err := doc.DataTo(&reg); err != nil {
			continue
		}
		reg.ID = doc.Ref.ID
		registrations = append(registrations, reg)
	}

	c.JSON(http.StatusOK, RegistrationResponse{
		Success:       true,
		Message:       "Registrations retrieved successfully",
		Registrations: registrations,
	})
}

// getUserRegistration retrieves a user's registration from Firestore
func (h *RegistrationHandler) getUserRegistration(ctx context.Context, userID string) (*models.Registration, error) {
	iter := h.firebaseClient.Firestore.Collection("registrations").Where("userId", "==", userID).Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err != nil {
		return nil, err
	}

	var registration models.Registration
	if err := doc.DataTo(&registration); err != nil {
		return nil, err
	}
	registration.ID = doc.Ref.ID

	return &registration, nil
}
