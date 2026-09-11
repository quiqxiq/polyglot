// Spec-builder tests (Fase 0 — PLAN-ISP-CORE-HARDENING.md).
// F2-1: seluruh jalur provisioner mencari queue dedicated bernama "dq-<username>",
// jadi spec builder harus menghasilkan nama yang sama.
package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
)

func TestBuildDedicatedProvisionSpec_QueueNameUsesDQPrefix(t *testing.T) {
	sub := domainSubscription.Subscription{
		ID:             "sub-ded-1",
		RemoteUsername: "bs1234",
		RemotePassword: "secret",
		ServiceType:    "DEDICATED",
	}
	pl := domainPlan.ServicePlan{
		ID:                    "plan-ded",
		Name:                  "DEDIC-10M",
		ServiceType:           domainPlan.TypeDedicated,
		BandwidthDownloadKbps: 10000,
		BandwidthUploadKbps:   10000,
	}

	spec := planUC.BuildDedicatedProvisionSpec(sub, pl)
	assert.Equal(t, "dq-bs1234", spec.Queue.QueueName)
}
