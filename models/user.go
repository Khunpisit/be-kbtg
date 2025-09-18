package models

import (
	"time"
)

// User represents an application user.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	FirstName    string    `gorm:"size:100" json:"first_name"`
	LastName     string    `gorm:"size:100" json:"last_name"`
	DisplayName  string    `gorm:"size:150" json:"display_name"`
	Phone        string    `gorm:"size:30" json:"phone"`
	AvatarURL    string    `gorm:"size:255" json:"avatar_url"`
	Bio          string    `gorm:"size:500" json:"bio"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName overrides default pluralized table name.
func (User) TableName() string { return "users" }
