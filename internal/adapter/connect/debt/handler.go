package debt

import (
	"context"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	domainDebt "github.com/quixiq/polyglot/internal/domain/debt"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	debtUC "github.com/quixiq/polyglot/internal/usecase/debt"
	"github.com/quixiq/polyglot/pkg/response"
)

// DebtConnectHandler implements polyglot.v1.DebtService.
type DebtConnectHandler struct {
	svc *debtUC.Service
}

// NewDebtConnectHandler creates a new DebtConnectHandler.
func NewDebtConnectHandler(svc *debtUC.Service) *DebtConnectHandler {
	return &DebtConnectHandler{svc: svc}
}

func toProtoDebt(d *domainDebt.Debt) *polyglotv1.Debt {
	if d == nil {
		return nil
	}
	var subID string
	if d.SubscriptionID != nil {
		subID = *d.SubscriptionID
	}
	var dueUnix, paidUnix int64
	if d.DueDate != nil {
		dueUnix = d.DueDate.Unix()
	}
	if d.PaidAt != nil {
		paidUnix = d.PaidAt.Unix()
	}
	return &polyglotv1.Debt{
		Id:              d.ID,
		CustomerId:      d.CustomerID,
		SubscriptionId:  subID,
		Amount:          d.Amount,
		RemainingAmount: d.RemainingAmount,
		Description:     d.Description,
		DebtDateUnix:    d.DebtDate.Unix(),
		DueDateUnix:     dueUnix,
		Status:          d.Status,
		PaidAtUnix:      paidUnix,
		CreatedAtUnix:   d.CreatedAt.Unix(),
	}
}

func toProtoDebtPayment(p *domainDebt.Payment) *polyglotv1.DebtPayment {
	if p == nil {
		return nil
	}
	var pmID string
	if p.PaymentMethodID != nil {
		pmID = *p.PaymentMethodID
	}
	return &polyglotv1.DebtPayment{
		Id:              p.ID,
		DebtId:          p.DebtID,
		Amount:          p.Amount,
		PaymentMethodId: pmID,
		PaymentDateUnix: p.PaymentDate.Unix(),
		Notes:           p.Notes,
	}
}

func (h *DebtConnectHandler) ListDebts(ctx context.Context, req *connect.Request[polyglotv1.ListDebtsRequest]) (*connect.Response[polyglotv1.ListDebtsResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	debts, total, err := h.svc.ListDebts(ctx, port.DebtFilter{
		TenantID:    domainTenant.DefaultID,
		CustomerID:  req.Msg.GetCustomerId(),
		Status:      req.Msg.GetStatus(),
		Search:      req.Msg.GetSearch(),
		OverdueOnly: req.Msg.GetOverdueOnly(),
		Limit:       int(req.Msg.GetLimit()),
		Offset:      int(req.Msg.GetOffset()),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.Debt, len(debts))
	for i := range debts {
		out[i] = toProtoDebt(&debts[i])
	}
	return connect.NewResponse(&polyglotv1.ListDebtsResponse{
		Debts: out,
		Total: int32(total),
	}), nil
}

func (h *DebtConnectHandler) GetDebt(ctx context.Context, req *connect.Request[polyglotv1.GetDebtRequest]) (*connect.Response[polyglotv1.GetDebtResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	d, err := h.svc.GetDebt(ctx, req.Msg.GetId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.GetDebtResponse{
		Debt: toProtoDebt(&d),
	}), nil
}

func (h *DebtConnectHandler) CreateDebt(ctx context.Context, req *connect.Request[polyglotv1.CreateDebtRequest]) (*connect.Response[polyglotv1.CreateDebtResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	var createdBy *uint
	if userID, _, exists := iauth.IdentityFromContext(ctx); exists && userID > 0 {
		createdBy = &userID
	}
	var subID *string
	if req.Msg.GetSubscriptionId() != "" {
		s := req.Msg.GetSubscriptionId()
		subID = &s
	}
	var debtDate time.Time
	if req.Msg.GetDebtDateUnix() > 0 {
		debtDate = time.Unix(req.Msg.GetDebtDateUnix(), 0).UTC()
	} else {
		debtDate = time.Now().UTC()
	}
	var dueDate *time.Time
	if req.Msg.GetDueDateUnix() > 0 {
		t := time.Unix(req.Msg.GetDueDateUnix(), 0).UTC()
		dueDate = &t
	}
	d, err := h.svc.CreateDebt(ctx, debtUC.CreateDebtInput{
		TenantID:       domainTenant.DefaultID,
		CustomerID:     req.Msg.GetCustomerId(),
		SubscriptionID: subID,
		Amount:         req.Msg.GetAmount(),
		Description:    req.Msg.GetDescription(),
		DebtDate:       debtDate,
		DueDate:        dueDate,
		CreatedBy:      createdBy,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.CreateDebtResponse{
		Debt: toProtoDebt(&d),
	}), nil
}

func (h *DebtConnectHandler) UpdateDebt(ctx context.Context, req *connect.Request[polyglotv1.UpdateDebtRequest]) (*connect.Response[polyglotv1.UpdateDebtResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	var dueDate *time.Time
	if req.Msg.GetDueDateUnix() > 0 {
		t := time.Unix(req.Msg.GetDueDateUnix(), 0).UTC()
		dueDate = &t
	}
	d, err := h.svc.UpdateDebt(ctx, req.Msg.GetId(), req.Msg.GetDescription(), dueDate)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.UpdateDebtResponse{
		Debt: toProtoDebt(&d),
	}), nil
}

func (h *DebtConnectHandler) CancelDebt(ctx context.Context, req *connect.Request[polyglotv1.CancelDebtRequest]) (*connect.Response[polyglotv1.CancelDebtResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	d, err := h.svc.CancelDebt(ctx, req.Msg.GetId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.CancelDebtResponse{
		Debt: toProtoDebt(&d),
	}), nil
}

func (h *DebtConnectHandler) PayDebt(ctx context.Context, req *connect.Request[polyglotv1.PayDebtRequest]) (*connect.Response[polyglotv1.PayDebtResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	var recordedBy *uint
	if userID, _, exists := iauth.IdentityFromContext(ctx); exists && userID > 0 {
		recordedBy = &userID
	}
	var pmID *string
	if req.Msg.GetPaymentMethodId() != "" {
		id := req.Msg.GetPaymentMethodId()
		pmID = &id
	}
	payment, err := h.svc.PayDebt(ctx, port.DebtPaymentCommand{
		TenantID:        domainTenant.DefaultID,
		DebtID:          req.Msg.GetDebtId(),
		Amount:          req.Msg.GetAmount(),
		PaymentMethodID: pmID,
		PaymentDate:     time.Now().UTC(),
		Notes:           req.Msg.GetNotes(),
		RecordedBy:      recordedBy,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	d, _ := h.svc.GetDebt(ctx, req.Msg.GetDebtId())
	return connect.NewResponse(&polyglotv1.PayDebtResponse{
		Payment: toProtoDebtPayment(&payment),
		Debt:    toProtoDebt(&d),
	}), nil
}

func (h *DebtConnectHandler) ListDebtPayments(ctx context.Context, req *connect.Request[polyglotv1.ListDebtPaymentsRequest]) (*connect.Response[polyglotv1.ListDebtPaymentsResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	payments, err := h.svc.ListDebtPayments(ctx, req.Msg.GetDebtId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.DebtPayment, len(payments))
	for i := range payments {
		out[i] = toProtoDebtPayment(&payments[i])
	}
	return connect.NewResponse(&polyglotv1.ListDebtPaymentsResponse{
		Payments: out,
	}), nil
}

func (h *DebtConnectHandler) GetDebtSummary(ctx context.Context, req *connect.Request[polyglotv1.GetDebtSummaryRequest]) (*connect.Response[polyglotv1.GetDebtSummaryResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	sum, err := h.svc.GetDebtSummary(ctx, domainTenant.DefaultID, req.Msg.GetCustomerId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.GetDebtSummaryResponse{
		Summary: &polyglotv1.DebtSummary{
			Count:       int32(sum.Count),
			TotalAmount: sum.TotalAmount,
			Outstanding: sum.Outstanding,
			Overdue:     int32(sum.Overdue),
		},
	}), nil
}

func (h *DebtConnectHandler) ExportDebts(ctx context.Context, req *connect.Request[polyglotv1.ExportDebtsRequest]) (*connect.Response[polyglotv1.ExportDebtsResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	var f port.DebtFilter
	if req.Msg.GetFilter() != nil {
		flt := req.Msg.GetFilter()
		f = port.DebtFilter{
			TenantID:    domainTenant.DefaultID,
			CustomerID:  flt.GetCustomerId(),
			Status:      flt.GetStatus(),
			Search:      flt.GetSearch(),
			OverdueOnly: flt.GetOverdueOnly(),
		}
	}
	payload, contentType, filename, err := h.svc.ExportDebts(ctx, f, req.Msg.GetFormat())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.ExportDebtsResponse{
		Payload:     payload,
		ContentType: contentType,
		Filename:    filename,
	}), nil
}
