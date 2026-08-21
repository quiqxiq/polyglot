package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
)

const (
	historyCacheKeyFmt = "history:conv:%d"
	summaryCacheKeyFmt = "summary:conv:%d"

	historyCacheTTL = 24 * 3600     // 1 hari
	summaryCacheTTL = 7 * 24 * 3600 // 7 hari

	maxCachedTurns     = 40 // maksimal pasangan user/bot yang disimpan di cache
	summarizeThreshold = 20 // jika jumlah pesan cache >= ambang ini, ringkas
)

// summarizePrompt instructs the LLM to condense a cached conversation turn
// window into a short Indonesian summary that survives cache expiry.
var summarizePrompt = "Anda adalah peringkas percakapan layanan pelanggan internet. Ringkas percakapan berikut dalam Bahasa Indonesia, maksimal 3 kalimat. Pertahankan konteks penting: nama/nomor pelanggan, keluhan, dan solusi yang sudah diberikan. Jangan menambahkan informasi baru:"

const DefaultSystemPrompt = "Kamu adalah asisten layanan pelanggan internet yang ramah, sopan, dan solutif."

// ContextManager builds LLM prompt context scoped to a conversation (convID).
// Cache keys are per-conversation, bukan per nomor telepon, sehingga dua
// percakapan berbeda dari kontak yang sama tidak saling mencampur konteks.
type ContextManager struct {
	cache port.CacheStore
}

func NewContextManager(cache port.CacheStore) *ContextManager {
	return &ContextManager{
		cache: cache,
	}
}

// BuildPromptContext assembles the system prompt (composite skills prompt +
// summarized earlier context) and the ordered chat history for the LLM.
// history harus dalam urutan kronologis (pesan terbaru di akhir).
func (cm *ContextManager) BuildPromptContext(
	ctx context.Context,
	convID uint,
	history []bot.Message,
	basePrompt string,
) (systemPrompt string, messages []llm.ChatMessage, err error) {
	var sb strings.Builder
	if strings.TrimSpace(basePrompt) != "" {
		sb.WriteString(basePrompt)
	} else {
		sb.WriteString(DefaultSystemPrompt)
	}

	sb.WriteString("\n\n")

	if cm.cache != nil {
		summary, _ := cm.cache.Get(ctx, fmt.Sprintf(summaryCacheKeyFmt, convID))
		if summary != "" {
			sb.WriteString("### RINGKASAN PERCAKAPAN SEBELUMNYA:\n")
			sb.WriteString(summary)
			sb.WriteString("\n\n")
		}
	}

	sb.WriteString("### ATURAN OUTPUT (WAJIB DIPATUHI):\n")
	sb.WriteString("- JANGAN PERNAH menampilkan proses berpikir internal, tag <think>...</think>, nomor langkah perancangan, atau draf teks.\n")
	sb.WriteString("- Berikan HANYA balasan akhir yang ramah, sopan, langsung dalam Bahasa Indonesia untuk pelanggan WhatsApp.\n\n")

	systemPrompt = sb.String()

	for _, m := range history {
		role := "user"
		switch m.SenderType {
		case bot.SenderBot, bot.SenderAgent:
			role = "assistant"
		}
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		messages = append(messages, llm.ChatMessage{Role: role, Content: m.Content})
	}

	return systemPrompt, messages, nil
}

// SaveMessageToSession appends the latest user/bot turn to the per-conversation
// rolling cache window, trimming the oldest turns past maxCachedTurns.
func (cm *ContextManager) SaveMessageToSession(ctx context.Context, convID uint, userMsg string, botMsg string) error {
	if cm.cache == nil {
		return nil
	}

	key := fmt.Sprintf(historyCacheKeyFmt, convID)
	raw, _ := cm.cache.Get(ctx, key)

	var msgs []llm.ChatMessage
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &msgs)
	}

	msgs = append(msgs,
		llm.ChatMessage{Role: "user", Content: userMsg},
		llm.ChatMessage{Role: "assistant", Content: botMsg},
	)
	if len(msgs) > maxCachedTurns {
		msgs = msgs[len(msgs)-maxCachedTurns:]
	}

	data, err := json.Marshal(msgs)
	if err != nil {
		return fmt.Errorf("marshal session history: %w", err)
	}
	return cm.cache.Set(ctx, key, string(data), historyCacheTTL)
}

// GetSummary returns the stored per-conversation summary (if any) — dipakai
// untuk expose konteks LLM ke dashboard agen.
func (cm *ContextManager) GetSummary(ctx context.Context, convID uint) string {
	if cm.cache == nil {
		return ""
	}
	summary, _ := cm.cache.Get(ctx, fmt.Sprintf(summaryCacheKeyFmt, convID))
	return summary
}

// SummarizeSessionIfLong condenses the cached turn window into a summary when
// it grows past summarizeThreshold, then clears the raw window. The summary is
// injected by BuildPromptContext on subsequent messages.
func (cm *ContextManager) SummarizeSessionIfLong(ctx context.Context, convID uint, provider port.LLMProvider) error {
	if cm.cache == nil || provider == nil {
		return nil
	}

	key := fmt.Sprintf(historyCacheKeyFmt, convID)
	raw, err := cm.cache.Get(ctx, key)
	if err != nil || raw == "" {
		return nil
	}

	var msgs []llm.ChatMessage
	if err := json.Unmarshal([]byte(raw), &msgs); err != nil {
		return nil
	}
	if len(msgs) < summarizeThreshold {
		return nil
	}

	resp, err := provider.Chat(ctx, summarizePrompt, msgs, 256)
	if err != nil {
		return fmt.Errorf("summarize session: %w", err)
	}

	if err := cm.cache.Set(ctx, fmt.Sprintf(summaryCacheKeyFmt, convID), resp.Content, summaryCacheTTL); err != nil {
		return fmt.Errorf("store session summary: %w", err)
	}
	return cm.cache.Delete(ctx, key)
}
