package skill

import "errors"

var (
	ErrSkillNotFound        = errors.New("skill not found")
	ErrSkillAlreadyExists    = errors.New("skill with this slug already exists")
	ErrInvalidSlug          = errors.New("invalid skill slug: only lowercase alphanumeric and dashes allowed")
	ErrInvalidSkillName     = errors.New("skill name cannot be empty")
	ErrFileNotFound         = errors.New("skill file not found")
	ErrInvalidFileName     = errors.New("invalid file name")
	ErrGlobalPromptNotFound = errors.New("global system prompt not found")
)
