package entities

import "time"

type Discount struct {
	ID        string
	ItemID    string
	Percent   float64
	Active    bool
	StartsAt  *time.Time
	EndsAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
