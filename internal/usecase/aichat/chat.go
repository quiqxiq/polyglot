// Package aichat implements the AI Router Ops chat agent: READ tools execute
// immediately, MUTATION tools only create approval requests.
package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/aichat"
	"github.com/quixiq/polyglot/internal/domain/approval"
	domainLLM "github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/domain/restorepoint"
	"github.com/quixiq/polyglot/internal/port"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/crypto"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// Event is one streamed chat event.
type Event struct {
	ID      int64
	Type    string
	Payload map[string]any
}

// Config holds chat behavior knobs.
type Config struct {
	MaxContextMessages int
	MaxOutputTokens    int
}

// Deps wires the chat usecase.
type Deps struct {
	Repo          port.AIChatRepository
	Configs       port.LLMConfigRepository
	EncryptionKey string
	Provider      func(ctx context.Context, cfg *domainLLM.Config, apiKey string) (port.LLMProvider, error)
	Devices       port.DeviceRepository
	ActiveDevice  ActiveDeviceResolver
	Ops           func(ctx context.Context, deviceID string) (port.RouterOps, error)
	Proposals     *networkUC.ProposalUseCase
	Approvals     port.ApprovalRepository
	RestorePoints *networkUC.RestorePointUseCase
	Config        Config
}

// ActiveDeviceResolver is the subset needed by chat.
type ActiveDeviceResolver interface {
	Resolve(ctx context.Context, tenantID, userID, sessionID string) (string, error)
}

// ChatInput is one chat turn.
type ChatInput struct {
	TenantID       string
	UserID         uint
	ConversationID string
	Message        string
	DeviceID       string
	ModelConfigID  string
	SinceEventID   int64
}

// UseCase runs the AI Router Ops chat agent.
type UseCase struct {
	deps Deps
}

// New builds the chat usecase.
func New(deps Deps) *UseCase {
	if deps.Config.MaxContextMessages <= 0 {
		deps.Config.MaxContextMessages = 50
	}
	if deps.Config.MaxOutputTokens <= 0 {
		deps.Config.MaxOutputTokens = 2048
	}
	return &UseCase{deps: deps}
}

// Chat starts one streaming turn. The returned channel closes after Done/Error.
func (uc *UseCase) Chat(ctx context.Context, in ChatInput) (<-chan Event, error) {
	if strings.TrimSpace(in.Message) == "" {
		return nil, fmt.Errorf("%w: message is required", aichat.ErrInvalidInput)
	}
	if in.TenantID == "" {
		in.TenantID = "tenant-default"
	}
	conv, err := uc.resolveConversation(ctx, in)
	if err != nil {
		return nil, err
	}
	modelCfg, apiKey, err := uc.resolveModel(ctx, in.ModelConfigID)
	if err != nil {
		return nil, err
	}
	provider, err := uc.deps.Provider(ctx, modelCfg, apiKey)
	if err != nil {
		return nil, err
	}
	ops, err := uc.deps.Ops(ctx, conv.DeviceID)
	if err != nil {
		return nil, err
	}

	history := uc.history(ctx, conv.ID)
	history = append(history, domainLLM.ChatMessage{Role: "user", Content: in.Message})
	if _, err := uc.deps.Repo.AppendEvent(ctx, conv.ID, aichat.EventText, mustJSON(map[string]any{"text": in.Message, "role": "user"})); err != nil {
		return nil, err
	}

	out := make(chan Event, 64)
	emit := func(eventType string, payload map[string]any) {
		persisted, perr := uc.deps.Repo.AppendEvent(ctx, conv.ID, eventType, mustJSON(payload))
		if perr == nil {
			out <- Event{ID: persisted.ID, Type: eventType, Payload: payload}
			return
		}
		out <- Event{Type: eventType, Payload: payload}
	}

	go func() {
		defer close(out)
		emit(aichat.EventPlanStep, map[string]any{"step": "Memahami permintaan & konteks router", "status": "running"})
		tools := uc.tools(ctx, conv, ops, emit)
		systemPrompt := fmt.Sprintf(`Anda adalah asisten operasional RouterOS untuk Polyglot.
Aturan wajib:
1. READ/inspeksi bebas. Selalu inspeksi sebelum mengusulkan perubahan.
2. Setiap MUTATION hanya boleh lewat propose_command/propose_config. JANGAN pernah mengklaim perubahan sudah diterapkan sebelum approval berstatus EXECUTED.
3. Jangan meminta/menampilkan kredensial.
4. Perangkat aktif: %s. Selalu sebut perangkat dalam jawaban.
5. Jika approval dibuat, jelaskan bahwa operator harus menyetujuinya di AI Ops Approval Inbox.`, conv.DeviceID)

		resp, err := provider.ChatWithTools(ctx, systemPrompt, history, tools, uc.deps.Config.MaxOutputTokens)
		if err != nil {
			emit(aichat.EventError, map[string]any{"message": err.Error()})
			return
		}
		emit(aichat.EventText, map[string]any{"text": resp.Content, "role": "assistant"})
		emit(aichat.EventUsage, map[string]any{"input_tokens": resp.TokenIn, "output_tokens": resp.TokenOut})
		last, _ := uc.deps.Repo.AppendEvent(ctx, conv.ID, aichat.EventDone, mustJSON(map[string]any{"conversation_id": conv.ID}))
		emit(aichat.EventDone, map[string]any{"conversation_id": conv.ID, "last_event_id": last.ID})
	}()
	return out, nil
}

func (uc *UseCase) resolveConversation(ctx context.Context, in ChatInput) (aichat.Conversation, error) {
	if in.ConversationID != "" {
		conv, err := uc.deps.Repo.FindConversation(ctx, in.ConversationID)
		if err != nil {
			return aichat.Conversation{}, err
		}
		if conv.UserID != 0 && conv.UserID != in.UserID {
			return aichat.Conversation{}, fmt.Errorf("%w: conversation belongs to another user", aichat.ErrInvalidInput)
		}
		return conv, nil
	}
	deviceID := in.DeviceID
	if deviceID == "" {
		if uc.deps.ActiveDevice == nil {
			return aichat.Conversation{}, fmt.Errorf("no active router selected: choose a router in the sidebar or pass device_id")
		}
		resolved, err := uc.deps.ActiveDevice.Resolve(ctx, in.TenantID, fmt.Sprintf("%d", in.UserID), "")
		if err != nil {
			return aichat.Conversation{}, err
		}
		deviceID = resolved
	}
	if _, err := uc.deps.Devices.FindByID(ctx, deviceID); err != nil {
		return aichat.Conversation{}, err
	}
	now := time.Now().UTC()
	conv := aichat.Conversation{
		ID: idgen.New("conv"), TenantID: in.TenantID, UserID: in.UserID, DeviceID: deviceID,
		Title: truncateTitle(in.Message), Status: aichat.StatusActive, CreatedAt: now, UpdatedAt: now,
	}
	if err := uc.deps.Repo.CreateConversation(ctx, conv); err != nil {
		return aichat.Conversation{}, err
	}
	return conv, nil
}

func (uc *UseCase) resolveModel(ctx context.Context, modelConfigID string) (*domainLLM.Config, string, error) {
	var (
		cfg *domainLLM.Config
		err error
	)
	if modelConfigID != "" {
		id := parseUint(modelConfigID)
		if id == 0 {
			return nil, "", fmt.Errorf("%w: invalid model_config_id", aichat.ErrInvalidInput)
		}
		cfg, err = uc.deps.Configs.FindByID(ctx, id)
	} else {
		cfg, err = uc.deps.Configs.FindActive(ctx)
	}
	if err != nil || cfg == nil {
		return nil, "", fmt.Errorf("%w: no LLM config available", aichat.ErrProviderFailed)
	}
	apiKey, err := crypto.Decrypt(cfg.APIKeyEncrypted, uc.deps.EncryptionKey)
	if err != nil {
		return nil, "", fmt.Errorf("%w: decrypt api key", aichat.ErrProviderFailed)
	}
	return cfg, apiKey, nil
}

func (uc *UseCase) history(ctx context.Context, conversationID string) []domainLLM.ChatMessage {
	events, err := uc.deps.Repo.ListEvents(ctx, conversationID, 0, uc.deps.Config.MaxContextMessages*2)
	if err != nil {
		return nil
	}
	out := make([]domainLLM.ChatMessage, 0, len(events))
	for _, e := range events {
		if e.Type != aichat.EventText {
			continue
		}
		var payload struct {
			Text string `json:"text"`
			Role string `json:"role"`
		}
		if json.Unmarshal(e.Payload, &payload) != nil || payload.Text == "" {
			continue
		}
		role := payload.Role
		if role == "" {
			role = "assistant"
		}
		out = append(out, domainLLM.ChatMessage{Role: role, Content: payload.Text})
	}
	if len(out) > uc.deps.Config.MaxContextMessages {
		out = out[len(out)-uc.deps.Config.MaxContextMessages:]
	}
	return out
}

func (uc *UseCase) tools(ctx context.Context, conv aichat.Conversation, ops port.RouterOps, emit func(string, map[string]any)) []domainLLM.Tool {
	tool := func(name, desc string, fn func(argsJSON string) (string, error)) domainLLM.Tool {
		return domainLLM.Tool{
			Name: name, Description: desc,
			Handler: func(_ context.Context, argsJSON string) (string, error) {
				start := time.Now()
				callID := idgen.New("tc")
				emit(aichat.EventToolCall, map[string]any{"id": callID, "name": name, "args_json": argsJSON})
				res, err := fn(argsJSON)
				dur := time.Since(start).Milliseconds()
				payload := map[string]any{"tool_call_id": callID, "duration_ms": dur, "result": res}
				if err != nil {
					payload["is_error"] = true
					payload["result"] = err.Error()
					emit(aichat.EventToolResult, payload)
					return toJSON(map[string]any{"error": err.Error()}), nil
				}
				emit(aichat.EventToolResult, payload)
				return res, nil
			},
		}
	}

	return []domainLLM.Tool{
		tool("router_capabilities", "RouterOS version and AI support mode", func(_ string) (string, error) {
			caps, err := ops.Capabilities(ctx)
			if err != nil {
				return "", err
			}
			return toJSON(caps), nil
		}),
		tool("router_read", "Read rows from a router menu (print detail). To read all rows, omit selector or pass empty selector. Example: {\"menu\":\"interface\"} or with filter {\"menu\":\"ip/address\",\"selector\":{\"interface\":\"ether1\"},\"limit\":100}", func(argsJSON string) (string, error) {
			var args struct {
				Menu     string            `json:"menu"`
				Selector map[string]string `json:"selector"`
				Limit    int               `json:"limit"`
			}
			if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
				return "", err
			}
			rows, err := ops.ReadMenu(ctx, args.Menu, args.Selector)
			if err != nil {
				return "", err
			}
			if args.Limit > 0 && len(rows) > args.Limit {
				rows = rows[:args.Limit]
			}
			return toJSON(rows), nil
		}),
		tool("inspect_validate", "Validate a command/script: {\"script\":\"...\"}", func(argsJSON string) (string, error) {
			var args struct {
				Script string `json:"script"`
			}
			if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
				return "", err
			}
			res, err := ops.ValidateScript(ctx, args.Script)
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		}),
		tool("list_approvals", "List recent approvals: {\"status\":\"PENDING\",\"limit\":10}", func(argsJSON string) (string, error) {
			var args struct {
				Status string `json:"status"`
				Limit  int    `json:"limit"`
			}
			_ = json.Unmarshal([]byte(argsJSON), &args)
			if args.Limit <= 0 {
				args.Limit = 10
			}
			items, _, err := uc.deps.Approvals.List(ctx, port.ApprovalFilter{TenantID: conv.TenantID, Status: args.Status, Limit: args.Limit})
			if err != nil {
				return "", err
			}
			out := make([]map[string]any, len(items))
			for i, a := range items {
				out[i] = map[string]any{"id": a.ID, "status": a.Status, "command": a.Command, "revision": a.Revision}
			}
			return toJSON(out), nil
		}),
		tool("list_restore_points", "List recent restore points: {\"limit\":10}", func(argsJSON string) (string, error) {
			var args struct {
				Limit int `json:"limit"`
			}
			_ = json.Unmarshal([]byte(argsJSON), &args)
			if args.Limit <= 0 {
				args.Limit = 10
			}
			items, _, err := uc.deps.RestorePoints.List(ctx, port.RestorePointFilter{TenantID: conv.TenantID, DeviceID: conv.DeviceID, Limit: args.Limit})
			if err != nil {
				return "", err
			}
			out := make([]map[string]any, len(items))
			for i, r := range items {
				out[i] = map[string]any{"id": r.ID, "status": r.Status, "trigger": r.Trigger, "size_bytes": r.TotalSizeBytes}
			}
			return toJSON(out), nil
		}),
		tool("propose_command", "Propose ONE mutation for human approval. Never executes. {\"command\":\"...\",\"args\":{...}}", func(argsJSON string) (string, error) {
			var args struct {
				Command string            `json:"command"`
				Args    map[string]string `json:"args"`
			}
			if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
				return "", err
			}
			res, err := uc.deps.Proposals.ProposeCommand(ctx, networkUC.ProposeCommandInput{
				TenantID: conv.TenantID, DeviceID: conv.DeviceID, ActorID: fmt.Sprintf("%d", conv.UserID),
				ActorType: approval.ActorAgent, Command: args.Command, Args: args.Args,
			})
			if err != nil {
				return "", err
			}
			emit(aichat.EventApproval, map[string]any{"approval_id": res.Approval.ID, "summary": res.Diff.Summary, "restore_point_id": res.RestorePoint.ID})
			return toJSON(map[string]any{
				"status": "pending_approval", "executed": false, "approval_id": res.Approval.ID,
				"restore_point_id": res.RestorePoint.ID, "restore_point_status": res.RestorePoint.Status,
				"diff": res.Diff, "validation": res.Validation,
			}), nil
		}),
		tool("propose_config", "Propose a multi-line configuration script for approval. Never executes. {\"script\":\"...\"}", func(argsJSON string) (string, error) {
			var args struct {
				Script string `json:"script"`
			}
			if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
				return "", err
			}
			res, err := uc.deps.Proposals.ProposeConfig(ctx, networkUC.ProposeCommandInput{
				TenantID: conv.TenantID, DeviceID: conv.DeviceID, ActorID: fmt.Sprintf("%d", conv.UserID),
				ActorType: approval.ActorAgent, Command: args.Script,
			})
			if err != nil {
				return "", err
			}
			emit(aichat.EventApproval, map[string]any{"approval_id": res.Approval.ID, "summary": "multi-line config", "restore_point_id": res.RestorePoint.ID})
			return toJSON(map[string]any{"status": "pending_approval", "executed": false, "approval_id": res.Approval.ID, "restore_point_id": res.RestorePoint.ID}), nil
		}),
	}
}

// RestorePointSummary is a small helper used by tests/UI.
func RestorePointSummary(rp restorepoint.RestorePoint) map[string]any {
	return map[string]any{"id": rp.ID, "status": rp.Status, "size_bytes": rp.TotalSizeBytes}
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func truncateTitle(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 60 {
		return s
	}
	return s[:60] + "…"
}

func parseUint(s string) uint {
	var v uint
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		v = v*10 + uint(r-'0')
	}
	return v
}
