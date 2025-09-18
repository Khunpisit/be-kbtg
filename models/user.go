package models

import (
	"time"
)

// User represents an application user.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName overrides default pluralized table name.
func (User) TableName() string { return "users" }
