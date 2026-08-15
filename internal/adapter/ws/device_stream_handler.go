package ws

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coder/websocket"

	"github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

type TerminalMessage struct {
	DeviceID  string `json:"device_id"`
	InputData string `json:"input_data"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
}

type TerminalHandler struct {
	openTermUC *network.OpenTerminalUseCase
}

func NewTerminalHandler(openTermUC *network.OpenTerminalUseCase) *TerminalHandler {
	return &TerminalHandler{
		openTermUC: openTermUC,
	}
}

func (h *TerminalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("id")
	if deviceID == "" {
		response.Fail(w, http.StatusBadRequest, "BAD_REQUEST", "Device ID is required in URL path", "")
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		logger.FromContext(r.Context()).WithError(err).Warn("WebSocket handshake failed")
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()

	if h.openTermUC == nil {
		_ = conn.Write(ctx, websocket.MessageText, []byte("\r\n\x1b[31m[Polyglot Terminal Error]: Terminal usecase is not initialized.\x1b[0m\r\n"))
		return
	}

	_ = conn.Write(ctx, websocket.MessageText, []byte(fmt.Sprintf("\r\n\x1b[38;5;39m[Polyglot Engine] Opening SSH PTY session for device '%s'...\x1b[0m\r\n", deviceID)))

	termSession, err := h.openTermUC.Execute(ctx, deviceID, 120, 35)
	if err != nil {
		_ = conn.Write(ctx, websocket.MessageText, []byte(fmt.Sprintf("\r\n\x1b[31m[SSH Connection Failed]: %v\x1b[0m\r\n\x1b[33m[Notice]: Ensure target device has SSH service enabled and credentials configured.\x1b[0m\r\n", err)))
		return
	}
	defer termSession.Close()

	_ = conn.Write(ctx, websocket.MessageText, []byte("\x1b[32m[SSH Connected - Native PTY Active]\x1b[0m\r\n\r\n"))

	errChan := make(chan error, 2)

	// Goroutine 1: Read stdout from real SSH PTY -> Write directly to WebSocket (xterm.js)
	go func() {
		buf := make([]byte, 4096)
		stdout := termSession.Stdout()
		for {
			n, rErr := stdout.Read(buf)
			if n > 0 {
				if wErr := conn.Write(ctx, websocket.MessageText, buf[:n]); wErr != nil {
					errChan <- wErr
					return
				}
			}
			if rErr != nil {
				errChan <- rErr
				return
			}
		}
	}()

	// Goroutine 2: Read raw keystrokes from WebSocket (xterm.js) -> Write directly to SSH PTY stdin
	go func() {
		stdin := termSession.Stdin()
		for {
			_, msgBytes, rErr := conn.Read(ctx)
			if rErr != nil {
				errChan <- rErr
				return
			}

			var msg TerminalMessage
			if err := json.Unmarshal(msgBytes, &msg); err == nil {
				if msg.InputData != "" {
					_, _ = stdin.Write([]byte(msg.InputData))
				}
				if msg.Cols > 0 && msg.Rows > 0 {
					_ = termSession.Resize(msg.Cols, msg.Rows)
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errChan:
		logger.FromContext(ctx).WithError(err).Debug("Terminal WebSocket session closed")
	}
}
