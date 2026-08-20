package bot

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/skill"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
	"github.com/quixiq/polyglot/pkg/response"
)

type SkillConnectHandler struct {
	skillUC *skillUC.ManageSkillUseCase
}

func NewSkillConnectHandler(skillUC *skillUC.ManageSkillUseCase) *SkillConnectHandler {
	return &SkillConnectHandler{
		skillUC: skillUC,
	}
}

func (h *SkillConnectHandler) ListSkills(ctx context.Context, req *connect.Request[devicepb.ListSkillsRequest]) (*connect.Response[devicepb.ListSkillsResponse], error) {
	skills, err := h.skillUC.ListSkills(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	var protoSkills []*devicepb.SkillItem
	for _, s := range skills {
		protoSkills = append(protoSkills, ToProtoSkill(&s))
	}

	return connect.NewResponse(&devicepb.ListSkillsResponse{
		Skills: protoSkills,
	}), nil
}

func (h *SkillConnectHandler) GetSkill(ctx context.Context, req *connect.Request[devicepb.GetSkillRequest]) (*connect.Response[devicepb.GetSkillResponse], error) {
	sk, err := h.skillUC.GetSkill(ctx, req.Msg.Slug)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetSkillResponse{
		Skill: ToProtoSkill(sk),
	}), nil
}

func (h *SkillConnectHandler) CreateSkill(ctx context.Context, req *connect.Request[devicepb.CreateSkillRequest]) (*connect.Response[devicepb.CreateSkillResponse], error) {
	sk, err := h.skillUC.CreateSkill(ctx, req.Msg.Slug, req.Msg.Name, req.Msg.Description)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.CreateSkillResponse{
		Skill: ToProtoSkill(sk),
	}), nil
}

func (h *SkillConnectHandler) SaveSkillFile(ctx context.Context, req *connect.Request[devicepb.SaveSkillFileRequest]) (*connect.Response[devicepb.SaveSkillFileResponse], error) {
	f, err := h.skillUC.SaveSkillFile(ctx, req.Msg.Slug, req.Msg.FilePath, req.Msg.Content, req.Msg.IsReference)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.SaveSkillFileResponse{
		File: ToProtoSkillFile(f),
	}), nil
}

func (h *SkillConnectHandler) DeleteSkill(ctx context.Context, req *connect.Request[devicepb.DeleteSkillRequest]) (*connect.Response[devicepb.DeleteSkillResponse], error) {
	if err := h.skillUC.DeleteSkill(ctx, uint(req.Msg.Id), req.Msg.Slug); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteSkillResponse{
		Success: true,
		Message: "Skill deleted successfully",
	}), nil
}

func (h *SkillConnectHandler) DeleteSkillFile(ctx context.Context, req *connect.Request[devicepb.DeleteSkillFileRequest]) (*connect.Response[devicepb.DeleteSkillFileResponse], error) {
	if err := h.skillUC.DeleteSkillFile(ctx, req.Msg.Slug, uint(req.Msg.FileId), req.Msg.FilePath); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteSkillFileResponse{
		Success: true,
		Message: "Skill file deleted successfully",
	}), nil
}

func (h *SkillConnectHandler) ToggleSkill(ctx context.Context, req *connect.Request[devicepb.ToggleSkillRequest]) (*connect.Response[devicepb.ToggleSkillResponse], error) {
	if err := h.skillUC.ToggleSkillEnabled(ctx, req.Msg.Slug, req.Msg.Enabled); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ToggleSkillResponse{
		Success: true,
		Message: fmt.Sprintf("Skill status updated to %v", req.Msg.Enabled),
	}), nil
}

func (h *SkillConnectHandler) SyncSkillsFromDisk(ctx context.Context, req *connect.Request[devicepb.SyncSkillsFromDiskRequest]) (*connect.Response[devicepb.SyncSkillsFromDiskResponse], error) {
	count, err := h.skillUC.SyncFromDisk(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.SyncSkillsFromDiskResponse{
		SyncedCount: int32(count),
		Message:     fmt.Sprintf("Successfully synced %d skills from disk", count),
	}), nil
}

func (h *SkillConnectHandler) GetGlobalPrompt(ctx context.Context, req *connect.Request[devicepb.GetGlobalPromptRequest]) (*connect.Response[devicepb.GetGlobalPromptResponse], error) {
	content, err := h.skillUC.GetGlobalSystemPrompt(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetGlobalPromptResponse{
		Content: content,
	}), nil
}

func (h *SkillConnectHandler) SaveGlobalPrompt(ctx context.Context, req *connect.Request[devicepb.SaveGlobalPromptRequest]) (*connect.Response[devicepb.SaveGlobalPromptResponse], error) {
	if err := h.skillUC.SaveGlobalSystemPrompt(ctx, req.Msg.Content); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.SaveGlobalPromptResponse{
		Success: true,
		Message: "Global system prompt saved successfully",
	}), nil
}

func ToProtoSkill(s *skill.Skill) *devicepb.SkillItem {
	if s == nil {
		return nil
	}
	var protoFiles []*devicepb.SkillFileItem
	for _, f := range s.Files {
		protoFiles = append(protoFiles, ToProtoSkillFile(&f))
	}
	return &devicepb.SkillItem{
		Id:          uint32(s.ID),
		Slug:        s.Slug,
		Name:        s.Name,
		Description: s.Description,
		IsEnabled:   s.IsEnabled,
		Files:       protoFiles,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
}

func ToProtoSkillFile(f *skill.SkillFile) *devicepb.SkillFileItem {
	if f == nil {
		return nil
	}
	return &devicepb.SkillFileItem{
		Id:          uint32(f.ID),
		SkillId:     uint32(f.SkillID),
		Name:        f.Name,
		FilePath:    f.FilePath,
		Content:     f.Content,
		IsReference: f.IsReference,
		UpdatedAt:   f.UpdatedAt.Format(time.RFC3339),
	}
}
