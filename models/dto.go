package models

import (
	"errors"
	"strings"
)

// MinPasswordLength specifies minimum password length.
const MinPasswordLength = 6

// RegisterRequest represents the payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate checks required fields.
func (r *RegisterRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	if r.Email == "" || r.Password == "" {
		return errors.New("email & password required")
	}
	if len(r.Password) < MinPasswordLength {
		return errors.New("password too short")
	}
	return nil
}

// LoginRequest represents login payload.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateProfileRequest represents updatable profile fields.
type UpdateProfileRequest struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	DisplayName *string `json:"display_name"`
	Phone       *string `json:"phone"`
	AvatarURL   *string `json:"avatar_url"`
	Bio         *string `json:"bio"`
}

// TokenResponse represents a JWT token response.
type TokenResponse struct {
	Token string `json:"token"`
}

// RegisterResponse minimal response after successful registration.
type RegisterResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	CreatedAt any    `json:"created_at"`
}
