// pickle/backend/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Maps     MapsConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	GRPCPort int
	HTTPPort int
	Host     string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	JWTSecret          string
}

// MapsConfig holds maps API configuration
type MapsConfig struct {
	APIKey string
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			GRPCPort: getEnvAsInt("GRPC_PORT", 50051),
			HTTPPort: getEnvAsInt("HTTP_PORT", 8080),
			Host:     getEnv("HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "pickle"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
			JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
		},
		Maps: MapsConfig{
			APIKey: getEnv("MAPS_API_KEY", ""),
		},
	}

	return config, nil
}

// GetDatabaseConnectionString returns the database connection string
func (c *Config) GetDatabaseConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetGRPCAddress returns the gRPC server address
func (c *Config) GetGRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.GRPCPort)
}

// GetHTTPAddress returns the HTTP server address
func (c *Config) GetHTTPAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.HTTPPort)
}

// Helper function to get environment variable or default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to get environment variable as int or default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
