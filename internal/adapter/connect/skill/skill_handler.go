package skill

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
	"github.com/quixiq/polyglot/pkg/response"
)

// SkillConnectHandler implements the SkillService ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type SkillConnectHandler struct {
	skillUC *skillUC.ManageSkillUseCase
}

// NewSkillConnectHandler constructs a SkillConnectHandler.
func NewSkillConnectHandler(skillUC *skillUC.ManageSkillUseCase) *SkillConnectHandler {
	return &SkillConnectHandler{
		skillUC: skillUC,
	}
}

// ListSkills returns all skills registered in the system.
func (h *SkillConnectHandler) ListSkills(ctx context.Context, req *connect.Request[devicepb.ListSkillsRequest]) (*connect.Response[devicepb.ListSkillsResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	skills, err := h.skillUC.ListSkills(ctx, req.Msg.UserId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListSkillsResponse{
		Skills: ToProtoSkillList(skills),
	}), nil
}

// GetSkill returns details of a single skill by ID.
func (h *SkillConnectHandler) GetSkill(ctx context.Context, req *connect.Request[devicepb.GetSkillRequest]) (*connect.Response[devicepb.GetSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	sk, err := h.skillUC.GetSkill(ctx, req.Msg.UserId, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetSkillResponse{
		Skill: ToProtoSkill(sk),
	}), nil
}

// CreateSkill creates a new skill with the given definition.
func (h *SkillConnectHandler) CreateSkill(ctx context.Context, req *connect.Request[devicepb.CreateSkillRequest]) (*connect.Response[devicepb.CreateSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
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
		Skill: ToProtoSkill(sk),
	}), nil
}

// UpdateSkill updates an existing skill's configuration and content.
func (h *SkillConnectHandler) UpdateSkill(ctx context.Context, req *connect.Request[devicepb.UpdateSkillRequest]) (*connect.Response[devicepb.UpdateSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
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
		Skill: ToProtoSkill(sk),
	}), nil
}

// DeleteSkill deletes a skill by ID.
func (h *SkillConnectHandler) DeleteSkill(ctx context.Context, req *connect.Request[devicepb.DeleteSkillRequest]) (*connect.Response[devicepb.DeleteSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.DeleteSkill(ctx, req.Msg.UserId, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteSkillResponse{
		Success: true,
		Message: "Skill deleted successfully",
	}), nil
}

// ExportSkill exports a skill as a zip archive.
func (h *SkillConnectHandler) ExportSkill(_ context.Context, req *connect.Request[devicepb.ExportSkillRequest]) (*connect.Response[devicepb.ExportSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	data, err := h.skillUC.ExportSkill(req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ExportSkillResponse{
		Archive:  data,
		Filename: req.Msg.Id + ".zip",
	}), nil
}

// ImportSkill imports a skill from a zip archive.
func (h *SkillConnectHandler) ImportSkill(ctx context.Context, req *connect.Request[devicepb.ImportSkillRequest]) (*connect.Response[devicepb.ImportSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	sk, err := h.skillUC.ImportSkill(ctx, req.Msg.UserId, req.Msg.Archive)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ImportSkillResponse{
		Skill: ToProtoSkill(sk),
	}), nil
}

// ListResources returns the static resource files for a skill.
func (h *SkillConnectHandler) ListResources(_ context.Context, req *connect.Request[devicepb.ListResourcesRequest]) (*connect.Response[devicepb.ListResourcesResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	res, err := h.skillUC.ListResources(req.Msg.SkillId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListResourcesResponse{
		Resources: ToProtoSkillResourceList(res),
	}), nil
}

// GetResource returns the content of a specific resource file.
func (h *SkillConnectHandler) GetResource(_ context.Context, req *connect.Request[devicepb.GetResourceRequest]) (*connect.Response[devicepb.GetResourceResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	content, meta, err := h.skillUC.GetResource(req.Msg.SkillId, req.Msg.Path)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetResourceResponse{
		Content:  ToProtoResourceContent(content),
		Metadata: ToProtoSkillResource(meta),
	}), nil
}

// SaveResource creates or updates a resource file within a skill.
func (h *SkillConnectHandler) SaveResource(_ context.Context, req *connect.Request[devicepb.SaveResourceRequest]) (*connect.Response[devicepb.SaveResourceResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.SaveResource(req.Msg.SkillId, req.Msg.Path, req.Msg.Data)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.SaveResourceResponse{
		Success: true,
		Message: "Resource saved successfully",
	}), nil
}

// DeleteResource deletes a resource file from a skill.
func (h *SkillConnectHandler) DeleteResource(_ context.Context, req *connect.Request[devicepb.DeleteResourceRequest]) (*connect.Response[devicepb.DeleteResourceResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.DeleteResource(req.Msg.SkillId, req.Msg.Path)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteResourceResponse{
		Success: true,
		Message: "Resource deleted successfully",
	}), nil
}

// ListGitRepos lists all registered skill Git repositories.
func (h *SkillConnectHandler) ListGitRepos(_ context.Context, _ *connect.Request[devicepb.ListGitReposRequest]) (*connect.Response[devicepb.ListGitReposResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	repos, err := h.skillUC.ListGitRepos()
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListGitReposResponse{
		Repos: ToProtoGitRepoInfoList(repos),
	}), nil
}

// AddGitRepo adds a new skill Git repository.
func (h *SkillConnectHandler) AddGitRepo(ctx context.Context, req *connect.Request[devicepb.AddGitRepoRequest]) (*connect.Response[devicepb.AddGitRepoResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	repo, err := h.skillUC.AddGitRepo(ctx, req.Msg.Url)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.AddGitRepoResponse{
		Repo: ToProtoGitRepoInfo(repo),
	}), nil
}

// UpdateGitRepo updates an existing skill Git repository configuration.
func (h *SkillConnectHandler) UpdateGitRepo(_ context.Context, req *connect.Request[devicepb.UpdateGitRepoRequest]) (*connect.Response[devicepb.UpdateGitRepoResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	repo, err := h.skillUC.UpdateGitRepo(req.Msg.Id, req.Msg.Url, req.Msg.Enabled)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdateGitRepoResponse{
		Repo: ToProtoGitRepoInfo(repo),
	}), nil
}

// DeleteGitRepo removes a skill Git repository.
func (h *SkillConnectHandler) DeleteGitRepo(ctx context.Context, req *connect.Request[devicepb.DeleteGitRepoRequest]) (*connect.Response[devicepb.DeleteGitRepoResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.DeleteGitRepo(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteGitRepoResponse{
		Success: true,
		Message: "Git repository deleted successfully",
	}), nil
}

// SyncGitRepo synchronizes skills from a Git repository.
func (h *SkillConnectHandler) SyncGitRepo(ctx context.Context, req *connect.Request[devicepb.SyncGitRepoRequest]) (*connect.Response[devicepb.SyncGitRepoResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.SyncGitRepo(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.SyncGitRepoResponse{
		Success: true,
		Message: "Git repository synced successfully",
	}), nil
}

// ToggleGitRepo enables or disables a Git repository.
func (h *SkillConnectHandler) ToggleGitRepo(_ context.Context, req *connect.Request[devicepb.ToggleGitRepoRequest]) (*connect.Response[devicepb.ToggleGitRepoResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	repo, err := h.skillUC.ToggleGitRepo(req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ToggleGitRepoResponse{
		Repo: ToProtoGitRepoInfo(repo),
	}), nil
}

// ToggleSkill enables or disables an individual skill.
func (h *SkillConnectHandler) ToggleSkill(ctx context.Context, req *connect.Request[devicepb.ToggleSkillRequest]) (*connect.Response[devicepb.ToggleSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.ToggleSkillEnabled(ctx, req.Msg.UserId, req.Msg.Id, req.Msg.Enabled)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ToggleSkillResponse{
		Success: true,
		Message: "Skill toggled successfully",
	}), nil
}

// GetGlobalPrompt returns the system-wide global prompt.
func (h *SkillConnectHandler) GetGlobalPrompt(ctx context.Context, _ *connect.Request[devicepb.GetGlobalPromptRequest]) (*connect.Response[devicepb.GetGlobalPromptResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	content, err := h.skillUC.GetGlobalSystemPrompt(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetGlobalPromptResponse{
		Content: content,
	}), nil
}

// SaveGlobalPrompt saves the system-wide global prompt.
func (h *SkillConnectHandler) SaveGlobalPrompt(ctx context.Context, req *connect.Request[devicepb.SaveGlobalPromptRequest]) (*connect.Response[devicepb.SaveGlobalPromptResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.SaveGlobalSystemPrompt(ctx, req.Msg.Content)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.SaveGlobalPromptResponse{
		Success: true,
		Message: "Global prompt saved successfully",
	}), nil
}
