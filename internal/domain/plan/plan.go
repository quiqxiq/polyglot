package plan

import "time"

// Plan represents an ISP internet service plan product.
type Plan struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	RateLimit   string    `json:"rate_limit"` // e.g. "50M/50M"
	MonthlyFee  float64   `json:"monthly_fee"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
