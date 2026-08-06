package customer

import "time"

// Customer represents an ISP subscriber entity.
type Customer struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Subscription represents an active plan subscription for a customer.
type Subscription struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	PlanID     string    `json:"plan_id"`
	Status     string    `json:"status"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Price      float64   `json:"price"`
}
