package command

// Decision is the policy outcome for a classified Command.
type Decision int

const (
	DecisionAutoApprove Decision = iota
	DecisionRequireApproval
	DecisionDeny
)

// Decide returns the policy decision for a Command given its risk Class.
// Pure logic, no I/O — does not take context.Context per CLAUDE.md §5.
// TODO: incorporate role/permission context per Polyglot-Architecture.md §6.1
// (command_policy table: allow/deny/ask per command pattern, per vendor).
func Decide(class Class) Decision {
	if class == ClassDestructive {
		return DecisionRequireApproval
	}
	return DecisionAutoApprove
}
