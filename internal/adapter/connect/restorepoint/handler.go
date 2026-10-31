package restorepoint

import (
	"context"
	"io"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	domainRestore "github.com/quixiq/polyglot/internal/domain/restorepoint"
	"github.com/quixiq/polyglot/internal/port"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/response"
)

// ConnectHandler implements polyglot.v1.RestorePointService.
type ConnectHandler struct {
	uc         *networkUC.RestorePointUseCase
	authorizer port.DeviceAuthorizer
}

// NewConnectHandler builds the restore point handler.
func NewConnectHandler(uc *networkUC.RestorePointUseCase, authorizer port.DeviceAuthorizer) *ConnectHandler {
	return &ConnectHandler{uc: uc, authorizer: authorizer}
}

func (h *ConnectHandler) ListRestorePoints(ctx context.Context, req *connect.Request[devicepb.ListRestorePointsRequest]) (*connect.Response[devicepb.ListRestorePointsResponse], error) {
	items, total, err := h.uc.List(ctx, port.RestorePointFilter{
		TenantID: "tenant-default", DeviceID: req.Msg.GetDeviceId(),
		Status: req.Msg.GetStatus(), Limit: int(req.Msg.GetLimit()), Offset: int(req.Msg.GetOffset()),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*devicepb.RestorePoint, len(items))
	for i := range items {
		out[i] = toProto(items[i])
	}
	return connect.NewResponse(&devicepb.ListRestorePointsResponse{RestorePoints: out, Total: int32(total)}), nil
}

func (h *ConnectHandler) GetRestorePoint(ctx context.Context, req *connect.Request[devicepb.GetRestorePointRequest]) (*connect.Response[devicepb.GetRestorePointResponse], error) {
	rp, err := h.uc.Get(ctx, req.Msg.GetId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetRestorePointResponse{RestorePoint: toProto(rp)}), nil
}

func (h *ConnectHandler) CreateRestorePoint(ctx context.Context, req *connect.Request[devicepb.CreateRestorePointRequest]) (*connect.Response[devicepb.CreateRestorePointResponse], error) {
	userID, roles, ok := iauth.IdentityFromContext(ctx)
	if !ok || userID == 0 {
		return nil, response.Unauthenticated("user session required")
	}
	if err := h.checkDevice(ctx, userID, roles, req.Msg.GetDeviceId()); err != nil {
		return nil, err
	}
	rp, err := h.uc.CreateManual(ctx, "tenant-default", req.Msg.GetDeviceId(), actorID(userID))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreateRestorePointResponse{RestorePoint: toProto(rp)}), nil
}

func (h *ConnectHandler) RestoreFromPoint(ctx context.Context, req *connect.Request[devicepb.RestoreFromPointRequest]) (*connect.Response[devicepb.RestoreFromPointResponse], error) {
	if _, _, ok := iauth.IdentityFromContext(ctx); !ok {
		return nil, response.Unauthenticated("user session required")
	}
	mode := port.RestoreInverse
	switch req.Msg.GetMode() {
	case devicepb.RestoreMode_RESTORE_MODE_SCOPED_EXPORT:
		mode = port.RestoreScopedExport
	case devicepb.RestoreMode_RESTORE_MODE_BINARY:
		mode = port.RestoreBinary
	}
	rp, err := h.uc.Restore(ctx, req.Msg.GetId(), mode, req.Msg.GetConfirmation())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RestoreFromPointResponse{RestorePoint: toProto(rp)}), nil
}

func (h *ConnectHandler) DownloadRestorePointArtifact(ctx context.Context, req *connect.Request[devicepb.DownloadRestorePointArtifactRequest]) (*connect.Response[devicepb.DownloadRestorePointArtifactResponse], error) {
	kind := domainRestore.ArtifactKind(req.Msg.GetKind())
	if kind == "" {
		kind = domainRestore.ArtifactInverseScript
	}
	rc, _, err := h.uc.OpenArtifact(ctx, req.Msg.GetId(), kind)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	defer func() { _ = rc.Close() }()
	data, err := io.ReadAll(io.LimitReader(rc, 50*1024*1024))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DownloadRestorePointArtifactResponse{
		Payload: data, ContentType: "application/octet-stream", Filename: "restore-point-" + req.Msg.GetId() + ".bin",
	}), nil
}

func (h *ConnectHandler) DeleteRestorePoint(ctx context.Context, req *connect.Request[devicepb.DeleteRestorePointRequest]) (*connect.Response[devicepb.DeleteRestorePointResponse], error) {
	if err := h.uc.Delete(ctx, req.Msg.GetId()); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteRestorePointResponse{Message: "Restore point dihapus"}), nil
}

func (h *ConnectHandler) checkDevice(ctx context.Context, userID uint, roles []string, deviceID string) error {
	if h.authorizer == nil || deviceID == "" {
		return nil
	}
	ok, err := h.authorizer.CanAccessDevice(ctx, userID, roles, deviceID)
	if err != nil {
		return response.MapDomainError(err)
	}
	if !ok {
		return response.PermissionDenied("no access to this device")
	}
	return nil
}

func actorID(userID uint) string {
	return "user:" + itoa(userID)
}

func itoa(v uint) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

func toProto(rp domainRestore.RestorePoint) *devicepb.RestorePoint {
	arts := make([]*devicepb.RestorePointArtifact, len(rp.Artifacts))
	for i, a := range rp.Artifacts {
		arts[i] = &devicepb.RestorePointArtifact{Kind: string(a.Kind), Scope: a.Scope, Sha256: a.SHA256, SizeBytes: a.SizeBytes}
	}
	var used int64
	if rp.UsedAt != nil {
		used = rp.UsedAt.Unix()
	}
	return &devicepb.RestorePoint{
		Id: rp.ID, DeviceId: rp.DeviceID, Trigger: string(rp.Trigger), Status: string(rp.Status),
		Scope: rp.Scope, RouterosVersion: rp.RouterOSVersion, TotalSizeBytes: rp.TotalSizeBytes,
		Artifacts: arts, CreatedAtUnix: rp.CreatedAt.Unix(), ExpiresAtUnix: rp.ExpiresAt.Unix(), UsedAtUnix: used,
	}
}
