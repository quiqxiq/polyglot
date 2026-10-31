package aichat_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/aichat"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	domainLLM "github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
	aichatUC "github.com/quixiq/polyglot/internal/usecase/aichat"
	"github.com/quixiq/polyglot/pkg/crypto"
)

const testKey = "01234567890123456789012345678901"

type fakeChatRepo struct {
	mu     sync.Mutex
	convs  map[string]aichat.Conversation
	events map[string][]aichat.Event
	seq    int64
}

func newFakeChatRepo() *fakeChatRepo {
	return &fakeChatRepo{convs: map[string]aichat.Conversation{}, events: map[string][]aichat.Event{}}
}
func (f *fakeChatRepo) CreateConversation(_ context.Context, c aichat.Conversation) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.convs[c.ID] = c
	return nil
}
func (f *fakeChatRepo) FindConversation(_ context.Context, id string) (aichat.Conversation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.convs[id]
	if !ok {
		return aichat.Conversation{}, aichat.ErrNotFound
	}
	return c, nil
}
func (f *fakeChatRepo) UpdateConversation(_ context.Context, c aichat.Conversation) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.convs[c.ID] = c
	return nil
}
func (f *fakeChatRepo) AppendEvent(_ context.Context, conversationID, eventType string, payload json.RawMessage) (aichat.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seq++
	ev := aichat.Event{ID: f.seq, ConversationID: conversationID, Seq: f.seq, Type: eventType, Payload: payload}
	f.events[conversationID] = append(f.events[conversationID], ev)
	return ev, nil
}
func (f *fakeChatRepo) DeleteEventsBefore(_ context.Context, cutoff time.Time) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	deleted := 0
	for id, events := range f.events {
		kept := events[:0]
		for _, e := range events {
			if e.CreatedAt.Before(cutoff) {
				deleted++
				continue
			}
			kept = append(kept, e)
		}
		f.events[id] = kept
	}
	return deleted, nil
}
func (f *fakeChatRepo) ListEvents(_ context.Context, conversationID string, sinceSeq int64, limit int) ([]aichat.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []aichat.Event
	for _, e := range f.events[conversationID] {
		if e.Seq > sinceSeq {
			out = append(out, e)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

type fakeConfigRepo struct{ cfg *domainLLM.Config }

func (f *fakeConfigRepo) Create(context.Context, *domainLLM.Config) error { return nil }
func (f *fakeConfigRepo) FindByID(context.Context, uint) (*domainLLM.Config, error) {
	return f.cfg, nil
}
func (f *fakeConfigRepo) FindActive(context.Context) (*domainLLM.Config, error) { return f.cfg, nil }
func (f *fakeConfigRepo) FindAll(context.Context) ([]domainLLM.Config, error)   { return nil, nil }
func (f *fakeConfigRepo) Update(context.Context, *domainLLM.Config) error       { return nil }
func (f *fakeConfigRepo) SetActive(context.Context, uint) error                 { return nil }
func (f *fakeConfigRepo) Delete(context.Context, uint) error                    { return nil }

type fakeDevices struct{ d device.Device }

func (f *fakeDevices) Save(context.Context, device.Device) error { return nil }
func (f *fakeDevices) FindByID(context.Context, string) (device.Device, error) {
	return f.d, nil
}
func (f *fakeDevices) FindAll(context.Context) ([]device.Device, error) {
	return []device.Device{f.d}, nil
}
func (f *fakeDevices) Update(context.Context, device.Device) error { return nil }
func (f *fakeDevices) Delete(context.Context, string) error        { return nil }
func (f *fakeDevices) FindByUserScope(context.Context, uint) ([]device.Device, error) {
	return nil, nil
}

type fakeActive struct{ id string }

func (f *fakeActive) Resolve(context.Context, string, string, string) (string, error) {
	return f.id, nil
}

type fakeOps struct{}

func (fakeOps) Capabilities(context.Context) (port.DeviceCapabilities, error) {
	return port.DeviceCapabilities{Version: "7.16.2", Major: 7, InspectSupported: true, AIMode: "full"}, nil
}
func (fakeOps) ValidateScript(context.Context, string) (port.ScriptValidation, error) {
	return port.ScriptValidation{Supported: true, Valid: true}, nil
}
func (fakeOps) ReadMenu(context.Context, string, map[string]string) ([]map[string]string, error) {
	return []map[string]string{{"name": "ether1"}}, nil
}
func (fakeOps) CreateArtifacts(context.Context, string, []string, []byte, int64) ([]port.RawArtifact, error) {
	return nil, nil
}
func (fakeOps) ExecuteRaw(context.Context, string, map[string]string) (command.Result, error) {
	return command.Result{}, nil
}
func (fakeOps) UploadFile(context.Context, string, []byte) error { return nil }
func (fakeOps) ImportScript(context.Context, string) error       { return nil }
func (fakeOps) LoadBinary(context.Context, string) error         { return nil }
func (fakeOps) RemoveFile(context.Context, string)               {}
func (fakeOps) PostVerify(context.Context, []string) (map[string][]map[string]string, error) {
	return nil, nil
}
func (fakeOps) ResourceInfo(context.Context) (string, int64, error) { return "7.16.2", 1 << 30, nil }
func (fakeOps) RouterNow(context.Context) (time.Time, error)        { return time.Now().UTC(), nil }

type fakeProvider struct{ calledTool bool }

func (f *fakeProvider) Chat(context.Context, string, []domainLLM.ChatMessage, int) (*domainLLM.ChatResponse, error) {
	return &domainLLM.ChatResponse{Content: "ok"}, nil
}
func (f *fakeProvider) ChatWithTools(_ context.Context, _ string, _ []domainLLM.ChatMessage, tools []domainLLM.Tool, _ int) (*domainLLM.ChatResponse, error) {
	for _, tool := range tools {
		if tool.Name == "router_capabilities" && !f.calledTool {
			f.calledTool = true
			_, _ = tool.Handler(context.Background(), "{}")
		}
	}
	return &domainLLM.ChatResponse{Content: "Router sehat.", TokenIn: 10, TokenOut: 5}, nil
}
func (f *fakeProvider) TestConnection(context.Context) error { return nil }

func TestChatStreamsReadToolsAndPersistsEvents(t *testing.T) {
	enc, err := crypto.Encrypt("sk-test", testKey)
	require.NoError(t, err)
	repo := newFakeChatRepo()
	cfgRepo := &fakeConfigRepo{cfg: &domainLLM.Config{ID: 1, Provider: "openai", Model: "gpt-4o-mini", APIKeyEncrypted: enc, IsActive: true}}
	provider := &fakeProvider{}
	uc := aichatUC.New(aichatUC.Deps{
		Repo: repo, Configs: cfgRepo, EncryptionKey: testKey,
		Provider:     func(context.Context, *domainLLM.Config, string) (port.LLMProvider, error) { return provider, nil },
		Devices:      &fakeDevices{d: device.Device{ID: "dev1", TenantID: "tenant-default", Name: "Router-A", Enabled: true}},
		ActiveDevice: &fakeActive{id: "dev1"},
		Ops:          func(context.Context, string) (port.RouterOps, error) { return fakeOps{}, nil },
		Config:       aichatUC.Config{MaxContextMessages: 10},
	})
	ch, err := uc.Chat(context.Background(), aichatUC.ChatInput{TenantID: "tenant-default", UserID: 1, Message: "cek router"})
	require.NoError(t, err)

	var types []string
	for ev := range ch {
		types = append(types, ev.Type)
	}
	require.True(t, provider.calledTool)
	assert.Contains(t, types, aichat.EventPlanStep)
	assert.Contains(t, types, aichat.EventToolCall)
	assert.Contains(t, types, aichat.EventToolResult)
	assert.Contains(t, types, aichat.EventText)
	assert.Equal(t, aichat.EventDone, types[len(types)-1])

	var convID string
	for id := range repo.convs {
		convID = id
	}
	require.NotEmpty(t, convID)
	assert.Equal(t, "dev1", repo.convs[convID].DeviceID)
	events, _ := repo.ListEvents(context.Background(), convID, 0, 100)
	var textCount int
	for _, e := range events {
		if e.Type == aichat.EventText {
			textCount++
		}
	}
	assert.GreaterOrEqual(t, textCount, 2)
}

func TestChatRequiresMessageAndDevice(t *testing.T) {
	uc := aichatUC.New(aichatUC.Deps{Repo: newFakeChatRepo()})
	if _, err := uc.Chat(context.Background(), aichatUC.ChatInput{Message: "  "}); err == nil {
		t.Fatal("empty message must be rejected")
	}
	if _, err := uc.Chat(context.Background(), aichatUC.ChatInput{Message: "halo", UserID: 1}); err == nil {
		t.Fatal("missing device/active binding must be rejected")
	}
}

type errorOps struct {
	fakeOps
}

func (errorOps) ReadMenu(context.Context, string, map[string]string) ([]map[string]string, error) {
	return nil, errors.New("unknown parameter .id")
}

type fakeProviderToolCustom struct {
	onTools func(tools []domainLLM.Tool)
}

func (f *fakeProviderToolCustom) Chat(context.Context, string, []domainLLM.ChatMessage, int) (*domainLLM.ChatResponse, error) {
	return &domainLLM.ChatResponse{Content: "ok"}, nil
}

func (f *fakeProviderToolCustom) ChatWithTools(_ context.Context, _ string, _ []domainLLM.ChatMessage, tools []domainLLM.Tool, _ int) (*domainLLM.ChatResponse, error) {
	if f.onTools != nil {
		f.onTools(tools)
	}
	return &domainLLM.ChatResponse{Content: "Maaf, terjadi error saat membaca interface.", TokenIn: 10, TokenOut: 5}, nil
}

func (f *fakeProviderToolCustom) TestConnection(context.Context) error { return nil }

func TestChatToolErrorReturnsJSONPayloadWithoutFailingTurn(t *testing.T) {
	enc, err := crypto.Encrypt("sk-test", testKey)
	require.NoError(t, err)
	repo := newFakeChatRepo()
	cfgRepo := &fakeConfigRepo{cfg: &domainLLM.Config{ID: 1, Provider: "openai", Model: "gpt-4o-mini", APIKeyEncrypted: enc, IsActive: true}}
	var toolRes string
	var toolErr error
	var handlerCalled bool
	provider := &fakeProviderToolCustom{
		onTools: func(tools []domainLLM.Tool) {
			for _, tool := range tools {
				if tool.Name == "router_read" {
					handlerCalled = true
					toolRes, toolErr = tool.Handler(context.Background(), `{"menu":"interface"}`)
				}
			}
		},
	}
	uc := aichatUC.New(aichatUC.Deps{
		Repo: repo, Configs: cfgRepo, EncryptionKey: testKey,
		Provider:     func(context.Context, *domainLLM.Config, string) (port.LLMProvider, error) { return provider, nil },
		Devices:      &fakeDevices{d: device.Device{ID: "dev1", TenantID: "tenant-default", Name: "Router-A", Enabled: true}},
		ActiveDevice: &fakeActive{id: "dev1"},
		Ops:          func(context.Context, string) (port.RouterOps, error) { return errorOps{}, nil },
		Config:       aichatUC.Config{MaxContextMessages: 10},
	})
	ch, err := uc.Chat(context.Background(), aichatUC.ChatInput{TenantID: "tenant-default", UserID: 1, Message: "cek interface"})
	require.NoError(t, err)

	var hasToolResultError bool
	for ev := range ch {
		if ev.Type == aichat.EventToolResult {
			if isErr, ok := ev.Payload["is_error"].(bool); ok && isErr {
				hasToolResultError = true
			}
		}
	}

	require.True(t, handlerCalled)
	require.NoError(t, toolErr, "tool handler must return nil error to avoid crashing LLM generator")
	assert.Contains(t, toolRes, "unknown parameter .id")
	assert.True(t, hasToolResultError, "tool result event must flag is_error=true for UI display")
}
