package whatsapp

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	domainWA "github.com/quixiq/polyglot/internal/domain/whatsapp"
)

const (
	// WhatsApp poll limits. The question and option limits are the practical
	// UI limits; the option cap keeps the message readable on a phone.
	maxPollOptions       = 12
	maxPollQuestionRunes = 255
	maxPollOptionRunes   = 100
)

var waPollBlockRe = regexp.MustCompile("(?s)```wa-poll[ \t]*\\r?\\n(.*?)```")

// pollDirective is a validated WhatsApp poll request returned by the harness.
type pollDirective struct {
	Question string
	Text     string
	Select   int
	Options  []pollOption
}

type pollOption struct {
	Label string
	Value string
}

type rawPollDirective struct {
	Question string            `json:"question"`
	Text     string            `json:"text"`
	Select   int               `json:"select"`
	Options  []json.RawMessage `json:"options"`
}

type rawPollOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// parsePollDirective extracts the first wa-poll block from a harness reply.
//
// It always returns the reply with every wa-poll block removed so an invalid
// directive never leaks JSON into the chat. When no block exists, directive is
// nil and prose is the original reply. When a block is invalid, directive is
// nil and err explains why so the caller can fall back to prose.
func parsePollDirective(reply string) (*pollDirective, string, error) {
	matches := waPollBlockRe.FindAllStringSubmatchIndex(reply, -1)
	if len(matches) == 0 {
		return nil, reply, nil
	}

	prose := waPollBlockRe.ReplaceAllString(reply, "")
	prose = strings.TrimSpace(prose)
	if len(matches) > 1 {
		return nil, prose, errors.New("poll directive appears more than once")
	}

	body := reply[matches[0][2]:matches[0][3]]
	directive, err := decodePollDirective(body)
	if err != nil {
		return nil, prose, err
	}
	return directive, prose, nil
}

func decodePollDirective(body string) (*pollDirective, error) {
	var raw rawPollDirective
	decoder := json.NewDecoder(strings.NewReader(body))
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode poll directive: %w", err)
	}

	question := normalizePollText(raw.Question)
	if question == "" {
		return nil, errors.New("poll question is empty")
	}
	if runeCount(question) > maxPollQuestionRunes {
		return nil, fmt.Errorf("poll question exceeds %d characters", maxPollQuestionRunes)
	}
	if len(raw.Options) < 2 {
		return nil, errors.New("poll requires at least two options")
	}
	if len(raw.Options) > maxPollOptions {
		return nil, fmt.Errorf("poll allows at most %d options", maxPollOptions)
	}

	options := make([]pollOption, 0, len(raw.Options))
	seen := make(map[string]struct{}, len(raw.Options))
	for i, rawOption := range raw.Options {
		option, err := decodePollOption(rawOption)
		if err != nil {
			return nil, fmt.Errorf("poll option %d: %w", i+1, err)
		}
		if runeCount(option.Label) > maxPollOptionRunes {
			return nil, fmt.Errorf("poll option %d label exceeds %d characters", i+1, maxPollOptionRunes)
		}
		key := strings.ToLower(option.Label)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("poll option %d duplicates another label", i+1)
		}
		seen[key] = struct{}{}
		options = append(options, option)
	}

	selectable := raw.Select
	if selectable == 0 {
		selectable = 1
	}
	if selectable < 1 || selectable > len(options) {
		return nil, fmt.Errorf("poll select must be between 1 and %d", len(options))
	}

	return &pollDirective{
		Question: question,
		Text:     strings.TrimSpace(raw.Text),
		Select:   selectable,
		Options:  options,
	}, nil
}

func decodePollOption(raw json.RawMessage) (pollOption, error) {
	var label string
	if err := json.Unmarshal(raw, &label); err == nil {
		label = normalizePollText(label)
		if label == "" {
			return pollOption{}, errors.New("label is empty")
		}
		return pollOption{Label: label, Value: label}, nil
	}

	var option rawPollOption
	if err := json.Unmarshal(raw, &option); err != nil {
		return pollOption{}, errors.New("label must be a string or an object")
	}
	option.Label = normalizePollText(option.Label)
	if option.Label == "" {
		return pollOption{}, errors.New("label is empty")
	}
	option.Value = strings.TrimSpace(option.Value)
	if option.Value == "" {
		option.Value = option.Label
	}
	return pollOption(option), nil
}

// normalizePollText collapses whitespace so a poll label or question cannot
// introduce line breaks into the WhatsApp poll UI.
func normalizePollText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func runeCount(text string) int {
	return len([]rune(text))
}

type pendingPoll struct {
	sessionID uint
	chatJID   string
	pollID    string
	request   string
	options   []pendingPollOption
	createdAt time.Time
}

type pendingPollOption struct {
	Label string
	Value string
	Hash  [sha256.Size]byte
}

func newPendingPoll(sessionID uint, chatJID string, pollID string, request string, options []pollOption) *pendingPoll {
	pending := &pendingPoll{
		sessionID: sessionID,
		chatJID:   chatJID,
		pollID:    pollID,
		request:   request,
		options:   make([]pendingPollOption, 0, len(options)),
		createdAt: time.Now(),
	}
	for _, option := range options {
		pending.options = append(pending.options, pendingPollOption{
			Label: option.Label,
			Value: option.Value,
			Hash:  sha256.Sum256([]byte(option.Label)),
		})
	}
	return pending
}

func matchPollOptions(options []pendingPollOption, hashes [][]byte) []pendingPollOption {
	selected := make([]pendingPollOption, 0, len(hashes))
	seen := make(map[int]struct{}, len(hashes))
	for _, hash := range hashes {
		for i, option := range options {
			if _, already := seen[i]; already {
				continue
			}
			if bytes.Equal(hash, option.Hash[:]) {
				selected = append(selected, option)
				seen[i] = struct{}{}
				break
			}
		}
	}
	return selected
}

func pollFallbackText(directive *pollDirective) string {
	var b strings.Builder
	b.WriteString(directive.Question)
	for _, option := range directive.Options {
		b.WriteString("\n• ")
		b.WriteString(option.Label)
	}
	return b.String()
}

func (a *Assistant) peekPendingPoll(pollID string) *pendingPoll {
	a.pollMu.Lock()
	defer a.pollMu.Unlock()
	a.prunePollsLocked(time.Now())

	pending, ok := a.polls[pollID]
	if !ok {
		return nil
	}
	return clonePendingPoll(pending)
}

func (a *Assistant) takePendingPoll(pollID string) *pendingPoll {
	a.pollMu.Lock()
	defer a.pollMu.Unlock()
	a.prunePollsLocked(time.Now())

	pending, ok := a.polls[pollID]
	if !ok {
		return nil
	}
	delete(a.polls, pollID)
	return clonePendingPoll(pending)
}

func (a *Assistant) storePendingPoll(pending *pendingPoll) {
	if pending == nil || pending.pollID == "" {
		return
	}

	a.pollMu.Lock()
	defer a.pollMu.Unlock()

	if a.polls == nil {
		a.polls = make(map[string]*pendingPoll)
	}
	a.prunePollsLocked(time.Now())

	if len(a.polls) >= maxPendingPolls {
		var oldestID string
		var oldest time.Time
		for id, candidate := range a.polls {
			if oldestID == "" || candidate.createdAt.Before(oldest) {
				oldestID = id
				oldest = candidate.createdAt
			}
		}
		if oldestID != "" {
			delete(a.polls, oldestID)
		}
	}

	a.polls[pending.pollID] = clonePendingPoll(pending)
}

func (a *Assistant) prunePollsLocked(now time.Time) {
	for id, pending := range a.polls {
		if now.Sub(pending.createdAt) > assistantPollTTL {
			delete(a.polls, id)
		}
	}
}

func clonePendingPoll(pending *pendingPoll) *pendingPoll {
	if pending == nil {
		return nil
	}
	cloned := *pending
	cloned.options = append([]pendingPollOption(nil), pending.options...)
	return &cloned
}

func buildPollFollowUpPrompt(identity *domainWA.SenderIdentity, request string, selected []pendingPollOption, now time.Time) string {
	name := strings.TrimSpace(identity.DisplayName)
	if name == "" {
		name = "Admin"
	}
	return fmt.Sprintf(`Pesan WhatsApp dari admin.
Nama: %s
Waktu: %s
Permintaan awal:
%s

Admin memilih opsi poll:
%s

Lanjutkan permintaan awal menggunakan pilihan tersebut. Periksa status terbaru dengan tool baca sebelum menjawab.

%s

%s`,
		name, now.Format(time.RFC3339), request, formatSelectedOptions(selected), assistantPollInstruction, assistantInstructions)
}

func pollSelectionContext(request string, selected []pendingPollOption) string {
	return fmt.Sprintf("%s\n\nPilihan yang sudah dibuat:\n%s", strings.TrimSpace(request), formatSelectedOptions(selected))
}

func formatSelectedOptions(selected []pendingPollOption) string {
	var b strings.Builder
	for _, option := range selected {
		b.WriteString("- ")
		b.WriteString(option.Label)
		if option.Value != "" && option.Value != option.Label {
			b.WriteString(" (nilai: ")
			b.WriteString(option.Value)
			b.WriteString(")")
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
