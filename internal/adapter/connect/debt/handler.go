package debt

import (
	"context"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	domainDebt "github.com/quixiq/polyglot/internal/domain/debt"
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

func toProtoDebt(d *domainDebt.Debt) *devicepb.Debt {
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
	return &devicepb.Debt{
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

func toProtoDebtPayment(p *domainDebt.Payment) *devicepb.DebtPayment {
	if p == nil {
		return nil
	}
	var pmID string
	if p.PaymentMethodID != nil {
		pmID = *p.PaymentMethodID
	}
	return &devicepb.DebtPayment{
		Id:              p.ID,
		DebtId:          p.DebtID,
		Amount:          p.Amount,
		PaymentMethodId: pmID,
		PaymentDateUnix: p.PaymentDate.Unix(),
		Notes:           p.Notes,
	}
}

func (h *DebtConnectHandler) ListDebts(ctx context.Context, req *connect.Request[devicepb.ListDebtsRequest]) (*connect.Response[devicepb.ListDebtsResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	debts, total, err := h.svc.ListDebts(ctx, port.DebtFilter{
		TenantID:    "tenant-default",
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
	out := make([]*devicepb.Debt, len(debts))
	for i := range debts {
		out[i] = toProtoDebt(&debts[i])
	}
	return connect.NewResponse(&devicepb.ListDebtsResponse{
		Debts: out,
		Total: int32(total),
	}), nil
}

func (h *DebtConnectHandler) GetDebt(ctx context.Context, req *connect.Request[devicepb.GetDebtRequest]) (*connect.Response[devicepb.GetDebtResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	d, err := h.svc.GetDebt(ctx, req.Msg.GetId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetDebtResponse{
		Debt: toProtoDebt(&d),
	}), nil
}

func (h *DebtConnectHandler) CreateDebt(ctx context.Context, req *connect.Request[devicepb.CreateDebtRequest]) (*connect.Response[devicepb.CreateDebtResponse], error) {
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
		TenantID:       "tenant-default",
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
	return connect.NewResponse(&devicepb.CreateDebtResponse{
		Debt: toProtoDebt(&d),
	}), nil
}

func (h *DebtConnectHandler) UpdateDebt(ctx context.Context, req *connect.Request[devicepb.UpdateDebtRequest]) (*connect.Response[devicepb.UpdateDebtResponse], error) {
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
	return connect.NewResponse(&devicepb.UpdateDebtResponse{
		Debt: toProtoDebt(&d),
	}), nil
}

func (h *DebtConnectHandler) CancelDebt(ctx context.Context, req *connect.Request[devicepb.CancelDebtRequest]) (*connect.Response[devicepb.CancelDebtResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	d, err := h.svc.CancelDebt(ctx, req.Msg.GetId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CancelDebtResponse{
		Debt: toProtoDebt(&d),
	}), nil
}

func (h *DebtConnectHandler) PayDebt(ctx context.Context, req *connect.Request[devicepb.PayDebtRequest]) (*connect.Response[devicepb.PayDebtResponse], error) {
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
		TenantID:        "tenant-default",
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
	return connect.NewResponse(&devicepb.PayDebtResponse{
		Payment: toProtoDebtPayment(&payment),
		Debt:    toProtoDebt(&d),
	}), nil
}

func (h *DebtConnectHandler) ListDebtPayments(ctx context.Context, req *connect.Request[devicepb.ListDebtPaymentsRequest]) (*connect.Response[devicepb.ListDebtPaymentsResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	payments, err := h.svc.ListDebtPayments(ctx, req.Msg.GetDebtId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*devicepb.DebtPayment, len(payments))
	for i := range payments {
		out[i] = toProtoDebtPayment(&payments[i])
	}
	return connect.NewResponse(&devicepb.ListDebtPaymentsResponse{
		Payments: out,
	}), nil
}

func (h *DebtConnectHandler) GetDebtSummary(ctx context.Context, req *connect.Request[devicepb.GetDebtSummaryRequest]) (*connect.Response[devicepb.GetDebtSummaryResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	sum, err := h.svc.GetDebtSummary(ctx, "tenant-default", req.Msg.GetCustomerId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetDebtSummaryResponse{
		Summary: &devicepb.DebtSummary{
			Count:       int32(sum.Count),
			TotalAmount: sum.TotalAmount,
			Outstanding: sum.Outstanding,
			Overdue:     int32(sum.Overdue),
		},
	}), nil
}

func (h *DebtConnectHandler) ExportDebts(ctx context.Context, req *connect.Request[devicepb.ExportDebtsRequest]) (*connect.Response[devicepb.ExportDebtsResponse], error) {
	if h.svc == nil {
		return nil, response.Unavailable("debt service unavailable")
	}
	var f port.DebtFilter
	if req.Msg.GetFilter() != nil {
		flt := req.Msg.GetFilter()
		f = port.DebtFilter{
			TenantID:    "tenant-default",
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
	return connect.NewResponse(&devicepb.ExportDebtsResponse{
		Payload:     payload,
		ContentType: contentType,
		Filename:    filename,
	}), nil
}
