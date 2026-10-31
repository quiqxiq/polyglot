package provisioner

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	domainCommand "github.com/quixiq/polyglot/internal/domain/command"
	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// ensureIsolirWebProxy mengaktifkan webproxy internal MikroTik dan memasang
// access rules isolir (blueprint §5.4, planning v2 D12): whitelist server
// billing, redirect semua request lain ke halaman bayar, izinkan port proxy.
func (p *Provisioner) ensureIsolirWebProxy(ctx context.Context, driver port.DeviceDriver, cfg domainDevice.IsolationConfig) error {
	if !cfg.WebProxyEnabled {
		return nil
	}
	proxyPort := cfg.WebProxyPort
	if proxyPort <= 0 {
		proxyPort = 3125
	}
	portStr := strconv.Itoa(proxyPort)

	if _, err := driver.Execute(ctx, domainCommand.Command{
		Raw: "/ip/proxy/set",
		Args: map[string]string{
			"enabled":        "yes",
			"port":           portStr,
			"cache-path":     "webproxy",
			"max-cache-size": "none",
		},
	}); err != nil {
		return fmt.Errorf("enable web proxy: %w", err)
	}

	billingHost := cfg.RedirectIP
	if len(cfg.BillingHosts) > 0 && cfg.BillingHosts[0] != "" {
		billingHost = cfg.BillingHosts[0]
	}
	if billingHost == "" {
		return nil
	}
	srcSubnet := subnet24(cfg.LocalAddress)

	// Bersihkan access rule lama milik aplikasi (idempotent re-apply).
	if res, err := driver.Execute(ctx, domainCommand.Command{Raw: "/ip/proxy/access/print"}); err == nil {
		for _, row := range res.Rows {
			if row["comment"] == "polyglot:isolir" {
				if id := row[".id"]; id != "" {
					_, _ = driver.Execute(ctx, domainCommand.Command{
						Raw:  "/ip/proxy/access/remove",
						Args: map[string]string{"numbers": id},
					})
				}
			}
		}
	}

	// 1. Izinkan akses ke server billing (whitelist).
	for _, host := range cfg.BillingHosts {
		if strings.TrimSpace(host) == "" {
			continue
		}
		_, _ = driver.Execute(ctx, domainCommand.Command{
			Raw: "/ip/proxy/access/add",
			Args: map[string]string{
				"action":      "allow",
				"dst-host":    host,
				"src-address": srcSubnet,
				"comment":     "polyglot:isolir",
			},
		})
	}
	// 2. Redirect semua request lain ke halaman bayar.
	if _, err := driver.Execute(ctx, domainCommand.Command{
		Raw: "/ip/proxy/access/add",
		Args: map[string]string{
			"action":      "deny",
			"src-address": srcSubnet,
			"redirect-to": billingHost + ":" + portStr,
			"comment":     "polyglot:isolir",
		},
	}); err != nil {
		return fmt.Errorf("proxy redirect rule: %w", err)
	}
	// 3. Izinkan koneksi ke port proxy itu sendiri.
	_, err := driver.Execute(ctx, domainCommand.Command{
		Raw: "/ip/proxy/access/add",
		Args: map[string]string{
			"action":      "allow",
			"src-address": srcSubnet,
			"dst-port":    portStr,
			"comment":     "polyglot:isolir",
		},
	})
	return err
}

// ensureBillingWhitelist menandai host server billing ke address-list
// IPBILL-GLOBAL agar tetap dapat diakses saat isolir.
func (p *Provisioner) ensureBillingWhitelist(ctx context.Context, driver port.DeviceDriver, hosts []string) error {
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}
		if err := p.fw.AddToAddressList(ctx, driver, "IPBILL-GLOBAL", host, "polyglot:isolir"); err != nil {
			return fmt.Errorf("whitelist billing %s: %w", host, err)
		}
	}
	return nil
}

// fetchIsolirErrorPage mengunduh error.html ke flash router via /tool fetch
// (mode self, blueprint §5.4.C). Best-effort: bila gagal, mode cloud tetap
// berjalan lewat portal isolir.
func (p *Provisioner) fetchIsolirErrorPage(ctx context.Context, driver port.DeviceDriver, pageURL string) {
	pageURL = strings.TrimSpace(pageURL)
	if pageURL == "" {
		return
	}
	for _, dst := range []string{"flash/webproxy/error.html", "webproxy/error.html"} {
		if _, err := driver.Execute(ctx, domainCommand.Command{
			Raw: "/tool/fetch",
			Args: map[string]string{
				"url":      pageURL,
				"dst-path": dst,
				"mode":     "overwrite",
			},
		}); err == nil {
			return
		}
	}
	logger.WithComponent("Provisioner").WithFields(map[string]any{
		"url": pageURL,
	}).Warn("gagal fetch error.html isolir; mode cloud tetap dipakai")
}

// subnet24 menurunkan subnet /24 dari alamat IPv4 lokal (mis. 10.100.0.1 →
// 10.100.0.0/24). Bila format tidak dikenal, kembalikan alamat apa adanya.
func subnet24(addr string) string {
	addr = strings.TrimSpace(addr)
	parts := strings.Split(addr, ".")
	if len(parts) == 4 {
		return parts[0] + "." + parts[1] + "." + parts[2] + ".0/24"
	}
	return addr
}
