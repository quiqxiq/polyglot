//go:build integration

// Integration test repositori ISP core (service_plans, customers,
// subscriptions, registrations, secrets) terhadap PostgreSQL nyata.
//
// Jalankan (docker compose dev postgres harus hidup):
//
//	make migrate-up
//	DATABASE_URL='postgres://postgres:netops@localhost:5432/netops?sslmode=disable' \
//	  go test -tags=integration ./test/integration/ -run TestISPCoreRepo -v
package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainReg "github.com/quixiq/polyglot/internal/domain/registration"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
)

func ispTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL tidak di-set — skip integration test Postgres")
	}
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func seedPlan(t *testing.T, db *gorm.DB, name string) domainPlan.Plan {
	t.Helper()
	repo := postgres.NewPlanRepository(db)
	p := domainPlan.Plan{
		ID: uuid.NewString(), Name: name, ServiceType: domainPlan.ServiceTypePPPoE,
		RateDownKbps: 10240, RateUpKbps: 10240, Price: 150000,
		IPPoolName: "PPPOE-POOL", IsActive: true,
	}
	require.NoError(t, repo.Save(context.Background(), p))
	return p
}

// seedDevice inserts one minimal devices row so subscription FKs resolve.
func seedDevice(t *testing.T, db *gorm.DB) string {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.WithContext(context.Background()).Exec(
		`INSERT INTO devices (id, tenant_id, name, vendor, driver_type, host, port)
		 VALUES (?, 'tenant-default', 'TEST-ROUTER', 'mikrotik', 'mikrotik', '10.0.0.1', 8728)`,
		id).Error)
	return id.String()
}

func TestISPCoreRepo_PlanCRUD(t *testing.T) {
	ctx := context.Background()
	db := ispTestDB(t)
	repo := postgres.NewPlanRepository(db)
	p := seedPlan(t, db, "TEST-PLAN-"+uuid.NewString()[:8])

	got, err := repo.FindByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, p.Name, got.Name)
	assert.Equal(t, "PPPOE", got.ServiceType)

	list, err := repo.List(ctx, "PPPOE", true, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, list)

	require.NoError(t, repo.Delete(ctx, p.ID))
	_, err = repo.FindByID(ctx, p.ID)
	assert.ErrorIs(t, err, domainPlan.ErrNotFound)
}

func TestISPCoreRepo_CustomerLifecycle(t *testing.T) {
	ctx := context.Background()
	db := ispTestDB(t)
	repo := postgres.NewCustomerRepository(db)

	code, err := repo.NextCustomerCode(ctx)
	require.NoError(t, err)
	assert.Regexp(t, `^\d{5}$`, code)

	c := domainCustomer.Customer{
		ID: uuid.NewString(), TenantID: "tenant-default",
		CustomerCode: code, Name: "Budi", Phone: "0812" + uuid.NewString()[:8],
		Address: "Jl. Test", Status: "ACTIVE",
	}
	require.NoError(t, repo.Save(ctx, c))

	found, err := repo.FindByPhone(ctx, c.Phone)
	require.NoError(t, err)
	assert.Equal(t, c.ID, found.ID)

	now := time.Now()
	require.NoError(t, repo.SoftDelete(ctx, c.ID, now))
	_, err = repo.FindByID(ctx, c.ID)
	assert.ErrorIs(t, err, domainCustomer.ErrNotFound, "soft-deleted harus tak terlihat")
}

func TestISPCoreRepo_SubscriptionMappingUnique(t *testing.T) {
	ctx := context.Background()
	db := ispTestDB(t)
	subRepo := postgres.NewSubscriptionRepository(db)
	custRepo := postgres.NewCustomerRepository(db)

	p := seedPlan(t, db, "TEST-SUB-"+uuid.NewString()[:8])
	deviceID := seedDevice(t, db)

	c := domainCustomer.Customer{
		ID: uuid.NewString(), TenantID: "tenant-default",
		Name: "Tini", Phone: "0813" + uuid.NewString()[:8], Status: "ACTIVE",
	}
	require.NoError(t, custRepo.Save(ctx, c))

	base := domainSub.Subscription{
		ID: uuid.NewString(), TenantID: "tenant-default",
		CustomerID: c.ID, PlanID: p.ID,
		DeviceID: deviceID, ServiceType: domainPlan.ServiceTypePPPoE,
		RemoteUsername: "UNIQ-" + uuid.NewString()[:6],
		BillingDay:     15, Status: domainSub.StatusActive,
	}
	require.NoError(t, subRepo.Save(ctx, base))

	// Mapping sama pada langganan kedua → ditolak unique index parsial.
	dup := base
	dup.ID = uuid.NewString()
	err := subRepo.Save(ctx, dup)
	require.Error(t, err, "device_id+remote_username harus unik untuk langganan aktif")

	// Soft delete yang pertama → username boleh dipakai lagi.
	now := time.Now()
	require.NoError(t, db.WithContext(ctx).Table("subscriptions").
		Where("id = ?", base.ID).Update("deleted_at", now).Error)

	dup.RemoteUsername = base.RemoteUsername
	require.NoError(t, subRepo.Save(ctx, dup), "setelah soft-delete, mapping boleh dipakai ulang")
}

func TestISPCoreRepo_SubscriptionIsolationState(t *testing.T) {
	ctx := context.Background()
	db := ispTestDB(t)
	subRepo := postgres.NewSubscriptionRepository(db)
	custRepo := postgres.NewCustomerRepository(db)

	p := seedPlan(t, db, "TEST-ISOLIR-"+uuid.NewString()[:8])
	c := domainCustomer.Customer{
		ID: uuid.NewString(), TenantID: "tenant-default",
		Name: "Joko", Phone: "0814" + uuid.NewString()[:8], Status: "ACTIVE",
	}
	require.NoError(t, custRepo.Save(ctx, c))
	sub := domainSub.Subscription{
		ID: uuid.NewString(), CustomerID: c.ID, PlanID: p.ID,
		ServiceType: domainPlan.ServiceTypeHotspot,
		RemoteUsername: "HS-" + uuid.NewString()[:6],
		Status:      domainSub.StatusActive,
	}
	require.NoError(t, subRepo.Save(ctx, sub))

	now := time.Now()
	require.NoError(t, subRepo.SetIsolation(ctx, sub.ID, domainSub.StatusIsolated, &now, "overdue 30 hari"))
	got, err := subRepo.FindByID(ctx, sub.ID)
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusIsolated, got.Status)
	require.NotNil(t, got.IsolatedAt)
	assert.Equal(t, "overdue 30 hari", got.IsolationReason)

	require.NoError(t, subRepo.SetIsolation(ctx, sub.ID, domainSub.StatusActive, nil, ""))
	got, err = subRepo.FindByID(ctx, sub.ID)
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusActive, got.Status)
	assert.Nil(t, got.IsolatedAt)
}

func TestISPCoreRepo_RegistrationFlowAndGuards(t *testing.T) {
	ctx := context.Background()
	db := ispTestDB(t)
	regRepo := postgres.NewRegistrationRepository(db)
	custRepo := postgres.NewCustomerRepository(db)

	p := seedPlan(t, db, "TEST-REG-"+uuid.NewString()[:8])
	reg := domainReg.Registration{
		ID: uuid.NewString(), TenantID: "tenant-default",
		RegistrationNo: "REG-TEST-" + uuid.NewString()[:8],
		PlanID:         p.ID, FullName: "Siti", Phone: "0815" + uuid.NewString()[:8],
		Address: "Dusun Test", Status: domainReg.StatusPending,
	}
	require.NoError(t, regRepo.Save(ctx, reg))

	active, err := regRepo.HasActiveByPhone(ctx, reg.Phone)
	require.NoError(t, err)
	assert.True(t, active)

	// Konversi: buat customer lalu tautkan; FK registrations.customer_id valid.
	c := domainCustomer.Customer{
		ID: uuid.NewString(), TenantID: "tenant-default",
		Name: reg.FullName, Phone: reg.Phone, Status: "ACTIVE",
	}
	require.NoError(t, custRepo.Save(ctx, c))
	require.NoError(t, db.WithContext(ctx).Table("registrations").
		Where("id = ?", reg.ID).
		Updates(map[string]any{"customer_id": c.ID, "status": domainReg.StatusInstalled}).Error)

	// Plan tidak boleh terhapus selama masih direferensikan registrasi.
	planRepo := postgres.NewPlanRepository(db)
	err = planRepo.Delete(ctx, p.ID)
	require.Error(t, err, "FK RESTRICT harus memblokir hapus plan yang dipakai")
}
