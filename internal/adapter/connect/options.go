package connect

import (
	"connectrpc.com/connect"
	"connectrpc.com/validate"
)

// DefaultHandlerOptions returns standard ConnectRPC handler options:
// JSONCodec with protojson support and protovalidate automatic request validation.
func DefaultHandlerOptions() []connect.HandlerOption {
	validateInterceptor := validate.NewInterceptor()
	return []connect.HandlerOption{
		connect.WithCodec(JSONCodec()),
		connect.WithInterceptors(validateInterceptor, requestIDInterceptor{}),
	}
}
