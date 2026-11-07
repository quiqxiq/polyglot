package skill

import (
	"context"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
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
func (h *SkillConnectHandler) ListSkills(ctx context.Context, req *connect.Request[polyglotv1.ListSkillsRequest]) (*connect.Response[polyglotv1.ListSkillsResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	skills, err := h.skillUC.ListSkills(ctx, req.Msg.UserId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ListSkillsResponse{
		Skills: ToProtoSkillList(skills),
	}), nil
}

// GetSkill returns details of a single skill by ID.
func (h *SkillConnectHandler) GetSkill(ctx context.Context, req *connect.Request[polyglotv1.GetSkillRequest]) (*connect.Response[polyglotv1.GetSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	sk, err := h.skillUC.GetSkill(ctx, req.Msg.UserId, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.GetSkillResponse{
		Skill: ToProtoSkill(sk),
	}), nil
}

// CreateSkill creates a new skill with the given definition.
func (h *SkillConnectHandler) CreateSkill(ctx context.Context, req *connect.Request[polyglotv1.CreateSkillRequest]) (*connect.Response[polyglotv1.CreateSkillResponse], error) {
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

	return connect.NewResponse(&polyglotv1.CreateSkillResponse{
		Skill: ToProtoSkill(sk),
	}), nil
}

// UpdateSkill updates an existing skill's configuration and content.
func (h *SkillConnectHandler) UpdateSkill(ctx context.Context, req *connect.Request[polyglotv1.UpdateSkillRequest]) (*connect.Response[polyglotv1.UpdateSkillResponse], error) {
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

	return connect.NewResponse(&polyglotv1.UpdateSkillResponse{
		Skill: ToProtoSkill(sk),
	}), nil
}

// DeleteSkill deletes a skill by ID.
func (h *SkillConnectHandler) DeleteSkill(ctx context.Context, req *connect.Request[polyglotv1.DeleteSkillRequest]) (*connect.Response[polyglotv1.DeleteSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.DeleteSkill(ctx, req.Msg.UserId, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.DeleteSkillResponse{
		Success: true,
		Message: "Skill deleted successfully",
	}), nil
}

// ExportSkill exports a skill as a zip archive.
func (h *SkillConnectHandler) ExportSkill(_ context.Context, req *connect.Request[polyglotv1.ExportSkillRequest]) (*connect.Response[polyglotv1.ExportSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	data, err := h.skillUC.ExportSkill(req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ExportSkillResponse{
		Archive:  data,
		Filename: req.Msg.Id + ".zip",
	}), nil
}

// ImportSkill imports a skill from a zip archive.
func (h *SkillConnectHandler) ImportSkill(ctx context.Context, req *connect.Request[polyglotv1.ImportSkillRequest]) (*connect.Response[polyglotv1.ImportSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	sk, err := h.skillUC.ImportSkill(ctx, req.Msg.UserId, req.Msg.Archive)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ImportSkillResponse{
		Skill: ToProtoSkill(sk),
	}), nil
}

// ListResources returns the static resource files for a skill.
func (h *SkillConnectHandler) ListResources(_ context.Context, req *connect.Request[polyglotv1.ListResourcesRequest]) (*connect.Response[polyglotv1.ListResourcesResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	res, err := h.skillUC.ListResources(req.Msg.SkillId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ListResourcesResponse{
		Resources: ToProtoSkillResourceList(res),
	}), nil
}

// GetResource returns the content of a specific resource file.
func (h *SkillConnectHandler) GetResource(_ context.Context, req *connect.Request[polyglotv1.GetResourceRequest]) (*connect.Response[polyglotv1.GetResourceResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	content, meta, err := h.skillUC.GetResource(req.Msg.SkillId, req.Msg.Path)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.GetResourceResponse{
		Content:  ToProtoResourceContent(content),
		Metadata: ToProtoSkillResource(meta),
	}), nil
}

// SaveResource creates or updates a resource file within a skill.
func (h *SkillConnectHandler) SaveResource(_ context.Context, req *connect.Request[polyglotv1.SaveResourceRequest]) (*connect.Response[polyglotv1.SaveResourceResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.SaveResource(req.Msg.SkillId, req.Msg.Path, req.Msg.Data)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.SaveResourceResponse{
		Success: true,
		Message: "Resource saved successfully",
	}), nil
}

// DeleteResource deletes a resource file from a skill.
func (h *SkillConnectHandler) DeleteResource(_ context.Context, req *connect.Request[polyglotv1.DeleteResourceRequest]) (*connect.Response[polyglotv1.DeleteResourceResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.DeleteResource(req.Msg.SkillId, req.Msg.Path)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.DeleteResourceResponse{
		Success: true,
		Message: "Resource deleted successfully",
	}), nil
}

// ListGitRepos lists all registered skill Git repositories.
func (h *SkillConnectHandler) ListGitRepos(_ context.Context, _ *connect.Request[polyglotv1.ListGitReposRequest]) (*connect.Response[polyglotv1.ListGitReposResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	repos, err := h.skillUC.ListGitRepos()
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ListGitReposResponse{
		Repos: ToProtoGitRepoInfoList(repos),
	}), nil
}

// AddGitRepo adds a new skill Git repository.
func (h *SkillConnectHandler) AddGitRepo(ctx context.Context, req *connect.Request[polyglotv1.AddGitRepoRequest]) (*connect.Response[polyglotv1.AddGitRepoResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	repo, err := h.skillUC.AddGitRepo(ctx, req.Msg.Url)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.AddGitRepoResponse{
		Repo: ToProtoGitRepoInfo(repo),
	}), nil
}

// UpdateGitRepo updates an existing skill Git repository configuration.
func (h *SkillConnectHandler) UpdateGitRepo(_ context.Context, req *connect.Request[polyglotv1.UpdateGitRepoRequest]) (*connect.Response[polyglotv1.UpdateGitRepoResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	repo, err := h.skillUC.UpdateGitRepo(req.Msg.Id, req.Msg.Url, req.Msg.Enabled)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.UpdateGitRepoResponse{
		Repo: ToProtoGitRepoInfo(repo),
	}), nil
}

// DeleteGitRepo removes a skill Git repository.
func (h *SkillConnectHandler) DeleteGitRepo(ctx context.Context, req *connect.Request[polyglotv1.DeleteGitRepoRequest]) (*connect.Response[polyglotv1.DeleteGitRepoResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.DeleteGitRepo(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.DeleteGitRepoResponse{
		Success: true,
		Message: "Git repository deleted successfully",
	}), nil
}

// SyncGitRepo synchronizes skills from a Git repository.
func (h *SkillConnectHandler) SyncGitRepo(ctx context.Context, req *connect.Request[polyglotv1.SyncGitRepoRequest]) (*connect.Response[polyglotv1.SyncGitRepoResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.SyncGitRepo(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.SyncGitRepoResponse{
		Success: true,
		Message: "Git repository synced successfully",
	}), nil
}

// ToggleGitRepo enables or disables a Git repository.
func (h *SkillConnectHandler) ToggleGitRepo(_ context.Context, req *connect.Request[polyglotv1.ToggleGitRepoRequest]) (*connect.Response[polyglotv1.ToggleGitRepoResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	repo, err := h.skillUC.ToggleGitRepo(req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ToggleGitRepoResponse{
		Repo: ToProtoGitRepoInfo(repo),
	}), nil
}

// ToggleSkill enables or disables an individual skill.
func (h *SkillConnectHandler) ToggleSkill(ctx context.Context, req *connect.Request[polyglotv1.ToggleSkillRequest]) (*connect.Response[polyglotv1.ToggleSkillResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.ToggleSkillEnabled(ctx, req.Msg.UserId, req.Msg.Id, req.Msg.Enabled)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ToggleSkillResponse{
		Success: true,
		Message: "Skill toggled successfully",
	}), nil
}

// GetGlobalPrompt returns the system-wide global prompt.
func (h *SkillConnectHandler) GetGlobalPrompt(ctx context.Context, _ *connect.Request[polyglotv1.GetGlobalPromptRequest]) (*connect.Response[polyglotv1.GetGlobalPromptResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	content, err := h.skillUC.GetGlobalSystemPrompt(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.GetGlobalPromptResponse{
		Content: content,
	}), nil
}

// SaveGlobalPrompt saves the system-wide global prompt.
func (h *SkillConnectHandler) SaveGlobalPrompt(ctx context.Context, req *connect.Request[polyglotv1.SaveGlobalPromptRequest]) (*connect.Response[polyglotv1.SaveGlobalPromptResponse], error) {
	if h.skillUC == nil {
		return nil, response.Unavailable("skill usecase unavailable")
	}
	err := h.skillUC.SaveGlobalSystemPrompt(ctx, req.Msg.Content)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.SaveGlobalPromptResponse{
		Success: true,
		Message: "Global prompt saved successfully",
	}), nil
}
