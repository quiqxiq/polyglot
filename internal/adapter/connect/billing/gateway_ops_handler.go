package billing

import (
	"context"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// ─── Gateway Operations ─────────────────────────────────────────────────────

// ListGatewayTransactions returns filtered gateway transactions.
func (h *BillingConnectHandler) ListGatewayTransactions(ctx context.Context, req *connect.Request[polyglotv1.ListGatewayTransactionsRequest]) (*connect.Response[polyglotv1.ListGatewayTransactionsResponse], error) {
	if h.gwTxRepo == nil {
		return nil, response.Unavailable("gateway transaction repository unavailable")
	}
	fRepo, ok := h.gwTxRepo.(port.GatewayTransactionFilteredRepository)
	if !ok {
		return nil, response.Unavailable("gateway transaction filtering unavailable")
	}
	txs, total, err := fRepo.ListFiltered(ctx, port.GatewayTransactionFilter{
		TenantID:  domainTenant.DefaultID,
		Gateway:   req.Msg.GetGateway(),
		Status:    req.Msg.GetStatus(),
		InvoiceID: req.Msg.GetInvoiceId(),
		Limit:     int(req.Msg.GetLimit()),
		Offset:    int(req.Msg.GetOffset()),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.GatewayTransaction, len(txs))
	for i, t := range txs {
		var paidAt int64
		if t.PaidAt != nil {
			paidAt = t.PaidAt.Unix()
		}
		var invID, payID string
		if t.InvoiceID != nil {
			invID = *t.InvoiceID
		}
		if t.PaymentID != nil {
			payID = *t.PaymentID
		}
		out[i] = &polyglotv1.GatewayTransaction{
			Id:             t.ID,
			Gateway:        t.Gateway,
			ExternalId:     t.ExternalID,
			InvoiceId:      invID,
			PaymentId:      payID,
			Amount:         t.Amount,
			FeeAmount:      t.FeeAmount,
			Status:         t.Status,
			PaymentChannel: t.PaymentChannel,
			PaymentUrl:     t.PaymentURL,
			CreatedAtUnix:  t.CreatedAt.Unix(),
			PaidAtUnix:     paidAt,
		}
	}
	return connect.NewResponse(&polyglotv1.ListGatewayTransactionsResponse{
		Transactions: out,
		Total:        int32(total),
	}), nil
}

// SyncGatewayTransactions pulls pending gateway transaction updates.
func (h *BillingConnectHandler) SyncGatewayTransactions(ctx context.Context, _ *connect.Request[polyglotv1.SyncGatewayTransactionsRequest]) (*connect.Response[polyglotv1.SyncGatewayTransactionsResponse], error) {
	if h.chargeUC == nil {
		return nil, response.Unavailable("gateway usecase unavailable")
	}
	synced, expired, err := h.chargeUC.SyncPendingTransactions(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.SyncGatewayTransactionsResponse{
		Synced:  int32(synced),
		Expired: int32(expired),
	}), nil
}

// ListGatewaySettlements returns recorded gateway settlements.
func (h *BillingConnectHandler) ListGatewaySettlements(ctx context.Context, req *connect.Request[polyglotv1.ListGatewaySettlementsRequest]) (*connect.Response[polyglotv1.ListGatewaySettlementsResponse], error) {
	if h.settlementRepo == nil {
		return nil, response.Unavailable("gateway settlement repository unavailable")
	}
	settlements, err := h.settlementRepo.List(ctx, port.GatewaySettlementFilter{
		TenantID: domainTenant.DefaultID,
		Gateway:  req.Msg.GetGateway(),
		Status:   req.Msg.GetStatus(),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.GatewaySettlement, len(settlements))
	for i, s := range settlements {
		var fromUnix, toUnix int64
		if s.PeriodFrom != nil {
			fromUnix = s.PeriodFrom.Unix()
		}
		if s.PeriodTo != nil {
			toUnix = s.PeriodTo.Unix()
		}
		out[i] = &polyglotv1.GatewaySettlement{
			Id:             s.ID,
			Gateway:        s.Gateway,
			SettlementRef:  s.SettlementRef,
			PeriodFromUnix: fromUnix,
			PeriodToUnix:   toUnix,
			GrossAmount:    s.GrossAmount,
			FeeAmount:      s.FeeAmount,
			NetAmount:      s.NetAmount,
			Status:         s.Status,
			Notes:          s.Notes,
		}
	}
	return connect.NewResponse(&polyglotv1.ListGatewaySettlementsResponse{
		Settlements: out,
	}), nil
}

// SaveGatewaySettlement creates or updates a gateway settlement.
func (h *BillingConnectHandler) SaveGatewaySettlement(ctx context.Context, req *connect.Request[polyglotv1.SaveGatewaySettlementRequest]) (*connect.Response[polyglotv1.SaveGatewaySettlementResponse], error) {
	if h.settlementRepo == nil {
		return nil, response.Unavailable("gateway settlement repository unavailable")
	}
	s := req.Msg.GetSettlement()
	var from, to *time.Time
	if s.GetPeriodFromUnix() > 0 {
		t := time.Unix(s.GetPeriodFromUnix(), 0).UTC()
		from = &t
	}
	if s.GetPeriodToUnix() > 0 {
		t := time.Unix(s.GetPeriodToUnix(), 0).UTC()
		to = &t
	}
	var createdBy *uint
	if userID, _, exists := iauth.IdentityFromContext(ctx); exists && userID > 0 {
		createdBy = &userID
	}
	domainSettlement := domainBilling.GatewaySettlement{
		ID:            s.GetId(),
		TenantID:      domainTenant.DefaultID,
		Gateway:       s.GetGateway(),
		SettlementRef: s.GetSettlementRef(),
		PeriodFrom:    from,
		PeriodTo:      to,
		GrossAmount:   s.GetGrossAmount(),
		FeeAmount:     s.GetFeeAmount(),
		NetAmount:     s.GetNetAmount(),
		Status:        s.GetStatus(),
		Notes:         s.GetNotes(),
		CreatedBy:     createdBy,
	}
	if err := h.settlementRepo.Save(ctx, domainSettlement); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.SaveGatewaySettlementResponse{
		Settlement: s,
	}), nil
}
