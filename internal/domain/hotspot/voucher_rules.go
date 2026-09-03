package hotspot

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// CharSet defines character set types used for generating voucher usernames and passwords.
type CharSet string

// Supported character sets for voucher code generation.
const (
	// CharSetNumeric uses digits only (0-9).
	CharSetNumeric CharSet = "numeric" // "1234567890"
	// CharSetLower uses lowercase letters only (a-z).
	CharSetLower CharSet = "lower" // "abcdefghijklmnopqrstuvwxyz"
	// CharSetUpper uses uppercase letters only (A-Z).
	CharSetUpper CharSet = "upper" // "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// CharSetLowerNum uses lowercase letters and digits.
	CharSetLowerNum CharSet = "lowernum" // "abcdefghijklmnopqrstuvwxyz1234567890"
	// CharSetUpperNum uses uppercase letters and digits.
	CharSetUpperNum CharSet = "uppernum" // "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	// CharSetMixed uses mixed-case letters and digits.
	CharSetMixed CharSet = "mixed" // "aBcDeFgHiJkLmNoPqRsTuVwXyZ1234567890"
)

var charSetMap = map[CharSet]string{
	CharSetNumeric:  "1234567890",
	CharSetLower:    "abcdefghijklmnopqrstuvwxyz",
	CharSetUpper:    "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	CharSetLowerNum: "abcdefghijklmnopqrstuvwxyz1234567890",
	CharSetUpperNum: "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890",
	CharSetMixed:    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890",
}

// VoucherGenerateParams holds parameters for mass-generating vouchers.
type VoucherGenerateParams struct {
	Server      string  // Hotspot server name (empty = "all")
	Profile     string  // Hotspot user profile name (required)
	Prefix      string  // Code prefix (e.g. "vc")
	UserLength  int     // Length of generated username
	PassLength  int     // Length of generated password (0 = username is password)
	CharSet     CharSet // Character set to use
	LimitUptime string  // Time limit (e.g. "1d", "3h")
	LimitBytes  string  // Data quota in bytes
	CommentTag  string  // Label tag in comment
}

// GenerateVoucherCode generates a cryptographically random string of given length
// using the specified character set rules.
func GenerateVoucherCode(length int, cs CharSet) string {
	if length <= 0 {
		length = 6
	}
	charset, ok := charSetMap[cs]
	if !ok {
		charset = charSetMap[CharSetUpperNum]
	}
	maxIdx := big.NewInt(int64(len(charset)))
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, maxIdx)
		if err != nil {
			result[i] = charset[0]
			continue
		}
		result[i] = charset[idx.Int64()]
	}
	return string(result)
}

// FormatPreLoginComment formats the initial comment string when creating a new voucher.
func FormatPreLoginComment(vcType, code, tag string, t time.Time) string {
	if vcType == "" {
		vcType = "vc"
	}
	dateStr := t.Format("01.02.06") // MM.DD.YY
	if tag != "" {
		return fmt.Sprintf("%s-%s-%s-%s", vcType, code, dateStr, tag)
	}
	return fmt.Sprintf("%s-%s-%s", vcType, code, dateStr)
}

// BuildCreateUserComment returns the initial comment for a newly created hotspot user.
func BuildCreateUserComment(name, password, comment string, now time.Time) string {
	if comment == "" {
		vcType := "up"
		if name == password {
			vcType = "vc"
		}
		code := GenerateVoucherCode(3, CharSetUpperNum)
		comment = FormatPreLoginComment(vcType, code, "", now)
	}
	return comment
}

// BuildUpdatedComment rebuilds a hotspot user comment when updating a user via
// the Mikhmon admin UI, mirroring legacy post_update_user.php rules:
//   - expireDate == "" && userCode == "" → newComment returned unchanged.
//   - expireDate == "" && userCode != "" → keep the legacy prefix ("vc"/"up"/
//     "X") and rebuild with a fresh code and today's date.
//   - expireDate != "" && userCode == "" → "<expireDate> <newComment>"
//     (persists the expiry date inside the comment).
func BuildUpdatedComment(expireDate, userCode, newComment string, now time.Time) string {
	if expireDate != "" && userCode == "" {
		return expireDate + " " + newComment
	}
	if expireDate == "" && userCode != "" {
		vcType := "up"
		switch {
		case strings.HasPrefix(userCode, "vc"):
			vcType = "vc"
		case strings.HasPrefix(userCode, "X"):
			vcType = "X"
		}
		code := GenerateVoucherCode(3, CharSetUpperNum)
		return FormatPreLoginComment(vcType, code, newComment, now)
	}
	return newComment
}

// ProfileMeta holds the structured metadata embedded in a profile's on-login script.
type ProfileMeta struct {
	ExpireMode   string
	Price        float64
	Validity     string
	SellingPrice float64
	LockUser     string
	LockServer   string
}

// ParseOnLoginScript extracts structured ProfileMeta from a profile's on-login script.
func ParseOnLoginScript(onLogin string) (ProfileMeta, error) {
	meta := ProfileMeta{}
	if strings.TrimSpace(onLogin) == "" {
		return meta, nil
	}

	start := strings.Index(onLogin, ":put (\",")
	if start == -1 {
		return meta, fmt.Errorf("on-login script does not contain Mikhmon metadata payload")
	}
	start += len(":put (\"")
	end := strings.Index(onLogin[start:], "\")")
	if end == -1 {
		return meta, fmt.Errorf("malformed Mikhmon metadata payload in on-login script")
	}
	payload := onLogin[start : start+end]

	parts := strings.Split(payload, ",")
	if len(parts) < 8 {
		return meta, fmt.Errorf("metadata payload has fewer fields than expected: %q", payload)
	}

	meta.ExpireMode = parts[1]
	if meta.ExpireMode == "" && len(parts) > 5 && parts[5] == "noexp" {
		meta.ExpireMode = "0"
	}
	meta.Validity = parts[3]
	meta.LockUser = parts[6]
	meta.LockServer = parts[7]

	if parts[2] != "" {
		if p, err := strconv.ParseFloat(parts[2], 64); err == nil {
			meta.Price = p
		}
	}
	if parts[4] != "" {
		if sp, err := strconv.ParseFloat(parts[4], 64); err == nil {
			meta.SellingPrice = sp
		}
	}

	return meta, nil
}

// NormalizeProfileName replaces all whitespace sequences with single hyphens.
func NormalizeProfileName(name string) string {
	return strings.Join(strings.Fields(name), "-")
}

// FilterInactiveUsers compares all registered Hotspot users against active sessions and returns inactive users.
func FilterInactiveUsers(users []HotspotUser, active []HotspotActiveSession) []HotspotUser {
	activeMap := make(map[string]bool, len(active))
	for _, s := range active {
		activeMap[s.User] = true
	}
	inactive := make([]HotspotUser, 0)
	for _, u := range users {
		if !activeMap[u.Name] {
			inactive = append(inactive, u)
		}
	}
	return inactive
}
