package connect

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

type requestIDInterceptor struct{}

func (requestIDInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		requestID := requestID(req.Header())
		response, err := next(ctx, req)
		if response != nil {
			response.Header().Set(requestIDHeader, requestID)
		}
		return response, err
	}
}

func (requestIDInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (requestIDInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, stream connect.StreamingHandlerConn) error {
		requestID := requestID(stream.RequestHeader())
		stream.ResponseHeader().Set(requestIDHeader, requestID)
		return next(ctx, stream)
	}
}

func requestID(header http.Header) string {
	value := strings.TrimSpace(header.Get(requestIDHeader))
	if value == "" || len(value) > 128 {
		return uuid.NewString()
	}
	return value
}
