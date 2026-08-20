package bot

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
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
	skills, err := h.skillUC.ListSkills(ctx, req.Msg.UserId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListSkillsResponse{
		Skills: toProtoSkillList(skills),
	}), nil
}

func (h *SkillConnectHandler) GetSkill(ctx context.Context, req *connect.Request[devicepb.GetSkillRequest]) (*connect.Response[devicepb.GetSkillResponse], error) {
	sk, err := h.skillUC.GetSkill(ctx, req.Msg.UserId, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetSkillResponse{
		Skill: toProtoSkill(sk),
	}), nil
}

func (h *SkillConnectHandler) CreateSkill(ctx context.Context, req *connect.Request[devicepb.CreateSkillRequest]) (*connect.Response[devicepb.CreateSkillResponse], error) {
	sk, err := h.skillUC.CreateSkill(
		ctx,
		req.Msg.UserId,
		req.Msg.Name,
		req.Msg.Description,
		req.Msg.Content,
		req.Msg.License,
		req.Msg.Compatibility,
		req.Msg.AllowedTools,
		req.Msg.Metadata,
	)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.CreateSkillResponse{
		Skill: toProtoSkill(sk),
	}), nil
}

func (h *SkillConnectHandler) UpdateSkill(ctx context.Context, req *connect.Request[devicepb.UpdateSkillRequest]) (*connect.Response[devicepb.UpdateSkillResponse], error) {
	sk, err := h.skillUC.UpdateSkill(
		ctx,
		req.Msg.UserId,
		req.Msg.Id,
		req.Msg.Description,
		req.Msg.Content,
		req.Msg.License,
		req.Msg.Compatibility,
		req.Msg.AllowedTools,
		req.Msg.Metadata,
	)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdateSkillResponse{
		Skill: toProtoSkill(sk),
	}), nil
}

func (h *SkillConnectHandler) DeleteSkill(ctx context.Context, req *connect.Request[devicepb.DeleteSkillRequest]) (*connect.Response[devicepb.DeleteSkillResponse], error) {
	if err := h.skillUC.DeleteSkill(ctx, req.Msg.UserId, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteSkillResponse{
		Success: true,
		Message: "Skill deleted successfully",
	}), nil
}

func (h *SkillConnectHandler) ExportSkill(ctx context.Context, req *connect.Request[devicepb.ExportSkillRequest]) (*connect.Response[devicepb.ExportSkillResponse], error) {
	archiveBytes, err := h.skillUC.ExportSkill(req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ExportSkillResponse{
		Archive:  archiveBytes,
		Filename: fmt.Sprintf("%s.zip", req.Msg.Id),
	}), nil
}

func (h *SkillConnectHandler) ImportSkill(ctx context.Context, req *connect.Request[devicepb.ImportSkillRequest]) (*connect.Response[devicepb.ImportSkillResponse], error) {
	sk, err := h.skillUC.ImportSkill(ctx, req.Msg.UserId, req.Msg.Archive)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ImportSkillResponse{
		Skill: toProtoSkill(sk),
	}), nil
}

func (h *SkillConnectHandler) ListResources(ctx context.Context, req *connect.Request[devicepb.ListResourcesRequest]) (*connect.Response[devicepb.ListResourcesResponse], error) {
	resources, err := h.skillUC.ListResources(req.Msg.SkillId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListResourcesResponse{
		Resources: toProtoSkillResourceList(resources),
	}), nil
}

func (h *SkillConnectHandler) GetResource(ctx context.Context, req *connect.Request[devicepb.GetResourceRequest]) (*connect.Response[devicepb.GetResourceResponse], error) {
	content, meta, err := h.skillUC.GetResource(req.Msg.SkillId, req.Msg.Path)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetResourceResponse{
		Content:  toProtoResourceContent(content),
		Metadata: toProtoSkillResource(meta),
	}), nil
}

func (h *SkillConnectHandler) SaveResource(ctx context.Context, req *connect.Request[devicepb.SaveResourceRequest]) (*connect.Response[devicepb.SaveResourceResponse], error) {
	if err := h.skillUC.SaveResource(req.Msg.SkillId, req.Msg.Path, req.Msg.Data); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.SaveResourceResponse{
		Success: true,
		Message: "Resource saved successfully",
	}), nil
}

func (h *SkillConnectHandler) DeleteResource(ctx context.Context, req *connect.Request[devicepb.DeleteResourceRequest]) (*connect.Response[devicepb.DeleteResourceResponse], error) {
	if err := h.skillUC.DeleteResource(req.Msg.SkillId, req.Msg.Path); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteResourceResponse{
		Success: true,
		Message: "Resource deleted successfully",
	}), nil
}

func (h *SkillConnectHandler) ListGitRepos(ctx context.Context, req *connect.Request[devicepb.ListGitReposRequest]) (*connect.Response[devicepb.ListGitReposResponse], error) {
	repos, err := h.skillUC.ListGitRepos()
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListGitReposResponse{
		Repos: toProtoGitRepoInfoList(repos),
	}), nil
}

func (h *SkillConnectHandler) AddGitRepo(ctx context.Context, req *connect.Request[devicepb.AddGitRepoRequest]) (*connect.Response[devicepb.AddGitRepoResponse], error) {
	repo, err := h.skillUC.AddGitRepo(ctx, req.Msg.Url)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.AddGitRepoResponse{
		Repo: toProtoGitRepoInfo(repo),
	}), nil
}

func (h *SkillConnectHandler) UpdateGitRepo(ctx context.Context, req *connect.Request[devicepb.UpdateGitRepoRequest]) (*connect.Response[devicepb.UpdateGitRepoResponse], error) {
	repo, err := h.skillUC.UpdateGitRepo(req.Msg.Id, req.Msg.Url, req.Msg.Enabled)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdateGitRepoResponse{
		Repo: toProtoGitRepoInfo(repo),
	}), nil
}

func (h *SkillConnectHandler) DeleteGitRepo(ctx context.Context, req *connect.Request[devicepb.DeleteGitRepoRequest]) (*connect.Response[devicepb.DeleteGitRepoResponse], error) {
	if err := h.skillUC.DeleteGitRepo(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteGitRepoResponse{
		Success: true,
		Message: "Git repo deleted successfully",
	}), nil
}

func (h *SkillConnectHandler) SyncGitRepo(ctx context.Context, req *connect.Request[devicepb.SyncGitRepoRequest]) (*connect.Response[devicepb.SyncGitRepoResponse], error) {
	if err := h.skillUC.SyncGitRepo(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.SyncGitRepoResponse{
		Success: true,
		Message: "Git repo sync initiated successfully",
	}), nil
}

func (h *SkillConnectHandler) ToggleGitRepo(ctx context.Context, req *connect.Request[devicepb.ToggleGitRepoRequest]) (*connect.Response[devicepb.ToggleGitRepoResponse], error) {
	repo, err := h.skillUC.ToggleGitRepo(req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ToggleGitRepoResponse{
		Repo: toProtoGitRepoInfo(repo),
	}), nil
}

func (h *SkillConnectHandler) ToggleSkill(ctx context.Context, req *connect.Request[devicepb.ToggleSkillRequest]) (*connect.Response[devicepb.ToggleSkillResponse], error) {
	if err := h.skillUC.ToggleSkillEnabled(ctx, req.Msg.UserId, req.Msg.Id, req.Msg.Enabled); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ToggleSkillResponse{
		Success: true,
		Message: fmt.Sprintf("Skill status updated to %v", req.Msg.Enabled),
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
