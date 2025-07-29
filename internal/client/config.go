package client

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	SERVER_URL    string
	CLIENT_SECRET string
	DEV_MODE      bool
)

// SetBuildTimeConfig allows setting configuration from build-time variables
func SetBuildTimeConfig(serverURL, clientSecret string) {
	if serverURL != "" {
		SERVER_URL = serverURL
	}
	if clientSecret != "" {
		CLIENT_SECRET = clientSecret
	}
}

func init() {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Set default values
	SERVER_URL = getEnv("SERVER_URL", "http://localhost:8080")
	CLIENT_SECRET = getEnv("CLIENT_SECRET", "")
	DEV_MODE = getEnv("DEV_MODE", "false") == "true"

	if DEV_MODE {
		log.Printf("Development mode enabled. Server URL: %s", SERVER_URL)
	}
}

// ValidateConfig checks if required configuration is available
func ValidateConfig() error {
	if CLIENT_SECRET == "" {
		return fmt.Errorf("CLIENT_SECRET is required. Set it via environment variable or build-time injection")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
