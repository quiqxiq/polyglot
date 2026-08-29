package ispadmin

import (
	"bytes"
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/importer"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/response"
)

type IspAdminConnectHandler struct {
	upsert     *importer.UpsertUseCase
	routerSrc  *importer.RouterSource
	reconciler *importer.Reconciler
	exporter   *importer.ExportUseCase
	resolve    func(ctx context.Context, deviceID string) (port.DeviceDriver, bool)
}

func NewIspAdminConnectHandler(
	upsert *importer.UpsertUseCase,
	routerSrc *importer.RouterSource,
	reconciler *importer.Reconciler,
	exporter *importer.ExportUseCase,
	resolver func(ctx context.Context, deviceID string) (port.DeviceDriver, bool),
) *IspAdminConnectHandler {
	return &IspAdminConnectHandler{
		upsert: upsert, routerSrc: routerSrc, reconciler: reconciler,
		exporter: exporter, resolve: resolver,
	}
}

func (h *IspAdminConnectHandler) ImportFile(ctx context.Context, req *connect.Request[devicepb.ImportFileRequest]) (*connect.Response[devicepb.ImportFileResponse], error) {
	if h.upsert == nil {
		return nil, response.Unavailable("importer unavailable")
	}
	var rows []importer.Row
	var err error
	if req.Msg.Format == devicepb.ImportFormat_IMPORT_FORMAT_XLSX {
		rows, err = importer.ParseXLSX(req.Msg.Payload)
	} else {
		rows, err = importer.ParseCSV(bytes.NewReader(req.Msg.Payload))
	}
	if err != nil {
		return nil, response.MapDomainError(fault.Wrap(fault.KindInvalidInput, err))
	}
	res, uerr := h.upsert.Import(ctx, rows)
	if uerr != nil {
		return nil, response.MapDomainError(uerr)
	}
	return connect.NewResponse(&devicepb.ImportFileResponse{
		Result: toProtoImportResult(res),
	}), nil
}

func (h *IspAdminConnectHandler) ImportRouter(ctx context.Context, req *connect.Request[devicepb.ImportRouterRequest]) (*connect.Response[devicepb.ImportRouterResponse], error) {
	driver, ok := h.resolve(ctx, req.Msg.DeviceId)
	if !ok || driver == nil {
		return nil, response.MapDomainError(device.ErrNotFound)
	}
	devName := req.Msg.DeviceName
	if devName == "" {
		devName = req.Msg.DeviceId
	}
	rows, err := h.routerSrc.PullPPPoERows(ctx, driver, devName)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	if verrs := importer.ValidateRows(rows); len(verrs) > 0 {
		errMsgs := make([]string, len(verrs))
		for i, e := range verrs {
			errMsgs[i] = e.Error()
		}
		return connect.NewResponse(&devicepb.ImportRouterResponse{
			ValidationErrors: errMsgs,
		}), nil
	}
	if req.Msg.DryRun {
		preview := make([]string, 0, 10)
		for i, r := range rows {
			if i >= 10 {
				break
			}
			preview = append(preview, r.Username+"@"+r.PlanName)
		}
		return connect.NewResponse(&devicepb.ImportRouterResponse{
			PreviewRows: preview,
			Result:      &devicepb.ImportResult{RowsTotal: int32(len(rows))},
		}), nil
	}
	res, uerr := h.upsert.Import(ctx, rows)
	if uerr != nil {
		return nil, response.MapDomainError(uerr)
	}
	return connect.NewResponse(&devicepb.ImportRouterResponse{
		Result: toProtoImportResult(res),
	}), nil
}

func (h *IspAdminConnectHandler) ExportCustomers(ctx context.Context, req *connect.Request[devicepb.ExportCustomersRequest]) (*connect.Response[devicepb.ExportCustomersResponse], error) {
	if h.exporter == nil {
		return nil, response.Unavailable("exporter unavailable")
	}
	format := "csv"
	contentType := "text/csv"
	filename := "pelanggan.csv"
	if req.Msg.Format == devicepb.ImportFormat_IMPORT_FORMAT_XLSX {
		format = "xlsx"
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "pelanggan.xlsx"
	}
	bytesOut, err := h.exporter.ExportAll(ctx, format)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ExportCustomersResponse{
		Payload: bytesOut, ContentType: contentType, Filename: filename,
	}), nil
}

func (h *IspAdminConnectHandler) Reconcile(ctx context.Context, req *connect.Request[devicepb.ReconcileRequest]) (*connect.Response[devicepb.ReconcileResponse], error) {
	driver, ok := h.resolve(ctx, req.Msg.DeviceId)
	if !ok || driver == nil {
		return nil, response.MapDomainError(device.ErrNotFound)
	}
	report, err := h.reconciler.Compare(ctx, req.Msg.DeviceId, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ReconcileResponse{
		MissingInDb:     report.MissingInDB,
		MissingInRouter: report.MissingInRoute,
		ProfileMismatch: report.ProfileMismatch,
	}), nil
}

func toProtoImportResult(res *importer.Result) *devicepb.ImportResult {
	if res == nil {
		return nil
	}
	return &devicepb.ImportResult{
		RowsTotal:            int32(res.RowsTotal),
		CustomersCreated:     int32(res.CustomersCreated),
		CustomersUpdated:     int32(res.CustomersUpdated),
		SubscriptionsCreated: int32(res.SubsCreated),
		PlansCreated:         int32(res.PlansCreated),
		Skipped:              res.Skipped,
	}
}
