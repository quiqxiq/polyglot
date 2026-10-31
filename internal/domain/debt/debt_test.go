package debt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDebtOutstanding(t *testing.T) {
	t.Parallel()
	d := Debt{Status: StatusOpen, Amount: 100000, RemainingAmount: 40000, DebtDate: time.Now()}
	assert.True(t, d.IsOutstanding())
	d.Status = StatusPaid
	assert.False(t, d.IsOutstanding())
	d.Status = StatusPartial
	d.RemainingAmount = 0
	assert.False(t, d.IsOutstanding())
}
