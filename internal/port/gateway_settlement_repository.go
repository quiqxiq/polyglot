package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// GatewaySettlementFilter menyaring daftar settlement gateway.
type GatewaySettlementFilter struct {
	TenantID string
	Gateway  string
	Status   string
	From     *time.Time
	To       *time.Time
	Limit    int
}

// GatewaySettlementRepository adalah persistensi settlement/pencairan gateway
// untuk rekonsiliasi — W4 PLAN-BILLING-OPS-V3.
type GatewaySettlementRepository interface {
	Save(ctx context.Context, s billing.GatewaySettlement) error
	FindByID(ctx context.Context, id string) (billing.GatewaySettlement, error)
	FindByRef(ctx context.Context, tenantID, gateway, ref string) (billing.GatewaySettlement, error)
	List(ctx context.Context, f GatewaySettlementFilter) ([]billing.GatewaySettlement, error)
}
