package skill

import "github.com/quixiq/polyglot/internal/domain/skill"

// --- Resource Management ---

func (u *ManageSkillUseCase) ListResources(skillName string) ([]skill.SkillResource, error) {
	return u.store.ListResources(skillName)
}

func (u *ManageSkillUseCase) GetResource(skillName, path string) (*skill.ResourceContent, *skill.SkillResource, error) {
	return u.store.ReadResource(skillName, path)
}

func (u *ManageSkillUseCase) SaveResource(skillName, path string, data []byte) error {
	return u.store.WriteResource(skillName, path, data)
}

func (u *ManageSkillUseCase) DeleteResource(skillName, path string) error {
	return u.store.DeleteResource(skillName, path)
}
