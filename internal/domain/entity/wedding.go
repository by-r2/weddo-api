package entity

import "time"

type Wedding struct {
	ID           string
	Slug         string
	Title        string
	Date         string
	Partner1Name string
	Partner2Name string
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
