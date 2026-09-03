package response

import (
	"errors"
	"testing"

	"connectrpc.com/connect"

	"github.com/quixiq/polyglot/pkg/fault"
)

func TestTransportErrorsUseExpectedCodes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want connect.Code
	}{
		{name: "unavailable", err: Unavailable("dependency unavailable"), want: connect.CodeUnavailable},
		{name: "invalid argument", err: InvalidArgument("invalid request"), want: connect.CodeInvalidArgument},
		{name: "unauthenticated", err: Unauthenticated("authentication required"), want: connect.CodeUnauthenticated},
		{name: "permission denied", err: PermissionDenied("permission denied"), want: connect.CodePermissionDenied},
		{name: "unimplemented", err: Unimplemented("not supported"), want: connect.CodeUnimplemented},
		{name: "internal", err: Internal("internal failure"), want: connect.CodeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var connectErr *connect.Error
			if !errors.As(tt.err, &connectErr) {
				t.Fatalf("%s error type = %T, want *connect.Error", tt.name, tt.err)
			}
			if got := connectErr.Code(); got != tt.want {
				t.Errorf("%s code = %s, want %s", tt.name, got, tt.want)
			}
		})
	}
}

func TestMapDomainError(t *testing.T) {
	if got := MapDomainError(nil); got != nil {
		t.Errorf("MapDomainError(nil) = %v, want nil", got)
	}

	tests := []struct {
		kind fault.Kind
		want connect.Code
	}{
		{kind: fault.KindNotFound, want: connect.CodeNotFound},
		{kind: fault.KindInvalidInput, want: connect.CodeInvalidArgument},
		{kind: fault.KindAlreadyExists, want: connect.CodeAlreadyExists},
		{kind: fault.KindPermissionDenied, want: connect.CodePermissionDenied},
		{kind: fault.KindUnauthenticated, want: connect.CodeUnauthenticated},
		{kind: fault.KindFailedPrecondition, want: connect.CodeFailedPrecondition},
		{kind: fault.KindConflict, want: connect.CodeAborted},
		{kind: fault.KindUnavailable, want: connect.CodeUnavailable},
		{kind: fault.KindResourceExhausted, want: connect.CodeResourceExhausted},
		{kind: fault.KindUnknown, want: connect.CodeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			domainErr := fault.New(tt.kind, "domain error")
			mapped := MapDomainError(domainErr)

			var connectErr *connect.Error
			if !errors.As(mapped, &connectErr) {
				t.Fatalf("mapped error type = %T, want *connect.Error", mapped)
			}
			if got := connectErr.Code(); got != tt.want {
				t.Errorf("kind %s mapped to %s, want %s", tt.kind, got, tt.want)
			}
		})
	}
}
