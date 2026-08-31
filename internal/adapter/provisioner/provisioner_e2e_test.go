//go:build mikrotik_e2e

package provisioner

// E2E test siklus hidup akun router terhadap MikroTik SUNGGUHAN.
//
// Gate:
//   - Hanya jalan bila MIKROTIK_TEST_HOST terisi (.env project sudah menyediakannya). Selain itu test di-skip.
//   - Semua artefak memakai prefix "e2e-" dan dibersihkan di cleanup.
//   - Rule NAT redirect dibuat Disabled=true (tak mengganggu trafik).
//
// Jalankan eksplisit: go test -tags=mikrotik_e2e ./internal/adapter/provisioner/ -run TestProvisioner_E2E -v

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
)

func e2eTarget(t *testing.T) device.Target {
	t.Helper()
	_ = godotenv.Load(filepath.Join("..", "..", "..", ".env"))
	host := os.Getenv("MIKROTIK_TEST_HOST")
	if host == "" {
		t.Skip("MIKROTIK_TEST_HOST tidak diset — lewati E2E real router")
	}
	p := 8728
	if v := os.Getenv("MIKROTIK_TEST_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			p = n
		}
	}
	return device.Target{
		Host:     host,
		Port:     p,
		Username: os.Getenv("MIKROTIK_TEST_USER"),
		Password: os.Getenv("MIKROTIK_TEST_PASS"),
		Timeout:  15 * time.Second,
	}
}

func TestProvisioner_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	target := e2eTarget(t)
	ctx := context.Background()

	drv, err := mikrotik.NewDriver(ctx, target)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "connect") {
		t.Skipf("router unreachable (%v) — lewati E2E", err)
	}
	requireNoErr(t, err, "connect driver")

	exec := networkUC.ExecuteCommandPreApproved

	gw := mikrotik.NewGateway(exec)
	mgr := NewWithResolver(
		func(context.Context, string) (port.DeviceDriver, error) { return drv, nil },
		gw,
		nil,
		gw,
		nil,
	)

	const (
		deviceID      = "e2e-direct"
		username      = "e2etestuser"
		password      = "e2epass123"
		planProfile   = "e2e-test-plan"
		isolirProfile = "e2e-test-isolir"
		addressList   = "E2E_ISOLIR_USERS"
	)
	acct := port.SubscriberAccount{
		Username: username, Password: password,
		Profile: planProfile, RateLimit: "2M/2M",
		Comment: "polyglot:e2e",
	}

	t.Cleanup(func() {
		cctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = gw.RemoveFromAddressListByComment(cctx, drv, addressList, "isolir:"+username)
		_ = mgr.Terminate(cctx, deviceID, "PPPOE", username)
		_ = removeProfileIfExists(cctx, gw, drv, planProfile)
		_ = removeProfileIfExists(cctx, gw, drv, isolirProfile)
	})

	// 1. Provision: auto-buat profil paket + secret.
	requireNoErr(t, mgr.Provision(ctx, deviceID, "PPPOE", acct), "provision")
	assertSecret(t, ctx, gw, drv, username, func(s port.PPPoESecret) bool {
		return s.Profile == planProfile && !s.Disabled
	}, "setelah provision")

	// Profil paket auto-terbuat.
	profiles, err := gw.ListProfiles(ctx, drv, planProfile)
	requireNoErr(t, err, "list profiles")
	foundPlan := false
	for _, pr := range profiles {
		if pr.Name == planProfile {
			foundPlan = true
		}
	}
	requireTrue(t, foundPlan, "profil paket %s harus auto-dibuat", planProfile)

	// 2. Isolate: profil isolir + rule redirect (disabled utk keamanan).
	opt := port.IsolationOptions{
		IsolirProfile: isolirProfile,
		AddressList:   addressList,
		Redirect: &port.IsolationRedirectConfig{
			SrcAddressList: addressList,
			PaymentHost:    "192.168.233.254",
			PaymentPort:    "8080",
			Protocols:      []string{"tcp"},
			DstPorts:       []string{"80", "443"},
			Disabled:       true,
		},
	}
	requireNoErr(t, mgr.Isolate(ctx, deviceID, "PPPOE", username, opt), "isolate")
	assertSecret(t, ctx, gw, drv, username, func(s port.PPPoESecret) bool {
		return s.Profile == isolirProfile
	}, "setelah isolate")

	// 3. Restore: kembali ke profil paket.
	requireNoErr(t, mgr.Restore(ctx, deviceID, "PPPOE", username, planProfile, addressList), "restore")
	assertSecret(t, ctx, gw, drv, username, func(s port.PPPoESecret) bool {
		return s.Profile == planProfile
	}, "setelah restore")

	// 4. Suspend: disabled=true.
	requireNoErr(t, mgr.Suspend(ctx, deviceID, "PPPOE", username), "suspend")
	assertSecret(t, ctx, gw, drv, username, func(s port.PPPoESecret) bool {
		return s.Disabled
	}, "setelah suspend")

	// 5. Terminate: secret terhapus.
	requireNoErr(t, mgr.Terminate(ctx, deviceID, "PPPOE", username), "terminate")
	gone, err := findSecretByName(ctx, gw, drv, username)
	requireNoErr(t, err, "find after terminate")
	requireTrue(t, gone == nil, "secret harus terhapus setelah terminate")

	// Rule redirect tetap ada (disabled) untuk audit.
	rules, err := gw.ListFirewallNATRules(ctx, drv, "dstnat", "", addressList)
	requireNoErr(t, err, "list nat")
	rules = mikrotik.FindIsolationRedirectRules(rules)
	requireTrue(t, len(rules) >= 1, "rule redirect milik app harus ada (disabled)")
	for _, r := range rules {
		requireTrue(t, r.Disabled, "rule redirect E2E wajib disabled")
	}
}

func requireNoErr(t *testing.T, err error, what string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", what, err)
	}
}

func requireTrue(t *testing.T, cond bool, format string, args ...any) {
	t.Helper()
	if !cond {
		t.Fatalf("E2E assertion gagal: "+format, args...)
	}
}

func findSecretByName(ctx context.Context, gw *mikrotik.Gateway, drv port.DeviceDriver, name string) (*port.PPPoESecret, error) {
	secrets, err := gw.ListSecrets(ctx, drv, name)
	if err != nil {
		return nil, err
	}
	for _, s := range secrets {
		if s.Name == name {
			cp := s
			return &cp, nil
		}
	}
	return nil, nil
}

func assertSecret(t *testing.T, ctx context.Context, gw *mikrotik.Gateway, drv port.DeviceDriver, name string, cond func(port.PPPoESecret) bool, when string) {
	t.Helper()
	s, err := findSecretByName(ctx, gw, drv, name)
	requireNoErr(t, err, "find secret "+when)
	requireTrue(t, s != nil, "secret %s harus ada %s", name, when)
	requireTrue(t, cond(*s), "kondisi secret salah %s: profile=%s disabled=%v",
		when, s.Profile, s.Disabled)
}

func removeProfileIfExists(ctx context.Context, gw *mikrotik.Gateway, drv port.DeviceDriver, profileName string) error {
	profiles, err := gw.ListProfiles(ctx, drv, profileName)
	if err != nil {
		return err
	}
	for _, pr := range profiles {
		if pr.Name == profileName {
			if _, err := gw.RemoveProfile(ctx, drv, pr.RosID); err != nil {
				return fmt.Errorf("remove profile: %w", err)
			}
		}
	}
	return nil
}
