package network

import (
	"encoding/json"
	"time"

	"github.com/quixiq/polyglot/internal/domain/audit"
)

// commandLog builds one detailed command log entry for the AI Router Ops
// audit trail. Summary is defensive: callers must mask secrets and truncate.
func commandLog(tenantID, deviceID, actorType, actorID, sessionID, mode, commandText, status, summary string) audit.CommandLog {
	args, _ := json.Marshal(map[string]string{})
	return audit.CommandLog{
		TenantID: tenantID, DeviceID: deviceID,
		ActorType: actorType, ActorID: actorID, SessionID: sessionID,
		Mode: mode, Command: commandText, CommandArgs: args,
		Status: status, OutputSummary: summary, CreatedAt: time.Now().UTC(),
	}
}
