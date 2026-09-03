package device

// SimpleQueue represents one row returned by /queue/simple/print.
//
// Counter fields (Bytes, Packets, Dropped, Rate, PacketRate) are runtime
// statistics — they reset on reboot and are useful for monitoring dashboards.
type SimpleQueue struct {
	RosID          string
	Name           string
	Target         string
	Parent         string
	MaxLimit       string
	LimitAt        string
	BurstLimit     string
	BurstThreshold string
	BurstTime      string
	Priority       string
	Queue          string // queue type: pfifo, fq_codel, etc.
	Bytes          string // cumulative bytes "rx/tx"
	Packets        string
	Dropped        string
	Rate           string // current rate bps "rx/tx"
	PacketRate     string
	Comment        string
	Disabled       bool
}
