package aichat

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	aichatUC "github.com/quixiq/polyglot/internal/usecase/aichat"
	"github.com/quixiq/polyglot/pkg/response"
)

// Handler implements polyglot.v1.AIChatService.
type Handler struct {
	uc *aichatUC.UseCase
}

// NewHandler builds the stream handler.
func NewHandler(uc *aichatUC.UseCase) *Handler { return &Handler{uc: uc} }

// Chat streams one AI Ops turn.
func (h *Handler) Chat(ctx context.Context, req *connect.Request[devicepb.ChatRequest], stream *connect.ServerStream[devicepb.ChatEvent]) error {
	userID, _, ok := iauth.IdentityFromContext(ctx)
	if !ok || userID == 0 {
		return response.Unauthenticated("user session required")
	}
	if h.uc == nil {
		return response.Unavailable("AI chat unavailable")
	}
	events, err := h.uc.Chat(ctx, aichatUC.ChatInput{
		TenantID: "tenant-default", UserID: userID,
		ConversationID: req.Msg.GetConversationId(), Message: req.Msg.GetMessage(),
		DeviceID: req.Msg.GetDeviceId(), ModelConfigID: req.Msg.GetModelConfigId(),
		SinceEventID: req.Msg.GetSinceEventId(),
	})
	if err != nil {
		return response.MapDomainError(err)
	}
	for ev := range events {
		if err := stream.Send(toProtoEvent(ev)); err != nil {
			return err
		}
	}
	return nil
}

func toProtoEvent(ev aichatUC.Event) *devicepb.ChatEvent {
	out := &devicepb.ChatEvent{EventId: ev.ID}
	payload := ev.Payload
	switch ev.Type {
	case "text":
		out.Event = &devicepb.ChatEvent_Text{Text: &devicepb.TextDelta{Text: str(payload["text"])}}
	case "reasoning":
		out.Event = &devicepb.ChatEvent_Reasoning{Reasoning: &devicepb.ReasoningDelta{Text: str(payload["text"])}}
	case "tool_call":
		out.Event = &devicepb.ChatEvent_ToolCall{ToolCall: &devicepb.ToolCallEvent{Id: str(payload["id"]), Name: str(payload["name"]), ArgsJson: str(payload["args_json"])}}
	case "tool_result":
		out.Event = &devicepb.ChatEvent_ToolResult{ToolResult: &devicepb.ToolResultEvent{ToolCallId: str(payload["tool_call_id"]), IsError: boolVal(payload["is_error"]), Result: str(payload["result"]), DurationMs: int64Val(payload["duration_ms"])}}
	case "plan_step":
		out.Event = &devicepb.ChatEvent_PlanStep{PlanStep: &devicepb.PlanStepEvent{Step: str(payload["step"]), Status: str(payload["status"]), Detail: str(payload["detail"])}}
	case "terminal":
		out.Event = &devicepb.ChatEvent_Terminal{Terminal: &devicepb.TerminalOutputEvent{Command: str(payload["command"]), Output: str(payload["output"]), IsError: boolVal(payload["is_error"])}}
	case "approval":
		out.Event = &devicepb.ChatEvent_Approval{Approval: &devicepb.ApprovalRequestedEvent{ApprovalId: str(payload["approval_id"]), Summary: str(payload["summary"]), RestorePointId: str(payload["restore_point_id"])}}
	case "usage":
		out.Event = &devicepb.ChatEvent_Usage{Usage: &devicepb.UsageEvent{InputTokens: int32(int64Val(payload["input_tokens"])), OutputTokens: int32(int64Val(payload["output_tokens"]))}}
	case "done":
		out.Event = &devicepb.ChatEvent_Done{Done: &devicepb.DoneEvent{ConversationId: str(payload["conversation_id"]), LastEventId: int64Val(payload["last_event_id"])}}
	case "error":
		out.Event = &devicepb.ChatEvent_Error{Error: &devicepb.ErrorEvent{Message: str(payload["message"])}}
	default:
		out.Event = &devicepb.ChatEvent_Text{Text: &devicepb.TextDelta{Text: str(payload["text"])}}
	}
	return out
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func boolVal(v any) bool {
	b, _ := v.(bool)
	return b
}

func int64Val(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case int32:
		return int64(n)
	default:
		return 0
	}
}
