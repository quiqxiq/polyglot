package mikrotik

import "errors"

// ErrNotConnected indicates Execute or Stream was called while this
// Driver's underlying connection was down and had not yet reconnected.
// Reconnection happens automatically in the background (see connect.go),
// but this driver never blocks a caller waiting for it — how long to wait
// or whether to retry is a usecase/caller policy decision, not this
// driver's to make.
var ErrNotConnected = errors.New("mikrotik: not connected")

// ErrStreamingCommand indicates Execute was called with a command that
// must be run via Stream instead — see isStreamingCommand in commands.go.
var ErrStreamingCommand = errors.New("mikrotik: command requires Stream, not Execute")

// ErrNotStreamingCommand indicates Stream was called with a command that
// is not recognized as streaming — use Execute instead.
var ErrNotStreamingCommand = errors.New("mikrotik: command is not a streaming command, use Execute")

// ErrSecretNotFound indicates a /ppp/secret lookup by name returned no
// matching entry on the device — used by FindPPPoESecretRosID in ppp.go.
var ErrSecretNotFound = errors.New("mikrotik: ppp secret not found")

// ErrProfileNotFound indicates a profile lookup by name returned no matching
// entry on the device — used by FindHotspotUserProfileRosID in hotspot_profile.go.
var ErrProfileNotFound = errors.New("mikrotik: profile not found")
