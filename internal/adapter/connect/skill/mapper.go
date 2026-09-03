package skill

import (
	"time"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/skill"
)

// ToProtoSkill maps domain Skill to proto Skill message.
func ToProtoSkill(s *skill.Skill) *devicepb.Skill {
	if s == nil {
		return &devicepb.Skill{}
	}
	return &devicepb.Skill{
		Id:            s.ID,
		Name:          s.Name,
		Description:   s.Description,
		Content:       s.Content,
		License:       s.License,
		Compatibility: s.Compatibility,
		AllowedTools:  s.AllowedTools,
		Metadata:      s.Metadata,
		ReadOnly:      s.ReadOnly,
		SourceType:    s.SourceType,
		SourceUrl:     s.SourceURL,
		CreatedAt:     s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     s.UpdatedAt.Format(time.RFC3339),
	}
}

// ToProtoSkillList converts a slice of domain Skills into proto messages.
func ToProtoSkillList(skills []skill.Skill) []*devicepb.Skill {
	items := make([]*devicepb.Skill, len(skills))
	for i := range skills {
		items[i] = ToProtoSkill(&skills[i])
	}
	return items
}

// ToProtoSkillResource maps domain SkillResource to proto SkillResource message.
func ToProtoSkillResource(r *skill.SkillResource) *devicepb.SkillResource {
	if r == nil {
		return &devicepb.SkillResource{}
	}
	return &devicepb.SkillResource{
		Path:     r.Path,
		Name:     r.Name,
		Type:     r.Type,
		Size:     r.Size,
		MimeType: r.MimeType,
		Readable: r.Readable,
		Modified: r.Modified.Format(time.RFC3339),
	}
}

// ToProtoSkillResourceList converts a slice of domain SkillResources into proto messages.
func ToProtoSkillResourceList(resources []skill.SkillResource) []*devicepb.SkillResource {
	items := make([]*devicepb.SkillResource, len(resources))
	for i := range resources {
		items[i] = ToProtoSkillResource(&resources[i])
	}
	return items
}

// ToProtoResourceContent maps domain ResourceContent to proto ResourceContent message.
func ToProtoResourceContent(c *skill.ResourceContent) *devicepb.ResourceContent {
	if c == nil {
		return &devicepb.ResourceContent{}
	}
	return &devicepb.ResourceContent{
		Content:  c.Content,
		Encoding: c.Encoding,
		MimeType: c.MimeType,
		Size:     c.Size,
	}
}

// ToProtoGitRepoInfo maps domain GitRepoInfo to proto GitRepoInfo message.
func ToProtoGitRepoInfo(g *skill.GitRepoInfo) *devicepb.GitRepoInfo {
	if g == nil {
		return &devicepb.GitRepoInfo{}
	}
	return &devicepb.GitRepoInfo{
		Id:           g.ID,
		Url:          g.URL,
		Name:         g.Name,
		Enabled:      g.Enabled,
		LastSyncedAt: g.LastSyncedAt.Format(time.RFC3339),
	}
}

// ToProtoGitRepoInfoList converts a slice of domain GitRepoInfos into proto messages.
func ToProtoGitRepoInfoList(repos []skill.GitRepoInfo) []*devicepb.GitRepoInfo {
	items := make([]*devicepb.GitRepoInfo, len(repos))
	for i := range repos {
		items[i] = ToProtoGitRepoInfo(&repos[i])
	}
	return items
}
