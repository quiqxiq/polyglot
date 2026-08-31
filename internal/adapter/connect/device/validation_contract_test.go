package device_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	validatepb "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"connectrpc.com/connect"
	"connectrpc.com/validate"
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	connectDevice "github.com/quixiq/polyglot/internal/adapter/connect/device"
)

func TestValidationInterceptorPreservesViolationDetails(t *testing.T) {
	interceptor := validate.NewInterceptor()
	handler := interceptor.WrapUnary(func(_ context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		return connect.NewResponse(&devicepb.GetDeviceResponse{}), nil
	})
	_, err := handler(context.Background(), connect.NewRequest(&devicepb.GetDeviceRequest{}))
	if err == nil {
		t.Fatal("validation interceptor error = nil, want invalid request")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("validation error type = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("validation error code = %s, want %s", connectErr.Code(), connect.CodeInvalidArgument)
	}
	if len(connectErr.Details()) != 1 {
		t.Fatalf("validation error details = %d, want 1", len(connectErr.Details()))
	}
	detail, err := connectErr.Details()[0].Value()
	if err != nil {
		t.Fatalf("decode validation detail: %v", err)
	}
	violations, ok := detail.(*validatepb.Violations)
	if !ok || len(violations.Violations) == 0 {
		t.Fatalf("validation detail = %T, want non-empty *validatepb.Violations", detail)
	}
}

func TestDeviceServiceValidationRejectsInvalidUnaryRequest(t *testing.T) {
	_, handler := connectDevice.NewDeviceServiceHandler(nil, nil, nil, nil, nil)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/polyglot.v1.DeviceService/GetDevice",
		bytes.NewBufferString(`{"id":""}`),
	)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	if got, want := recorder.Code, http.StatusBadRequest; got != want {
		t.Errorf("GetDevice invalid request status = %d, want %d", got, want)
	}
}

func TestDeviceServiceValidationRejectsInvalidStreamingRequest(t *testing.T) {
	_, handler := connectDevice.NewDeviceServiceHandler(nil, nil, nil, nil, nil)

	recorder := httptest.NewRecorder()
	payload := []byte(`{"id":""}`)
	body := make([]byte, 5+len(payload))
	body[0] = 0
	binary.BigEndian.PutUint32(body[1:5], uint32(len(payload)))
	copy(body[5:], payload)
	req := httptest.NewRequest(
		http.MethodPost,
		"/polyglot.v1.DeviceService/StreamPing",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/connect+json")
	req.Header.Set("Connect-Protocol-Version", "1")

	handler.ServeHTTP(recorder, req)

	if got, want := recorder.Code, http.StatusOK; got != want {
		t.Errorf("StreamPing invalid request status = %d, want %d", got, want)
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"code":"invalid_argument"`)) {
		t.Errorf("StreamPing invalid response = %s, want invalid_argument stream error", recorder.Body.Bytes())
	}
}
