package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/port"
)

// migrationsDir resolves the absolute path of ../../migrations from this file.
func migrationsDir(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", "..", "migrations"))
	require.NoError(t, err)
	return filepath.ToSlash(abs)
}

// startMigratedPostgres spins an ephemeral Postgres and applies all SQL
// migrations. Auto-skip bila Docker tidak tersedia agar `go test ./...`
// tetap hijau di mesin tanpa container runtime.
func startMigratedPostgres(t *testing.T) (*sql.DB, *migrate.Migrate, string) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "netops_it",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(120 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil && isDockerUnavailable(err.Error()) {
		t.Skip("docker tidak tersedia — lewati smoke test migrasi PG")
	}
	require.NoError(t, err, "start postgres container")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	mapped, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/netops_it?sslmode=disable", host, mapped.Port())

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	var ready int
	for i := 0; i < 30; i++ {
		if db.Ping() == nil {
			ready = 1
			break
		}
		time.Sleep(time.Second)
	}
	require.Equal(t, 1, ready, "postgres siap dalam 30s")

	m, err := migrate.New("file://"+migrationsDir(t), dsn)
	require.NoError(t, err)
	require.NoError(t, m.Up())
	return db, m, dsn
}

func isDockerUnavailable(msg string) bool {
	lower := strings.ToLower(msg)
	for _, marker := range []string{"docker", "daemon", "cannot connect", "connection refused"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func TestPostgresMigrationsSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	db, m, dsn := startMigratedPostgres(t)

	// ── Seed data awal masuk. ────────────────────────────────────────────
	assertSeedCount(t, db, "payment_methods", 4)
	assertSeedCount(t, db, "cash_categories", 4)
	assertSeedCount(t, db, "cash_accounts", 1)
	assertSeedCount(t, db, "notification_templates", 5)

	// Migrasi 000018: seed settings ISP (kolom provisi diuji setelah fixture FK siap).
	var ispKeys int
	require.NoError(t, db.QueryRow(
		`SELECT COUNT(*) FROM system_settings WHERE key LIKE 'isp.%'`).Scan(&ispKeys))
	assert.Equal(t, 9, ispKeys, "seed key isp.*")

	// ── Unique constraint (tenant_id, name) pada service_plans. ────────
	seedPlan := func(id, name string) error {
		_, err := db.Exec(`INSERT INTO service_plans
			(id, name, service_type, bandwidth_download_kbps, bandwidth_upload_kbps, price)
			VALUES ($1,$2,'PPPOE',5000,5000,100000)`, id, name)
		return err
	}
	require.NoError(t, seedPlan("dup-1", "DUP"))
	err := seedPlan("dup-2", "DUP")
	require.Error(t, err, "duplikat (tenant_id,name) harus ditolak")

	// ── CHECK chk_reg_status menolak status liar. ───────────────────────
	require.NoError(t, seedPlan("plan-reg", "PLAN-REG"))
	_, err = db.Exec(`INSERT INTO registrations (id, registration_no, plan_id, full_name, phone, address, status)
		VALUES ('reg-x','REG-X-1','plan-reg','Budi','085200000001','Jl. X','WEIRD')`)
	require.ErrorContains(t, err, "chk_reg_status")

	// ── FK RESTRICT: plan terreferensi subscription tak bisa dihapus. ──
	devID := insertDeviceForFK(t, db)
	_, err = db.Exec(`INSERT INTO customers (id, customer_code, name, phone, address, portal_access_code)
		VALUES ('cust-fk','CUST-FK','Budi','085200000002','Jl. FK','87654321')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO subscriptions (id, customer_id, plan_id, device_id, remote_username, remote_password_cipher, provision_status, router_profile)
		VALUES ('sub-fk','cust-fk','plan-reg',$1,'BUDI','enc:x','PENDING','PLAN-REG')`, devID)
	require.NoError(t, err)
	var pstatus, rprofile string
	require.NoError(t, db.QueryRow(
		`SELECT provision_status, router_profile FROM subscriptions WHERE id='sub-fk'`).
		Scan(&pstatus, &rprofile))
	assert.Equal(t, "PENDING", pstatus)
	assert.Equal(t, "PLAN-REG", rprofile)
	_, err = db.Exec(`DELETE FROM service_plans WHERE id = 'plan-reg'`)
	require.Error(t, err, "RESTRICT harus memblokir hapus plan terreferensi")

	// ── Atomicity: kegagalan FK mengembalikan SELURUH transaksi. ───────
	insertAtomicInvoice(t, db)

	// Processor GORM menunjuk DB yang sama via driver postgres.
	gormDB, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	proc := postgres.NewPaymentProcessor(gormDB)
	_, err = proc.ProcessCashPayment(context.Background(), port.CashPaymentCommand{
		TenantID:         "tenant-default",
		InvoiceID:        "inv-atomic",
		Amount:           110000,
		CashAccountID:    "rekening-palsu", // melanggar FK cash_transactions.account_id
		IncomeCategoryID: "cc-tagihan",
	})
	require.Error(t, err, "pembayaran dengan rekening palsu harus gagal")

	var status string
	require.NoError(t, db.QueryRow(`SELECT status FROM invoices WHERE id='inv-atomic'`).Scan(&status))
	assert.Equal(t, "UNPAID", status, "rollback penuh: invoice tetap UNPAID")
	assertZero(t, db, "payments", "invoice_id='inv-atomic'")
	assertZero(t, db, "cash_transactions", "description LIKE '%INV-ATOMIC-1%'")
	assertZero(t, db, "wa_notifications", "invoice_id='inv-atomic'")

	// Pembayaran valid → lunas + jurnal kas masuk.
	insertOKInvoice(t, db)
	pay, perr := proc.ProcessCashPayment(context.Background(), port.CashPaymentCommand{
		TenantID: "tenant-default", InvoiceID: "inv-ok", Amount: 100000,
		CashAccountID: "ca-1001-kas-kantor", IncomeCategoryID: "cc-tagihan",
	})
	require.NoError(t, perr)
	require.NotEmpty(t, pay.ID)

	var paidStatus string
	require.NoError(t, db.QueryRow(`SELECT status FROM invoices WHERE id='inv-ok'`).Scan(&paidStatus))
	assert.Equal(t, "PAID", paidStatus)
	assertOne(t, db, "cash_transactions", fmt.Sprintf("source_id='%s' AND direction='IN'", pay.ID))

	// ── Down 1 lalu Up lagi: siklus migrasi sehat. ──────────────────────
	require.NoError(t, m.Steps(-1))
	require.NoError(t, m.Up())
}

func insertAtomicInvoice(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO invoices (id, invoice_number, customer_id, period, subtotal, tax_amount,
		total, paid_amount, due_date, status, qr_payload, manual_payment_code)
		VALUES ('inv-atomic','INV-ATOMIC-1','cust-fk','2026-08',100000,0,110000,0,CURRENT_DATE,'UNPAID',
		'polyglot://invoice/inv-atomic','PAY-ATOMIC1')`)
	require.NoError(t, err)
}

func insertOKInvoice(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO invoices (id, invoice_number, customer_id, period, subtotal, tax_amount,
		total, paid_amount, due_date, status, qr_payload, manual_payment_code)
		VALUES ('inv-ok','INV-ATOMIC-2','cust-fk','2026-09',100000,0,100000,0,CURRENT_DATE,'UNPAID',
		'polyglot://invoice/inv-ok','PAY-ATOMIC2')`)
	require.NoError(t, err)
}

func assertSeedCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()
	var n int
	require.NoError(t, db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&n))
	assert.Equal(t, want, n, "seed %s", table)
}

func assertZero(t *testing.T, db *sql.DB, table, where string) {
	t.Helper()
	var n int
	require.NoError(t, db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where)).Scan(&n))
	assert.Equal(t, 0, n, "%s WHERE %s", table, where)
}

func assertOne(t *testing.T, db *sql.DB, table, where string) {
	t.Helper()
	var n int
	require.NoError(t, db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where)).Scan(&n))
	assert.Equal(t, 1, n, "%s WHERE %s", table, where)
}

func insertDeviceForFK(t *testing.T, db *sql.DB) string {
	t.Helper()
	var id string
	require.NoError(t, db.QueryRow(
		`INSERT INTO devices (name, vendor, driver_type, host) VALUES ('BRAS-1','mikrotik','mikrotik','10.0.0.1') RETURNING id`,
	).Scan(&id))
	return id
}
