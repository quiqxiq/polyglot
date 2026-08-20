package mikrotik

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// IPPool is the vendor-neutral IP pool row.
// Canonical definition lives in internal/port (see port.IPPool docs).
type IPPool = port.IPPool

// NewPrintIPPoolsCommand builds the command.Command for /ip/pool/print.
// Pass a non-empty nameFilter to look up one IP pool by name.
func NewPrintIPPoolsCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ip/pool/print",
		Args: args,
	}
}

// ParseIPPools converts command.Result rows from /ip/pool/print into typed IPPool values.
// Rows missing ".id" or "name" are skipped.
func ParseIPPools(result command.Result) []IPPool {
	pools := make([]IPPool, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		pools = append(pools, IPPool{
			RosID:    id,
			Name:     name,
			Ranges:   row["ranges"],
			NextPool: row["next-pool"],
			Comment:  row["comment"],
		})
	}
	return pools
}
