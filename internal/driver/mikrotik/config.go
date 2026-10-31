package mikrotik

import (
	"strconv"
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/core/roslimit"
)

// DriverConfig holds runtime configuration and tunable intervals for a MikroTik driver.
type DriverConfig struct {
	PollInterval     time.Duration
	FastPollInterval time.Duration
	SlowPollInterval time.Duration
	ConcurrencyLimit int
	DialTimeout      time.Duration
	UseTLS           bool
	InsecureTLS      bool
}

// DefaultDriverConfig returns sensible defaults for RouterOS interaction.
func DefaultDriverConfig() DriverConfig {
	return DriverConfig{
		PollInterval:     5 * time.Second,
		FastPollInterval: 1 * time.Second,
		SlowPollInterval: 30 * time.Second,
		ConcurrencyLimit: roslimit.DefaultMax,
		DialTimeout:      10 * time.Second,
		UseTLS:           false,
		InsecureTLS:      true,
	}
}

// ConfigFromTarget builds DriverConfig by overlaying target.Extra parameters over defaults.
func ConfigFromTarget(target device.Target) DriverConfig {
	cfg := DefaultDriverConfig()

	if target.Timeout > 0 {
		cfg.DialTimeout = target.Timeout
	}

	if target.Extra != nil {
		if tlsVal, ok := target.Extra["tls"]; ok {
			cfg.UseTLS = tlsVal == "true" || tlsVal == "1" || tlsVal == "yes"
		} else if useTLS, ok := target.Extra["use_tls"]; ok {
			cfg.UseTLS = useTLS == "true" || useTLS == "1" || useTLS == "yes"
		}

		if pollMsStr, ok := target.Extra["poll_interval_ms"]; ok {
			if ms, err := strconv.Atoi(pollMsStr); err == nil && ms > 0 {
				cfg.PollInterval = time.Duration(ms) * time.Millisecond
			}
		} else if pollSecStr, ok := target.Extra["poll_interval_sec"]; ok {
			if sec, err := strconv.Atoi(pollSecStr); err == nil && sec > 0 {
				cfg.PollInterval = time.Duration(sec) * time.Second
			}
		}

		if limitStr, ok := target.Extra["concurrency_limit"]; ok {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				cfg.ConcurrencyLimit = limit
			}
		}
	}

	return cfg
}
