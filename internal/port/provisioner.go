package port

import (
	"context"

	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
)

// SubscriberAccount alias to domain subscription model.
type SubscriberAccount = domainSub.SubscriberAccount

// IsolationOptions membawa parameter isolation yang berasal dari
// system_settings (bukan konstanta kode).
type IsolationOptions struct {
	IsolirProfile string                   // profil tujuan saat isolir
	AddressList   string                   // penanda IP pelanggan terisolir
	Redirect      *IsolationRedirectConfig // nil = tanpa rule dst-nat global
}

// RouterAccountManager mengelola siklus hidup akun pelanggan di router BRAS
// dengan semantik ISP nyata:
//   - Provision : buat akun + pastikan profil paket ada (auto dari plan).
//   - Update    : ganti profil (ganti paket / isolir / restore).
//   - Suspend   : disabled=true — cuti/penghentian sementara.
//   - Terminate : hapus akun permanen.
//
// ISOLIR ≠ disable: akun dipindah ke profil isolir, sesi di-kick, dan IP
// ditandai address-list agar rule dst-nat redirect ke halaman bayar bekerja.
type RouterAccountManager interface {
	Provision(ctx context.Context, deviceID, serviceType string, acct SubscriberAccount) error
	ProvisionPPPoE(ctx context.Context, deviceID string, spec domainSub.PPPoEProvisionSpec) error
	ProvisionHotspot(ctx context.Context, deviceID string, spec domainSub.HotspotProvisionSpec) error
	ProvisionDedicated(ctx context.Context, deviceID string, spec domainSub.DedicatedProvisionSpec) error
	UpdateAccount(ctx context.Context, deviceID, serviceType, username, newProfile string) error
	// EnsureProfile memastikan profil bernama profileName ada di router
	// dengan rate tertentu (auto-create bila belum ada). Dipakai sebelum
	// memindahkan akun ke profil tersebut (mis. saat ganti paket).
	EnsureProfile(ctx context.Context, deviceID, serviceType, profileName, rateLimit string) error
	Isolate(ctx context.Context, deviceID, serviceType, username string, opt IsolationOptions) error
	Restore(ctx context.Context, deviceID, serviceType, username, normalProfile, addressList string) error
	Suspend(ctx context.Context, deviceID, serviceType, username string) error
	Terminate(ctx context.Context, deviceID, serviceType, username string) error

	// Manajemen Infrastruktur Isolir & Integrasi Script
	EnsureIsolationInfrastructure(ctx context.Context, deviceID string, cfg domainDevice.IsolationConfig) error
	GetIsolationInfrastructureStatus(ctx context.Context, deviceID string) (domainDevice.IsolationStatus, error)
	ApplyIntegrationScript(ctx context.Context, deviceID, profileName, serviceType, scriptType, script string) error

	// Sinkronisasi Profil Paket Layanan ke Router
	SyncPlanProfile(ctx context.Context, deviceID string, plan domainPlan.ServicePlan) error
	DeletePlanProfile(ctx context.Context, deviceID string, serviceType, profileName string) error
}
