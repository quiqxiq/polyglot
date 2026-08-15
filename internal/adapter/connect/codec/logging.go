package codec

import (
	"context"
	"time"

	"connectrpc.com/connect"

	"github.com/quixiq/polyglot/pkg/logger"
)

// LoggingInterceptor returns a ConnectRPC interceptor that logs procedure executions with Logrus.
func LoggingInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			procedure := req.Spec().Procedure

			entry := logger.FromContext(ctx).WithFields(logger.Fields{
				"rpc_procedure":   procedure,
				"rpc_stream_type": req.Spec().StreamType.String(),
			})

			resp, err := next(ctx, req)
			duration := time.Since(start)

			if err != nil {
				entry.WithFields(logger.Fields{
					"duration_ms": duration.Milliseconds(),
					"error":       err,
				}).Error("ConnectRPC procedure failed")
			} else {
				entry.WithFields(logger.Fields{
					"duration_ms": duration.Milliseconds(),
				}).Info("ConnectRPC procedure completed")
			}

			return resp, err
		}
	}
}
