-- pickle/backend/db/setup.sql
-- Run this script in Postico on your PostgreSQL database

-- Create database (if not already created)
-- CREATE DATABASE pickle;

-- Connect to database
\c pickle;

-- Create required extensions
CREATE EXTENSION IF NOT EXISTS earthdistance CASCADE;
CREATE EXTENSION IF NOT EXISTS cube CASCADE;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    picture TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create courts table
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
);

-- Create bookings table
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
);

-- Insert sample court data
INSERT INTO courts (id, name, address, latitude, longitude, number_of_courts, amenities, image_url, created_at)
VALUES 
    ('court-1', 'Downtown Padel Club', '123 Main St, Seattle, WA 98101', 47.6062, -122.3321, 4, 
     ARRAY['Parking', 'Restrooms', 'Pro Shop', 'Lessons'], 'https://example.com/downtown.jpg', CURRENT_TIMESTAMP),
    ('court-2', 'Eastside Padel Center', '456 Park Ave, Bellevue, WA 98004', 47.6101, -122.2015, 6, 
     ARRAY['Parking', 'Restrooms', 'Pro Shop', 'Lessons', 'Cafe'], 'https://example.com/eastside.jpg', CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Insert sample user data
INSERT INTO users (id, email, name, picture, created_at)
VALUES 
    ('user-1', 'alice@example.com', 'Alice Smith', 'https://example.com/alice.jpg', CURRENT_TIMESTAMP),
    ('user-2', 'bob@example.com', 'Bob Johnson', 'https://example.com/bob.jpg', CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;