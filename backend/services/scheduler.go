// pickle/backend/services/scheduler.go
package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	// These will be available after proto generation
	// "github.com/carlostbanks/pickle/proto"
)

// Temporary type definitions until proto generation is done
// These will be replaced by the generated protobuf types

// BookingStatus represents the status of a booking
type BookingStatus int32

const (
	BookingStatus_PENDING   BookingStatus = 0
	BookingStatus_CONFIRMED BookingStatus = 1
	BookingStatus_CANCELLED BookingStatus = 2
)

// Court represents a padel court
type Court struct {
	Id             string
	Name           string
	Address        string
	Latitude       float64
	Longitude      float64
	NumberOfCourts int32
	Amenities      []string
	ImageUrl       string
}

// GetCourtsRequest represents a request to get courts
type GetCourtsRequest struct {
	City      string
	Latitude  float64
	Longitude float64
	RadiusKm  int32
}

// GetCourtsResponse represents a response with courts
type GetCourtsResponse struct {
	Courts []*Court
}

// GetCourtRequest represents a request to get a court
type GetCourtRequest struct {
	CourtId string
}

// Booking represents a court booking
type Booking struct {
	Id              string
	CourtId         string
	UserId          string
	Date            string
	StartTime       string
	EndTime         string
	NumberOfPlayers int32
	PlayerEmails    []string
	Status          BookingStatus
	CreatedAt       string
	UpdatedAt       string
}

// CreateBookingRequest represents a request to create a booking
type CreateBookingRequest struct {
	CourtId         string
	Date            string
	StartTime       string
	EndTime         string
	NumberOfPlayers int32
	PlayerEmails    []string
}

// GetBookingsRequest represents a request to get bookings
type GetBookingsRequest struct {
	UserId  string
	CourtId string
	Date    string
}

// GetBookingsResponse represents a response with bookings
type GetBookingsResponse struct {
	Bookings []*Booking
}

// UpdateBookingRequest represents a request to update a booking
type UpdateBookingRequest struct {
	BookingId       string
	StartTime       string
	EndTime         string
	NumberOfPlayers int32
	PlayerEmails    []string
}

// CancelBookingRequest represents a request to cancel a booking
type CancelBookingRequest struct {
	BookingId string
}

// CancelBookingResponse represents a response to a booking cancellation
type CancelBookingResponse struct {
	Success bool
	Message string
}

// SchedulerServer implements the SchedulerService gRPC service
type SchedulerServer struct {
	// This field will be added after proto generation
	// proto.UnimplementedSchedulerServiceServer
	db *sql.DB
}

// NewSchedulerServer creates a new scheduler server
func NewSchedulerServer(db *sql.DB) *SchedulerServer {
	return &SchedulerServer{db: db}
}

// GetCourts returns courts based on search criteria
func (s *SchedulerServer) GetCourts(ctx context.Context, req *GetCourtsRequest) (*GetCourtsResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you'd query a database or external API

	// Query courts from database based on location
	rows, err := s.db.Query(`
		SELECT id, name, address, latitude, longitude, number_of_courts, amenities, image_url
		FROM courts
		WHERE 
			($1 = '' OR city = $1)
			AND ($2 = 0 OR 
				(
					earth_distance(
						ll_to_earth($2, $3),
						ll_to_earth(latitude, longitude)
					) <= $4 * 1000
				)
			)
	`, req.City, req.Latitude, req.Longitude, req.RadiusKm)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courts []*Court
	for rows.Next() {
		var court Court
		var amenitiesArray []string

		err := rows.Scan(
			&court.Id,
			&court.Name,
			&court.Address,
			&court.Latitude,
			&court.Longitude,
			&court.NumberOfCourts,
			&amenitiesArray,
			&court.ImageUrl,
		)

		if err != nil {
			return nil, err
		}

		court.Amenities = amenitiesArray
		courts = append(courts, &court)
	}

	return &GetCourtsResponse{Courts: courts}, nil
}

// GetCourt returns a specific court by ID
func (s *SchedulerServer) GetCourt(ctx context.Context, req *GetCourtRequest) (*Court, error) {
	var court Court
	var amenitiesArray []string

	err := s.db.QueryRow(`
		SELECT id, name, address, latitude, longitude, number_of_courts, amenities, image_url
		FROM courts
		WHERE id = $1
	`, req.CourtId).Scan(
		&court.Id,
		&court.Name,
		&court.Address,
		&court.Latitude,
		&court.Longitude,
		&court.NumberOfCourts,
		&amenitiesArray,
		&court.ImageUrl,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("court not found")
		}
		return nil, err
	}

	court.Amenities = amenitiesArray
	return &court, nil
}

// CreateBooking creates a new booking
func (s *SchedulerServer) CreateBooking(ctx context.Context, req *CreateBookingRequest) (*Booking, error) {
	// Generate a new UUID for the booking
	bookingID := uuid.New().String()

	// Get user ID from context
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, errors.New("user not authenticated")
	}

	// Validate booking times
	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, errors.New("invalid start time format")
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, errors.New("invalid end time format")
	}

	if endTime.Before(startTime) {
		return nil, errors.New("end time must be after start time")
	}

	// Check if court exists
	var courtExists bool
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM courts WHERE id = $1)", req.CourtId).Scan(&courtExists)
	if err != nil {
		return nil, err
	}

	if !courtExists {
		return nil, errors.New("court not found")
	}

	// Check for conflicting bookings
	var conflictExists bool
	err = s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM bookings 
			WHERE court_id = $1 
			AND date = $2 
			AND status != 'CANCELLED'
			AND (
				(start_time <= $3 AND end_time > $3) OR
				(start_time < $4 AND end_time >= $4) OR
				(start_time >= $3 AND end_time <= $4)
			)
		)
	`, req.CourtId, req.Date, req.StartTime, req.EndTime).Scan(&conflictExists)

	if err != nil {
		return nil, err
	}

	if conflictExists {
		return nil, errors.New("booking time conflicts with existing booking")
	}

	// Create booking in database
	now := time.Now().Format(time.RFC3339)

	_, err = s.db.Exec(`
		INSERT INTO bookings (
			id, court_id, user_id, date, start_time, end_time, 
			number_of_players, player_emails, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, bookingID, req.CourtId, userID, req.Date, req.StartTime, req.EndTime,
		req.NumberOfPlayers, req.PlayerEmails, "CONFIRMED", now, now)

	if err != nil {
		return nil, err
	}

	// Return the created booking
	booking := &Booking{
		Id:              bookingID,
		CourtId:         req.CourtId,
		UserId:          userID,
		Date:            req.Date,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		NumberOfPlayers: req.NumberOfPlayers,
		PlayerEmails:    req.PlayerEmails,
		Status:          BookingStatus_CONFIRMED,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return booking, nil
}

// GetBookings retrieves bookings based on filter criteria
func (s *SchedulerServer) GetBookings(ctx context.Context, req *GetBookingsRequest) (*GetBookingsResponse, error) {
	// Build query based on filters
	query := `
		SELECT id, court_id, user_id, date, start_time, end_time, 
			   number_of_players, player_emails, status, created_at, updated_at
		FROM bookings
		WHERE 1=1
	`
	var args []interface{}
	var argCount int = 1

	if req.UserId != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, req.UserId)
		argCount++
	}

	if req.CourtId != "" {
		query += fmt.Sprintf(" AND court_id = $%d", argCount)
		args = append(args, req.CourtId)
		argCount++
	}

	if req.Date != "" {
		query += fmt.Sprintf(" AND date = $%d", argCount)
		args = append(args, req.Date)
		argCount++
	}

	query += " ORDER BY date, start_time"

	// Execute query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results
	var bookings []*Booking
	for rows.Next() {
		var booking Booking
		var statusStr string
		var playerEmailsArray []string

		err := rows.Scan(
			&booking.Id,
			&booking.CourtId,
			&booking.UserId,
			&booking.Date,
			&booking.StartTime,
			&booking.EndTime,
			&booking.NumberOfPlayers,
			&playerEmailsArray,
			&statusStr,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Map status string to enum
		switch statusStr {
		case "PENDING":
			booking.Status = BookingStatus_PENDING
		case "CONFIRMED":
			booking.Status = BookingStatus_CONFIRMED
		case "CANCELLED":
			booking.Status = BookingStatus_CANCELLED
		}

		booking.PlayerEmails = playerEmailsArray
		bookings = append(bookings, &booking)
	}

	return &GetBookingsResponse{Bookings: bookings}, nil
}

// UpdateBooking updates an existing booking
func (s *SchedulerServer) UpdateBooking(ctx context.Context, req *UpdateBookingRequest) (*Booking, error) {
	// Get user ID from context
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, errors.New("user not authenticated")
	}

	// Check if booking exists and belongs to user
	var booking Booking
	var statusStr string
	var playerEmailsArray []string
	var courtID string
	var dateStr string

	err := s.db.QueryRow(`
		SELECT id, court_id, user_id, date, start_time, end_time, 
			   number_of_players, player_emails, status, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`, req.BookingId).Scan(
		&booking.Id,
		&courtID,
		&booking.UserId,
		&dateStr,
		&booking.StartTime,
		&booking.EndTime,
		&booking.NumberOfPlayers,
		&playerEmailsArray,
		&statusStr,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("booking not found")
		}
		return nil, err
	}

	if booking.UserId != userID {
		return nil, errors.New("not authorized to update this booking")
	}

	// Check for conflicting bookings (excluding this booking)
	var conflictExists bool
	err = s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM bookings 
			WHERE court_id = $1 
			AND date = $2 
			AND id != $3
			AND status != 'CANCELLED'
			AND (
				(start_time <= $4 AND end_time > $4) OR
				(start_time < $5 AND end_time >= $5) OR
				(start_time >= $4 AND end_time <= $5)
			)
		)
	`, courtID, dateStr, req.BookingId, req.StartTime, req.EndTime).Scan(&conflictExists)

	if err != nil {
		return nil, err
	}

	if conflictExists {
		return nil, errors.New("booking time conflicts with existing booking")
	}

	// Update booking
	now := time.Now().Format(time.RFC3339)

	_, err = s.db.Exec(`
		UPDATE bookings
		SET start_time = $1, end_time = $2, number_of_players = $3, 
			player_emails = $4, updated_at = $5
		WHERE id = $6
	`, req.StartTime, req.EndTime, req.NumberOfPlayers, req.PlayerEmails, now, req.BookingId)

	if err != nil {
		return nil, err
	}

	// Return updated booking
	booking.StartTime = req.StartTime
	booking.EndTime = req.EndTime
	booking.NumberOfPlayers = req.NumberOfPlayers
	booking.PlayerEmails = req.PlayerEmails
	booking.UpdatedAt = now

	// Map status string to enum
	switch statusStr {
	case "PENDING":
		booking.Status = BookingStatus_PENDING
	case "CONFIRMED":
		booking.Status = BookingStatus_CONFIRMED
	case "CANCELLED":
		booking.Status = BookingStatus_CANCELLED
	}

	return &booking, nil
}

// CancelBooking cancels an existing booking
func (s *SchedulerServer) CancelBooking(ctx context.Context, req *CancelBookingRequest) (*CancelBookingResponse, error) {
	// Get user ID from context
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, errors.New("user not authenticated")
	}

	// Check if booking exists and belongs to user
	var bookingUserID string
	err := s.db.QueryRow("SELECT user_id FROM bookings WHERE id = $1", req.BookingId).Scan(&bookingUserID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("booking not found")
		}
		return nil, err
	}

	if bookingUserID != userID {
		return nil, errors.New("not authorized to cancel this booking")
	}

	// Update booking status to cancelled
	now := time.Now().Format(time.RFC3339)

	_, err = s.db.Exec(`
		UPDATE bookings
		SET status = 'CANCELLED', updated_at = $1
		WHERE id = $2
	`, now, req.BookingId)

	if err != nil {
		return nil, err
	}

	return &CancelBookingResponse{
		Success: true,
		Message: "Booking cancelled successfully",
	}, nil
}

// Helper function to get user ID from context
// In a real implementation, this would retrieve the user ID from the JWT token
func getUserIDFromContext(ctx context.Context) string {
	// This is a placeholder
	// In a real implementation, you'd extract the user ID from the context
	// which would be set by the auth middleware
	return "sample-user-id"
}
