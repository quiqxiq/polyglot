package knowledge

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/knowledge"
	"github.com/stretchr/testify/require"
)

// ─── Fakes ────────────────────────────────────────────────────────────────

// fakeRepo meniru semantics repo Postgres (Create mengisi ID, Update replace
// per-ID, FindByID mengembalikan copy) tanpa database nyata.
type fakeRepo struct {
	mu      sync.Mutex
	entries map[uint]*knowledge.KnowledgeEntry
	nextID  uint
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{entries: map[uint]*knowledge.KnowledgeEntry{}, nextID: 1}
}

func (f *fakeRepo) Create(e *knowledge.KnowledgeEntry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	e.ID = f.nextID
	f.nextID++
	cp := *e
	f.entries[e.ID] = &cp
	return nil
}

func (f *fakeRepo) FindByID(id uint) (*knowledge.KnowledgeEntry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.entries[id]
	if !ok {
		return nil, knowledge.ErrNotFound
	}
	cp := *e
	return &cp, nil
}

func (f *fakeRepo) FindAll() ([]knowledge.KnowledgeEntry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	res := make([]knowledge.KnowledgeEntry, 0, len(f.entries))
	for _, e := range f.entries {
		res = append(res, *e)
	}
	return res, nil
}

func (f *fakeRepo) Update(e *knowledge.KnowledgeEntry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.entries[e.ID]; !ok {
		return knowledge.ErrNotFound
	}
	cp := *e
	f.entries[e.ID] = &cp
	return nil
}

func (f *fakeRepo) Delete(id uint) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.entries[id]; !ok {
		return knowledge.ErrNotFound
	}
	delete(f.entries, id)
	return nil
}

func (f *fakeRepo) SearchByTags([]string) ([]knowledge.KnowledgeEntry, error) {
	return f.FindAll()
}

type upsertCall struct {
	docName  string
	title    string
	markdown string
}

// fakeManager meniru port.KnowledgeDocumentManager; error bisa diprogram per
// test lewat field upsertErr/deleteErr.
type fakeManager struct {
	upsertCalls []upsertCall
	deleteCalls []string
	upsertErr   error
	deleteErr   error
	callCount   int
}

func (f *fakeManager) UpsertDocument(_ context.Context, docName, title, markdown string) (string, error) {
	f.upsertCalls = append(f.upsertCalls, upsertCall{docName: docName, title: title, markdown: markdown})
	if f.upsertErr != nil {
		return "", f.upsertErr
	}
	f.callCount++
	return fmt.Sprintf("custom-documents/raw-doc-%d.json", f.callCount), nil
}

func (f *fakeManager) DeleteDocument(_ context.Context, docName string) error {
	f.deleteCalls = append(f.deleteCalls, docName)
	return f.deleteErr
}

// ─── Helpers ──────────────────────────────────────────────────────────────

func mustCreate(t *testing.T, m *DocumentManager, p CreateParams) *knowledge.KnowledgeEntry {
	t.Helper()
	e, err := m.CreateDocument(context.Background(), p)
	require.NoError(t, err)
	return e
}

// ─── CreateDocument ───────────────────────────────────────────────────────

func TestCreateDocumentValidation(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{}
	m := NewDocumentManager(repo, mgr)

	_, err := m.CreateDocument(context.Background(), CreateParams{Title: "  ", Content: "isi"})
	require.ErrorIs(t, err, ErrInvalidTitle)

	_, err = m.CreateDocument(context.Background(), CreateParams{Title: "Judul", Content: "", EmbedToLLM: true})
	require.ErrorIs(t, err, ErrEmptyContent)

	// embed tanpa manager (AnythingLLM tidak dikonfigurasi) → ditolak sebelum save
	mNoManager := NewDocumentManager(repo, nil)
	_, err = mNoManager.CreateDocument(context.Background(), CreateParams{Title: "Judul", Content: "isi", EmbedToLLM: true})
	require.ErrorIs(t, err, ErrEmbedNotConfigured)
	require.Empty(t, repo.entries, "entry tidak boleh tersimpan kalau embed ditolak")
}

func TestCreateDocumentLocalOnly(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{}
	m := NewDocumentManager(repo, mgr)

	entry, err := m.CreateDocument(context.Background(), CreateParams{
		Title: "Prosedur Reset", Content: "Langkah 1..2", Tags: []string{"mikrotik", "reset"},
	})
	require.NoError(t, err)
	require.Equal(t, uint(1), entry.ID)
	require.False(t, entry.EmbedToLLM)
	require.Equal(t, knowledge.EmbedStatusNone, entry.EmbedStatus)
	require.Equal(t, DefaultCategory, entry.Category)
	require.Empty(t, mgr.upsertCalls, "dokumen lokal tidak boleh menyentuh AnythingLLM")
	require.Empty(t, mgr.deleteCalls)
}

func TestCreateDocumentEmbed(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{}
	m := NewDocumentManager(repo, mgr)

	entry, err := m.CreateDocument(context.Background(), CreateParams{
		Title: "Harga Paket", Content: "20 Mbps Rp250.000", Category: "layanan", EmbedToLLM: true,
	})
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusEmbedded, entry.EmbedStatus)
	require.Equal(t, "custom-documents/raw-doc-1.json", entry.AnythingLLMDocName)
	require.True(t, entry.EmbedToLLM)

	require.Len(t, mgr.upsertCalls, 1)
	call := mgr.upsertCalls[0]
	require.Equal(t, "", call.docName, "create tidak punya doc lama")
	require.Equal(t, "Harga Paket", call.title)
	require.Equal(t, "20 Mbps Rp250.000", call.markdown)

	// state ter-persist di repo, bukan cuma di memori
	stored, err := repo.FindByID(entry.ID)
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusEmbedded, stored.EmbedStatus)
}

func TestCreateDocumentEmbedFailureKeepsEntry(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{upsertErr: errors.New("collector offline")}
	m := NewDocumentManager(repo, mgr)

	entry, err := m.CreateDocument(context.Background(), CreateParams{
		Title: "Harga Paket", Content: "20 Mbps Rp250.000", EmbedToLLM: true,
	})
	require.ErrorIs(t, err, ErrEmbedSync)
	require.NotNil(t, entry, "entry tetap tersimpan walau embed gagal")
	require.Equal(t, knowledge.EmbedStatusFailed, entry.EmbedStatus)

	stored, err := repo.FindByID(entry.ID)
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusFailed, stored.EmbedStatus, "status failed harus tersimpan untuk tombol retry")
}

// ─── UpdateDocument ───────────────────────────────────────────────────────

func TestUpdateDocumentValidation(t *testing.T) {
	repo := newFakeRepo()
	m := NewDocumentManager(repo, &fakeManager{})

	_, err := m.UpdateDocument(context.Background(), UpdateParams{ID: 1, Title: ""})
	require.ErrorIs(t, err, ErrInvalidTitle)

	entry := mustCreate(t, m, CreateParams{Title: "Judul", Content: "isi"})
	_, err = m.UpdateDocument(context.Background(), UpdateParams{ID: entry.ID, Title: "Judul", EmbedToLLM: true, Content: ""})
	require.ErrorIs(t, err, ErrEmptyContent)
}

func TestUpdateDocumentNotFound(t *testing.T) {
	m := NewDocumentManager(newFakeRepo(), &fakeManager{})
	_, err := m.UpdateDocument(context.Background(), UpdateParams{ID: 999, Title: "Judul"})
	require.ErrorIs(t, err, knowledge.ErrNotFound)
}

func TestUpdateDocumentEmbedOnFromLocal(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{}
	m := NewDocumentManager(repo, mgr)
	entry := mustCreate(t, m, CreateParams{Title: "Judul", Content: "isi lama", Tags: []string{"a"}})
	require.Equal(t, knowledge.EmbedStatusNone, entry.EmbedStatus)

	updated, err := m.UpdateDocument(context.Background(), UpdateParams{
		ID: entry.ID, Title: "Judul", Content: "isi lama", Tags: []string{"a"}, EmbedToLLM: true,
	})
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusEmbedded, updated.EmbedStatus)
	require.Len(t, mgr.upsertCalls, 1)
	require.Equal(t, "", mgr.upsertCalls[0].docName, "dari lokal ke embed → tidak ada doc lama")
}

func TestUpdateDocumentReembedWhenContentChanged(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{}
	m := NewDocumentManager(repo, mgr)
	entry := mustCreate(t, m, CreateParams{Title: "Judul", Content: "isi lama", EmbedToLLM: true})
	require.Equal(t, "custom-documents/raw-doc-1.json", entry.AnythingLLMDocName)

	updated, err := m.UpdateDocument(context.Background(), UpdateParams{
		ID: entry.ID, Title: "Judul", Content: "isi BARU", EmbedToLLM: true,
	})
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusEmbedded, updated.EmbedStatus)
	require.Equal(t, "custom-documents/raw-doc-2.json", updated.AnythingLLMDocName)
	require.Len(t, mgr.upsertCalls, 2)
	require.Equal(t, "custom-documents/raw-doc-1.json", mgr.upsertCalls[1].docName, "re-embed harus delete doc lama dulu")
	require.Equal(t, "isi BARU", mgr.upsertCalls[1].markdown)
}

func TestUpdateDocumentSkipsReembedWhenUnchanged(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{}
	m := NewDocumentManager(repo, mgr)
	entry := mustCreate(t, m, CreateParams{Title: "Judul", Content: "isi", EmbedToLLM: true})

	// Hanya kategori/tags yang berubah — title & content sama → tidak re-embed.
	updated, err := m.UpdateDocument(context.Background(), UpdateParams{
		ID: entry.ID, Title: "Judul", Content: "isi", Category: "layanan", Tags: []string{"x"}, EmbedToLLM: true,
	})
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusEmbedded, updated.EmbedStatus)
	require.Equal(t, "custom-documents/raw-doc-1.json", updated.AnythingLLMDocName)
	require.Len(t, mgr.upsertCalls, 1, "tanpa perubahan content/title tidak boleh re-embed")
}

func TestUpdateDocumentToggleEmbedOff(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{}
	m := NewDocumentManager(repo, mgr)
	entry := mustCreate(t, m, CreateParams{Title: "Judul", Content: "isi", EmbedToLLM: true})

	updated, err := m.UpdateDocument(context.Background(), UpdateParams{
		ID: entry.ID, Title: "Judul", Content: "isi", EmbedToLLM: false,
	})
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusNone, updated.EmbedStatus)
	require.Equal(t, "", updated.AnythingLLMDocName)
	require.False(t, updated.EmbedToLLM)
	require.Equal(t, []string{"custom-documents/raw-doc-1.json"}, mgr.deleteCalls)
}

func TestUpdateDocumentToggleOffFailureKeepsDocName(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{deleteErr: errors.New("server error")}
	m := NewDocumentManager(repo, mgr)
	entry := mustCreate(t, m, CreateParams{Title: "Judul", Content: "isi", EmbedToLLM: true})

	updated, err := m.UpdateDocument(context.Background(), UpdateParams{
		ID: entry.ID, Title: "Judul", Content: "isi", EmbedToLLM: false,
	})
	require.ErrorIs(t, err, ErrEmbedSync)
	require.NotNil(t, updated)
	require.Equal(t, knowledge.EmbedStatusFailed, updated.EmbedStatus)
	require.Equal(t, "custom-documents/raw-doc-1.json", updated.AnythingLLMDocName, "doc name dipertahankan supaya RetryEmbed bisa membersihkan")
}

// ─── DeleteDocument ───────────────────────────────────────────────────────

func TestDeleteDocumentNotFound(t *testing.T) {
	m := NewDocumentManager(newFakeRepo(), &fakeManager{})
	err := m.DeleteDocument(context.Background(), 999)
	require.ErrorIs(t, err, knowledge.ErrNotFound)
}

func TestDeleteDocumentLocal(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{}
	m := NewDocumentManager(repo, mgr)
	entry := mustCreate(t, m, CreateParams{Title: "Judul", Content: "isi"})

	require.NoError(t, m.DeleteDocument(context.Background(), entry.ID))
	_, err := repo.FindByID(entry.ID)
	require.ErrorIs(t, err, knowledge.ErrNotFound, "entry harus terhapus dari repo")
	require.Empty(t, mgr.deleteCalls, "dokumen lokal tidak punya doc AnythingLLM")
}

func TestDeleteDocumentEmbedded(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{}
	m := NewDocumentManager(repo, mgr)
	entry := mustCreate(t, m, CreateParams{Title: "Judul", Content: "isi", EmbedToLLM: true})

	require.NoError(t, m.DeleteDocument(context.Background(), entry.ID))
	require.Equal(t, []string{"custom-documents/raw-doc-1.json"}, mgr.deleteCalls)
	_, err := repo.FindByID(entry.ID)
	require.ErrorIs(t, err, knowledge.ErrNotFound)
}

func TestDeleteDocumentEmbeddedDeleteFailsIsOrphan(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{deleteErr: errors.New("server error")}
	m := NewDocumentManager(repo, mgr)
	entry := mustCreate(t, m, CreateParams{Title: "Judul", Content: "isi", EmbedToLLM: true})

	err := m.DeleteDocument(context.Background(), entry.ID)
	require.ErrorIs(t, err, ErrEmbedSync, "entry terhapus tapi sisa di AnythingLLM — kabarkan sebagai orphan")
	_, repoErr := repo.FindByID(entry.ID)
	require.ErrorIs(t, repoErr, knowledge.ErrNotFound, "entry tetap terhapus dari Postgres")
}

// ─── RetryEmbed ───────────────────────────────────────────────────────────

func TestRetryEmbedNotFound(t *testing.T) {
	m := NewDocumentManager(newFakeRepo(), &fakeManager{})
	_, err := m.RetryEmbed(context.Background(), 999)
	require.ErrorIs(t, err, knowledge.ErrNotFound)
}

func TestRetryEmbedNoManager(t *testing.T) {
	m := NewDocumentManager(newFakeRepo(), nil)
	_, err := m.RetryEmbed(context.Background(), 1)
	require.ErrorIs(t, err, ErrEmbedNotConfigured)
}

func TestRetryEmbedFailedEmbed(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{upsertErr: errors.New("collector offline")}
	m := NewDocumentManager(repo, mgr)
	entry, err := m.CreateDocument(context.Background(), CreateParams{Title: "Judul", Content: "isi", EmbedToLLM: true})
	require.ErrorIs(t, err, ErrEmbedSync, "create dengan embed gagal harus melaporkan ErrEmbedSync")
	require.NotNil(t, entry)
	require.Equal(t, knowledge.EmbedStatusFailed, entry.EmbedStatus)

	// LLM sudah pulih → retry berhasil
	mgr.upsertErr = nil
	updated, err := m.RetryEmbed(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusEmbedded, updated.EmbedStatus)
	require.Equal(t, "custom-documents/raw-doc-1.json", updated.AnythingLLMDocName)
	require.Len(t, mgr.upsertCalls, 2, "retry = satu upload lagi")
}

func TestRetryEmbedUnembedLeftover(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{}
	m := NewDocumentManager(repo, mgr)
	entry := mustCreate(t, m, CreateParams{Title: "Judul", Content: "isi", EmbedToLLM: true})

	// toggle off gagal → embed=false tapi doc name masih ada
	mgr.deleteErr = errors.New("server error")
	_, err := m.UpdateDocument(context.Background(), UpdateParams{ID: entry.ID, Title: "Judul", Content: "isi", EmbedToLLM: false})
	require.ErrorIs(t, err, ErrEmbedSync)

	// retry → delete sisa berhasil → none + doc name dibersihkan
	mgr.deleteErr = nil
	updated, err := m.RetryEmbed(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusNone, updated.EmbedStatus)
	require.Equal(t, "", updated.AnythingLLMDocName)
	require.Len(t, mgr.deleteCalls, 2)
}

func TestRetryEmbedNoOp(t *testing.T) {
	repo := newFakeRepo()
	mgr := &fakeManager{}
	m := NewDocumentManager(repo, mgr)
	entry := mustCreate(t, m, CreateParams{Title: "Judul", Content: "isi"})

	updated, err := m.RetryEmbed(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, knowledge.EmbedStatusNone, updated.EmbedStatus)
	require.Empty(t, mgr.upsertCalls)
	require.Empty(t, mgr.deleteCalls)
}

// ─── List / Get ───────────────────────────────────────────────────────────

func TestListAndGetDocument(t *testing.T) {
	repo := newFakeRepo()
	m := NewDocumentManager(repo, &fakeManager{})
	entry := mustCreate(t, m, CreateParams{Title: "Satu", Content: "isi 1"})
	_ = mustCreate(t, m, CreateParams{Title: "Dua", Content: "isi 2"})

	all, err := m.ListDocuments(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 2)

	got, err := m.GetDocument(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, "Satu", got.Title)

	_, err = m.GetDocument(context.Background(), 999)
	require.ErrorIs(t, err, knowledge.ErrNotFound)
}
