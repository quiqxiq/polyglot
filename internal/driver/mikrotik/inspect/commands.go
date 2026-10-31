package inspect

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// Field names returned by /console/inspect.
const (
	proplistChild      = ".id,name,node-type,type,symbol,symbol-type,text"
	proplistSyntax     = "type,symbol,symbol-type,nested,nonorm,text"
	proplistCompletion = "type,completion,style,offset,preference,show,text"
	proplistHighlight  = "type,text,start,end"
)

// NewChildCommand builds /console/inspect request=child.
func NewChildCommand(path string) command.Command {
	return command.Command{
		Raw: "/console/inspect",
		Args: map[string]string{
			"request":   "child",
			"path":      path,
			".proplist": proplistChild,
		},
	}
}

// NewSyntaxCommand builds /console/inspect request=syntax.
func NewSyntaxCommand(path string) command.Command {
	return command.Command{
		Raw: "/console/inspect",
		Args: map[string]string{
			"request":   "syntax",
			"path":      path,
			".proplist": proplistSyntax,
		},
	}
}

// NewCompletionCommand builds /console/inspect request=completion.
func NewCompletionCommand(input, path string) command.Command {
	args := map[string]string{
		"request":   "completion",
		"input":     input,
		".proplist": proplistCompletion,
	}
	if strings.TrimSpace(path) != "" {
		args["path"] = path
	}
	return command.Command{Raw: "/console/inspect", Args: args}
}

// NewHighlightCommand builds /console/inspect request=highlight. This is the
// gate used for pre-approval syntax validation.
func NewHighlightCommand(input string) command.Command {
	return command.Command{
		Raw: "/console/inspect",
		Args: map[string]string{
			"request":   "highlight",
			"input":     input,
			".proplist": proplistHighlight,
		},
	}
}
