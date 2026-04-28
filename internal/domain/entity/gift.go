package entity

import "time"

type GiftKind string

const (
	GiftKindCatalog       GiftKind = "catalog"
	GiftKindCashTemplate  GiftKind = "cash_template"
)

type GiftStatus string

const (
	GiftStatusAvailable GiftStatus = "available"
	GiftStatusPurchased GiftStatus = "purchased"
)

type Gift struct {
	ID          string
	WeddingID   string
	Name        string
	Description string
	Price       float64
	ImageURL    string
	Category    string
	Status      GiftStatus
	Kind        GiftKind
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
