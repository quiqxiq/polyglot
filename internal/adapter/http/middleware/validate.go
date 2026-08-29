package middleware

import (
	"connectrpc.com/connect"
	"connectrpc.com/validate"
)

// NewValidationInterceptor creates a ConnectRPC interceptor that automatically
// validates all incoming requests against protovalidate constraints defined in .proto files.
// When validation fails, it returns a ConnectError with CodeInvalidArgument.
func NewValidationInterceptor() connect.UnaryInterceptorFunc {
	interceptor := validate.NewInterceptor()
	return interceptor.WrapUnary
}
