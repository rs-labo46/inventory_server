package domain

import "time"

type Inventory struct {
	ProductID string
	Quantity  int
	UpdatedAt time.Time
}
