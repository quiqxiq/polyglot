package port

// IPPool represents one row returned by /ip/pool/print.
type IPPool struct {
	RosID    string
	Name     string
	Ranges   string
	NextPool string
	Comment  string
}
