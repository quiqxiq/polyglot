package reporting

import "github.com/quixiq/polyglot/pkg/fault"

var (
	// ErrNotFound indicates the requested financial report was not found.
	ErrNotFound = fault.New(fault.KindNotFound, "reporting: report not found")
	// ErrInvalidInput indicates invalid date range or reporting parameters.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "reporting: invalid date range or parameters")
)
