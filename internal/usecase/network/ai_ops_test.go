package network_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/approval"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/domain/restorepoint"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
)

type fakeOps struct {
	mu          sync.Mutex
	executed    []command.Command
	menus       map[string][]map[string]string
	invalid     bool
	restoreFail bool
	uploaded    map[string][]byte
	imported    []string
	loaded      []string
}

func (f *fakeOps) Capabilities(context.Context) (port.DeviceCapabilities, error) {
	return port.DeviceCapabilities{Version: "7.16.2", Major: 7, InspectSupported: true, AIMode: "full"}, nil
}
func (f *fakeOps) ValidateScript(_ context.Context, _ string) (port.ScriptValidation, error) {
	return port.ScriptValidation{Supported: true, Valid: !f.invalid}, nil
}
func (f *fakeOps) ReadMenu(_ context.Context, menu string, _ map[string]string) ([]map[string]string, error) {
	return f.menus[menu], nil
}
func (f *fakeOps) CreateArtifacts(_ context.Context, _ string, scopes []string, inverse []byte, _ int64) ([]port.RawArtifact, error) {
	if f.restoreFail {
		return nil, assert.AnError
	}
	out := []port.RawArtifact{}
	for _, scope := range scopes {
		data := []byte("# export " + scope + "\n")
		out = append(out, port.RawArtifact{Kind: restorepoint.ArtifactScopedExport, Scope: []string{scope}, Name: "scoped.rsc", Data: io.NopCloser(bytes.NewReader(data)), Size: int64(len(data))})
	}
	bin := []byte("BINARY")
	out = append(out, port.RawArtifact{Kind: restorepoint.ArtifactBinaryBackup, Name: "binary.backup", Data: io.NopCloser(bytes.NewReader(bin)), Size: int64(len(bin))})
	if len(inverse) > 0 {
		out = append(out, port.RawArtifact{Kind: restorepoint.ArtifactInverseScript, Name: "inverse.rsc", Data: io.NopCloser(bytes.NewReader(inverse)), Size: int64(len(inverse))})
	}
	return out, nil
}
func (f *fakeOps) ExecuteRaw(_ context.Context, raw string, _ map[string]string) (command.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.executed = append(f.executed, command.Command{Raw: raw})
	return command.Result{Output: "ok", Rows: []map[string]string{{"status": "ok"}}}, nil
}
func (f *fakeOps) UploadFile(_ context.Context, name string, data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.uploaded == nil {
		f.uploaded = map[string][]byte{}
	}
	f.uploaded[name] = data
	return nil
}
func (f *fakeOps) ImportScript(_ context.Context, name string) error {
	f.imported = append(f.imported, name)
	return nil
}
func (f *fakeOps) LoadBinary(_ context.Context, name string) error {
	f.loaded = append(f.loaded, name)
	return nil
}
func (f *fakeOps) RemoveFile(context.Context, string) {}
func (f *fakeOps) PostVerify(context.Context, []string) (map[string][]map[string]string, error) {
	return map[string][]map[string]string{}, nil
}
func (f *fakeOps) ResourceInfo(context.Context) (string, int64, error) {
	return "7.16.2", 1 << 30, nil
}
func (f *fakeOps) RouterNow(context.Context) (time.Time, error) { return time.Now().UTC(), nil }

type fakeStorage struct {
	mu        sync.Mutex
	artifacts map[string][]byte
	manifest  map[string]any
	saved     map[restorepoint.ArtifactKind]int
	links     int
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{artifacts: map[string][]byte{}, saved: map[restorepoint.ArtifactKind]int{}}
}

func (s *fakeStorage) Save(_ context.Context, _, _, rpID string, kind restorepoint.ArtifactKind, scope []string, name string, data io.Reader, _ int64) (restorepoint.Artifact, error) {
	b, err := io.ReadAll(data)
	if err != nil {
		return restorepoint.Artifact{}, err
	}
	sum := sha256.Sum256(b)
	path := rpID + "/" + string(kind) + "/" + name
	s.mu.Lock()
	s.artifacts[path] = b
	s.saved[kind]++
	s.mu.Unlock()
	return restorepoint.Artifact{Kind: kind, Scope: scope, StoragePath: path, SHA256: hex.EncodeToString(sum[:]), SizeBytes: int64(len(b))}, nil
}
func (s *fakeStorage) SaveManifest(_ context.Context, _, _, _ string, m any) error {
	s.mu.Lock()
	s.manifest = map[string]any{"m": m}
	s.mu.Unlock()
	return nil
}
func (s *fakeStorage) Open(_ context.Context, a restorepoint.Artifact) (io.ReadCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.artifacts[a.StoragePath]
	if !ok {
		return nil, assert.AnError
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}
func (s *fakeStorage) LinkArtifact(_ context.Context, _, _, _ string, src restorepoint.Artifact) (restorepoint.Artifact, error) {
	s.mu.Lock()
	s.links++
	s.mu.Unlock()
	return src, nil
}
func (s *fakeStorage) DeleteBundle(context.Context, string, string, string) error { return nil }

type fakeDevices struct {
	byID map[string]device.Device
}

func (f *fakeDevices) Save(context.Context, device.Device) error { return nil }
func (f *fakeDevices) FindByID(_ context.Context, id string) (device.Device, error) {
	d, ok := f.byID[id]
	if !ok {
		return device.Device{}, device.ErrNotFound
	}
	return d, nil
}
func (f *fakeDevices) FindAll(context.Context) ([]device.Device, error) { return nil, nil }
func (f *fakeDevices) Update(context.Context, device.Device) error      { return nil }
func (f *fakeDevices) Delete(context.Context, string) error             { return nil }
func (f *fakeDevices) FindByUserScope(context.Context, uint) ([]device.Device, error) {
	return nil, nil
}

type fakeDriver struct{}

func (fakeDriver) Execute(context.Context, command.Command) (command.Result, error) {
	return command.Result{}, nil
}
func (fakeDriver) Classify(cmd command.Command) command.Classification {
	if len(cmd.Raw) > 0 && (contains(cmd.Raw, "/add") || contains(cmd.Raw, "/set") || contains(cmd.Raw, "/remove") || contains(cmd.Raw, "/reboot")) {
		return command.Classification{Mode: command.ModeMutation}
	}
	return command.Classification{Mode: command.ModeRead}
}
func (fakeDriver) Translate(command.Operation) (command.Command, error) {
	return command.Command{}, nil
}
func (fakeDriver) Close() error { return nil }

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

type harness struct {
	approvals *mocktest.FakeApprovalRepo
	restores  *mocktest.FakeRestorePointRepo
	logs      *mocktest.FakeCommandLogRepo
	ops       *fakeOps
	storage   *fakeStorage
	devices   *fakeDevices
	restoreUC *networkUC.RestorePointUseCase
	proposal  *networkUC.ProposalUseCase
	execute   *networkUC.ExecutionUseCase
	decide    *networkUC.DecisionUseCase
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		approvals: mocktest.NewFakeApprovalRepo(),
		restores:  mocktest.NewFakeRestorePointRepo(),
		logs:      mocktest.NewFakeCommandLogRepo(),
		ops:       &fakeOps{menus: map[string][]map[string]string{"ip/firewall/filter": {{".id": "*A", "action": "accept", "chain": "input"}}}},
		storage:   newFakeStorage(),
		devices:   &fakeDevices{byID: map[string]device.Device{"dev1": {ID: "dev1", TenantID: "tenant-default", Name: "Router-A", Enabled: true}}},
	}
	ops := func(context.Context, string) (port.RouterOps, error) { return h.ops, nil }
	drv := func(context.Context, string) (port.DeviceDriver, error) { return fakeDriver{}, nil }
	h.restoreUC = networkUC.NewRestorePointUseCase(networkUC.RestorePointDeps{
		Repo: h.restores, Storage: h.storage, Devices: h.devices, Ops: ops,
		Config: networkUC.RestorePointConfig{TTL: time.Hour, MaxBytes: 1 << 20, Dedupe: true},
	})
	h.proposal = networkUC.NewProposalUseCase(networkUC.ProposalDeps{
		Approvals: h.approvals, RestorePoints: h.restoreUC, Devices: h.devices,
		Driver: drv, Ops: ops, CommandLogs: h.logs,
		Config: networkUC.ProposalConfig{TTL: 15 * time.Minute, MaxScriptLines: 50, V6Mutations: true},
	})
	h.execute = networkUC.NewExecutionUseCase(networkUC.ExecutionDeps{
		Approvals: h.approvals, RestorePoints: h.restores, Devices: h.devices,
		Driver: drv, Ops: ops, CommandLogs: h.logs,
	})
	h.decide = networkUC.NewDecisionUseCase(networkUC.DecisionDeps{
		Approvals: h.approvals, RestorePoints: h.restoreUC, Devices: h.devices,
		Driver: drv, Ops: ops, Execution: h.execute, RequireRejectReason: true,
	})
	return h
}

func TestProposalCreatesApprovalWithRestorePoint(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	in := networkUC.ProposeCommandInput{
		TenantID: "tenant-default", DeviceID: "dev1", ActorID: "agent:1", ActorType: approval.ActorAgent,
		Command: "/ip/firewall/filter/add", Args: map[string]string{"chain": "input", "action": "drop", "comment": "polyglot-test"},
	}
	p, err := h.proposal.ProposeCommand(ctx, in)
	require.NoError(t, err)
	assert.Equal(t, approval.StatusPending, p.Approval.Status)
	assert.NotEmpty(t, p.Approval.ID)
	assert.NotEmpty(t, p.RestorePoint.ID)
	assert.Equal(t, restorepoint.StatusAvailable, p.RestorePoint.Status)
	assert.True(t, p.RestorePoint.HasKind(restorepoint.ArtifactBinaryBackup))
	assert.True(t, p.RestorePoint.HasKind(restorepoint.ArtifactInverseScript))
	assert.NotEmpty(t, p.Diff.Lines)
	assert.Equal(t, "dev1", p.Approval.DeviceID)

	stored, err := h.approvals.FindByID(ctx, p.Approval.ID)
	require.NoError(t, err)
	assert.Equal(t, approval.StatusPending, stored.Status)

	// Idempotent: same proposal returns the existing PENDING approval.
	p2, err := h.proposal.ProposeCommand(ctx, in)
	require.NoError(t, err)
	assert.Equal(t, p.Approval.ID, p2.Approval.ID)
}

func TestProposalDedupesBinaryBackupOnNonCriticalMenu(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	in := networkUC.ProposeCommandInput{
		DeviceID: "dev1", Command: "/queue/simple/add", Args: map[string]string{"name": "q1", "max-limit": "1M/1M", "comment": "a"},
		ActorType: approval.ActorAgent,
	}
	first, err := h.proposal.ProposeCommand(ctx, in)
	require.NoError(t, err)
	require.True(t, first.RestorePoint.HasKind(restorepoint.ArtifactBinaryBackup))

	in.Args = map[string]string{"name": "q1", "max-limit": "1M/1M", "comment": "b"}
	second, err := h.proposal.ProposeCommand(ctx, in)
	require.NoError(t, err)
	require.True(t, second.RestorePoint.HasKind(restorepoint.ArtifactBinaryBackup))

	h.storage.mu.Lock()
	binarySaves := h.storage.saved[restorepoint.ArtifactBinaryBackup]
	links := h.storage.links
	h.storage.mu.Unlock()
	assert.Equal(t, 1, binarySaves, "second non-critical mutation must reuse the binary artifact")
	assert.Equal(t, 1, links, "reused artifact must be linked into the new bundle")
}

func TestProposalRejectsReadAndMissingSelector(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()

	_, err := h.proposal.ProposeCommand(ctx, networkUC.ProposeCommandInput{
		DeviceID: "dev1", Command: "/ip/firewall/filter/print", ActorType: approval.ActorAgent,
	})
	require.ErrorContains(t, err, "READ")

	_, err = h.proposal.ProposeCommand(ctx, networkUC.ProposeCommandInput{
		DeviceID: "dev1", Command: "/ip/firewall/filter/set", Args: map[string]string{"action": "drop"}, ActorType: approval.ActorAgent,
	})
	require.ErrorContains(t, err, "selector")
}

func TestProposalFailsClosedWhenRestorePointFails(t *testing.T) {
	h := newHarness(t)
	h.ops.restoreFail = true
	_, err := h.proposal.ProposeCommand(context.Background(), networkUC.ProposeCommandInput{
		DeviceID: "dev1", Command: "/ip/firewall/filter/add", Args: map[string]string{"chain": "input", "action": "drop"}, ActorType: approval.ActorAgent,
	})
	require.Error(t, err)
	list, total, _ := h.approvals.List(context.Background(), port.ApprovalFilter{TenantID: "tenant-default"})
	assert.Equal(t, 0, total)
	assert.Empty(t, list)
}

func TestProposalRejectsInvalidScript(t *testing.T) {
	h := newHarness(t)
	h.ops.invalid = true
	_, err := h.proposal.ProposeCommand(context.Background(), networkUC.ProposeCommandInput{
		DeviceID: "dev1", Command: "/ip/firewall/filter/add", Args: map[string]string{"chain": "input", "action": "drop"}, ActorType: approval.ActorAgent,
	})
	require.ErrorContains(t, err, "syntax invalid")
}

func TestDecideRejectAndEdit(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	p, err := h.proposal.ProposeCommand(ctx, networkUC.ProposeCommandInput{
		DeviceID: "dev1", Command: "/ip/firewall/filter/set", Args: map[string]string{".id": "*A", "action": "drop"}, ActorType: approval.ActorAgent,
	})
	require.NoError(t, err)

	rejected, err := h.decide.Decide(ctx, networkUC.DecideInput{ApprovalID: p.Approval.ID, Decision: networkUC.DecisionReject, ActorID: 7, Reason: "not now", Revision: 1})
	require.NoError(t, err)
	assert.Equal(t, approval.StatusRejected, rejected.Status)

	// Fresh proposal for edit.
	p2, err := h.proposal.ProposeCommand(ctx, networkUC.ProposeCommandInput{
		DeviceID: "dev1", Command: "/ip/firewall/filter/add", Args: map[string]string{"chain": "input", "action": "drop", "comment": "x"}, ActorType: approval.ActorAgent,
	})
	require.NoError(t, err)
	edited, err := h.decide.Decide(ctx, networkUC.DecideInput{
		ApprovalID: p2.Approval.ID, Decision: networkUC.DecisionEdit, ActorID: 7,
		EditedCommand: "/ip/firewall/filter/add", EditedArgs: map[string]string{"chain": "input", "action": "accept", "comment": "y"},
		Revision: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, approval.StatusPending, edited.Status)
	assert.Equal(t, 2, edited.Revision)
	assert.Equal(t, 1, edited.EditCount)
}

func TestApproveExecutesOnce(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	p, err := h.proposal.ProposeCommand(ctx, networkUC.ProposeCommandInput{
		DeviceID: "dev1", Command: "/ip/firewall/filter/add", Args: map[string]string{"chain": "input", "action": "drop", "comment": "z"}, ActorType: approval.ActorAgent,
	})
	require.NoError(t, err)

	executed, err := h.decide.Decide(ctx, networkUC.DecideInput{ApprovalID: p.Approval.ID, Decision: networkUC.DecisionApprove, ActorID: 9, Revision: 1})
	require.NoError(t, err)
	assert.Equal(t, approval.StatusExecuted, executed.Status)
	h.ops.mu.Lock()
	assert.Len(t, h.ops.executed, 1)
	h.ops.mu.Unlock()

	// Second approve is rejected (already EXECUTED).
	_, err = h.decide.Decide(ctx, networkUC.DecideInput{ApprovalID: p.Approval.ID, Decision: networkUC.DecisionApprove, ActorID: 9, Revision: 1})
	require.Error(t, err)
	h.ops.mu.Lock()
	assert.Len(t, h.ops.executed, 1)
	h.ops.mu.Unlock()
}

func TestExecutionFailsWhenRestorePointNotAvailable(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	p, err := h.proposal.ProposeCommand(ctx, networkUC.ProposeCommandInput{
		DeviceID: "dev1", Command: "/ip/firewall/filter/add", Args: map[string]string{"chain": "input", "action": "drop", "comment": "q"}, ActorType: approval.ActorAgent,
	})
	require.NoError(t, err)
	rp, _ := h.restores.FindByID(ctx, p.RestorePoint.ID)
	rp.Status = restorepoint.StatusPartial
	require.NoError(t, h.restores.UpdateStatus(ctx, rp.ID, restorepoint.StatusPartial, nil))

	req, err := h.approvals.FindByID(ctx, p.Approval.ID)
	require.NoError(t, err)
	req.Status = approval.StatusApproved
	ok, err := h.approvals.UpdateIfRevision(ctx, req, req.Revision)
	require.NoError(t, err)
	require.True(t, ok)

	_, err = h.execute.Execute(ctx, p.Approval.ID)
	require.Error(t, err)
	final, _ := h.approvals.FindByID(ctx, p.Approval.ID)
	assert.Equal(t, approval.StatusFailed, final.Status)
}

func TestRestoreTypedConfirmation(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	p, err := h.proposal.ProposeCommand(ctx, networkUC.ProposeCommandInput{
		DeviceID: "dev1", Command: "/ip/firewall/filter/add", Args: map[string]string{"chain": "input", "action": "drop", "comment": "r"}, ActorType: approval.ActorAgent,
	})
	require.NoError(t, err)

	_, err = h.restoreUC.Restore(ctx, p.RestorePoint.ID, port.RestoreInverse, "wrong-name")
	require.Error(t, err)

	rp, err := h.restoreUC.Restore(ctx, p.RestorePoint.ID, port.RestoreInverse, "Router-A")
	require.NoError(t, err)
	assert.Equal(t, restorepoint.StatusUsed, rp.Status)
	assert.NotEmpty(t, h.ops.imported)
}
