package notification

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
)

func toProtoTemplate(t *domainNotification.NotificationTemplate) *devicepb.NotificationTemplate {
	if t == nil {
		return nil
	}
	return &devicepb.NotificationTemplate{
		Id: t.ID, TemplateKey: t.TemplateKey, Name: t.Name,
		Content: t.Content, VariablesJson: t.VariablesJSON, IsActive: t.IsActive,
	}
}

func toProtoTemplateList(list []domainNotification.NotificationTemplate) []*devicepb.NotificationTemplate {
	out := make([]*devicepb.NotificationTemplate, len(list))
	for i := range list {
		out[i] = toProtoTemplate(&list[i])
	}
	return out
}

func toProtoWANotification(n *domainNotification.WANotification) *devicepb.WANotification {
	if n == nil {
		return nil
	}
	var tplID, custID, invID string
	if n.TemplateID != nil {
		tplID = *n.TemplateID
	}
	if n.CustomerID != nil {
		custID = *n.CustomerID
	}
	if n.InvoiceID != nil {
		invID = *n.InvoiceID
	}
	var sentAtUnix int64
	if n.SentAt != nil {
		sentAtUnix = n.SentAt.Unix()
	}
	return &devicepb.WANotification{
		Id: n.ID, TemplateId: tplID, CustomerId: custID, InvoiceId: invID,
		RecipientPhone: n.RecipientPhone, MessageType: n.MessageType,
		MessageContent: n.MessageContent, Status: n.Status,
		ErrorMessage: n.ErrorMessage, Attempts: int32(n.Attempts),
		SentAtUnix: sentAtUnix, CreatedAtUnix: n.CreatedAt.Unix(),
	}
}

func toProtoWANotificationList(list []domainNotification.WANotification) []*devicepb.WANotification {
	out := make([]*devicepb.WANotification, len(list))
	for i := range list {
		out[i] = toProtoWANotification(&list[i])
	}
	return out
}
