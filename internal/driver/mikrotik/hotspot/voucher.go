package hotspot

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
)

// CharSet defines character set types used for generating voucher usernames and passwords.
// Canonical definition lives in internal/port (see port.CharSet).
type CharSet = port.CharSet

const (
	CharSetNumeric  CharSet = "numeric"  // "1234567890"
	CharSetLower    CharSet = "lower"    // "abcdefghijklmnopqrstuvwxyz"
	CharSetUpper    CharSet = "upper"    // "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharSetLowerNum CharSet = "lowernum" // "abcdefghijklmnopqrstuvwxyz1234567890"
	CharSetUpperNum CharSet = "uppernum" // "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	CharSetMixed    CharSet = "mixed"    // "aBcDeFgHiJkLmNoPqRsTuVwXyZ1234567890"
)

var charSetMap = map[CharSet]string{
	CharSetNumeric:  "1234567890",
	CharSetLower:    "abcdefghijklmnopqrstuvwxyz",
	CharSetUpper:    "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	CharSetLowerNum: "abcdefghijklmnopqrstuvwxyz1234567890",
	CharSetUpperNum: "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890",
	CharSetMixed:    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890",
}

// VoucherGenerateParams is the vendor-neutral voucher generation parameter set.
// Canonical definition lives in internal/port (see port.VoucherGenerateParams docs).
type VoucherGenerateParams = port.VoucherGenerateParams

// GeneratedVoucher holds the details of one generated voucher.
type GeneratedVoucher struct {
	Username string
	Password string
	Comment  string
	Command  command.Command
}

// VoucherBatch holds the result of a mass voucher generation.
type VoucherBatch struct {
	Vouchers []GeneratedVoucher
	Commands []command.Command
}

// GenerateVoucherCode generates a cryptographically random string of given length
// using the specified character set rules.
func GenerateVoucherCode(length int, cs CharSet) string {
	if length <= 0 {
		length = 6
	}
	charset, ok := charSetMap[cs]
	if !ok {
		charset = charSetMap[CharSetLowerNum]
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			result[i] = charset[i%len(charset)]
		} else {
			result[i] = charset[n.Int64()]
		}
	}
	return string(result)
}

// ParseDataLimit converts human-readable data limit strings like "100m", "1g", "500k"
// into integer byte counts as required by RouterOS limit-bytes-total / limit-bytes-out.
func ParseDataLimit(s string) int64 {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0
	}
	unit := s[len(s)-1]
	valStr := s[:len(s)-1]
	var multiplier int64 = 1
	switch unit {
	case 'k':
		multiplier = 1024
	case 'm':
		multiplier = 1024 * 1024
	case 'g':
		multiplier = 1024 * 1024 * 1024
	default:
		valStr = s
	}
	var val int64
	_, _ = fmt.Sscanf(valStr, "%d", &val)
	return val * multiplier
}

// NewAddMikhmonVoucherCommand builds a command.Command for /ip/hotspot/user/add
// using Mikhmon-formatted comments. Reuses mikrotik.NewAddHotspotUserCommand.
func NewAddMikhmonVoucherCommand(p VoucherGenerateParams, username, password string) command.Command {
	code := GenerateVoucherCode(3, CharSetUpperNum)
	comment := FormatPreLoginComment("vc", code, p.CommentTag, time.Now())

	limitBytesStr := p.LimitBytes
	if bytes := ParseDataLimit(p.LimitBytes); bytes > 0 {
		limitBytesStr = fmt.Sprintf("%d", bytes)
	}

	return mikrotik.NewAddHotspotUserCommand(mikrotik.HotspotUserParams{
		Name:          username,
		Password:      password,
		Profile:       p.Profile,
		Server:        p.Server,
		LimitUptime:   p.LimitUptime,
		LimitBytesOut: limitBytesStr,
		Comment:       comment,
		Disabled:      false,
	})
}

// NewGenerateVoucherBatchCommands generates a batch of count vouchers with
// random usernames/passwords according to params.
func NewGenerateVoucherBatchCommands(p VoucherGenerateParams, count int) VoucherBatch {
	if count <= 0 {
		count = 1
	}
	userLen := p.UserLength
	if userLen <= 0 {
		userLen = 6
	}

	limitBytesStr := p.LimitBytes
	if bytes := ParseDataLimit(p.LimitBytes); bytes > 0 {
		limitBytesStr = fmt.Sprintf("%d", bytes)
	}

	batch := VoucherBatch{
		Vouchers: make([]GeneratedVoucher, 0, count),
		Commands: make([]command.Command, 0, count),
	}

	now := time.Now()
	for i := 0; i < count; i++ {
		code := GenerateVoucherCode(3, CharSetUpperNum)
		uname := p.Prefix + GenerateVoucherCode(userLen, p.CharSet)
		pass := uname
		if p.PassLength > 0 {
			pass = GenerateVoucherCode(p.PassLength, p.CharSet)
		}
		comment := FormatPreLoginComment("vc", code, p.CommentTag, now)

		cmd := mikrotik.NewAddHotspotUserCommand(mikrotik.HotspotUserParams{
			Name:          uname,
			Password:      pass,
			Profile:       p.Profile,
			Server:        p.Server,
			LimitUptime:   p.LimitUptime,
			LimitBytesOut: limitBytesStr,
			Comment:       comment,
			Disabled:      false,
		})

		v := GeneratedVoucher{
			Username: uname,
			Password: pass,
			Comment:  comment,
			Command:  cmd,
		}

		batch.Vouchers = append(batch.Vouchers, v)
		batch.Commands = append(batch.Commands, cmd)
	}

	return batch
}
