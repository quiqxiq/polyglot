package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/subscription"
)

type customerLookupArgs struct {
	Phone      string `json:"phone,omitempty" jsonschema:"WhatsApp or phone number of the customer"`
	Name       string `json:"name,omitempty" jsonschema:"customer name (full or partial)"`
	CustomerID string `json:"customer_id,omitempty" jsonschema:"exact customer ID"`
}

type customerDetail struct {
	ID            string                      `json:"id"`
	Name          string                      `json:"name"`
	Email         string                      `json:"email"`
	Phone         string                      `json:"phone"`
	Address       string                      `json:"address"`
	Status        string                      `json:"status"`
	Subscriptions []subscription.Subscription `json:"subscriptions,omitempty"`
}

type customerLookupOutput struct {
	Status    string           `json:"status"`
	Summary   string           `json:"summary"`
	Customers []customerDetail `json:"customers,omitempty"`
}

func (s *Server) customerLookup(ctx context.Context, _ *mcp.CallToolRequest, args customerLookupArgs) (*mcp.CallToolResult, customerLookupOutput, error) {
	if args.Phone == "" && args.Name == "" && args.CustomerID == "" {
		return toolError(customerLookupOutput{Status: "error", Summary: "at least one search criterion (phone, name, or customer_id) is required"})
	}

	if s.customerRepo == nil {
		return toolOK(customerLookupOutput{
			Status:  "success",
			Summary: "Customer repository not configured in server mode",
		})
	}

	var matched []customer.Customer
	if args.CustomerID != "" {
		c, err := s.customerRepo.FindByID(ctx, args.CustomerID)
		if err == nil {
			matched = append(matched, c)
		}
	} else {
		all, err := s.customerRepo.FindAll(ctx)
		if err != nil {
			return toolError(customerLookupOutput{Status: "error", Summary: err.Error()})
		}
		for _, c := range all {
			matchPhone := args.Phone != "" && strings.Contains(c.Phone, args.Phone)
			matchName := args.Name != "" && strings.Contains(strings.ToLower(c.Name), strings.ToLower(args.Name))
			if matchPhone || matchName {
				matched = append(matched, c)
			}
		}
	}

	if len(matched) == 0 {
		return toolOK(customerLookupOutput{
			Status:  "success",
			Summary: "No customers matched the search criteria",
		})
	}

	details := make([]customerDetail, len(matched))
	for i, c := range matched {
		subs, _ := s.customerRepo.FindSubscriptions(ctx, c.ID)
		details[i] = customerDetail{
			ID:            c.ID,
			Name:          c.Name,
			Email:         c.Email,
			Phone:         c.Phone,
			Address:       c.Address,
			Status:        c.Status,
			Subscriptions: subs,
		}
	}

	summary := fmt.Sprintf("Found %d customer(s) matching query", len(details))
	return toolOK(customerLookupOutput{
		Status:    "success",
		Summary:   summary,
		Customers: details,
	})
}
