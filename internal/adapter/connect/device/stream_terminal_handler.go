package device

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

// StreamTerminal establishes an interactive bidirectional PTY terminal session.
func (h *DeviceConnectHandler) StreamTerminal(
	ctx context.Context,
	stream *connect.BidiStream[devicepb.TerminalFrame, devicepb.TerminalFrame],
) error {
	firstFrame, err := stream.Receive()
	if err != nil {
		return err
	}
	if firstFrame.DeviceId == "" {
		return response.InvalidArgument("device_id is required in the initial frame")
	}

	if h.openTermUC == nil {
		return response.Unavailable("open terminal use case is not initialized")
	}

	cols := int(firstFrame.Cols)
	rows := int(firstFrame.Rows)
	if cols <= 0 {
		cols = 120
	}
	if rows <= 0 {
		rows = 35
	}

	session, err := h.openTermUC.Execute(ctx, firstFrame.DeviceId, cols, rows)
	if err != nil {
		logger.WithComponent("DeviceConnectHandler").WithError(err).WithField("device_id", firstFrame.DeviceId).Error("SSH PTY connection failed")
		return response.MapDomainError(fault.Wrap(fault.KindUnavailable, err))
	}
	defer func() { _ = session.Close() }()

	errChan := make(chan error, 2)

	// Goroutine 1: PTY stdout -> Client stream
	go func() {
		buf := make([]byte, 4096)
		stdout := session.Stdout()
		for {
			n, rErr := stdout.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				if sErr := stream.Send(&devicepb.TerminalFrame{
					DeviceId:   firstFrame.DeviceId,
					OutputData: chunk,
				}); sErr != nil {
					errChan <- sErr
					return
				}
			}
			if rErr != nil {
				errChan <- rErr
				return
			}
		}
	}()

	// Goroutine 2: Client stream -> PTY stdin
	go func() {
		stdin := session.Stdin()
		for {
			req, rErr := stream.Receive()
			if rErr != nil {
				errChan <- rErr
				return
			}
			if len(req.InputData) > 0 {
				if _, wErr := stdin.Write(req.InputData); wErr != nil {
					errChan <- wErr
					return
				}
			}
			if req.Cols > 0 && req.Rows > 0 {
				// best-effort: gagal resize tidak menghentikan aliran data terminal.
				_ = session.Resize(int(req.Cols), int(req.Rows))
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
