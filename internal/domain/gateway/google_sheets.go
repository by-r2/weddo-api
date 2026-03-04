package gateway

import (
	"context"
	"time"
)

type GoogleToken struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

type GoogleSheetsClient interface {
	WriteTab(ctx context.Context, tabName string, values [][]any) error
	ReadTab(ctx context.Context, tabName string) ([][]string, error)
}

type GoogleSheetsOAuthProvider interface {
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (*GoogleToken, error)
	NewClient(ctx context.Context, token *GoogleToken, spreadsheetID string) (GoogleSheetsClient, error)
	CreateSpreadsheet(ctx context.Context, token *GoogleToken, title string) (id string, url string, err error)
}
