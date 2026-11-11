package netops

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// SessionCookieName carries the signed NetOps browser session. Its path is
	// scoped to the /netops mount so it never reaches the rest of Polyglot.
	SessionCookieName = "polyglot_netops"
	sessionVersion    = "v1"
)

var errInvalidSession = errors.New("invalid netops session")

// SessionCodec signs and verifies stateless NetOps browser sessions. The
// payload binds the user ID to an absolute expiry; the HMAC prevents a client
// from changing either field.
type SessionCodec struct {
	secret []byte
	ttl    time.Duration
}

// NewSessionCodec returns a codec with the given HMAC key and session lifetime.
func NewSessionCodec(secret string, ttl time.Duration) (*SessionCodec, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("netops session secret is empty")
	}
	if ttl <= 0 {
		return nil, errors.New("netops session lifetime must be positive")
	}
	return &SessionCodec{secret: []byte(secret), ttl: ttl}, nil
}

// Issue returns a signed session value for userID.
func (c *SessionCodec) Issue(userID uint, now time.Time) string {
	expiresAt := now.Add(c.ttl).Unix()
	payload := fmt.Sprintf("%s.%d.%d", sessionVersion, userID, expiresAt)
	return payload + "." + c.sign(payload)
}

// Verify returns the user ID carried by value when the signature is valid and
// the session has not expired.
func (c *SessionCodec) Verify(value string, now time.Time) (uint, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 4 || parts[0] != sessionVersion {
		return 0, errInvalidSession
	}
	payload := strings.Join(parts[:3], ".")
	expected := c.sign(payload)
	if !hmac.Equal([]byte(parts[3]), []byte(expected)) {
		return 0, errInvalidSession
	}

	userID, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || userID == 0 {
		return 0, errInvalidSession
	}
	expiresAt, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || now.Unix() >= expiresAt {
		return 0, errInvalidSession
	}
	return uint(userID), nil
}

// Cookie builds the browser cookie for value. Secure must be true whenever the
// browser reaches Polyglot over HTTPS.
func (c *SessionCodec) Cookie(value string, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     MountPath,
		MaxAge:   int(c.ttl.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// ClearCookie expires the NetOps session cookie.
func (c *SessionCodec) ClearCookie(secure bool) *http.Cookie {
	cookie := c.Cookie("", secure)
	cookie.MaxAge = -1
	cookie.Expires = time.Unix(1, 0)
	return cookie
}

func (c *SessionCodec) sign(payload string) string {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
