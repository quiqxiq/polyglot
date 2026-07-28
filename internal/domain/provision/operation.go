// Package provision defines abstract, vendor-neutral provisioning intents.
// A provisioning Operation describes WHAT should exist on a device (e.g. "list
// the PPPoE secrets", "create a PPPoE secret") without naming any vendor's
// concrete command syntax — that translation lives in each vendor's driver
// (see port.ProvisioningDriver and internal/driver/<vendor>/commands.go).
//
// This package is domain-pure per AGENTS.md §1.2: no I/O, no external imports,
// only data. It is intentionally separate from domain/command, whose
// command.Operation (a string type for parameterless ops like get_status) is a
// different, non-parameterized concern.
package provision

// Operation is an abstract, vendor-neutral provisioning intent. A vendor's
// port.ProvisioningDriver.TranslateProvision maps it to that vendor's native
// command sequence. The interface is sealed by the unexported marker method so
// only this package can define valid operations — a driver's exhaustive type
// switch can rely on the set being closed.
type Operation interface {
	isProvisionOperation()
}

// ListPPPSecrets requests the device's PPPoE secrets. Fields optionally
// restricts which columns the device returns (mapped by the driver to its
// native field-projection mechanism, e.g. RouterOS .proplist); an empty Fields
// means "return every field, raw".
type ListPPPSecrets struct {
	Fields []string
}

func (ListPPPSecrets) isProvisionOperation() {}

// ListPPPProfiles requests the device's PPPoE profiles (the rate-limit /
// address-pool templates a secret points at). Fields behaves exactly like
// ListPPPSecrets.Fields: empty means "return every field, raw".
type ListPPPProfiles struct {
	Fields []string
}

func (ListPPPProfiles) isProvisionOperation() {}

// ListActivePPP requests the sessions currently online (not the stored
// secrets) — useful to see who is connected right now before a change that
// would disconnect them. Fields behaves exactly like ListPPPSecrets.Fields.
type ListActivePPP struct {
	Fields []string
}

func (ListActivePPP) isProvisionOperation() {}

// CreatePPPSecret requests that a new PPPoE secret (a subscriber login) exist
// on the device. Name and Password are required; the driver rejects the
// operation if either is empty. Profile names the rate-limit/address template
// the secret uses (empty = device default). Service is the PPP service type
// (typically "pppoe"; empty = device default, usually "any"). RemoteAddress
// and Comment are optional and omitted from the native command when empty.
//
// This is a state-changing operation: unlike the List* reads, a driver
// classifies the command it produces as requiring human approval (HITL), so it
// only executes through the approved path (see usecase/network
// ExecuteCommandPreApproved), never auto-approved.
type CreatePPPSecret struct {
	Name          string
	Password      string
	Profile       string
	Service       string
	RemoteAddress string
	Comment       string
}

func (CreatePPPSecret) isProvisionOperation() {}

// CreatePPPProfile requests that a new PPPoE profile (a rate-limit /
// address template that secrets point at) exist on the device. Name is
// required; the driver rejects the operation if it is empty. RateLimit is the
// vendor-neutral bandwidth spec (RouterOS form "rx/tx", e.g. "10M/10M");
// LocalAddress and RemoteAddress name the gateway address and the client
// address/pool. All fields except Name are optional and omitted from the native
// command when empty. Like CreatePPPSecret, this is state-changing and only
// runs through the approved path.
type CreatePPPProfile struct {
	Name          string
	RateLimit     string
	LocalAddress  string
	RemoteAddress string
	Comment       string
}

func (CreatePPPProfile) isProvisionOperation() {}

// ChangeProfile requests that the existing PPPoE subscriber identified by
// Username be moved to Profile. Both fields are required; the driver rejects the
// operation if either is empty.
//
// On RouterOS a subscriber's new profile does not take effect while they are
// online — it applies on the next dial. So this is inherently a SEQUENCE, not a
// single command: change the stored profile, then drop the active session (if
// any) so the subscriber redials onto the new profile. Dropping a session that
// isn't there (subscriber offline) is a no-op, not a failure — the profile
// change still stands. Like the other writes this is state-changing and only
// runs through the approved path.
type ChangeProfile struct {
	Username string
	Profile  string
}

func (ChangeProfile) isProvisionOperation() {}
