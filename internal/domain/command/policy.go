package command

// Decision is the policy outcome for a classified Command.
type Decision int

const (
	// DecisionAutoApprove permits a command without extra approval.
	DecisionAutoApprove Decision = iota
	// DecisionRequireApproval pauses a command pending approval.
	DecisionRequireApproval
	// DecisionDeny rejects a command.
	DecisionDeny
)

// Decide returns the policy decision for a classified Command.
// Pure logic, no I/O — does not take context.Context per CLAUDE.md §5.
// Precedence: hard deny > mutation (approval) > read (auto).
func Decide(c Classification) Decision {
	if c.IsDenied() {
		return DecisionDeny
	}
	if c.Mode == ModeMutation {
		return DecisionRequireApproval
	}
	return DecisionAutoApprove
}
