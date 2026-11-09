package hotspot

import (
	"strconv"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	domainHotspot "github.com/quixiq/polyglot/internal/domain/hotspot"
)

// toProtoHotspotReports converts Mikhmon transaction records to proto list,
// converting the Price string into a float64 (proto field is double).
func toProtoHotspotReports(records []domainHotspot.MikhmonTransaction) []*polyglotv1.HotspotReport {
	pbReports := make([]*polyglotv1.HotspotReport, len(records))
	for i, r := range records {
		price, _ := strconv.ParseFloat(r.Price, 64)
		pbReports[i] = &polyglotv1.HotspotReport{
			Id:       r.RosID,
			Date:     r.Date,
			Time:     r.Time,
			Username: r.Username,
			Profile:  r.Profile,
			Price:    price,
			Comment:  r.Comment,
		}
	}
	return pbReports
}

// SumReportIncome sums the parsed Price values of the records. Records with
// an unparsable price are skipped (they still count toward the total count).
func SumReportIncome(records []domainHotspot.MikhmonTransaction) float64 {
	var total float64
	for _, r := range records {
		price, err := strconv.ParseFloat(r.Price, 64)
		if err != nil {
			continue
		}
		total += price
	}
	return total
}
