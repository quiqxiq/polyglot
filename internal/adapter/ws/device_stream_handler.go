package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

type TerminalMessage struct {
	DeviceID  string `json:"device_id"`
	InputData string `json:"input_data"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
}

type TerminalHandler struct {
	driverGetter   func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
	targetResolver func(ctx context.Context, deviceID string) (*device.Target, error)
}

func NewTerminalHandler(
	driverGetter func(ctx context.Context, deviceID string) (port.DeviceDriver, error),
	targetResolver func(ctx context.Context, deviceID string) (*device.Target, error),
) *TerminalHandler {
	return &TerminalHandler{
		driverGetter:   driverGetter,
		targetResolver: targetResolver,
	}
}

func (h *TerminalHandler) ServeHTTP(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device id is required"})
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	ctx := c.Request.Context()

	// 1. Resolve Real Device Connection Target (Host, Port, User, Password)
	var target *device.Target
	if h.targetResolver != nil {
		if t, err := h.targetResolver(ctx, deviceID); err == nil && t != nil {
			target = t
		}
	}

	// 2. Try Opening Real Direct SSH PTY Session to the Target Host
	if target != nil && target.Host != "" {
		sshPort := "22"
		if target.Port > 0 && target.Port != 8728 {
			sshPort = strconv.Itoa(target.Port)
		}
		if extraPort, ok := target.Extra["ssh_port"]; ok && extraPort != "" {
			sshPort = extraPort
		}

		user := target.Username
		if user == "" {
			user = "admin"
		}
		pass := target.Password

		_ = conn.Write(ctx, websocket.MessageText, []byte(fmt.Sprintf("\r\n\x1b[38;5;39m[Polyglot Engine] Dialing SSH PTY to %s@%s:%s...\x1b[0m\r\n", user, target.Host, sshPort)))

		sshConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.Password(pass),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         7 * time.Second,
		}

		sshClient, dialErr := ssh.Dial("tcp", net.JoinHostPort(target.Host, sshPort), sshConfig)
		if dialErr == nil {
			defer sshClient.Close()

			session, sessErr := sshClient.NewSession()
			if sessErr == nil {
				defer session.Close()

				modes := ssh.TerminalModes{
					ssh.ECHO:          1,
					ssh.TTY_OP_ISPEED: 14400,
					ssh.TTY_OP_OSPEED: 14400,
				}

				if ptyErr := session.RequestPty("xterm-256color", 35, 120, modes); ptyErr == nil {
					stdin, stdinErr := session.StdinPipe()
					stdout, stdoutErr := session.StdoutPipe()

					if stdinErr == nil && stdoutErr == nil {
						if shellErr := session.Shell(); shellErr == nil {
							_ = conn.Write(ctx, websocket.MessageText, []byte("\x1b[32m[SSH Connected - Real RouterOS PTY Active]\x1b[0m\r\n\r\n"))

							errChan := make(chan error, 2)

							// Goroutine 1: Read stdout from real SSH PTY -> Write directly to WebSocket (xterm.js)
							go func() {
								buf := make([]byte, 4096)
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
											_ = session.WindowChange(msg.Rows, msg.Cols)
										}
									} else {
										_, _ = stdin.Write(msgBytes)
									}
								}
							}()

							select {
							case <-ctx.Done():
							case <-errChan:
							}
							return
						}
					}
				}
			}
		}

		_ = conn.Write(ctx, websocket.MessageText, []byte(fmt.Sprintf("\r\n\x1b[31m[SSH Connection Failed]: %v\x1b[0m\r\n\x1b[33m[Notice]: Ensure SSH service is enabled on target device (%s:%s) with user '%s'.\x1b[0m\r\n", dialErr, target.Host, sshPort, user)))
		return
	}

	// Fallback error if target is not found
	_ = conn.Write(ctx, websocket.MessageText, []byte(fmt.Sprintf("\r\n\x1b[31m[Polyglot Terminal Error]: Target details for device ID '%s' not found in inventory.\x1b[0m\r\n", deviceID)))
}
