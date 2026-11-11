package ont

import "github.com/quixiq/polyglot/internal/port"

// Handler holds dependencies for ONT/ONU read-only MCP tools.
type Handler struct {
	GetONTRepo func() port.ONTRepository
}

// NewHandler creates a new ONT tool handler.
func NewHandler(repoGetter func() port.ONTRepository) *Handler {
	return &Handler{GetONTRepo: repoGetter}
}

func (h *Handler) repo() port.ONTRepository {
	if h.GetONTRepo != nil {
		return h.GetONTRepo()
	}
	return nil
}
