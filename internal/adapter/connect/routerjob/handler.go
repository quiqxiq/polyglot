// Package routerjob memasang ConnectRPC RouterJobService: monitoring antrean
// job router, retry manual, dan ekspor script .rsc (planning v2 Fase 4).
package routerjob

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
	"github.com/quixiq/polyglot/internal/port"
	jobuc "github.com/quixiq/polyglot/internal/usecase/routerjob"
	"github.com/quixiq/polyglot/pkg/response"
)

// ConnectHandler implements RouterJobService.
type ConnectHandler struct {
	jobs     port.RouterJobRepository
	exporter *jobuc.Exporter
}

// NewConnectHandler wires the router job handler.
func NewConnectHandler(jobs port.RouterJobRepository, exporter *jobuc.Exporter) *ConnectHandler {
	return &ConnectHandler{jobs: jobs, exporter: exporter}
}

func toProtoJob(j domainRouterJob.Job) *devicepb.RouterJob {
	payload := "{}"
	if len(j.Payload) > 0 {
		payload = string(j.Payload)
	}
	return &devicepb.RouterJob{
		Id:             j.ID,
		TenantId:       j.TenantID,
		DeviceId:       j.DeviceID,
		JobType:        string(j.JobType),
		PayloadJson:    payload,
		IdempotencyKey: j.IdempotencyKey,
		Status:         string(j.Status),
		Attempts:       int32(j.Attempts),
		MaxAttempts:    int32(j.MaxAttempts),
		NextRunAtUnix:  j.NextRunAt.Unix(),
		LastError:      j.LastError,
		CreatedAtUnix:  j.CreatedAt.Unix(),
		UpdatedAtUnix:  j.UpdatedAt.Unix(),
	}
}

// ListRouterJobs mengembalikan antrean job router (opsional per status).
func (h *ConnectHandler) ListRouterJobs(ctx context.Context, req *connect.Request[devicepb.ListRouterJobsRequest]) (*connect.Response[devicepb.ListRouterJobsResponse], error) {
	if h.jobs == nil {
		return nil, response.Unavailable("router job repository unavailable")
	}
	list, err := h.jobs.List(ctx, port.RouterJobFilter{Status: req.Msg.Status, Limit: 200})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*devicepb.RouterJob, len(list))
	for i := range list {
		out[i] = toProtoJob(list[i])
	}
	return connect.NewResponse(&devicepb.ListRouterJobsResponse{Jobs: out}), nil
}

// RetryRouterJob mengembalikan job DEAD/FAILED ke antrean (retry manual).
func (h *ConnectHandler) RetryRouterJob(ctx context.Context, req *connect.Request[devicepb.RetryRouterJobRequest]) (*connect.Response[devicepb.RetryRouterJobResponse], error) {
	if h.jobs == nil {
		return nil, response.Unavailable("router job repository unavailable")
	}
	job, err := h.jobs.FindByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(domainRouterJob.ErrNotFound)
	}
	switch job.Status {
	case domainRouterJob.StatusDead, domainRouterJob.StatusFailed:
		if err := h.jobs.Requeue(ctx, job.ID); err != nil {
			return nil, response.MapDomainError(err)
		}
		job, err = h.jobs.FindByID(ctx, job.ID)
		if err != nil {
			return nil, response.MapDomainError(err)
		}
	case domainRouterJob.StatusQueued, domainRouterJob.StatusRetryWait, domainRouterJob.StatusRunning:
		// Sudah dalam antrean — kembalikan apa adanya (idempoten).
	default:
		return nil, response.MapDomainError(fmt.Errorf("%w: job %s sudah %s",
			domainRouterJob.ErrInvalidTransition, job.ID, job.Status))
	}
	return connect.NewResponse(&devicepb.RetryRouterJobResponse{Job: toProtoJob(job)}), nil
}

// ExportRouterJobScript menghasilkan file .rsc fallback untuk dieksekusi manual.
func (h *ConnectHandler) ExportRouterJobScript(ctx context.Context, req *connect.Request[devicepb.ExportRouterJobScriptRequest]) (*connect.Response[devicepb.ExportRouterJobScriptResponse], error) {
	if h.exporter == nil {
		return nil, response.Unavailable("router job exporter unavailable")
	}
	script, err := h.exporter.ExportScript(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ExportRouterJobScriptResponse{
		Filename: req.Msg.Id + ".rsc",
		Script:   script,
	}), nil
}
