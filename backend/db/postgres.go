// pickle/backend/db/postgres.go
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() {
	var err error

	// Get connection details from environment variables
	// or use default values for local development
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "pickle")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open doesn't actually connect, it just validates arguments
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Verify connection
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Successfully connected to database")

	// Create tables if they don't exist
	createTables()
}

// Helper function to get environment variable or default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// createTables creates the necessary tables if they don't exist
func createTables() {
	// Users table
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			picture TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	// Courts table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS courts (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			address TEXT NOT NULL,
			latitude FLOAT NOT NULL,
			longitude FLOAT NOT NULL,
			number_of_courts INT NOT NULL,
			amenities TEXT[],
			image_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create courts table: %v", err)
	}

	// Bookings table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS bookings (
			id VARCHAR(255) PRIMARY KEY,
			court_id VARCHAR(255) REFERENCES courts(id),
			user_id VARCHAR(255) REFERENCES users(id),
			date DATE NOT NULL,
			start_time TIME NOT NULL,
			end_time TIME NOT NULL,
			number_of_players INT NOT NULL,
			player_emails TEXT[],
			status VARCHAR(20) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(court_id, date, start_time)
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create bookings table: %v", err)
	}
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}
