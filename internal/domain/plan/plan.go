package plan

import "context"

// New creates a placeholder Plan entity (ISP service plan).
// Named New, not NewPlan — CLAUDE.md §2.1 (avoid package name stutter).
// The package itself is named "plan", not "package" — "package" is a
// reserved Go keyword. See CLAUDE.md §1.1 catatan penamaan and
// NetOps-Architecture.md §7.2 (table named `plans`, not `packages`, for
// the same reason).
// TODO: implement per NetOps-Architecture.md business domain rules.
func New(ctx context.Context) error {
	return nil
}
