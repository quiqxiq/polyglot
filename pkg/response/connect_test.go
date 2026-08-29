package response

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
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
