package domain

import "time"

type StockMovement struct {
	ID        string
	ProductID string
	Delta     int
	Reason    string
	CreatedBy string
	CreatedAt time.Time
}
