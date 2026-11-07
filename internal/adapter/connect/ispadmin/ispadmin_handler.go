package ispadmin

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
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
func (h *ISPAdminConnectHandler) ImportFile(ctx context.Context, req *connect.Request[polyglotv1.ImportFileRequest]) (*connect.Response[polyglotv1.ImportFileResponse], error) {
	if h.upsert == nil {
		return nil, response.Unavailable("importer unavailable")
	}
	var rows []importer.Row
	var err error
	if req.Msg.Format == polyglotv1.ImportFormat_IMPORT_FORMAT_XLSX {
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
		return connect.NewResponse(&polyglotv1.ImportFileResponse{
			Result:           &polyglotv1.ImportResult{RowsTotal: int32(len(rows))},
			PreviewRows:      preview,
			ValidationErrors: errMsgs,
		}), nil
	}

	if len(verrs) > 0 {
		return connect.NewResponse(&polyglotv1.ImportFileResponse{
			Result:           &polyglotv1.ImportResult{RowsTotal: int32(len(rows))},
			ValidationErrors: errMsgs,
		}), nil
	}

	res, uerr := h.upsert.ImportWithDevice(ctx, rows, req.Msg.DefaultDeviceId)
	if uerr != nil {
		return nil, response.MapDomainError(uerr)
	}
	return connect.NewResponse(&polyglotv1.ImportFileResponse{
		Result: toProtoImportResult(res),
	}), nil
}

// ImportRouter pulls PPPoE and/or Hotspot accounts from a connected MikroTik router and registers them.
func (h *ISPAdminConnectHandler) ImportRouter(ctx context.Context, req *connect.Request[polyglotv1.ImportRouterRequest]) (*connect.Response[polyglotv1.ImportRouterResponse], error) {
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

	for i := range rows {
		if strings.TrimSpace(rows[i].Address) == "" {
			rows[i].Address = fmt.Sprintf("Area Router %s", devName)
		}
	}
	verrs := importer.ValidateRouterRows(rows)
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
		return connect.NewResponse(&polyglotv1.ImportRouterResponse{
			Result:                   &polyglotv1.ImportResult{RowsTotal: int32(len(rows))},
			PreviewRows:              preview,
			ValidationErrors:         errMsgs,
			PppoeDetected:            int32(pullRes.PPPoEDetected),
			HotspotPermanentDetected: int32(pullRes.HotspotPermanentDetected),
			HotspotIpBindingDetected: int32(pullRes.HotspotIPBindingDetected),
			VouchersSkipped:          int32(pullRes.VouchersSkipped),
		}), nil
	}

	if len(verrs) > 0 {
		return connect.NewResponse(&polyglotv1.ImportRouterResponse{
			Result:           &polyglotv1.ImportResult{RowsTotal: int32(len(rows))},
			ValidationErrors: errMsgs,
		}), nil
	}

	res, uerr := h.upsert.ImportWithDevice(ctx, rows, req.Msg.DeviceId)
	if uerr != nil {
		return nil, response.MapDomainError(uerr)
	}
	return connect.NewResponse(&polyglotv1.ImportRouterResponse{
		Result:                   toProtoImportResult(res),
		PppoeDetected:            int32(pullRes.PPPoEDetected),
		HotspotPermanentDetected: int32(pullRes.HotspotPermanentDetected),
		HotspotIpBindingDetected: int32(pullRes.HotspotIPBindingDetected),
		VouchersSkipped:          int32(pullRes.VouchersSkipped),
	}), nil
}

// ExportCustomers dumps customer records to CSV or XLSX format.
func (h *ISPAdminConnectHandler) ExportCustomers(ctx context.Context, req *connect.Request[polyglotv1.ExportCustomersRequest]) (*connect.Response[polyglotv1.ExportCustomersResponse], error) {
	if h.exporter == nil {
		return nil, response.Unavailable("exporter unavailable")
	}
	format := "csv"
	contentType := "text/csv"
	filename := "pelanggan.csv"
	if req.Msg.Format == polyglotv1.ImportFormat_IMPORT_FORMAT_XLSX {
		format = "xlsx"
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "pelanggan.xlsx"
	}
	bytesOut, err := h.exporter.ExportAll(ctx, format, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.ExportCustomersResponse{
		Payload: bytesOut, ContentType: contentType, Filename: filename,
	}), nil
}

// Reconcile compares router state against database subscriptions and flags mismatches.
func (h *ISPAdminConnectHandler) Reconcile(ctx context.Context, req *connect.Request[polyglotv1.ReconcileRequest]) (*connect.Response[polyglotv1.ReconcileResponse], error) {
	driver, ok := h.resolve(ctx, req.Msg.DeviceId)
	if !ok || driver == nil {
		return nil, response.MapDomainError(device.ErrNotFound)
	}
	report, err := h.reconciler.Compare(ctx, req.Msg.DeviceId, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.ReconcileResponse{
		MissingInDb:     report.MissingInDB,
		MissingInRouter: report.MissingInRoute,
		ProfileMismatch: report.ProfileMismatch,
	}), nil
}

func toProtoImportResult(res *importer.Result) *polyglotv1.ImportResult {
	if res == nil {
		return nil
	}
	return &polyglotv1.ImportResult{
		RowsTotal:            int32(res.RowsTotal),
		CustomersCreated:     int32(res.CustomersCreated),
		CustomersUpdated:     int32(res.CustomersUpdated),
		SubscriptionsCreated: int32(res.SubsCreated),
		PlansCreated:         int32(res.PlansCreated),
		Skipped:              res.Skipped,
	}
}

// PullRouterPlans mengambil profil PPP dan/atau Hotspot dari router untuk ditampilkan di preview table.
func (h *ISPAdminConnectHandler) PullRouterPlans(
	ctx context.Context,
	req *connect.Request[polyglotv1.PullRouterPlansRequest],
) (*connect.Response[polyglotv1.PullRouterPlansResponse], error) {
	driver, ok := h.resolve(ctx, req.Msg.DeviceId)
	if !ok || driver == nil {
		return nil, response.MapDomainError(device.ErrNotFound)
	}

	rows, pppoeCount, hotspotCount, err := h.routerSrc.PullRouterPlans(ctx, driver, req.Msg.ServiceType)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	protoRows := make([]*polyglotv1.PlanImportRow, 0, len(rows))
	for _, r := range rows {
		protoRows = append(protoRows, &polyglotv1.PlanImportRow{
			Name:                  r.Name,
			ServiceType:           r.ServiceType,
			RateLimit:             r.RateLimit,
			BandwidthDownloadKbps: int32(r.BandwidthDownloadKbps),
			BandwidthUploadKbps:   int32(r.BandwidthUploadKbps),
			Price:                 r.Price,
			ParentQueue:           r.ParentQueue,
			AddressList:           r.AddressList,
			IpPoolName:            r.IPPoolName,
			SharedUsers:           int32(r.SharedUsers),
			SessionTimeout:        r.SessionTimeout,
			IdleTimeout:           r.IdleTimeout,
			IsNew:                 r.IsNew,
			Selected:              r.Selected,
			RouterProfile:         r.RouterProfile,
		})
	}

	return connect.NewResponse(&polyglotv1.PullRouterPlansResponse{
		Rows:            protoRows,
		PppoeDetected:   int32(pppoeCount),
		HotspotDetected: int32(hotspotCount),
	}), nil
}

// CommitPlans mengupsert daftar paket yang telah diedit/dikonfirmasi oleh admin.
func (h *ISPAdminConnectHandler) CommitPlans(
	ctx context.Context,
	req *connect.Request[polyglotv1.CommitPlansRequest],
) (*connect.Response[polyglotv1.CommitPlansResponse], error) {
	if h.upsert == nil {
		return nil, response.Unavailable("importer unavailable")
	}

	rows := make([]importer.PlanRow, 0, len(req.Msg.Rows))
	for _, r := range req.Msg.Rows {
		rows = append(rows, importer.PlanRow{
			Name:                  r.Name,
			ServiceType:           r.ServiceType,
			RateLimit:             r.RateLimit,
			BandwidthDownloadKbps: int(r.BandwidthDownloadKbps),
			BandwidthUploadKbps:   int(r.BandwidthUploadKbps),
			Price:                 r.Price,
			ParentQueue:           r.ParentQueue,
			AddressList:           r.AddressList,
			IPPoolName:            r.IpPoolName,
			SharedUsers:           int(r.SharedUsers),
			SessionTimeout:        r.SessionTimeout,
			IdleTimeout:           r.IdleTimeout,
			Selected:              r.Selected,
			RouterProfile:         r.RouterProfile,
		})
	}

	res, err := h.upsert.CommitPlans(ctx, rows)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.CommitPlansResponse{
		PlansCreated: int32(res.PlansCreated),
		PlansUpdated: int32(res.PlansUpdated),
		Errors:       res.Errors,
	}), nil
}

// PullRouterCustomers mengambil akun pelanggan & langganan terstruktur dari router.
func (h *ISPAdminConnectHandler) PullRouterCustomers(
	ctx context.Context,
	req *connect.Request[polyglotv1.PullRouterCustomersRequest],
) (*connect.Response[polyglotv1.PullRouterCustomersResponse], error) {
	driver, ok := h.resolve(ctx, req.Msg.DeviceId)
	if !ok || driver == nil {
		return nil, response.MapDomainError(device.ErrNotFound)
	}

	opts := importer.PullOptions{
		ServiceType:       req.Msg.ServiceType,
		IncludeIPBindings: req.Msg.IncludeIpBindings,
		IncludeVouchers:   req.Msg.IncludeVouchers,
	}

	rows, pppoeCount, permCount, ipbCount, vSkipped, err := h.routerSrc.PullRouterCustomerSubscriptionRows(
		ctx, driver, req.Msg.DeviceId, req.Msg.DeviceId, opts,
	)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	protoRows := make([]*polyglotv1.CustomerSubscriptionImportRow, 0, len(rows))
	for _, r := range rows {
		protoRows = append(protoRows, &polyglotv1.CustomerSubscriptionImportRow{
			CustomerCode:       r.CustomerCode,
			Name:               r.Name,
			Phone:              r.Phone,
			Email:              r.Email,
			Address:            r.Address,
			ServiceType:        r.ServiceType,
			Username:           r.Username,
			Password:           r.Password,
			PlanName:           r.PlanName,
			PlanId:             r.PlanID,
			Price:              r.Price,
			RateLimit:          r.RateLimit,
			LocalAddress:       r.LocalAddress,
			RemoteAddress:      r.RemoteAddress,
			MacAddress:         r.MACAddress,
			HotspotType:        r.HotspotType,
			BillingDay:         int32(r.BillingDay),
			DeviceId:           r.DeviceID,
			DeviceName:         r.DeviceName,
			Selected:           r.Selected,
			ValidationWarnings: r.ValidationWarnings,
			RouterProfile:      r.RouterProfile,
		})
	}

	return connect.NewResponse(&polyglotv1.PullRouterCustomersResponse{
		Rows:                     protoRows,
		PppoeDetected:            int32(pppoeCount),
		HotspotPermanentDetected: int32(permCount),
		HotspotIpBindingDetected: int32(ipbCount),
		VouchersSkipped:          int32(vSkipped),
	}), nil
}

// CommitCustomers mengupsert pelanggan & langganan yang telah diedit/dikonfirmasi dari UI.
func (h *ISPAdminConnectHandler) CommitCustomers(
	ctx context.Context,
	req *connect.Request[polyglotv1.CommitCustomersRequest],
) (*connect.Response[polyglotv1.CommitCustomersResponse], error) {
	if h.upsert == nil {
		return nil, response.Unavailable("importer unavailable")
	}

	rows := make([]importer.CustomerSubscriptionRow, 0, len(req.Msg.Rows))
	for _, r := range req.Msg.Rows {
		rows = append(rows, importer.CustomerSubscriptionRow{
			CustomerCode:  r.CustomerCode,
			Name:          r.Name,
			Phone:         r.Phone,
			Email:         r.Email,
			Address:       r.Address,
			ServiceType:   r.ServiceType,
			Username:      r.Username,
			Password:      r.Password,
			PlanName:      r.PlanName,
			PlanID:        r.PlanId,
			Price:         r.Price,
			RateLimit:     r.RateLimit,
			LocalAddress:  r.LocalAddress,
			RemoteAddress: r.RemoteAddress,
			MACAddress:    r.MacAddress,
			HotspotType:   r.HotspotType,
			BillingDay:    int(r.BillingDay),
			DeviceID:      r.DeviceId,
			DeviceName:    r.DeviceName,
			Selected:      r.Selected,
			RouterProfile: r.RouterProfile,
		})
	}

	res, err := h.upsert.CommitCustomerSubscriptionRows(ctx, req.Msg.DeviceId, rows)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.CommitCustomersResponse{
		CustomersCreated:     int32(res.CustomersCreated),
		CustomersUpdated:     int32(res.CustomersUpdated),
		SubscriptionsCreated: int32(res.SubscriptionsCreated),
		SubscriptionsUpdated: int32(res.SubscriptionsUpdated),
		PlansCreated:         int32(res.PlansCreated),
		Errors:               res.Errors,
	}), nil
}
