package utils

import (
	"time"
)

type User struct {
	ID        int64     `json:"id"`         // GitHub user ID
	Login     string    `json:"login"`      // GitHub username
	AvatarURL string    `json:"avatar_url"` // GitHub user avatar URL
	IsAdmin   bool      `json:"is_admin"`   // Active administrator status
	IsStaff   bool      `json:"is_staff"`   // Active staff status
	Verified  bool      `json:"verified"`   // Trusted status
	Banned    bool      `json:"banned"`     // Banned status
	Created   time.Time `json:"created_at"` // First created
	Updated   time.Time `json:"updated_at"` // Last updated
}
