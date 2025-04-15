# Pickleball Court Scheduling System

A full-stack application for finding and booking pickleball/padel ball courts. This project uses a Go backend with GORM and a React/TypeScript frontend.

https://i.ibb.co/rG1pgJ95/Screenshot-2025-04-14-at-10-37-54-PM.png

## Features

- **Court Search**: Find courts by city or location radius
- **Court Details**: View court information, amenities, and availability
- **Booking System**: Book courts for specific dates and times
- **User Authentication**: Sign up and log in with Google OAuth

## Tech Stack

### Backend
- **Go**: Backend language
- **GORM**: ORM for database operations
- **PostgreSQL**: Database for storing courts, users, and bookings
- **JWT**: Authentication tokens

### Frontend
- **React**: Frontend library
- **TypeScript**: Type safety for JavaScript
- **React Router**: Client-side routing

## Getting Started

### Prerequisites
- Go 1.19 or later
- PostgreSQL
- Node.js 18+ (for frontend)

### Backend Setup

1. Clone the repository:
```bash
git clone https://github.com/carlostbanks/pickleball.git
cd pickleball
```

2. Install Go dependencies:
```bash
cd backend
go mod tidy
```

3. Set up the database:
```bash
# Create a PostgreSQL database named 'pickle'
# Then run the database setup script
psql -U postgres -d pickle -f db/setup.sql
```

4. Start the backend server:
```bash
go run server.go
```

### API Endpoints

- `GET /health`: Health check endpoint
- `GET /api/courts`: Get all courts, optionally filtered by city
- `GET /api/courts/{id}`: Get a specific court by ID
- `GET /api/bookings`: Get bookings, filtered by user_id, court_id, or date
- `POST /api/bookings`: Create a new booking

## License

MIT
