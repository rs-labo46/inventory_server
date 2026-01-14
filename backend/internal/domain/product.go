package domain

import "time"

type Product struct {
	ID        string
	Name      string
	CreatedAt time.Time
}
