package knowledge

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/knowledge"
	"github.com/quixiq/polyglot/internal/port"
)

// DefaultCategory dipakai saat params.Category kosong — kategori opsional di
// form admin (keputusan fase 1), default 'umum' konsisten dengan migrasi
// 000008 (column default).
const DefaultCategory = "umum"

// DocumentManager mengorkestrasi manajemen dokumen knowledge: Postgres adalah
// satu-satunya sumber kebenaran, AnythingLLM hanya proyeksi untuk vector
// search (embed per-dokumen lewat flag EmbedToLLM). Alur sync:
//
//	create/update dengan embed  → pending → embedded | failed
//	toggle embed off           → delete dari AnythingLLM → none
//
// Semua kegagalan sync bersifat fail-open: dokumen tetap tersimpan di
// Postgres dengan EmbedStatusFailed dan bisa di-retry dari UI (RetryEmbed).
type DocumentManager struct {
	repo    port.KnowledgeRepository
	manager port.KnowledgeDocumentManager // nil = AnythingLLM tidak dikonfigurasi
}

// NewDocumentManager membangun usecase. manager boleh nil — embed hanya
// berfungsi kalau AnythingLLM dikonfigurasi; tanpa manager, dokumen tetap
// bisa dikelola sebagai knowledge lokal.
func NewDocumentManager(repo port.KnowledgeRepository, manager port.KnowledgeDocumentManager) *DocumentManager {
	return &DocumentManager{repo: repo, manager: manager}
}

// CreateParams adalah input untuk CreateDocument.
type CreateParams struct {
	Title      string
	Content    string
	Category   string
	Tags       []string
	EmbedToLLM bool
}

// UpdateParams adalah input untuk UpdateDocument.
type UpdateParams struct {
	ID         uint
	Title      string
	Content    string
	Category   string
	Tags       []string
	EmbedToLLM bool
}

// CreateDocument menyimpan entry baru ke Postgres lalu, kalau EmbedToLLM,
// meng-embed isinya ke AnythingLLM. Error ErrEmbedSync berarti dokumen
// tersimpan tapi sync gagal (status failed).
func (m *DocumentManager) CreateDocument(ctx context.Context, p CreateParams) (*knowledge.Entry, error) {
	if strings.TrimSpace(p.Title) == "" {
		return nil, ErrInvalidTitle
	}
	if p.EmbedToLLM {
		if m.manager == nil {
			return nil, ErrEmbedNotConfigured
		}
		if strings.TrimSpace(p.Content) == "" {
			return nil, ErrEmptyContent
		}
	}

	entry := &knowledge.Entry{
		Title:       strings.TrimSpace(p.Title),
		Content:     p.Content,
		Category:    normalizeCategory(p.Category),
		Tags:        strings.Join(p.Tags, ","),
		EmbedToLLM:  p.EmbedToLLM,
		EmbedStatus: knowledge.EmbedStatusNone,
	}
	if p.EmbedToLLM {
		entry.EmbedStatus = knowledge.EmbedStatusPending
	}
	if err := m.repo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("knowledge: create entry: %w", err)
	}

	if p.EmbedToLLM {
		if err := m.syncEmbed(ctx, entry, ""); err != nil {
			return entry, err
		}
	}
	return entry, nil
}

// UpdateDocument memperbarui entry di Postgres lalu menyinkronkan status
// embed sesuai permintaan terbaru:
//
//	embed on  → embed ulang (delete + upload) kalau title/content berubah,
//	            atau kalau belum pernah embedded sebelumnya.
//	embed off → hapus dari AnythingLLM (toggle off).
//
// Error ErrEmbedSync berarti update tersimpan tapi sync gagal (status failed,
// bisa RetryEmbed).
func (m *DocumentManager) UpdateDocument(ctx context.Context, p UpdateParams) (*knowledge.Entry, error) {
	if strings.TrimSpace(p.Title) == "" {
		return nil, ErrInvalidTitle
	}
	if p.EmbedToLLM {
		if m.manager == nil {
			return nil, ErrEmbedNotConfigured
		}
		if strings.TrimSpace(p.Content) == "" {
			return nil, ErrEmptyContent
		}
	}

	entry, err := m.repo.FindByID(ctx, p.ID)
	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			return nil, knowledge.ErrNotFound
		}
		return nil, fmt.Errorf("knowledge: find entry %d: %w", p.ID, err)
	}

	wasEmbedded := entry.AnythingLLMDocName != ""
	contentChanged := entry.Title != strings.TrimSpace(p.Title) || entry.Content != p.Content

	entry.Title = strings.TrimSpace(p.Title)
	entry.Content = p.Content
	entry.Category = normalizeCategory(p.Category)
	entry.Tags = strings.Join(p.Tags, ",")
	entry.EmbedToLLM = p.EmbedToLLM

	switch {
	case p.EmbedToLLM:
		entry.EmbedStatus = knowledge.EmbedStatusPending
		if err := m.repo.Update(ctx, entry); err != nil {
			return nil, fmt.Errorf("knowledge: update entry: %w", err)
		}
		if !wasEmbedded || contentChanged {
			if err := m.syncEmbed(ctx, entry, entry.AnythingLLMDocName); err != nil {
				return entry, err
			}
		} else {
			entry.EmbedStatus = knowledge.EmbedStatusEmbedded
			if err := m.repo.Update(ctx, entry); err != nil {
				return nil, fmt.Errorf("knowledge: save embed state: %w", err)
			}
		}

	case wasEmbedded:
		// Toggle embed off: hapus dari AnythingLLM.
		entry.EmbedStatus = knowledge.EmbedStatusPending
		if err := m.repo.Update(ctx, entry); err != nil {
			return nil, fmt.Errorf("knowledge: update entry: %w", err)
		}
		if err := m.syncUnembed(ctx, entry); err != nil {
			return entry, err
		}

	default:
		entry.EmbedStatus = knowledge.EmbedStatusNone
		if err := m.repo.Update(ctx, entry); err != nil {
			return nil, fmt.Errorf("knowledge: update entry: %w", err)
		}
	}
	return entry, nil
}

// DeleteDocument menghapus entry dari Postgres, lalu best-effort menghapus
// dokumen terkait dari AnythingLLM. Kalau delete dari AnythingLLM gagal,
// entry sudah terhapus tapi sisa dokumen jadi orphan — dikabarkan lewat
// ErrEmbedSync supaya bisa dibersihkan manual.
func (m *DocumentManager) DeleteDocument(ctx context.Context, id uint) error {
	entry, err := m.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			return knowledge.ErrNotFound
		}
		return fmt.Errorf("knowledge: find entry %d: %w", id, err)
	}

	if err := m.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("knowledge: delete entry %d: %w", id, err)
	}

	if entry.AnythingLLMDocName != "" && m.manager != nil {
		if err := m.manager.DeleteDocument(ctx, entry.AnythingLLMDocName); err != nil {
			return fmt.Errorf("%w: %w", ErrEmbedSync, err)
		}
	}
	return nil
}

// RetryEmbed menjalankan ulang sinkronisasi embed untuk satu entry, sesuai
// state saat ini:
//
//	embed on            → embed ulang (delete + upload)
//	embed off + doc ada → hapus sisa dokumen dari AnythingLLM
//	embed off + tanpa doc → reset status none
//
// Ini tombol "Retry" di UI untuk entry berstatus failed.
func (m *DocumentManager) RetryEmbed(ctx context.Context, id uint) (*knowledge.Entry, error) {
	if m.manager == nil {
		return nil, ErrEmbedNotConfigured
	}
	entry, err := m.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			return nil, knowledge.ErrNotFound
		}
		return nil, fmt.Errorf("knowledge: find entry %d: %w", id, err)
	}

	switch {
	case entry.EmbedToLLM:
		entry.EmbedStatus = knowledge.EmbedStatusPending
		if err := m.repo.Update(ctx, entry); err != nil {
			return nil, fmt.Errorf("knowledge: update entry: %w", err)
		}
		if err := m.syncEmbed(ctx, entry, entry.AnythingLLMDocName); err != nil {
			return entry, err
		}
	case entry.AnythingLLMDocName != "":
		if err := m.syncUnembed(ctx, entry); err != nil {
			return entry, err
		}
	default:
		entry.EmbedStatus = knowledge.EmbedStatusNone
		if err := m.repo.Update(ctx, entry); err != nil {
			return nil, fmt.Errorf("knowledge: reset embed status: %w", err)
		}
	}
	return entry, nil
}

// ListDocuments mengembalikan semua entry knowledge (admin dashboard).
func (m *DocumentManager) ListDocuments(ctx context.Context) ([]knowledge.Entry, error) {
	entries, err := m.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("knowledge: list entries: %w", err)
	}
	return entries, nil
}

// GetDocument mengembalikan satu entry knowledge berdasarkan ID.
func (m *DocumentManager) GetDocument(ctx context.Context, id uint) (*knowledge.Entry, error) {
	entry, err := m.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			return nil, knowledge.ErrNotFound
		}
		return nil, fmt.Errorf("knowledge: find entry %d: %w", id, err)
	}
	return entry, nil
}

// syncEmbed meng-upload/re-embed dokumen ke AnythingLLM dan mencatat hasilnya
// ke entry. oldDocName = doc name JSON yang sudah ada ("" kalau belum pernah).
// Gagal → status failed (bisa retry); sukses → embedded + doc name terbaru.
func (m *DocumentManager) syncEmbed(ctx context.Context, entry *knowledge.Entry, oldDocName string) error {
	newDocName, err := m.manager.UpsertDocument(ctx, oldDocName, entry.Title, entry.Content)
	if err != nil {
		entry.EmbedStatus = knowledge.EmbedStatusFailed
		if repoErr := m.repo.Update(ctx, entry); repoErr != nil {
			return fmt.Errorf("knowledge: record embed failure: %w", repoErr)
		}
		return fmt.Errorf("%w: %w", ErrEmbedSync, err)
	}
	entry.EmbedToLLM = true
	entry.EmbedStatus = knowledge.EmbedStatusEmbedded
	entry.AnythingLLMDocName = newDocName
	if err := m.repo.Update(ctx, entry); err != nil {
		return fmt.Errorf("knowledge: save embed state: %w", err)
	}
	return nil
}

// syncUnembed menghapus dokumen dari AnythingLLM (toggle embed off). Gagal →
// status failed + doc name dipertahankan (RetryEmbed masih bisa membersihkan).
func (m *DocumentManager) syncUnembed(ctx context.Context, entry *knowledge.Entry) error {
	if entry.AnythingLLMDocName == "" {
		entry.EmbedStatus = knowledge.EmbedStatusNone
		entry.EmbedToLLM = false
		return m.repo.Update(ctx, entry)
	}
	if err := m.manager.DeleteDocument(ctx, entry.AnythingLLMDocName); err != nil {
		entry.EmbedStatus = knowledge.EmbedStatusFailed
		if repoErr := m.repo.Update(ctx, entry); repoErr != nil {
			return fmt.Errorf("knowledge: record unembed failure: %w", repoErr)
		}
		return fmt.Errorf("%w: %w", ErrEmbedSync, err)
	}
	entry.EmbedStatus = knowledge.EmbedStatusNone
	entry.AnythingLLMDocName = ""
	entry.EmbedToLLM = false
	if err := m.repo.Update(ctx, entry); err != nil {
		return fmt.Errorf("knowledge: save unembed state: %w", err)
	}
	return nil
}

func normalizeCategory(category string) string {
	if strings.TrimSpace(category) == "" {
		return DefaultCategory
	}
	return strings.TrimSpace(category)
}
