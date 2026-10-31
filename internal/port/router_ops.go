package port

import (
	"context"
	"io"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/restorepoint"
)

// DeviceCapabilities is the vendor-neutral AI capability profile for a router.
type DeviceCapabilities struct {
	Version          string
	Major            int
	InspectSupported bool
	AIMode           string // full | degraded
	Warning          string
}

// ScriptValidation is the vendor-neutral result of a pre-approval syntax check.
type ScriptValidation struct {
	Supported bool
	Valid     bool
	Errors    []string
	Warnings  []string
	Raw       map[string]any
}

// RawArtifact is an in-memory restore-point file produced by the device
// adapter. The usecase streams it into encrypted storage and then removes the
// temporary file from the router.
type RawArtifact struct {
	Kind  restorepoint.ArtifactKind
	Scope []string
	Name  string
	Data  io.ReadCloser
	Size  int64
}

// RestoreMode selects how a restore point is applied.
type RestoreMode string

const (
	RestoreInverse      RestoreMode = "INVERSE"
	RestoreScopedExport RestoreMode = "SCOPED_EXPORT"
	RestoreBinary       RestoreMode = "BINARY"
)

// RouterOps is the vendor-neutral device surface required by AI Router Ops.
// MikroTik implements it; the usecase never imports a driver package.
type RouterOps interface {
	// Capabilities reports RouterOS/version/AI-support profile.
	Capabilities(ctx context.Context) (DeviceCapabilities, error)
	// ValidateScript pre-validates a command/script (v7); v6 degrades.
	ValidateScript(ctx context.Context, script string) (ScriptValidation, error)
	// ReadMenu reads rows from a menu with an optional unique selector
	// (e.g. map[".id"]"*A"), using "print detail".
	ReadMenu(ctx context.Context, menu string, selector map[string]string) ([]map[string]string, error)
	// CreateArtifacts produces scoped exports + binary backup + inverse
	// script for the given scopes. maxBytes bounds each downloaded artifact.
	// The caller owns Data.Close.
	CreateArtifacts(ctx context.Context, restorePointID string, scopes []string, inverseScript []byte, maxBytes int64) ([]RawArtifact, error)
	// ExecuteRaw executes one vendor-native command with args.
	ExecuteRaw(ctx context.Context, raw string, args map[string]string) (command.Result, error)
	// UploadFile writes a temporary script/backup onto the device.
	UploadFile(ctx context.Context, name string, data []byte) error
	// ImportScript runs /import for an uploaded file.
	ImportScript(ctx context.Context, name string) error
	// LoadBinary runs /system/backup/load for an uploaded file and reboots.
	LoadBinary(ctx context.Context, name string) error
	// RemoveFile deletes a temporary router file (best effort).
	RemoveFile(ctx context.Context, name string)
	// PostVerify re-reads the affected menus and returns their rows.
	PostVerify(ctx context.Context, menus []string) (map[string][]map[string]string, error)
	// ResourceInfo returns router version and free disk space for preflight.
	ResourceInfo(ctx context.Context) (version string, freeBytes int64, err error)
	// RouterNow returns the device clock (watchdog uses router time).
	RouterNow(ctx context.Context) (time.Time, error)
}

// RestorePointStorage persists encrypted restore-point bundles.
type RestorePointStorage interface {
	Save(ctx context.Context, tenantID, deviceID, restorePointID string, kind restorepoint.ArtifactKind, scope []string, name string, data io.Reader, maxBytes int64) (restorepoint.Artifact, error)
	SaveManifest(ctx context.Context, tenantID, deviceID, restorePointID string, manifest any) error
	Open(ctx context.Context, artifact restorepoint.Artifact) (io.ReadCloser, error)
	// LinkArtifact reuses an existing encrypted artifact inside a new bundle
	// (hardlink, copy fallback) so deduplicated binary backups do not consume
	// a second copy on disk.
	LinkArtifact(ctx context.Context, tenantID, deviceID, newRestorePointID string, src restorepoint.Artifact) (restorepoint.Artifact, error)
	DeleteBundle(ctx context.Context, tenantID, deviceID, restorePointID string) error
}

// RouterInspector is an optional capability for vendors that expose
// introspection (MikroTik /console/inspect). Rows are returned as flat maps
// so MCP/API layers stay vendor-agnostic.
type RouterInspector interface {
	InspectChildren(ctx context.Context, path string) ([]map[string]string, error)
	InspectSyntax(ctx context.Context, path string) ([]map[string]string, error)
	InspectCompletion(ctx context.Context, input, path string) ([]map[string]string, error)
	InspectValidate(ctx context.Context, script string) (ScriptValidation, error)
}

// RollbackWatchdog is an optional capability: arm a router-side scheduler
// that runs the inverse script unless the post-verify cancels it. Only used
// when WATCHDOG_ENABLED=true (default off; must be tested on CHR first).
type RollbackWatchdog interface {
	ArmWatchdog(ctx context.Context, approvalID string, inverseScript []byte, seconds int) error
	DisarmWatchdog(ctx context.Context, approvalID string) error
}
