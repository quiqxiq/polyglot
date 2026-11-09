package connect_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/pkg/logger"
)

func TestRequestIDInterceptorUnary(t *testing.T) {
	opts := iconnect.DefaultHandlerOptions()
	var capturedID string

	handler := connect.NewUnaryHandler(
		"/polyglot.v1.RouterJobService/ListRouterJobs",
		func(ctx context.Context, _ *connect.Request[polyglotv1.ListRouterJobsRequest]) (*connect.Response[polyglotv1.ListRouterJobsResponse], error) {
			capturedID = logger.RequestIDFromContext(ctx)
			return connect.NewResponse(&polyglotv1.ListRouterJobsResponse{}), nil
		},
		opts...,
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.RouterJobService/ListRouterJobs", strings.NewReader("{}"))
	req.Header.Set("X-Request-ID", "custom-req-id-123")
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
	if capturedID != "custom-req-id-123" {
		t.Errorf("captured request ID in context = %q, want custom-req-id-123", capturedID)
	}
	if got := rec.Header().Get("X-Request-ID"); got != "custom-req-id-123" {
		t.Errorf("response X-Request-ID header = %q, want custom-req-id-123", got)
	}
}

func TestRequestIDInterceptorUnaryGeneratesID(t *testing.T) {
	opts := iconnect.DefaultHandlerOptions()
	var capturedID string

	handler := connect.NewUnaryHandler(
		"/polyglot.v1.RouterJobService/ListRouterJobs",
		func(ctx context.Context, _ *connect.Request[polyglotv1.ListRouterJobsRequest]) (*connect.Response[polyglotv1.ListRouterJobsResponse], error) {
			capturedID = logger.RequestIDFromContext(ctx)
			return connect.NewResponse(&polyglotv1.ListRouterJobsResponse{}), nil
		},
		opts...,
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.RouterJobService/ListRouterJobs", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
	if capturedID == "" {
		t.Error("expected generated request ID in context, got empty")
	}
	if got := rec.Header().Get("X-Request-ID"); got != capturedID {
		t.Errorf("response X-Request-ID = %q, want context ID = %q", got, capturedID)
	}
}
