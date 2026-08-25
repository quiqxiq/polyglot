package registration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainReg "github.com/quixiq/polyglot/internal/domain/registration"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
)

type fakeRegRepo struct {
	regs map[string]domainReg.Registration
}

func newFakeRegRepo() *fakeRegRepo { return &fakeRegRepo{regs: map[string]domainReg.Registration{}} }

func (f *fakeRegRepo) Save(ctx context.Context, r domainReg.Registration) error {
	f.regs[r.ID] = r
	return nil
}

func (f *fakeRegRepo) FindByID(ctx context.Context, id string) (domainReg.Registration, error) {
	r, ok := f.regs[id]
	if !ok {
		return domainReg.Registration{}, domainReg.ErrNotFound
	}
	return r, nil
}

func (f *fakeRegRepo) List(ctx context.Context, status string, limit int) ([]domainReg.Registration, error) {
	out := make([]domainReg.Registration, 0)
	for _, r := range f.regs {
		if status == "" || r.Status == status {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeRegRepo) HasActiveByPhone(ctx context.Context, phone string) (bool, error) {
	for _, r := range f.regs {
		if r.Phone == phone && (r.Status == domainReg.StatusPending || r.Status == domainReg.StatusApproved) {
			return true, nil
		}
	}
	return false, nil
}

type fakePlans struct{ plans map[string]domainPlan.Plan }

func (f *fakePlans) FindByID(ctx context.Context, id string) (domainPlan.Plan, error) {
	p, ok := f.plans[id]
	if !ok {
		return domainPlan.Plan{}, domainPlan.ErrNotFound
	}
	return p, nil
}

type fakeCustomers struct{ byPhone map[string]domainCustomer.Customer }

func newFakeCustomers() *fakeCustomers { return &fakeCustomers{byPhone: map[string]domainCustomer.Customer{}} }

func (f *fakeCustomers) Save(ctx context.Context, c domainCustomer.Customer) error {
	f.byPhone[c.Phone] = c
	return nil
}

func (f *fakeCustomers) FindByPhone(ctx context.Context, phone string) (domainCustomer.Customer, error) {
	c, ok := f.byPhone[phone]
	if !ok {
		return domainCustomer.Customer{}, domainCustomer.ErrNotFound
	}
	return c, nil
}

func (f *fakeCustomers) NextCustomerCode(ctx context.Context) (string, error) {
	return "00042", nil
}

type fakeSubs struct{ subs map[string]domainSub.Subscription }

func newFakeSubs() *fakeSubs { return &fakeSubs{subs: map[string]domainSub.Subscription{}} }

func (f *fakeSubs) Save(ctx context.Context, s domainSub.Subscription) error {
	f.subs[s.ID] = s
	return nil
}

func (f *fakeSubs) FindByDeviceAndUsername(ctx context.Context, deviceID, username string) (domainSub.Subscription, error) {
	for _, s := range f.subs {
		if s.DeviceID == deviceID && s.RemoteUsername == username && s.CustomerID != "" {
			return s, nil
		}
	}
	return domainSub.Subscription{}, domainSub.ErrNotFound
}

type fakeProvisioner struct {
	calls     int
	passwords []string
}

func (f *fakeProvisioner) Provision(ctx context.Context, sub domainSub.Subscription, password string) (domainSub.Subscription, error) {
	f.calls++
	sub.RemoteID = "*99"
	sub.Status = domainSub.StatusActive
	f.passwords = append(f.passwords, password)
	return sub, nil
}

const testDeviceID2 = "22222222-2222-2222-2222-222222222222"

func fixtureService(t *testing.T) (*RegistrationService, *fakeRegRepo, *fakeSubs, *fakeProvisioner) {
	t.Helper()
	regRepo := newFakeRegRepo()
	plans := &fakePlans{plans: map[string]domainPlan.Plan{
		"plan-1": {ID: "plan-1", Name: "100 RB", ServiceType: domainPlan.ServiceTypePPPoE, RateDownKbps: 10240, RateUpKbps: 10240},
	}}
	customers := newFakeCustomers()
	subs := newFakeSubs()
	prov := &fakeProvisioner{}
	svc := NewRegistrationService(regRepo, plans, customers, subs, prov)
	return svc, regRepo, subs, prov
}

func intakeForm(phone string) domainReg.Registration {
	return domainReg.Registration{
		FullName: "Budi Santoso", Phone: phone, Address: "Dusun Krajan",
		PlanID: "plan-1",
	}
}

func TestRegistration_Flow(t *testing.T) {
	ctx := context.Background()

	t.Run("create → approve → install memprovisi dan mengkonversi", func(t *testing.T) {
		svc, _, subs, prov := fixtureService(t)

		created, err := svc.Create(ctx, intakeForm("081234567001"))
		require.NoError(t, err)
		assert.Equal(t, domainReg.StatusPending, created.Status)
		assert.NotEmpty(t, created.RegistrationNo)
		assert.Len(t, created.RegistrationNo, len("REG-202608-xxxx"))

		reviewed, err := svc.Review(ctx, created.ID, true, 7, "oke", nil, 0)
		require.NoError(t, err)
		assert.Equal(t, domainReg.StatusApproved, reviewed.Status)
		assert.NotNil(t, reviewed.ReviewedBy)

		res, err := svc.Install(ctx, MarkInstalledInput{ID: created.ID, DeviceID: testDeviceID2})
		require.NoError(t, err)
		assert.Equal(t, domainReg.StatusActive, res.Registration.Status)
		assert.NotEmpty(t, res.CustomerID)
		assert.NotEmpty(t, res.SubscriptionID)
		assert.Equal(t, 1, prov.calls)
		assert.Len(t, prov.passwords, 1)
		assert.NotEmpty(t, prov.passwords[0], "password digenerate otomatis")

		sub := subs.subs[res.SubscriptionID]
		assert.Equal(t, "PPPOE", sub.ServiceType, "service_type snapshot dari plan")
		assert.Equal(t, "BUDI-SANTOSO", sub.RemoteUsername)
	})

	t.Run("install tanpa approval ditolak", func(t *testing.T) {
		svc, _, _, _ := fixtureService(t)
		created, err := svc.Create(ctx, intakeForm("081234567002"))
		require.NoError(t, err)
		_, err = svc.Install(ctx, MarkInstalledInput{ID: created.ID, DeviceID: testDeviceID2})
		assert.ErrorIs(t, err, domainReg.ErrInvalidTransition)
	})

	t.Run("telepon duplikat saat masih aktif ditolak", func(t *testing.T) {
		svc, _, _, _ := fixtureService(t)
		_, err := svc.Create(ctx, intakeForm("081234567003"))
		require.NoError(t, err)
		_, err = svc.Create(ctx, intakeForm("081234567003"))
		assert.ErrorIs(t, err, domainReg.ErrAlreadyPending)
	})

	t.Run("username bentrok di device diberi suffix", func(t *testing.T) {
		svc, _, subs, _ := fixtureService(t)
		// Langganan lain sudah memakai BUDI-SANTOSO di device yang sama.
		subs.subs["existing"] = domainSub.Subscription{
			ID: "existing", CustomerID: "other-cust", PlanID: "plan-1",
			DeviceID: testDeviceID2, RemoteUsername: "BUDI-SANTOSO",
		}
		created, err := svc.Create(ctx, intakeForm("081234567004"))
		require.NoError(t, err)
		_, err = svc.Review(ctx, created.ID, true, 1, "", nil, 0)
		require.NoError(t, err)
		res, err := svc.Install(ctx, MarkInstalledInput{ID: created.ID, DeviceID: testDeviceID2})
		require.NoError(t, err)
		sub := subs.subs[res.SubscriptionID]
		assert.NotEqual(t, "BUDI-SANTOSO", sub.RemoteUsername)
		assert.Contains(t, sub.RemoteUsername, "BUDI-SANTOSO-")
	})
}
