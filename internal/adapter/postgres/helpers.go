package postgres

import (
	"gorm.io/gorm/clause"

	"github.com/quixiq/polyglot/pkg/idgen"
)

// lockingClause returns SELECT ... FOR UPDATE untuk mencegah race
// double-payment pada baris invoice.
func lockingClause() clause.Locking {
	return clause.Locking{Strength: "UPDATE"}
}

func newID(prefix string) string {
	return idgen.New(prefix)
}

func orDefault(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}

func strPtr(s string) *string {
	return &s
}
