package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
	knowledgeuc "github.com/quixiq/polyglot/internal/usecase/knowledge"
)

func (h *KnowledgeConnectHandler) ListKnowledge(ctx context.Context, req *connect.Request[devicepb.ListKnowledgeRequest]) (*connect.Response[devicepb.ListKnowledgeResponse], error) {
	entries, err := h.documents.ListDocuments(ctx)
	if err != nil {
		return nil, mapKnowledgeError(err)
	}

	category := strings.TrimSpace(req.Msg.GetCategory())
	query := strings.ToLower(strings.TrimSpace(req.Msg.GetSearchQuery()))
	items := make([]*devicepb.KnowledgeItem, 0, len(entries))
	for i := range entries {
		e := entries[i]
		if category != "" && !strings.EqualFold(e.Category, category) {
			continue
		}
		if query != "" && !entryMatchesQuery(&e, query) {
			continue
		}
		items = append(items, toKnowledgeItem(&e))
	}
	return connect.NewResponse(&devicepb.ListKnowledgeResponse{Items: items}), nil
}

func (h *KnowledgeConnectHandler) GetKnowledge(ctx context.Context, req *connect.Request[devicepb.GetKnowledgeRequest]) (*connect.Response[devicepb.GetKnowledgeResponse], error) {
	id, err := parseKnowledgeID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	entry, err := h.documents.GetDocument(ctx, id)
	if err != nil {
		return nil, mapKnowledgeError(err)
	}
	return connect.NewResponse(&devicepb.GetKnowledgeResponse{Item: toKnowledgeItem(entry)}), nil
}

func (h *KnowledgeConnectHandler) CreateKnowledge(ctx context.Context, req *connect.Request[devicepb.CreateKnowledgeRequest]) (*connect.Response[devicepb.CreateKnowledgeResponse], error) {
	entry, err := h.documents.CreateDocument(ctx, knowledgeuc.CreateParams{
		Title:      req.Msg.GetTitle(),
		Content:    req.Msg.GetContent(),
		Category:   req.Msg.GetCategory(),
		Tags:       req.Msg.GetTags(),
		EmbedToLLM: req.Msg.GetEmbedToLlm(),
	})
	if err != nil {
		// ErrEmbedSync = dokumen TERSIMPAN dengan status failed — kembalikan
		// item supaya UI menampilkan badge failed + tombol retry.
		if !errors.Is(err, knowledgeuc.ErrEmbedSync) {
			return nil, mapKnowledgeError(err)
		}
	}
	return connect.NewResponse(&devicepb.CreateKnowledgeResponse{Item: toKnowledgeItem(entry)}), nil
}

func (h *KnowledgeConnectHandler) UpdateKnowledge(ctx context.Context, req *connect.Request[devicepb.UpdateKnowledgeRequest]) (*connect.Response[devicepb.UpdateKnowledgeResponse], error) {
	id, err := parseKnowledgeID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	entry, err := h.documents.UpdateDocument(ctx, knowledgeuc.UpdateParams{
		ID:         id,
		Title:      req.Msg.GetTitle(),
		Content:    req.Msg.GetContent(),
		Category:   req.Msg.GetCategory(),
		Tags:       req.Msg.GetTags(),
		EmbedToLLM: req.Msg.GetEmbedToLlm(),
	})
	if err != nil {
		if !errors.Is(err, knowledgeuc.ErrEmbedSync) {
			return nil, mapKnowledgeError(err)
		}
	}
	return connect.NewResponse(&devicepb.UpdateKnowledgeResponse{Item: toKnowledgeItem(entry)}), nil
}

func (h *KnowledgeConnectHandler) DeleteKnowledge(ctx context.Context, req *connect.Request[devicepb.DeleteKnowledgeRequest]) (*connect.Response[devicepb.DeleteKnowledgeResponse], error) {
	id, err := parseKnowledgeID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	err = h.documents.DeleteDocument(ctx, id)
	if err != nil {
		if errors.Is(err, knowledgeuc.ErrEmbedSync) {
			// Entry sudah terhapus dari Postgres; sisa di AnythingLLM jadi
			// orphan — operasi tetap sukses, dicatat di message.
			return connect.NewResponse(&devicepb.DeleteKnowledgeResponse{
				Message: "knowledge item deleted; note: document may still exist in AnythingLLM",
			}), nil
		}
		return nil, mapKnowledgeError(err)
	}
	return connect.NewResponse(&devicepb.DeleteKnowledgeResponse{
		Message: "knowledge item deleted successfully",
	}), nil
}

func (h *KnowledgeConnectHandler) RetryEmbed(ctx context.Context, req *connect.Request[devicepb.RetryEmbedRequest]) (*connect.Response[devicepb.RetryEmbedResponse], error) {
	id, err := parseKnowledgeID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	entry, err := h.documents.RetryEmbed(ctx, id)
	if err != nil {
		if !errors.Is(err, knowledgeuc.ErrEmbedSync) {
			return nil, mapKnowledgeError(err)
		}
	}
	return connect.NewResponse(&devicepb.RetryEmbedResponse{Item: toKnowledgeItem(entry)}), nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────

func parseKnowledgeID(raw string) (uint, error) {
	id, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid knowledge id %q", raw)
	}
	return uint(id), nil
}

func toKnowledgeItem(e *knowledge.KnowledgeEntry) *devicepb.KnowledgeItem {
	if e == nil {
		return &devicepb.KnowledgeItem{}
	}
	return &devicepb.KnowledgeItem{
		Id:                 strconv.FormatUint(uint64(e.ID), 10),
		Title:              e.Title,
		Content:            e.Content,
		Category:           e.Category,
		Tags:               splitTags(e.Tags),
		CreatedAt:          e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          e.UpdatedAt.Format(time.RFC3339),
		EmbedToLlm:         e.EmbedToLLM,
		EmbedStatus:        e.EmbedStatus,
		AnythingllmDocName: e.AnythingLLMDocName,
	}
}

// splitTags memecah tags yang disimpan comma-separated di Postgres menjadi
// slice untuk proto, membuang elemen kosong.
func splitTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// entryMatchesQuery mencocokkan query (sudah lowercase) terhadap title,
// content, category, dan tags entry.
func entryMatchesQuery(e *knowledge.KnowledgeEntry, query string) bool {
	haystacks := []string{
		strings.ToLower(e.Title),
		strings.ToLower(e.Content),
		strings.ToLower(e.Category),
		strings.ToLower(e.Tags),
	}
	for _, hay := range haystacks {
		if strings.Contains(hay, query) {
			return true
		}
	}
	return false
}

// mapKnowledgeError memetakan error usecase/domain ke connect error code
// yang bisa dipakai frontend untuk menampilkan pesan yang tepat.
func mapKnowledgeError(err error) error {
	switch {
	case errors.Is(err, knowledgeuc.ErrInvalidTitle),
		errors.Is(err, knowledgeuc.ErrEmptyContent):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, knowledgeuc.ErrEmbedNotConfigured):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, knowledge.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
