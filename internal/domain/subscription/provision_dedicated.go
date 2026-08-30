package subscription

// DedicatedQueueSpec mendefinisikan konfigurasi bandwidth dedicated CIR / MIR pada /queue/simple.
type DedicatedQueueSpec struct {
	QueueName   string
	Target      string // IP atau subnet target pelanggan
	MaxLimit    string // "50M/50M"
	LimitAt     string // Guaranteed CIR "50M/50M"
	Priority    string // "1/1" (highest)
	ParentQueue string
	Comment     string
}

// DedicatedProvisionSpec mendefinisikan parameter provisi untuk layanan Dedicated / Leased Line.
type DedicatedProvisionSpec struct {
	PPPoE PPPoEProvisionSpec
	Queue DedicatedQueueSpec
}
