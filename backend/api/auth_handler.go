// pickle/backend/api/auth_handler.go
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/carlostbanks/pickle/auth"
	"github.com/carlostbanks/pickle/db"
	"github.com/google/uuid"
)

// AuthHandler handles all authentication related endpoints
type AuthHandler struct{}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// RegisterRoutes registers all authentication routes
func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/auth/google/login", h.googleLogin)
	mux.HandleFunc("/auth/google/callback", h.googleCallback)
	mux.HandleFunc("/auth/refresh", h.refreshToken)
	mux.HandleFunc("/auth/logout", h.logout)
}

// googleLogin initiates the Google OAuth flow
func (h *AuthHandler) googleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate a state token to prevent CSRF
	state := uuid.New().String()

	// Create a cookie to store the state
	http.SetCookie(w, &http.Cookie{
		Name:     "oauthState",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set secure flag if using HTTPS
		MaxAge:   int(time.Now().Add(5 * time.Minute).Unix()),
	})

	// Redirect to Google's OAuth page
	url := auth.GetGoogleLoginURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// googleCallback handles the callback from Google OAuth
func (h *AuthHandler) googleCallback(w http.ResponseWriter, r *http.Request) {
	// Get the state from the cookie
	stateCookie, err := r.Cookie("oauthState")
	if err != nil {
		http.Error(w, "State not found", http.StatusBadRequest)
		return
	}

	// Compare the state from the cookie with the state from the callback
	if r.URL.Query().Get("state") != stateCookie.Value {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Get the authorization code from the callback
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	// Exchange the code for a token
	token, err := auth.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		log.Printf("Failed to exchange code for token: %v", err)
		return
	}

	// Get the user info from Google
	googleUser, err := auth.GetUserInfo(token)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		log.Printf("Failed to get user info: %v", err)
		return
	}

	// Check if the user exists in the database
	var userExists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", googleUser.ID).Scan(&userExists)
	if err != nil {
		http.Error(w, "Failed to check if user exists", http.StatusInternalServerError)
		log.Printf("Failed to check if user exists: %v", err)
		return
	}

	// If the user doesn't exist, create a new user
	if !userExists {
		_, err = db.DB.Exec(`
			INSERT INTO users (id, email, name, picture, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, googleUser.ID, googleUser.Email, googleUser.Name, googleUser.Picture, time.Now().Format(time.RFC3339))
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			log.Printf("Failed to create user: %v", err)
			return
		}
	}

	// Generate a JWT token
	jwtToken, err := auth.GenerateJWT(googleUser)
	if err != nil {
		http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
		log.Printf("Failed to generate JWT: %v", err)
		return
	}

	// Set the JWT token as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set secure flag if using HTTPS
		MaxAge:   int(time.Now().Add(24 * time.Hour).Unix()),
	})

	// Redirect to the frontend
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// refreshToken refreshes the JWT token
func (h *AuthHandler) refreshToken(w http.ResponseWriter, r *http.Request) {
	// Get the token from the cookie
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Token not found", http.StatusUnauthorized)
		return
	}

	// Validate the token
	claims, err := auth.ValidateJWT(tokenCookie.Value)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get the user from the database
	var user auth.User
	err = db.DB.QueryRow("SELECT id, email, name, picture FROM users WHERE id = $1", claims.UserID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Picture)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Generate a new JWT token
	jwtToken, err := auth.GenerateJWT(&user)
	if err != nil {
		http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
		return
	}

	// Set the new JWT token as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set secure flag if using HTTPS
		MaxAge:   int(time.Now().Add(24 * time.Hour).Unix()),
	})

	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// logout logs the user out
func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	// Clear the token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set secure flag if using HTTPS
		MaxAge:   -1,
	})

	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
