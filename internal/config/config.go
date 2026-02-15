package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	ServerPort string
	ServerHost string

	// Firebase configuration
	FirebaseCredentialsFile string
	FirebaseProjectID       string

	// Google OAuth configuration
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Session configuration
	SessionSecret string

	// Environment
	Environment string

	// Frontend URL for CORS and redirects
	FrontendURL string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),

		// Firebase
		FirebaseCredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", "firebase-service-account.json"),
		FirebaseProjectID:       getEnv("FIREBASE_PROJECT_ID", ""),

		// Google OAuth
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),

		// Session
		SessionSecret: getEnv("SESSION_SECRET", "your-secret-key-change-in-production"),

		// Environment
		Environment: getEnv("ENVIRONMENT", "development"),

		// Frontend
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
