package plan

import "time"

// Plan represents an ISP service offering.
type Plan struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	SpeedMbps   int       `json:"speed_mbps"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

