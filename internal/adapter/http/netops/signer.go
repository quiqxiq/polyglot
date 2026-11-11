package netops

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// TrustedProxyHeader asserts to the harness that the request already passed
	// Polyglot's admin session check. The harness may verify it with the shared
	// proxy secret.
	TrustedProxyHeader = "X-DSH-Trusted-Proxy"
	proxyVersion       = "v1"
)

var errInvalidProxyHeader = errors.New("invalid trusted proxy header")

// ProxySigner signs the trusted-proxy assertion with a secret shared between
// Polyglot and the harness.
type ProxySigner struct {
	secret []byte
}

// NewProxySigner returns a signer with the given HMAC key.
func NewProxySigner(secret string) (*ProxySigner, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("netops proxy secret is empty")
	}
	return &ProxySigner{secret: []byte(secret)}, nil
}

// Header returns the signed assertion for one forwarded request.
func (s *ProxySigner) Header(userID uint, now time.Time, method string, path string) string {
	timestamp := now.Unix()
	payload := proxyPayload(userID, timestamp, method, path)
	return fmt.Sprintf("%s;user=%d;ts=%d;sig=%s", proxyVersion, userID, timestamp, s.sign(payload))
}

// Verify checks a header produced by Header. The harness can use this to
// reject requests that did not pass through the Polyglot proxy.
func (s *ProxySigner) Verify(header string, userID uint, now time.Time, method string, path string, tolerance time.Duration) error {
	fields := strings.Split(header, ";")
	if len(fields) != 4 || fields[0] != proxyVersion {
		return errInvalidProxyHeader
	}
	gotUserID, err := strconv.ParseUint(strings.TrimPrefix(fields[1], "user="), 10, 64)
	if err != nil || uint(gotUserID) != userID {
		return errInvalidProxyHeader
	}
	timestamp, err := strconv.ParseInt(strings.TrimPrefix(fields[2], "ts="), 10, 64)
	if err != nil {
		return errInvalidProxyHeader
	}
	signedAt := time.Unix(timestamp, 0)
	if tolerance > 0 {
		delta := now.Sub(signedAt)
		if delta < -tolerance || delta > tolerance {
			return errInvalidProxyHeader
		}
	}
	expected := s.sign(proxyPayload(userID, timestamp, method, path))
	if !hmac.Equal([]byte(strings.TrimPrefix(fields[3], "sig=")), []byte(expected)) {
		return errInvalidProxyHeader
	}
	return nil
}

func (s *ProxySigner) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func proxyPayload(userID uint, timestamp int64, method string, path string) string {
	return fmt.Sprintf("%s|%d|%d|%s|%s", proxyVersion, userID, timestamp, method, path)
}
