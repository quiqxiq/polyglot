package skill

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the skill domain. Message convention:
// "skill: <description>", lowercase English.
var (
	ErrSkillNotFound        = fault.New(fault.KindNotFound, "skill: not found")
	ErrSkillAlreadyExists   = fault.New(fault.KindAlreadyExists, "skill: already exists")
	ErrInvalidSkillName     = fault.New(fault.KindInvalidInput, "skill: invalid name: only lowercase alphanumeric and dashes allowed")
	ErrInvalidSlug          = fault.New(fault.KindInvalidInput, "skill: invalid slug: only lowercase alphanumeric and dashes allowed")
	ErrReadOnlySkill        = fault.New(fault.KindFailedPrecondition, "skill: cannot modify read-only skill from git repository")
	ErrResourceNotFound     = fault.New(fault.KindNotFound, "skill: resource not found")
	ErrInvalidResourcePath  = fault.New(fault.KindInvalidInput, "skill: invalid resource path: must be a safe relative path")
	ErrInvalidGitURL        = fault.New(fault.KindInvalidInput, "skill: invalid git URL: must start with http://, https://, or git@")
	ErrGitRepoNotFound      = fault.New(fault.KindNotFound, "skill: git repository not found")
	ErrGitRepoAlreadyExists = fault.New(fault.KindAlreadyExists, "skill: git repository already exists")
	ErrGlobalPromptNotFound = fault.New(fault.KindNotFound, "skill: global system prompt not found")
)
