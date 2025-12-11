package utils

import "time"

// Database row for images listing
type Img struct {
	ID       uint64    `json:"id"`         // Image ID
	UserID   uint64    `json:"user_id"`    // Owner GitHub user ID
	ImageURL string    `json:"image_url"`  // URL to the image image
	Created  time.Time `json:"created_at"` // First created
	Pending  bool      `json:"pending"`    // Under review
	Login    string    `json:"login"`      // Owner branding image
}
