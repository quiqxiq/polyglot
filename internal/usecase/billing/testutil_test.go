package billing_test

import (
	"errors"
	"time"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
)

// timeAfter returns now + days (negatif = lampau).
func timeAfter(days int) time.Time {
	return time.Now().AddDate(0, 0, days)
}

func assert_AnError() error { return errors.New("simulated router failure") }

func newPlan(id, name string) domainPlan.ServicePlan {
	return domainPlan.ServicePlan{
		ID:                    id,
		TenantID:              "tenant-default",
		Name:                  name,
		ServiceType:           domainPlan.TypePPPoE,
		BandwidthDownloadKbps: 5120,
		BandwidthUploadKbps:   5120,
		Price:                 100000,
		IsActive:              true,
	}
}

func customerWithPortal(id, portal string) domainCustomer.Customer {
	return domainCustomer.Customer{
		ID:               id,
		TenantID:         "tenant-default",
		CustomerCode:     "CUST-" + id,
		Name:             "Pelanggan " + id,
		Phone:            "085200000000",
		Address:          "Jl. Test",
		PortalAccessCode: portal,
		Status:           domainCustomer.StatusActive,
	}
}
