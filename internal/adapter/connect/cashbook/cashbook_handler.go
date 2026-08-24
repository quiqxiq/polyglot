package cashbook

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainCashbook "github.com/quixiq/polyglot/internal/domain/cashbook"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/response"
)

type CashbookConnectHandler struct {
	repo port.CashbookRepository
}

func NewCashbookConnectHandler(repo port.CashbookRepository) *CashbookConnectHandler {
	return &CashbookConnectHandler{repo: repo}
}

func (h *CashbookConnectHandler) ListAccounts(ctx context.Context, req *connect.Request[devicepb.ListAccountsRequest]) (*connect.Response[devicepb.ListAccountsResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("cashbook repository unavailable"))
	}
	list, err := h.repo.FindAccounts(ctx, req.Msg.ActiveOnly)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListAccountsResponse{
		Accounts: toProtoAccountList(list),
	}), nil
}

func (h *CashbookConnectHandler) SaveAccount(ctx context.Context, req *connect.Request[devicepb.SaveAccountRequest]) (*connect.Response[devicepb.SaveAccountResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("cashbook repository unavailable"))
	}
	pb := req.Msg.Account
	if pb == nil || pb.Name == "" || pb.AccountCode == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("account_code and name are required"))
	}
	id := pb.Id
	if id == "" {
		id = idgen.New("ca")
	}
	a := domainCashbook.CashAccount{
		ID: id, TenantID: "tenant-default", AccountCode: pb.AccountCode,
		Name: pb.Name, Type: pb.Type, IsActive: pb.IsActive,
	}
	if a.Type == "" {
		a.Type = domainCashbook.AccountTypeCash
	}
	if err := h.repo.SaveAccount(ctx, a); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SaveAccountResponse{
		Account: toProtoAccount(&a),
	}), nil
}

func (h *CashbookConnectHandler) ListCategories(ctx context.Context, req *connect.Request[devicepb.ListCategoriesRequest]) (*connect.Response[devicepb.ListCategoriesResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("cashbook repository unavailable"))
	}
	list, err := h.repo.FindCategories(ctx, req.Msg.ActiveOnly)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListCategoriesResponse{
		Categories: toProtoCategoryList(list),
	}), nil
}

func (h *CashbookConnectHandler) SaveCategory(ctx context.Context, req *connect.Request[devicepb.SaveCategoryRequest]) (*connect.Response[devicepb.SaveCategoryResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("cashbook repository unavailable"))
	}
	pb := req.Msg.Category
	if pb == nil || pb.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("category name is required"))
	}
	id := pb.Id
	if id == "" {
		id = idgen.New("cc")
	}
	c := domainCashbook.CashCategory{
		ID: id, TenantID: "tenant-default", Name: pb.Name,
		Type: pb.Type, IsActive: pb.IsActive,
	}
	if c.Type == "" {
		c.Type = domainCashbook.CategoryTypeExpense
	}
	if err := h.repo.SaveCategory(ctx, c); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SaveCategoryResponse{
		Category: toProtoCategory(&c),
	}), nil
}

func (h *CashbookConnectHandler) AddTransaction(ctx context.Context, req *connect.Request[devicepb.AddTransactionRequest]) (*connect.Response[devicepb.AddTransactionResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("cashbook repository unavailable"))
	}
	if req.Msg.AccountId == "" || req.Msg.CategoryId == "" || req.Msg.Amount <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("account_id, category_id and amount > 0 are required"))
	}
	now := time.Now()
	dir := req.Msg.Direction
	if dir == "" {
		dir = domainCashbook.DirectionOut
	}
	t := domainCashbook.CashTransaction{
		ID:            idgen.New("trx"),
		TenantID:      "tenant-default",
		TransactionNo: fmt.Sprintf("TRX-%s-%06d", now.Format("200601"), now.UnixNano()%1000000),
		AccountID:     req.Msg.AccountId,
		CategoryID:    req.Msg.CategoryId,
		Direction:     dir,
		Amount:        req.Msg.Amount,
		TrxDate:       now,
		SourceType:    domainCashbook.SourceExpense,
		Description:   req.Msg.Description,
		CreatedAt:     now,
	}
	if err := h.repo.SaveTransaction(ctx, t); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.AddTransactionResponse{
		Transaction: toProtoTransaction(&t),
	}), nil
}

func (h *CashbookConnectHandler) ListTransactions(ctx context.Context, req *connect.Request[devicepb.ListTransactionsFilter]) (*connect.Response[devicepb.ListTransactionsResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("cashbook repository unavailable"))
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
	list, err := h.repo.FindTransactions(ctx, f)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListTransactionsResponse{
		Transactions: toProtoTransactionList(list),
	}), nil
}

func (h *CashbookConnectHandler) Balances(ctx context.Context, req *connect.Request[devicepb.BalancesRequest]) (*connect.Response[devicepb.BalancesResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("cashbook repository unavailable"))
	}
	f := port.CashTransactionFilter{}
	if req.Msg.FromUnix > 0 {
		f.From = time.Unix(req.Msg.FromUnix, 0)
	}
	if req.Msg.ToUnix > 0 {
		f.To = time.Unix(req.Msg.ToUnix, 0)
	}
	balances, err := h.repo.BalanceByAccounts(ctx, f)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.BalancesResponse{
		BalanceByAccount: balances,
	}), nil
}
