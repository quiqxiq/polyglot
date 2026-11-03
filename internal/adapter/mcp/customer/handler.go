package customer

import "github.com/quixiq/polyglot/internal/port"

// Handler holds dependencies for Customer MCP tools.
type Handler struct {
	GetCustomerRepo func() port.CustomerRepository
}

// NewHandler creates a new Customer tool handler.
func NewHandler(repoGetter func() port.CustomerRepository) *Handler {
	return &Handler{
		GetCustomerRepo: repoGetter,
	}
}

func (h *Handler) repo() port.CustomerRepository {
	if h.GetCustomerRepo != nil {
		return h.GetCustomerRepo()
	}
	return nil
}
