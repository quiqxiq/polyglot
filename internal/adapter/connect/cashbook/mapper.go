package cashbook

import (
	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	domainCashbook "github.com/quixiq/polyglot/internal/domain/cashbook"
)

func toProtoAccount(a *domainCashbook.CashAccount) *polyglotv1.CashAccount {
	if a == nil {
		return nil
	}
	return &polyglotv1.CashAccount{
		Id: a.ID, AccountCode: a.AccountCode, Name: a.Name, Type: a.Type, IsActive: a.IsActive,
	}
}

func toProtoAccountList(list []domainCashbook.CashAccount) []*polyglotv1.CashAccount {
	out := make([]*polyglotv1.CashAccount, len(list))
	for i := range list {
		out[i] = toProtoAccount(&list[i])
	}
	return out
}

func toProtoCategory(c *domainCashbook.CashCategory) *polyglotv1.CashCategory {
	if c == nil {
		return nil
	}
	return &polyglotv1.CashCategory{
		Id: c.ID, Name: c.Name, Type: c.Type, IsActive: c.IsActive,
	}
}

func toProtoCategoryList(list []domainCashbook.CashCategory) []*polyglotv1.CashCategory {
	out := make([]*polyglotv1.CashCategory, len(list))
	for i := range list {
		out[i] = toProtoCategory(&list[i])
	}
	return out
}

func toProtoTransaction(t *domainCashbook.CashTransaction) *polyglotv1.CashTransaction {
	if t == nil {
		return nil
	}
	return &polyglotv1.CashTransaction{
		Id: t.ID, TransactionNo: t.TransactionNo, AccountId: t.AccountID,
		CategoryId: t.CategoryID, Direction: t.Direction, Amount: t.Amount,
		TrxDateUnix: t.TrxDate.Unix(), SourceType: t.SourceType,
		SourceId: t.SourceID, Description: t.Description,
	}
}

func toProtoTransactionList(list []domainCashbook.CashTransaction) []*polyglotv1.CashTransaction {
	out := make([]*polyglotv1.CashTransaction, len(list))
	for i := range list {
		out[i] = toProtoTransaction(&list[i])
	}
	return out
}
