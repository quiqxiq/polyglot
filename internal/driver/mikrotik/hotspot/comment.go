package hotspot

import (
	"errors"
	"fmt"
	"strings"
	"time"

	domainHotspot "github.com/quixiq/polyglot/internal/domain/hotspot"
	"github.com/quixiq/polyglot/internal/port"
)

// ErrInvalidMikhmonComment indicates a comment string could not be parsed
// into a valid Mikhmon pre-login or post-login metadata comment structure.
var ErrInvalidMikhmonComment = errors.New("mikhmon: invalid comment format")

// MikhmonComment is the vendor-neutral parsed Mikhmon comment metadata.
// Canonical definition lives in internal/port (see port.MikhmonComment docs).
type MikhmonComment = port.MikhmonComment

// FormatPreLoginComment delegates to domainHotspot.FormatPreLoginComment.
func FormatPreLoginComment(vcType, code, tag string, t time.Time) string {
	return domainHotspot.FormatPreLoginComment(vcType, code, tag, t)
}

// BuildCreateUserComment delegates to domainHotspot.BuildCreateUserComment.
func BuildCreateUserComment(name, password, comment string, now time.Time) string {
	return domainHotspot.BuildCreateUserComment(name, password, comment, now)
}

// BuildUpdatedComment delegates to domainHotspot.BuildUpdatedComment.
func BuildUpdatedComment(expireDate, userCode, newComment string, now time.Time) string {
	return domainHotspot.BuildUpdatedComment(expireDate, userCode, newComment, now)
}

// ParseMikhmonComment parses a MikroTik Hotspot user comment string into a typed
// MikhmonComment struct. It detects whether the comment is in pre-login or post-login format.
func ParseMikhmonComment(commentStr string) (MikhmonComment, error) {
	c := MikhmonComment{RawComment: commentStr}
	trimmed := strings.TrimSpace(commentStr)
	if trimmed == "" {
		return c, ErrInvalidMikhmonComment
	}

	parts := strings.Fields(trimmed)

	// Post-login format check: "DD/MM/YYYY HH:MM:SS <mode> <old-comment>"
	// Has at least 4 space-separated tokens and first token contains '/'
	if len(parts) >= 4 && strings.Contains(parts[0], "/") && strings.Contains(parts[1], ":") {
		c.IsActivated = true
		c.ExpireDate = parts[0]
		c.ExpireTime = parts[1]
		c.ExpireMode = parts[2]

		// The rest is old-comment
		oldComment := strings.Join(parts[3:], " ")
		parsePreLoginParts(oldComment, &c)
		return c, nil
	}

	// Pre-login format check: "vc-A3X-08.03.26-MyTag"
	if parsePreLoginParts(trimmed, &c) {
		return c, nil
	}

	return c, ErrInvalidMikhmonComment
}

// parsePreLoginParts parses "<type>-<code>-<date>-<tag>" string into MikhmonComment fields.
func parsePreLoginParts(s string, c *MikhmonComment) bool {
	dashParts := strings.Split(s, "-")
	if len(dashParts) < 3 {
		return false
	}
	t := dashParts[0]
	if t != "vc" && t != "up" {
		return false
	}
	c.Type = t
	c.Code = dashParts[1]
	c.CreatedDate = dashParts[2]
	if len(dashParts) > 3 {
		c.Tag = strings.Join(dashParts[3:], "-")
	}
	return true
}

// IsExpired reports whether a user comment indicates the voucher has passed its
// expiry date and time relative to now. Returns false if the voucher is not yet activated.
func IsExpired(commentStr string, now time.Time) bool {
	parsed, err := ParseMikhmonComment(commentStr)
	if err != nil || !parsed.IsActivated {
		return false
	}

	// ExpireDate format: "DD/MM/YYYY" or "DD/Mon/YYYY"
	// ExpireTime format: "HH:MM:SS"
	dateTimeStr := fmt.Sprintf("%s %s", parsed.ExpireDate, parsed.ExpireTime)

	var expireTime time.Time
	// Try standard DD/MM/YYYY HH:MM:SS
	t, err := time.Parse("02/01/2006 15:04:05", dateTimeStr)
	if err != nil {
		// Try RouterOS date format DD/Jan/2006 15:04:05
		t, err = time.Parse("02/Jan/2006 15:04:05", dateTimeStr)
		if err != nil {
			return false
		}
	}
	expireTime = t

	return now.After(expireTime)
}

// IsMikhmonComment reports whether the given comment string is in a recognised
// Mikhmon format (either pre-login "vc/up-code-date[-tag]" or post-login
// "DD/MM/YYYY HH:MM:SS mode old-comment"). Returns false for empty strings
// or plain labels.
func IsMikhmonComment(comment string) bool {
	_, err := ParseMikhmonComment(comment)
	return err == nil
}
