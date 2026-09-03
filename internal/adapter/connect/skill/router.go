package skill

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
)

// NewSkillServiceHandler mounts SkillService Connect handlers.
func NewSkillServiceHandler(
	skillUC *skillUC.ManageSkillUseCase,
) (string, http.Handler) {
	handler := NewSkillConnectHandler(skillUC)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.SkillService"

	mux.Handle("/"+serviceName+"/ListSkills", connect.NewUnaryHandler("/"+serviceName+"/ListSkills", handler.ListSkills, opts...))
	mux.Handle("/"+serviceName+"/GetSkill", connect.NewUnaryHandler("/"+serviceName+"/GetSkill", handler.GetSkill, opts...))
	mux.Handle("/"+serviceName+"/CreateSkill", connect.NewUnaryHandler("/"+serviceName+"/CreateSkill", handler.CreateSkill, opts...))
	mux.Handle("/"+serviceName+"/UpdateSkill", connect.NewUnaryHandler("/"+serviceName+"/UpdateSkill", handler.UpdateSkill, opts...))
	mux.Handle("/"+serviceName+"/DeleteSkill", connect.NewUnaryHandler("/"+serviceName+"/DeleteSkill", handler.DeleteSkill, opts...))
	mux.Handle("/"+serviceName+"/ExportSkill", connect.NewUnaryHandler("/"+serviceName+"/ExportSkill", handler.ExportSkill, opts...))
	mux.Handle("/"+serviceName+"/ImportSkill", connect.NewUnaryHandler("/"+serviceName+"/ImportSkill", handler.ImportSkill, opts...))
	mux.Handle("/"+serviceName+"/ListResources", connect.NewUnaryHandler("/"+serviceName+"/ListResources", handler.ListResources, opts...))
	mux.Handle("/"+serviceName+"/GetResource", connect.NewUnaryHandler("/"+serviceName+"/GetResource", handler.GetResource, opts...))
	mux.Handle("/"+serviceName+"/SaveResource", connect.NewUnaryHandler("/"+serviceName+"/SaveResource", handler.SaveResource, opts...))
	mux.Handle("/"+serviceName+"/DeleteResource", connect.NewUnaryHandler("/"+serviceName+"/DeleteResource", handler.DeleteResource, opts...))
	mux.Handle("/"+serviceName+"/ListGitRepos", connect.NewUnaryHandler("/"+serviceName+"/ListGitRepos", handler.ListGitRepos, opts...))
	mux.Handle("/"+serviceName+"/AddGitRepo", connect.NewUnaryHandler("/"+serviceName+"/AddGitRepo", handler.AddGitRepo, opts...))
	mux.Handle("/"+serviceName+"/UpdateGitRepo", connect.NewUnaryHandler("/"+serviceName+"/UpdateGitRepo", handler.UpdateGitRepo, opts...))
	mux.Handle("/"+serviceName+"/DeleteGitRepo", connect.NewUnaryHandler("/"+serviceName+"/DeleteGitRepo", handler.DeleteGitRepo, opts...))
	mux.Handle("/"+serviceName+"/SyncGitRepo", connect.NewUnaryHandler("/"+serviceName+"/SyncGitRepo", handler.SyncGitRepo, opts...))
	mux.Handle("/"+serviceName+"/ToggleGitRepo", connect.NewUnaryHandler("/"+serviceName+"/ToggleGitRepo", handler.ToggleGitRepo, opts...))
	mux.Handle("/"+serviceName+"/ToggleSkill", connect.NewUnaryHandler("/"+serviceName+"/ToggleSkill", handler.ToggleSkill, opts...))
	mux.Handle("/"+serviceName+"/GetGlobalPrompt", connect.NewUnaryHandler("/"+serviceName+"/GetGlobalPrompt", handler.GetGlobalPrompt, opts...))
	mux.Handle("/"+serviceName+"/SaveGlobalPrompt", connect.NewUnaryHandler("/"+serviceName+"/SaveGlobalPrompt", handler.SaveGlobalPrompt, opts...))

	return "/" + serviceName + "/", mux
}
