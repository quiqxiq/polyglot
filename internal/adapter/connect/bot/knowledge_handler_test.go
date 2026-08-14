package bot

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
	knowledgeuc "github.com/quixiq/polyglot/internal/usecase/knowledge"
	"github.com/stretchr/testify/require"
)

// ─── Fakes (ringkas, single-threaded) ─────────────────────────────────────

type fakeKnowledgeRepo struct {
	entries map[uint]*knowledge.KnowledgeEntry
	nextID  uint
}

func newFakeKnowledgeRepo() *fakeKnowledgeRepo {
	return &fakeKnowledgeRepo{entries: map[uint]*knowledge.KnowledgeEntry{}, nextID: 1}
}

func (f *fakeKnowledgeRepo) Create(e *knowledge.KnowledgeEntry) error {
	e.ID = f.nextID
	f.nextID++
	cp := *e
	f.entries[e.ID] = &cp
	return nil
}

func (f *fakeKnowledgeRepo) FindByID(id uint) (*knowledge.KnowledgeEntry, error) {
	e, ok := f.entries[id]
	if !ok {
		return nil, knowledge.ErrNotFound
	}
	cp := *e
	return &cp, nil
}

func (f *fakeKnowledgeRepo) FindAll() ([]knowledge.KnowledgeEntry, error) {
	res := make([]knowledge.KnowledgeEntry, 0, len(f.entries))
	for _, e := range f.entries {
		res = append(res, *e)
	}
	return res, nil
}

func (f *fakeKnowledgeRepo) Update(e *knowledge.KnowledgeEntry) error {
	if _, ok := f.entries[e.ID]; !ok {
		return knowledge.ErrNotFound
	}
	cp := *e
	f.entries[e.ID] = &cp
	return nil
}

func (f *fakeKnowledgeRepo) Delete(id uint) error {
	if _, ok := f.entries[id]; !ok {
		return knowledge.ErrNotFound
	}
	delete(f.entries, id)
	return nil
}

func (f *fakeKnowledgeRepo) SearchByTags([]string) ([]knowledge.KnowledgeEntry, error) {
	return f.FindAll()
}

type fakeKnowledgeManager struct {
	upsertErr error
	deleteErr error
}

func (f *fakeKnowledgeManager) UpsertDocument(_ context.Context, _, _, _ string) (string, error) {
	if f.upsertErr != nil {
		return "", f.upsertErr
	}
	return "custom-documents/raw-doc-1.json", nil
}

func (f *fakeKnowledgeManager) DeleteDocument(_ context.Context, _ string) error {
	return f.deleteErr
}

// ─── Helpers ──────────────────────────────────────────────────────────────

func newTestKnowledgeHandler(repo *fakeKnowledgeRepo, mgr *fakeKnowledgeManager) *KnowledgeConnectHandler {
	return NewKnowledgeConnectHandler(knowledgeuc.NewDocumentManager(repo, mgr))
}

func createViaHandler(t *testing.T, h *KnowledgeConnectHandler, title string) *devicepb.KnowledgeItem {
	t.Helper()
	resp, err := h.CreateKnowledge(context.Background(), connect.NewRequest(&devicepb.CreateKnowledgeRequest{
		Title: title, Content: "isi", Category: "layanan", Tags: []string{"a", "b"},
	}))
	require.NoError(t, err)
	return resp.Msg.Item
}

// ─── Tests ────────────────────────────────────────────────────────────────

func TestKnowledgeHandlerList(t *testing.T) {
	repo := newFakeKnowledgeRepo()
	h := newTestKnowledgeHandler(repo, &fakeKnowledgeManager{})
	_ = createViaHandler(t, h, "Prosedur Reset")
	_ = createViaHandler(t, h, "Harga Paket")

	resp, err := h.ListKnowledge(context.Background(), connect.NewRequest(&devicepb.ListKnowledgeRequest{}))
	require.NoError(t, err)
	require.Len(t, resp.Msg.Items, 2)

	item := resp.Msg.Items[0]
	require.NotEmpty(t, item.Id)
	require.Equal(t, "Prosedur Reset", item.Title)
	require.Equal(t, "layanan", item.Category)
	require.Equal(t, []string{"a", "b"}, item.Tags)
	require.Equal(t, knowledge.EmbedStatusNone, item.EmbedStatus)
	require.False(t, item.EmbedToLlm)
	require.NotEmpty(t, item.CreatedAt, "timestamp harus terformat RFC3339")
}

func TestKnowledgeHandlerListFilters(t *testing.T) {
	repo := newFakeKnowledgeRepo()
	h := newTestKnowledgeHandler(repo, &fakeKnowledgeManager{})
	_ = createViaHandler(t, h, "Reset Router")
	_ = createViaHandler(t, h, "Harga Paket")

	byCategory, err := h.ListKnowledge(context.Background(), connect.NewRequest(&devicepb.ListKnowledgeRequest{Category: "layanan"}))
	require.NoError(t, err)
	require.Len(t, byCategory.Msg.Items, 2)

	byQuery, err := h.ListKnowledge(context.Background(), connect.NewRequest(&devicepb.ListKnowledgeRequest{SearchQuery: "harga"}))
	require.NoError(t, err)
	require.Len(t, byQuery.Msg.Items, 1)
	require.Equal(t, "Harga Paket", byQuery.Msg.Items[0].Title)
}

func TestKnowledgeHandlerGet(t *testing.T) {
	repo := newFakeKnowledgeRepo()
	h := newTestKnowledgeHandler(repo, &fakeKnowledgeManager{})
	item := createViaHandler(t, h, "Prosedur Reset")

	got, err := h.GetKnowledge(context.Background(), connect.NewRequest(&devicepb.GetKnowledgeRequest{Id: item.Id}))
	require.NoError(t, err)
	require.Equal(t, "Prosedur Reset", got.Msg.Item.Title)

	_, err = h.GetKnowledge(context.Background(), connect.NewRequest(&devicepb.GetKnowledgeRequest{Id: "999"}))
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))

	_, err = h.GetKnowledge(context.Background(), connect.NewRequest(&devicepb.GetKnowledgeRequest{Id: "abc"}))
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestKnowledgeHandlerCreateValidation(t *testing.T) {
	repo := newFakeKnowledgeRepo()
	h := newTestKnowledgeHandler(repo, &fakeKnowledgeManager{})

	_, err := h.CreateKnowledge(context.Background(), connect.NewRequest(&devicepb.CreateKnowledgeRequest{Title: "  "}))
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	require.Empty(t, repo.entries)
}

func TestKnowledgeHandlerCreateEmbedSync(t *testing.T) {
	repo := newFakeKnowledgeRepo()
	h := newTestKnowledgeHandler(repo, &fakeKnowledgeManager{})

	resp, err := h.CreateKnowledge(context.Background(), connect.NewRequest(&devicepb.CreateKnowledgeRequest{
		Title: "Harga Paket", Content: "20 Mbps Rp250.000", EmbedToLlm: true,
	}))
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusEmbedded, resp.Msg.Item.EmbedStatus)
	require.True(t, resp.Msg.Item.EmbedToLlm)
}

func TestKnowledgeHandlerCreateEmbedFailureStillReturnsItem(t *testing.T) {
	repo := newFakeKnowledgeRepo()
	h := newTestKnowledgeHandler(repo, &fakeKnowledgeManager{upsertErr: errors.New("collector offline")})

	resp, err := h.CreateKnowledge(context.Background(), connect.NewRequest(&devicepb.CreateKnowledgeRequest{
		Title: "Harga Paket", Content: "isi", EmbedToLlm: true,
	}))
	require.NoError(t, err, "ErrEmbedSync harus sukses dengan item berstatus failed")
	require.Equal(t, knowledge.EmbedStatusFailed, resp.Msg.Item.EmbedStatus)
	require.NotEmpty(t, resp.Msg.Item.Id, "dokumen tetap tersimpan walau embed gagal")
}

func TestKnowledgeHandlerCreateNoManager(t *testing.T) {
	repo := newFakeKnowledgeRepo()
	h := NewKnowledgeConnectHandler(knowledgeuc.NewDocumentManager(repo, nil))

	_, err := h.CreateKnowledge(context.Background(), connect.NewRequest(&devicepb.CreateKnowledgeRequest{
		Title: "Judul", Content: "isi", EmbedToLlm: true,
	}))
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Empty(t, repo.entries, "ditolak sebelum save kalau AnythingLLM belum dikonfigurasi")
}

func TestKnowledgeHandlerUpdate(t *testing.T) {
	repo := newFakeKnowledgeRepo()
	h := newTestKnowledgeHandler(repo, &fakeKnowledgeManager{})
	item := createViaHandler(t, h, "Judul Lama")

	resp, err := h.UpdateKnowledge(context.Background(), connect.NewRequest(&devicepb.UpdateKnowledgeRequest{
		Id: item.Id, Title: "Judul Baru", Content: "isi baru",
	}))
	require.NoError(t, err)
	require.Equal(t, "Judul Baru", resp.Msg.Item.Title)

	_, err = h.UpdateKnowledge(context.Background(), connect.NewRequest(&devicepb.UpdateKnowledgeRequest{
		Id: "999", Title: "X", Content: "y",
	}))
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestKnowledgeHandlerDelete(t *testing.T) {
	repo := newFakeKnowledgeRepo()
	h := newTestKnowledgeHandler(repo, &fakeKnowledgeManager{})
	item := createViaHandler(t, h, "Judul")

	resp, err := h.DeleteKnowledge(context.Background(), connect.NewRequest(&devicepb.DeleteKnowledgeRequest{Id: item.Id}))
	require.NoError(t, err)
	require.Contains(t, resp.Msg.Message, "deleted")
	require.Empty(t, repo.entries)

	_, err = h.DeleteKnowledge(context.Background(), connect.NewRequest(&devicepb.DeleteKnowledgeRequest{Id: "999"}))
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestKnowledgeHandlerDeleteOrphanNote(t *testing.T) {
	repo := newFakeKnowledgeRepo()
	mgr := &fakeKnowledgeManager{}
	h := newTestKnowledgeHandler(repo, mgr)
	item := createViaHandler(t, h, "Judul")
	// set embed supaya ada doc name
	_, err := h.UpdateKnowledge(context.Background(), connect.NewRequest(&devicepb.UpdateKnowledgeRequest{
		Id: item.Id, Title: "Judul", Content: "isi", EmbedToLlm: true,
	}))
	require.NoError(t, err)

	mgr.deleteErr = errors.New("server error")
	resp, err := h.DeleteKnowledge(context.Background(), connect.NewRequest(&devicepb.DeleteKnowledgeRequest{Id: item.Id}))
	require.NoError(t, err, "delete tetap sukses; orphan dicatat di message")
	require.Contains(t, resp.Msg.Message, "AnythingLLM")
	require.Empty(t, repo.entries, "entry tetap terhapus dari Postgres")
}

func TestKnowledgeHandlerRetryEmbed(t *testing.T) {
	repo := newFakeKnowledgeRepo()
	mgr := &fakeKnowledgeManager{upsertErr: errors.New("collector offline")}
	h := newTestKnowledgeHandler(repo, mgr)
	resp, err := h.CreateKnowledge(context.Background(), connect.NewRequest(&devicepb.CreateKnowledgeRequest{
		Title: "Judul", Content: "isi", EmbedToLlm: true,
	}))
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusFailed, resp.Msg.Item.EmbedStatus)

	mgr.upsertErr = nil
	retry, err := h.RetryEmbed(context.Background(), connect.NewRequest(&devicepb.RetryEmbedRequest{Id: resp.Msg.Item.Id}))
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusEmbedded, retry.Msg.Item.EmbedStatus)

	_, err = h.RetryEmbed(context.Background(), connect.NewRequest(&devicepb.RetryEmbedRequest{Id: "999"}))
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}
