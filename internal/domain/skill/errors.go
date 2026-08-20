package skill

import "errors"

var (
	ErrSkillNotFound         = errors.New("skill not found")
	ErrSkillAlreadyExists     = errors.New("skill already exists")
	ErrInvalidSkillName      = errors.New("invalid skill name: only lowercase alphanumeric and dashes allowed")
	ErrInvalidSlug           = errors.New("invalid skill slug: only lowercase alphanumeric and dashes allowed")
	ErrReadOnlySkill         = errors.New("cannot modify read-only skill from git repository")
	ErrResourceNotFound      = errors.New("skill resource not found")
	ErrInvalidResourcePath   = errors.New("invalid resource path: must be a safe relative path")
	ErrInvalidGitURL         = errors.New("invalid git URL: must start with http://, https://, or git@")
	ErrGitRepoNotFound       = errors.New("git repository not found")
	ErrGitRepoAlreadyExists  = errors.New("git repository already exists")
	ErrGlobalPromptNotFound  = errors.New("global system prompt not found")
)
