package skill

import (
	"time"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/skill"
)

// ToProtoSkill maps domain Skill to proto Skill message.
func ToProtoSkill(s *skill.Skill) *polyglotv1.Skill {
	if s == nil {
		return &polyglotv1.Skill{}
	}
	return &polyglotv1.Skill{
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
func ToProtoSkillList(skills []skill.Skill) []*polyglotv1.Skill {
	items := make([]*polyglotv1.Skill, len(skills))
	for i := range skills {
		items[i] = ToProtoSkill(&skills[i])
	}
	return items
}

// ToProtoSkillResource maps domain SkillResource to proto SkillResource message.
func ToProtoSkillResource(r *skill.SkillResource) *polyglotv1.SkillResource {
	if r == nil {
		return &polyglotv1.SkillResource{}
	}
	return &polyglotv1.SkillResource{
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
func ToProtoSkillResourceList(resources []skill.SkillResource) []*polyglotv1.SkillResource {
	items := make([]*polyglotv1.SkillResource, len(resources))
	for i := range resources {
		items[i] = ToProtoSkillResource(&resources[i])
	}
	return items
}

// ToProtoResourceContent maps domain ResourceContent to proto ResourceContent message.
func ToProtoResourceContent(c *skill.ResourceContent) *polyglotv1.ResourceContent {
	if c == nil {
		return &polyglotv1.ResourceContent{}
	}
	return &polyglotv1.ResourceContent{
		Content:  c.Content,
		Encoding: c.Encoding,
		MimeType: c.MimeType,
		Size:     c.Size,
	}
}

// ToProtoGitRepoInfo maps domain GitRepoInfo to proto GitRepoInfo message.
func ToProtoGitRepoInfo(g *skill.GitRepoInfo) *polyglotv1.GitRepoInfo {
	if g == nil {
		return &polyglotv1.GitRepoInfo{}
	}
	return &polyglotv1.GitRepoInfo{
		Id:           g.ID,
		Url:          g.URL,
		Name:         g.Name,
		Enabled:      g.Enabled,
		LastSyncedAt: g.LastSyncedAt.Format(time.RFC3339),
	}
}

// ToProtoGitRepoInfoList converts a slice of domain GitRepoInfos into proto messages.
func ToProtoGitRepoInfoList(repos []skill.GitRepoInfo) []*polyglotv1.GitRepoInfo {
	items := make([]*polyglotv1.GitRepoInfo, len(repos))
	for i := range repos {
		items[i] = ToProtoGitRepoInfo(&repos[i])
	}
	return items
}
