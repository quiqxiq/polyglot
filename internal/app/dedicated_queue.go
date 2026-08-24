package app

import (
	"strings"

	"github.com/quixiq/polyglot/internal/port"
)

// dedicatedQueuePrefix menandai queue milik langganan DEDICATED polyglot.
const dedicatedQueuePrefix = "dq-"

// dedicatedQueueName: nama simple queue deterministik per pelanggan.
func dedicatedQueueName(username string) string {
	return dedicatedQueuePrefix + username
}

// isDedicated melaporkan apakah tipe layanan adalah DEDICATED.
func isDedicated(serviceType string) bool {
	return strings.EqualFold(serviceType, "DEDICATED")
}

// dedicatedQueueFromAccount membangun params queue dari akun terpetakan.
// LimitAt = rate dasar plan (CIR dijamin); MaxLimit = rate penuh
// (termasuk segmen burst bila kolom burst plan terisi).
func dedicatedQueueFromAccount(a port.SubscriberAccount, comment string) port.DedicatedQueueParams {
	p := port.DedicatedQueueParams{
		Name:     dedicatedQueueName(a.Username),
		Target:   a.Username,
		MaxLimit: a.RateLimit,
		LimitAt:  a.BaseRateLimit,
		Comment:  comment,
	}
	// Rate 8-segmen RouterOS: rx/tx/rx-burst/tx-burst/rx-thr/tx-thr/rx-time/tx-time.
	// Burst hanya dianggap ada bila minimal segmen burst lengkap (>= 4 pasang).
	seg := strings.Split(p.MaxLimit, "/")
	if len(seg) >= 4 {
		p.BurstLimit = seg[2] + "/" + seg[3]
	}
	if len(seg) >= 6 {
		p.BurstThreshold = seg[4] + "/" + seg[5]
	}
	if len(seg) >= 8 {
		p.BurstTime = seg[6] + "/" + seg[7]
	}
	return p
}
