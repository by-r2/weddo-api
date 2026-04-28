package entity

import "time"

type PaymentItem struct {
	ID                string
	PaymentID         string
	GiftID            string
	Amount            float64
	CustomName        string
	CustomDescription string
	CreatedAt         time.Time
}
