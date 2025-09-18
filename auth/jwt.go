package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Manager handles JWT token creation and validation.
type Manager struct {
	secret    string
	expiresIn time.Duration
}

// NewManager returns a configured Manager.
func NewManager(secret string, expiresIn time.Duration) Manager {
	return Manager{secret: secret, expiresIn: expiresIn}
}

// ErrInvalidToken indicates token parsing or validation failed.
var ErrInvalidToken = errors.New("invalid token")

// Generate issues a signed JWT for a user.
func (m Manager) Generate(userID uint, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(m.expiresIn).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

// Parse validates the token string and returns claims.
func (m Manager) Parse(tokenStr string) (jwt.MapClaims, error) {
	tkn, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return []byte(m.secret), nil
	})
	if err != nil || !tkn.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := tkn.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
