package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type searchKnowledgeArgs struct {
	Query string `json:"query" jsonschema:"the search query or issue description (e.g. WiFi lemot, voucher tidak bisa login)"`
}

type knowledgeItem struct {
	ID        uint    `json:"id"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Tags      string  `json:"tags"`
	Relevance float64 `json:"relevance,omitempty"`
}

type searchKnowledgeOutput struct {
	Status  string          `json:"status"`
	Summary string          `json:"summary"`
	Entries []knowledgeItem `json:"entries,omitempty"`
}

func (s *Server) searchKnowledge(ctx context.Context, _ *mcp.CallToolRequest, args searchKnowledgeArgs) (*mcp.CallToolResult, searchKnowledgeOutput, error) {
	if args.Query == "" {
		return toolError(searchKnowledgeOutput{Status: "error", Summary: "search query is required"})
	}

	if s.knowledgeRetriever == nil {
		return toolOK(searchKnowledgeOutput{
			Status:  "success",
			Summary: "Knowledge retriever not configured",
		})
	}

	entries, err := s.knowledgeRetriever.Retrieve(ctx, args.Query)
	if err != nil {
		return toolError(searchKnowledgeOutput{Status: "error", Summary: err.Error()})
	}

	if len(entries) == 0 {
		return toolOK(searchKnowledgeOutput{
			Status:  "success",
			Summary: fmt.Sprintf("No knowledge base articles found for query %q", args.Query),
		})
	}

	items := make([]knowledgeItem, len(entries))
	for i, e := range entries {
		items[i] = knowledgeItem{
			ID:      e.ID,
			Title:   e.Title,
			Content: e.Content,
			Tags:    e.Tags,
		}
	}

	summary := fmt.Sprintf("Found %d relevant knowledge base article(s) for query %q", len(items), args.Query)
	return toolOK(searchKnowledgeOutput{
		Status:  "success",
		Summary: summary,
		Entries: items,
	})
}
