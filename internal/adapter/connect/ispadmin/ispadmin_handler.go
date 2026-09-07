package ispadmin

import (
	"bytes"
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/importer"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/response"
)

// ISPAdminConnectHandler implements ConnectRPC procedures for ISP administration
// (customer import, reconciliation, router data pull, and bulk export).
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type ISPAdminConnectHandler struct {
	upsert     *importer.UpsertUseCase
	routerSrc  *importer.RouterSource
	reconciler *importer.Reconciler
	exporter   *importer.ExportUseCase
	resolve    func(ctx context.Context, deviceID string) (port.DeviceDriver, bool)
}

// NewISPAdminConnectHandler constructs an ISP admin ConnectRPC handler.
func NewISPAdminConnectHandler(
	upsert *importer.UpsertUseCase,
	routerSrc *importer.RouterSource,
	reconciler *importer.Reconciler,
	exporter *importer.ExportUseCase,
	resolver func(ctx context.Context, deviceID string) (port.DeviceDriver, bool),
) *ISPAdminConnectHandler {
	return &ISPAdminConnectHandler{
		upsert:     upsert,
		routerSrc:  routerSrc,
		reconciler: reconciler,
		exporter:   exporter,
		resolve:    resolver,
	}
}

// ImportFile parses an uploaded CSV or XLSX file and upserts subscriber records.
func (h *ISPAdminConnectHandler) ImportFile(ctx context.Context, req *connect.Request[devicepb.ImportFileRequest]) (*connect.Response[devicepb.ImportFileResponse], error) {
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

	verrs := importer.ValidateRows(rows)
	errMsgs := make([]string, 0, len(verrs))
	for _, e := range verrs {
		errMsgs = append(errMsgs, e.Error())
	}

	if req.Msg.DryRun {
		preview := make([]string, 0, 20)
		for i, r := range rows {
			if i >= 20 {
				break
			}
			preview = append(preview, fmt.Sprintf("%s | %s | %s | %s", r.Name, r.Username, r.PlanName, r.DeviceName))
		}
		return connect.NewResponse(&devicepb.ImportFileResponse{
			Result:           &devicepb.ImportResult{RowsTotal: int32(len(rows))},
			PreviewRows:      preview,
			ValidationErrors: errMsgs,
		}), nil
	}

	if len(verrs) > 0 {
		return connect.NewResponse(&devicepb.ImportFileResponse{
			Result:           &devicepb.ImportResult{RowsTotal: int32(len(rows))},
			ValidationErrors: errMsgs,
		}), nil
	}

	res, uerr := h.upsert.ImportWithDevice(ctx, rows, req.Msg.DefaultDeviceId)
	if uerr != nil {
		return nil, response.MapDomainError(uerr)
	}
	return connect.NewResponse(&devicepb.ImportFileResponse{
		Result: toProtoImportResult(res),
	}), nil
}

// ImportRouter pulls PPPoE and/or Hotspot accounts from a connected MikroTik router and registers them.
func (h *ISPAdminConnectHandler) ImportRouter(ctx context.Context, req *connect.Request[devicepb.ImportRouterRequest]) (*connect.Response[devicepb.ImportRouterResponse], error) {
	driver, ok := h.resolve(ctx, req.Msg.DeviceId)
	if !ok || driver == nil {
		return nil, response.MapDomainError(device.ErrNotFound)
	}
	devName := req.Msg.DeviceName
	if devName == "" {
		devName = req.Msg.DeviceId
	}

	opts := importer.PullOptions{
		ServiceType:       req.Msg.ServiceType,
		IncludeIPBindings: req.Msg.IncludeIpBindings,
		IncludeVouchers:   req.Msg.IncludeVouchers,
	}

	pullRes, err := h.routerSrc.PullRouterRows(ctx, driver, devName, opts)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	rows := pullRes.Rows

	verrs := importer.ValidateRows(rows)
	errMsgs := make([]string, 0, len(verrs))
	for _, e := range verrs {
		errMsgs = append(errMsgs, e.Error())
	}

	if req.Msg.DryRun {
		preview := make([]string, 0, 20)
		for i, r := range rows {
			if i >= 20 {
				break
			}
			preview = append(preview, fmt.Sprintf("[%s] %s @ %s (%s)", r.ServiceType, r.Username, r.PlanName, r.Status))
		}
		return connect.NewResponse(&devicepb.ImportRouterResponse{
			Result:                   &devicepb.ImportResult{RowsTotal: int32(len(rows))},
			PreviewRows:              preview,
			ValidationErrors:         errMsgs,
			PppoeDetected:            int32(pullRes.PPPoEDetected),
			HotspotPermanentDetected: int32(pullRes.HotspotPermanentDetected),
			HotspotIpBindingDetected: int32(pullRes.HotspotIPBindingDetected),
			VouchersSkipped:          int32(pullRes.VouchersSkipped),
		}), nil
	}

	if len(verrs) > 0 {
		return connect.NewResponse(&devicepb.ImportRouterResponse{
			Result:           &devicepb.ImportResult{RowsTotal: int32(len(rows))},
			ValidationErrors: errMsgs,
		}), nil
	}

	res, uerr := h.upsert.ImportWithDevice(ctx, rows, req.Msg.DeviceId)
	if uerr != nil {
		return nil, response.MapDomainError(uerr)
	}
	return connect.NewResponse(&devicepb.ImportRouterResponse{
		Result:                   toProtoImportResult(res),
		PppoeDetected:            int32(pullRes.PPPoEDetected),
		HotspotPermanentDetected: int32(pullRes.HotspotPermanentDetected),
		HotspotIpBindingDetected: int32(pullRes.HotspotIPBindingDetected),
		VouchersSkipped:          int32(pullRes.VouchersSkipped),
	}), nil
}

// ExportCustomers dumps customer records to CSV or XLSX format.
func (h *ISPAdminConnectHandler) ExportCustomers(ctx context.Context, req *connect.Request[devicepb.ExportCustomersRequest]) (*connect.Response[devicepb.ExportCustomersResponse], error) {
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
	bytesOut, err := h.exporter.ExportAll(ctx, format, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ExportCustomersResponse{
		Payload: bytesOut, ContentType: contentType, Filename: filename,
	}), nil
}

// Reconcile compares router state against database subscriptions and flags mismatches.
func (h *ISPAdminConnectHandler) Reconcile(ctx context.Context, req *connect.Request[devicepb.ReconcileRequest]) (*connect.Response[devicepb.ReconcileResponse], error) {
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
