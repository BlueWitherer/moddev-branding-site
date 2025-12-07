package utils

import "time"

// Database row for images listing
type Img struct {
	ImgID    int64     `json:"img_id"`     // Image ID
	UserID   string    `json:"user_id"`    // Owner GitHub user ID
	ImageURL string    `json:"image_url"`  // URL to the image image
	Created  time.Time `json:"created_at"` // First created
	Pending  bool      `json:"pending"`    // Under review
}
