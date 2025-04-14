// pickle/backend/server.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

var (
	googleOAuthConfig *oauth2.Config
	oauthStateString  = "random-state" // In production, generate a random state string
)

var jwtSecret = []byte("your-secret-key")

func main() {
	// Connect to database
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
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

	// Initialize OAuth config
	googleOAuthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	// Set up HTTP routes with logging
	setupRoutes()

	// Set up CORS wrapper
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(http.DefaultServeMux)

	// Start HTTP server with CORS middleware
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting HTTP server on port %s", port)
	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
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
	// Your existing routes...
	http.HandleFunc("/health", logMiddleware(healthHandler))
	http.HandleFunc("/api/courts", logMiddleware(courtsHandler))
	http.HandleFunc("/api/courts/", logMiddleware(courtDetailHandler))
	http.HandleFunc("/api/bookings", logMiddleware(bookingsHandler))

	// Add Google OAuth routes
	http.HandleFunc("/auth/google/login", logMiddleware(handleGoogleLogin))
	http.HandleFunc("/auth/google/callback", logMiddleware(handleGoogleCallback))
	http.HandleFunc("/auth/logout", logMiddleware(logoutHandler))

	// Add user API endpoint
	http.HandleFunc("/api/users/me", logMiddleware(getCurrentUser))
}

// Google OAuth login handler
func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOAuthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Google OAuth callback handler
func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state
	state := r.FormValue("state")
	if state != oauthStateString {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	code := r.FormValue("code")
	token, err := googleOAuthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("Code exchange failed: %v", err)
		http.Error(w, "Code exchange failed", http.StatusInternalServerError)
		return
	}

	// Get user info
	client := googleOAuthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse user info
	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("Failed to parse user info: %v", err)
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Create or update user in database
	user := User{
		ID:        userInfo.ID,
		Email:     userInfo.Email,
		Name:      userInfo.Name,
		Picture:   userInfo.Picture,
		CreatedAt: time.Now(),
	}

	// Check if user exists
	var existingUser User
	result := db.Where("id = ?", user.ID).First(&existingUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new user
			if err := db.Create(&user).Error; err != nil {
				log.Printf("Failed to create user: %v", err)
				http.Error(w, "Failed to create user", http.StatusInternalServerError)
				return
			}
		} else {
			log.Printf("Database error: %v", result.Error)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing user
		existingUser.Email = user.Email
		existingUser.Name = user.Name
		existingUser.Picture = user.Picture
		if err := db.Save(&existingUser).Error; err != nil {
			log.Printf("Failed to update user: %v", err)
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}

		// Use the updated user
		user = existingUser
	}

	// Create JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.RegisteredClaims{
		Subject:   user.ID,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString(jwtSecret)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set JWT token in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,     // Set to true in production with HTTPS
		MaxAge:   3600 * 24, // 24 hours
	})

	// Redirect back to frontend with token in URL
	// This allows the frontend to capture the token
	redirectURL := fmt.Sprintf("http://localhost:3000/auth-callback?token=%s", tokenString)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// Get current user handler
func getCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Try to get from cookie as fallback
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Not authenticated", http.StatusUnauthorized)
			return
		}
		authHeader = "Bearer " + cookie.Value
	}

	// Extract token from Bearer prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Extract user ID from claims
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	userID := claims.Subject

	// Get user from database
	var user User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("Database error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Return user as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Error encoding user: %v", err)
	}
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

	// Get user ID from token
	userID := getUserIDFromRequest(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var input struct {
		CourtID         string   `json:"courtId"` // Changed from court_id to match frontend
		Date            string   `json:"date"`
		StartTime       string   `json:"startTime"`       // Changed from start_time
		EndTime         string   `json:"endTime"`         // Changed from end_time
		NumberOfPlayers int      `json:"numberOfPlayers"` // Changed from number_of_players
		PlayerEmails    []string `json:"playerEmails"`    // Changed from player_emails
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
		UserID:            userID, // Use the authenticated user's ID
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

// Add this helper function to extract user ID from JWT token
func getUserIDFromRequest(r *http.Request) string {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Try to get from cookie as fallback
		cookie, err := r.Cookie("token")
		if err != nil {
			return ""
		}
		authHeader = "Bearer " + cookie.Value
	}

	// Extract token from Bearer prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return ""
	}

	// Extract user ID from claims
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return ""
	}

	return claims.Subject
}

// logoutHandler handles POST requests to log out a user
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		MaxAge:   -1,    // Delete the cookie
	})

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
