package cashbook

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainCashbook "github.com/quixiq/polyglot/internal/domain/cashbook"
)

func toProtoAccount(a *domainCashbook.CashAccount) *devicepb.CashAccount {
	if a == nil {
		return nil
	}
	return &devicepb.CashAccount{
		Id: a.ID, AccountCode: a.AccountCode, Name: a.Name, Type: a.Type, IsActive: a.IsActive,
	}
}

func toProtoAccountList(list []domainCashbook.CashAccount) []*devicepb.CashAccount {
	out := make([]*devicepb.CashAccount, len(list))
	for i := range list {
		out[i] = toProtoAccount(&list[i])
	}
	return out
}

func toProtoCategory(c *domainCashbook.CashCategory) *devicepb.CashCategory {
	if c == nil {
		return nil
	}
	return &devicepb.CashCategory{
		Id: c.ID, Name: c.Name, Type: c.Type, IsActive: c.IsActive,
	}
}

func toProtoCategoryList(list []domainCashbook.CashCategory) []*devicepb.CashCategory {
	out := make([]*devicepb.CashCategory, len(list))
	for i := range list {
		out[i] = toProtoCategory(&list[i])
	}
	return out
}

func toProtoTransaction(t *domainCashbook.CashTransaction) *devicepb.CashTransaction {
	if t == nil {
		return nil
	}
	return &devicepb.CashTransaction{
		Id: t.ID, TransactionNo: t.TransactionNo, AccountId: t.AccountID,
		CategoryId: t.CategoryID, Direction: t.Direction, Amount: t.Amount,
		TrxDateUnix: t.TrxDate.Unix(), SourceType: t.SourceType,
		SourceId: t.SourceID, Description: t.Description,
	}
}

func toProtoTransactionList(list []domainCashbook.CashTransaction) []*devicepb.CashTransaction {
	out := make([]*devicepb.CashTransaction, len(list))
	for i := range list {
		out[i] = toProtoTransaction(&list[i])
	}
	return out
}
