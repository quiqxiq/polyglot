package cashbook

import (
	"context"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainCashbook "github.com/quixiq/polyglot/internal/domain/cashbook"
	"github.com/quixiq/polyglot/internal/port"
	cashbookUC "github.com/quixiq/polyglot/internal/usecase/cashbook"
	"github.com/quixiq/polyglot/pkg/response"
)

// CashbookConnectHandler implements the cashbook ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type CashbookConnectHandler struct {
	useCase *cashbookUC.ManageCashbookUseCase
}

// NewCashbookConnectHandler constructs a cashbook ConnectRPC handler.
func NewCashbookConnectHandler(useCase *cashbookUC.ManageCashbookUseCase) *CashbookConnectHandler {
	return &CashbookConnectHandler{useCase: useCase}
}

// ListAccounts returns cash accounts.
func (h *CashbookConnectHandler) ListAccounts(ctx context.Context, req *connect.Request[devicepb.ListAccountsRequest]) (*connect.Response[devicepb.ListAccountsResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("cashbook usecase unavailable")
	}
	list, err := h.useCase.ListAccounts(ctx, req.Msg.ActiveOnly)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListAccountsResponse{
		Accounts: toProtoAccountList(list),
	}), nil
}

// SaveAccount creates or updates a cash account.
func (h *CashbookConnectHandler) SaveAccount(ctx context.Context, req *connect.Request[devicepb.SaveAccountRequest]) (*connect.Response[devicepb.SaveAccountResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("cashbook usecase unavailable")
	}
	pb := req.Msg.Account
	a, err := h.useCase.SaveAccount(ctx, domainCashbook.CashAccount{
		ID:          pb.Id,
		AccountCode: pb.AccountCode,
		Name:        pb.Name,
		Type:        pb.Type,
		IsActive:    pb.IsActive,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SaveAccountResponse{
		Account: toProtoAccount(a),
	}), nil
}

// ListCategories returns cash categories.
func (h *CashbookConnectHandler) ListCategories(ctx context.Context, req *connect.Request[devicepb.ListCategoriesRequest]) (*connect.Response[devicepb.ListCategoriesResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("cashbook usecase unavailable")
	}
	list, err := h.useCase.ListCategories(ctx, req.Msg.ActiveOnly)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListCategoriesResponse{
		Categories: toProtoCategoryList(list),
	}), nil
}

// SaveCategory creates or updates a cash category.
func (h *CashbookConnectHandler) SaveCategory(ctx context.Context, req *connect.Request[devicepb.SaveCategoryRequest]) (*connect.Response[devicepb.SaveCategoryResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("cashbook usecase unavailable")
	}
	pb := req.Msg.Category
	c, err := h.useCase.SaveCategory(ctx, domainCashbook.CashCategory{
		ID:       pb.Id,
		Name:     pb.Name,
		Type:     pb.Type,
		IsActive: pb.IsActive,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SaveCategoryResponse{
		Category: toProtoCategory(c),
	}), nil
}

// AddTransaction records a cash transaction.
func (h *CashbookConnectHandler) AddTransaction(ctx context.Context, req *connect.Request[devicepb.AddTransactionRequest]) (*connect.Response[devicepb.AddTransactionResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("cashbook usecase unavailable")
	}
	t, err := h.useCase.AddTransaction(ctx, cashbookUC.AddTransactionInput{
		AccountID:   req.Msg.AccountId,
		CategoryID:  req.Msg.CategoryId,
		Direction:   req.Msg.Direction,
		Amount:      req.Msg.Amount,
		Description: req.Msg.Description,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.AddTransactionResponse{
		Transaction: toProtoTransaction(t),
	}), nil
}

// ListTransactions returns cash transactions matching the filter.
func (h *CashbookConnectHandler) ListTransactions(ctx context.Context, req *connect.Request[devicepb.ListTransactionsFilter]) (*connect.Response[devicepb.ListTransactionsResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("cashbook usecase unavailable")
	}
	f := port.CashTransactionFilter{
		AccountID:  req.Msg.AccountId,
		CategoryID: req.Msg.CategoryId,
		Direction:  req.Msg.Direction,
		Limit:      int(req.Msg.Limit),
	}
	if req.Msg.FromUnix > 0 {
		f.From = time.Unix(req.Msg.FromUnix, 0)
	}
	if req.Msg.ToUnix > 0 {
		f.To = time.Unix(req.Msg.ToUnix, 0)
	}
	list, err := h.useCase.ListTransactions(ctx, f)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListTransactionsResponse{
		Transactions: toProtoTransactionList(list),
	}), nil
}

// Balances returns account balances for the selected period.
func (h *CashbookConnectHandler) Balances(ctx context.Context, req *connect.Request[devicepb.BalancesRequest]) (*connect.Response[devicepb.BalancesResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("cashbook usecase unavailable")
	}
	f := port.CashTransactionFilter{}
	if req.Msg.FromUnix > 0 {
		f.From = time.Unix(req.Msg.FromUnix, 0)
	}
	if req.Msg.ToUnix > 0 {
		f.To = time.Unix(req.Msg.ToUnix, 0)
	}
	balances, err := h.useCase.Balances(ctx, f)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.BalancesResponse{
		BalanceByAccount: balances,
	}), nil
}
