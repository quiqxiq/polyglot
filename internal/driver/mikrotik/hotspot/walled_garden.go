package hotspot

import (
	"context"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// EnsureWalledGarden implements port.HotspotGateway.
// It idempotently provisions /ip/hotspot/walled-garden entries for domains
// and /ip/hotspot/walled-garden/ip for the portal destination.
func (g *Gateway) EnsureWalledGarden(ctx context.Context, driver port.DeviceDriver, domains []string, portalHost, portalPort string) error {
	// 1. Ensure domain entries in /ip/hotspot/walled-garden
	res, err := g.exec(ctx, driver, command.Command{Raw: "/ip/hotspot/walled-garden/print"})
	if err != nil {
		return fmt.Errorf("list walled-garden: %w", err)
	}
	existingDomains := make(map[string]bool)
	for _, row := range res.Rows {
		host := row["dst-host"]
		host = strings.TrimPrefix(host, "*")
		existingDomains[host] = true
	}

	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}
		clean := strings.TrimPrefix(domain, "*")
		if !existingDomains[clean] {
			dstHost := domain
			if !strings.HasPrefix(dstHost, "*") {
				dstHost = "*" + dstHost
			}
			cmd := command.Command{
				Raw: "/ip/hotspot/walled-garden/add",
				Args: map[string]string{
					"dst-host": dstHost,
					"action":   "allow",
					"comment":  "polyglot:isolation",
				},
			}
			if _, err := g.exec(ctx, driver, cmd); err != nil {
				return fmt.Errorf("add walled-garden domain %s: %w", domain, err)
			}
		}
	}

	// 2. If portalHost/Port provided, ensure /ip/hotspot/walled-garden/ip
	if portalHost != "" {
		if portalPort == "" {
			portalPort = "80"
		}
		resIP, err := g.exec(ctx, driver, command.Command{Raw: "/ip/hotspot/walled-garden/ip/print"})
		if err == nil {
			found := false
			for _, row := range resIP.Rows {
				if row["dst-address"] == portalHost && (row["dst-port"] == portalPort || portalPort == "") {
					found = true
					break
				}
			}
			if !found {
				cmd := command.Command{
					Raw: "/ip/hotspot/walled-garden/ip/add",
					Args: map[string]string{
						"dst-address": portalHost,
						"dst-port":    portalPort,
						"protocol":    "tcp",
						"action":      "accept",
						"comment":     "polyglot:portal",
					},
				}
				_, _ = g.exec(ctx, driver, cmd)
			}
		}
	}

	return nil
}
