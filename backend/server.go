// pickle/backend/server.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Court represents a padel court
type Court struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	Name           string    `json:"name"`
	Address        string    `json:"address"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	NumberOfCourts int       `json:"number_of_courts"`
	Amenities      []string  `json:"amenities" gorm:"-"` // Handled separately
	AmenitiesArray string    `json:"-" gorm:"column:amenities"`
	ImageURL       string    `json:"image_url" gorm:"column:image_url"`
	CreatedAt      time.Time `json:"created_at"`
}

// TableName sets the table name for Court model
func (Court) TableName() string {
	return "courts"
}

// Booking represents a court booking
type Booking struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	CourtID           string    `json:"court_id" gorm:"column:court_id"`
	UserID            string    `json:"user_id" gorm:"column:user_id"`
	Date              string    `json:"date"`
	StartTime         string    `json:"start_time" gorm:"column:start_time"`
	EndTime           string    `json:"end_time" gorm:"column:end_time"`
	NumberOfPlayers   int       `json:"number_of_players" gorm:"column:number_of_players"`
	PlayerEmails      []string  `json:"player_emails" gorm:"-"`
	PlayerEmailsArray string    `json:"-" gorm:"column:player_emails"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName sets the table name for Booking model
func (Booking) TableName() string {
	return "bookings"
}

// User represents a user
type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName sets the table name for User model
func (User) TableName() string {
	return "users"
}

var db *gorm.DB

func main() {
	// Connect to database
	var err error
	dsn := "host=localhost user=postgres password=postgres dbname=pickle port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get the underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get DB: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to database successfully")

	// Set up HTTP routes with logging
	setupRoutes()

	// Start HTTP server
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting HTTP server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// logMiddleware creates a middleware that logs request details
func logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		log.Printf("Request: %s %s", r.Method, r.URL.Path)

		// Call the handler
		next(w, r)

		duration := time.Since(startTime)
		log.Printf("Response: %s %s - took %v", r.Method, r.URL.Path, duration)
	}
}

// setupRoutes sets up all HTTP routes with logging middleware
func setupRoutes() {
	http.HandleFunc("/health", logMiddleware(healthHandler))
	http.HandleFunc("/api/courts", logMiddleware(courtsHandler))
	http.HandleFunc("/api/courts/", logMiddleware(courtDetailHandler))
	http.HandleFunc("/api/bookings", logMiddleware(bookingsHandler))
}

// healthHandler is a simple health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Server is healthy"))
}

// courtsHandler handles GET requests for courts
func courtsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	city := r.URL.Query().Get("city")

	// Initialize query with GORM
	query := db

	// Add filters if provided
	if city != "" {
		query = query.Where("address LIKE ?", "%"+city+"%")
	}

	// Fetch courts
	var courts []Court
	if err := query.Find(&courts).Error; err != nil {
		log.Printf("Error querying courts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Process amenities for each court (convert PostgreSQL array to Go slice)
	for i := range courts {
		amenitiesStr := strings.Trim(courts[i].AmenitiesArray, "{}")
		if amenitiesStr != "" {
			courts[i].Amenities = strings.Split(amenitiesStr, ",")
		} else {
			courts[i].Amenities = []string{}
		}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"courts": courts,
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// courtDetailHandler handles GET requests for a specific court
func courtDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract court ID from URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid court ID", http.StatusBadRequest)
		return
	}
	courtID := parts[len(parts)-1]

	// Query court with GORM
	var court Court
	if err := db.First(&court, "id = ?", courtID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Court not found", http.StatusNotFound)
		} else {
			log.Printf("Error querying court: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Process amenities
	amenitiesStr := strings.Trim(court.AmenitiesArray, "{}")
	if amenitiesStr != "" {
		court.Amenities = strings.Split(amenitiesStr, ",")
	} else {
		court.Amenities = []string{}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(court); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// bookingsHandler handles GET, POST requests for bookings
func bookingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getBookingsHandler(w, r)
	case http.MethodPost:
		createBookingHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getBookingsHandler handles GET requests for bookings
func getBookingsHandler(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	userID := r.URL.Query().Get("user_id")
	courtID := r.URL.Query().Get("court_id")
	date := r.URL.Query().Get("date")

	// Initialize query with GORM
	query := db

	// Add filters if provided
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if courtID != "" {
		query = query.Where("court_id = ?", courtID)
	}
	if date != "" {
		query = query.Where("date = ?", date)
	}

	// Add ordering
	query = query.Order("date").Order("start_time")

	// Fetch bookings
	var bookings []Booking
	if err := query.Find(&bookings).Error; err != nil {
		log.Printf("Error querying bookings: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Process player emails for each booking
	for i := range bookings {
		emailsStr := strings.Trim(bookings[i].PlayerEmailsArray, "{}")
		if emailsStr != "" {
			bookings[i].PlayerEmails = strings.Split(emailsStr, ",")
		} else {
			bookings[i].PlayerEmails = []string{}
		}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"bookings": bookings,
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// createBookingHandler handles POST requests to create a booking
func createBookingHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var input struct {
		CourtID         string   `json:"court_id"`
		UserID          string   `json:"user_id"`
		Date            string   `json:"date"`
		StartTime       string   `json:"start_time"`
		EndTime         string   `json:"end_time"`
		NumberOfPlayers int      `json:"number_of_players"`
		PlayerEmails    []string `json:"player_emails"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if input.CourtID == "" || input.Date == "" || input.StartTime == "" || input.EndTime == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Check if court exists
	var courtCount int64
	if err := db.Model(&Court{}).Where("id = ?", input.CourtID).Count(&courtCount).Error; err != nil {
		log.Printf("Error checking court: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if courtCount == 0 {
		http.Error(w, "Court not found", http.StatusNotFound)
		return
	}

	// Check for conflicting bookings
	var conflictCount int64
	if err := db.Model(&Booking{}).
		Where("court_id = ? AND date = ? AND status != 'CANCELLED'", input.CourtID, input.Date).
		Where("(start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?) OR (start_time >= ? AND end_time <= ?)",
			input.StartTime, input.StartTime, input.EndTime, input.EndTime, input.StartTime, input.EndTime).
		Count(&conflictCount).Error; err != nil {
		log.Printf("Error checking conflicts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if conflictCount > 0 {
		http.Error(w, "Time slot is already booked", http.StatusConflict)
		return
	}

	// Create booking
	booking := Booking{
		ID:                uuid.New().String(),
		CourtID:           input.CourtID,
		UserID:            input.UserID,
		Date:              input.Date,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
		NumberOfPlayers:   input.NumberOfPlayers,
		PlayerEmailsArray: "{" + strings.Join(input.PlayerEmails, ",") + "}",
		Status:            "CONFIRMED",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := db.Create(&booking).Error; err != nil {
		log.Printf("Error creating booking: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set email field for response
	booking.PlayerEmails = input.PlayerEmails

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(booking); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
