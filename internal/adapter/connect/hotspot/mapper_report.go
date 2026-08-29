package hotspot

import (
	"strconv"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
)

// toProtoHotspotReports converts Mikhmon transaction records to proto list,
// converting the Price string into a float64 (proto field is double).
func toProtoHotspotReports(records []port.MikhmonTransaction) []*devicepb.HotspotReport {
	pbReports := make([]*devicepb.HotspotReport, len(records))
	for i, r := range records {
		price, _ := strconv.ParseFloat(r.Price, 64)
		pbReports[i] = &devicepb.HotspotReport{
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
func SumReportIncome(records []port.MikhmonTransaction) float64 {
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
