package entity

import "time"

type GoogleIntegration struct {
	WeddingID             string
	SpreadsheetID         string
	SpreadsheetURL        string
	EncryptedAccessToken  string
	EncryptedRefreshToken string
	TokenExpiry           *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
