// pickle/backend/scripts/mock_data.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/carlostbanks/pickle/config"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Court represents a padel ball court
type Court struct {
	ID             string
	Name           string
	Address        string
	Latitude       float64
	Longitude      float64
	NumberOfCourts int
	Amenities      []string
	ImageURL       string
}

// Sample court data
var sampleCourts = []Court{
	{
		Name:           "Downtown Padel Club",
		Address:        "123 Main St, Seattle, WA 98101",
		Latitude:       47.6062,
		Longitude:      -122.3321,
		NumberOfCourts: 4,
		Amenities:      []string{"Parking", "Restrooms", "Pro Shop", "Lessons"},
		ImageURL:       "https://example.com/downtown.jpg",
	},
	{
		Name:           "Eastside Padel Center",
		Address:        "456 Park Ave, Bellevue, WA 98004",
		Latitude:       47.6101,
		Longitude:      -122.2015,
		NumberOfCourts: 6,
		Amenities:      []string{"Parking", "Restrooms", "Pro Shop", "Lessons", "Cafe"},
		ImageURL:       "https://example.com/eastside.jpg",
	},
	{
		Name:           "Northgate Padel",
		Address:        "789 North Way, Seattle, WA 98125",
		Latitude:       47.7062,
		Longitude:      -122.3259,
		NumberOfCourts: 2,
		Amenities:      []string{"Parking", "Restrooms"},
		ImageURL:       "https://example.com/northgate.jpg",
	},
	{
		Name:           "South Seattle Padel",
		Address:        "101 South Blvd, Seattle, WA 98118",
		Latitude:       47.5380,
		Longitude:      -122.3359,
		NumberOfCourts: 3,
		Amenities:      []string{"Parking", "Restrooms", "Pro Shop"},
		ImageURL:       "https://example.com/southseattle.jpg",
	},
	{
		Name:           "Redmond Padel Club",
		Address:        "202 Redmond Way, Redmond, WA 98052",
		Latitude:       47.6740,
		Longitude:      -122.1215,
		NumberOfCourts: 5,
		Amenities:      []string{"Parking", "Restrooms", "Pro Shop", "Lessons", "Cafe", "Locker Rooms"},
		ImageURL:       "https://example.com/redmond.jpg",
	},
}

// Sample users
var sampleUsers = []struct {
	ID      string
	Email   string
	Name    string
	Picture string
}{
	{
		ID:      "user-1",
		Email:   "alice@example.com",
		Name:    "Alice Smith",
		Picture: "https://example.com/alice.jpg",
	},
	{
		ID:      "user-2",
		Email:   "bob@example.com",
		Name:    "Bob Johnson",
		Picture: "https://example.com/bob.jpg",
	},
	{
		ID:      "user-3",
		Email:   "charlie@example.com",
		Name:    "Charlie Brown",
		Picture: "https://example.com/charlie.jpg",
	},
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.GetDatabaseConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Insert courts
	for _, court := range sampleCourts {
		id := uuid.New().String()
		_, err := db.Exec(`
			INSERT INTO courts (id, name, address, latitude, longitude, number_of_courts, amenities, image_url, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO NOTHING
		`, id, court.Name, court.Address, court.Latitude, court.Longitude, court.NumberOfCourts, pq.Array(court.Amenities), court.ImageURL, time.Now())
		if err != nil {
			log.Printf("Error inserting court %s: %v", court.Name, err)
		} else {
			log.Printf("Inserted court: %s", court.Name)
		}
	}

	// Insert users
	for _, user := range sampleUsers {
		_, err := db.Exec(`
			INSERT INTO users (id, email, name, picture, created_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO NOTHING
		`, user.ID, user.Email, user.Name, user.Picture, time.Now())
		if err != nil {
			log.Printf("Error inserting user %s: %v", user.Name, err)
		} else {
			log.Printf("Inserted user: %s", user.Name)
		}
	}

	// Insert bookings
	// Get court IDs
	rows, err := db.Query("SELECT id FROM courts")
	if err != nil {
		log.Fatalf("Failed to get court IDs: %v", err)
	}
	defer rows.Close()

	var courtIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Printf("Error scanning court ID: %v", err)
			continue
		}
		courtIDs = append(courtIDs, id)
	}

	// Generate bookings for the next 7 days
	for day := 0; day < 7; day++ {
		date := time.Now().AddDate(0, 0, day).Format("2006-01-02")

		for _, courtID := range courtIDs {
			// Generate 1-3 bookings per court per day
			numBookings := rand.Intn(3) + 1

			for i := 0; i < numBookings; i++ {
				// Generate random booking time
				hour := rand.Intn(12) + 8 // 8 AM to 8 PM
				startTime := fmt.Sprintf("%02d:00", hour)
				endTime := fmt.Sprintf("%02d:00", hour+1)

				// Pick a random user
				userID := sampleUsers[rand.Intn(len(sampleUsers))].ID

				// Generate random number of players (2 or 4)
				numPlayers := (rand.Intn(2) + 1) * 2

				// Generate random player emails
				playerEmails := []string{sampleUsers[rand.Intn(len(sampleUsers))].Email}
				for len(playerEmails) < numPlayers-1 {
					email := sampleUsers[rand.Intn(len(sampleUsers))].Email
					if !contains(playerEmails, email) {
						playerEmails = append(playerEmails, email)
					}
				}

				// Generate booking ID
				bookingID := uuid.New().String()

				// Insert booking
				_, err := db.Exec(`
					INSERT INTO bookings (id, court_id, user_id, date, start_time, end_time, number_of_players, player_emails, status, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
					ON CONFLICT (court_id, date, start_time) DO NOTHING
				`, bookingID, courtID, userID, date, startTime, endTime, numPlayers, pq.Array(playerEmails), "CONFIRMED", time.Now(), time.Now())

				if err != nil {
					log.Printf("Error inserting booking: %v", err)
				} else {
					log.Printf("Inserted booking for court %s on %s at %s", courtID, date, startTime)
				}
			}
		}
	}

	log.Println("Mock data generation complete")
}

// Helper function to check if a slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// Note: We're not using a custom pq.Array type anymore
// because it's already provided by the pq package
