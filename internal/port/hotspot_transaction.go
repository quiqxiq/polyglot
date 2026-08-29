package port

import (
	domainHotspot "github.com/quixiq/polyglot/internal/domain/hotspot"
)

// MikhmonTransaction alias to domain model.
type MikhmonTransaction = domainHotspot.MikhmonTransaction

// ReportFilter selects Mikhmon transaction report records from
// /system/script. Filter values follow the legacy Mikhmon frontend:
//   - Day   : exact date string ("aug/17/2026") → RouterOS ?source= (get_report)
//   - Month : owner value "<mon><year>" ("aug2026") → RouterOS ?owner= (get_livereport)
//   - Year  : "2026" → post-filter on the date suffix (combined with day/month
//     or applied over all mikhmon records when day/month are empty)
type ReportFilter struct {
	Day   string
	Month string
	Year  string
}
