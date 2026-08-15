package genieacs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// DefaultPollInterval is the gap between GET /tasks/{id} polls when a POSTed task returned 202.
const DefaultPollInterval = 500 * time.Millisecond

// defaultFaultChannel is the default GenieACS fault channel checked after task completion.
const defaultFaultChannel = "default"

// Driver implements port.DeviceDriver for TR-069/ACS via the GenieACS NBI.
type Driver struct {
	baseURL      url.URL
	deviceID     string
	apiKey       string
	httpClient   *http.Client
	pollInterval time.Duration
	faultChannel string
}

// Compile-time proof that *Driver satisfies port.DeviceDriver.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver builds a REST client for the GenieACS NBI bound to one CPE.
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	deviceID := target.Extra["device_id"]
	if deviceID == "" {
		return nil, ErrDeviceIDMissing
	}

	port := target.Port
	if port == 0 {
		port = 7557
	}

	scheme := "http"
	if target.Extra["use_tls"] == "true" {
		scheme = "https"
	}

	pollInterval := DefaultPollInterval
	if raw := target.Extra["poll_interval"]; raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			pollInterval = d
		}
	}

	faultChannel := target.Extra["fault_channel"]
	if faultChannel == "" {
		faultChannel = defaultFaultChannel
	}

	d := &Driver{
		baseURL: url.URL{
			Scheme: scheme,
			Host:   fmt.Sprintf("%s:%d", target.Host, port),
		},
		deviceID:     deviceID,
		apiKey:       target.Extra["api_key"],
		httpClient:   &http.Client{Timeout: target.Timeout},
		pollInterval: pollInterval,
		faultChannel: faultChannel,
	}

	if err := d.ping(ctx); err != nil {
		return nil, fmt.Errorf("genieacs: NBI unreachable: %w", err)
	}
	return d, nil
}

func (d *Driver) ping(ctx context.Context) error {
	q := url.Values{}
	q.Set("query", `{"_id":"__ping__"}`)
	resp, err := d.do(ctx, http.MethodGet, "/devices/", q, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("NBI responded %d", resp.StatusCode)
	}
	return nil
}

func (d *Driver) Close() error {
	d.httpClient.CloseIdleConnections()
	return nil
}

func (d *Driver) do(ctx context.Context, method, path string, query url.Values, body []byte) (*http.Response, error) {
	u := d.baseURL.JoinPath(path)
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if d.apiKey != "" {
		req.Header.Set("x-api-key", d.apiKey)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	return resp, nil
}
