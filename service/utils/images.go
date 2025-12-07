package utils

import "time"

// Database row for images listing
type Img struct {
	ImgID    int64     `json:"img_id"`     // Image ID
	UserID   string    `json:"user_id"`    // Owner GitHub user ID
	Type     int       `json:"type"`       // Type of image
	ImageURL string    `json:"image_url"`  // URL to the image image
	Created  time.Time `json:"created_at"` // First created
	Expiry   int64     `json:"expiry"`     // Unix time of expiration
	Pending  bool      `json:"pending"`    // Under review
}
