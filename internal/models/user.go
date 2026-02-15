package models

import "time"

// User represents a user in the system (linked to Firebase Auth)
type User struct {
	UID           string    `json:"uid" firestore:"uid"`
	Email         string    `json:"email" firestore:"email"`
	DisplayName   string    `json:"displayName" firestore:"displayName"`
	PhotoURL      string    `json:"photoUrl" firestore:"photoUrl"`
	Provider      string    `json:"provider" firestore:"provider"` // google, email, etc.
	EmailVerified bool      `json:"emailVerified" firestore:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt" firestore:"updatedAt"`
	LastLoginAt   time.Time `json:"lastLoginAt" firestore:"lastLoginAt"`
}

// Registration represents a conference registration
type Registration struct {
	ID               string    `json:"id" firestore:"-"`
	UserID           string    `json:"userId" firestore:"userId"`
	FirstName        string    `json:"firstName" firestore:"firstName"`
	LastName         string    `json:"lastName" firestore:"lastName"`
	Email            string    `json:"email" firestore:"email"`
	Phone            string    `json:"phone" firestore:"phone"`
	Organization     string    `json:"organization" firestore:"organization"`
	JobTitle         string    `json:"jobTitle" firestore:"jobTitle"`
	Country          string    `json:"country" firestore:"country"`
	City             string    `json:"city" firestore:"city"`
	DietaryReqs      string    `json:"dietaryRequirements" firestore:"dietaryRequirements"`
	SpecialNeeds     string    `json:"specialNeeds" firestore:"specialNeeds"`
	TicketType       string    `json:"ticketType" firestore:"ticketType"` // standard, vip, student, etc.
	SessionsOfInt    []string  `json:"sessionsOfInterest" firestore:"sessionsOfInterest"`
	PaymentStatus    string    `json:"paymentStatus" firestore:"paymentStatus"` // pending, completed, refunded
	RegistrationDate time.Time `json:"registrationDate" firestore:"registrationDate"`
	CreatedAt        time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt" firestore:"updatedAt"`
}

// RegistrationInput is used for creating/updating registrations
type RegistrationInput struct {
	FirstName     string   `json:"firstName" binding:"required"`
	LastName      string   `json:"lastName" binding:"required"`
	Email         string   `json:"email" binding:"required,email"`
	Phone         string   `json:"phone"`
	Organization  string   `json:"organization"`
	JobTitle      string   `json:"jobTitle"`
	Country       string   `json:"country" binding:"required"`
	City          string   `json:"city"`
	DietaryReqs   string   `json:"dietaryRequirements"`
	SpecialNeeds  string   `json:"specialNeeds"`
	TicketType    string   `json:"ticketType" binding:"required"`
	SessionsOfInt []string `json:"sessionsOfInterest"`
}

// Session represents a conference session
type Session struct {
	ID          string    `json:"id" firestore:"-"`
	Title       string    `json:"title" firestore:"title"`
	Description string    `json:"description" firestore:"description"`
	Speaker     string    `json:"speaker" firestore:"speaker"`
	SpeakerBio  string    `json:"speakerBio" firestore:"speakerBio"`
	StartTime   time.Time `json:"startTime" firestore:"startTime"`
	EndTime     time.Time `json:"endTime" firestore:"endTime"`
	Location    string    `json:"location" firestore:"location"`
	Capacity    int       `json:"capacity" firestore:"capacity"`
	Track       string    `json:"track" firestore:"track"` // technical, business, workshop, etc.
	Tags        []string  `json:"tags" firestore:"tags"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" firestore:"updatedAt"`
}
