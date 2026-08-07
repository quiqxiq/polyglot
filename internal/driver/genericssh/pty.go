package genericssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// SSHPtySession implements port.TerminalSession over a real SSH PTY connection.
type SSHPtySession struct {
	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader

	closeOnce sync.Once
}

var _ port.TerminalSession = (*SSHPtySession)(nil)

func (s *SSHPtySession) Stdin() io.Writer {
	return s.stdin
}

func (s *SSHPtySession) Stdout() io.Reader {
	return s.stdout
}

func (s *SSHPtySession) Resize(cols, rows int) error {
	if s.session == nil {
		return fmt.Errorf("genericssh: ssh session is nil")
	}
	return s.session.WindowChange(rows, cols)
}

func (s *SSHPtySession) Close() error {
	var errs []error
	s.closeOnce.Do(func() {
		if s.session != nil {
			if err := s.session.Close(); err != nil && err != io.EOF {
				errs = append(errs, err)
			}
		}
		if s.client != nil {
			if err := s.client.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	})
	if len(errs) > 0 {
		return fmt.Errorf("genericssh: close ssh pty session: %v", errs)
	}
	return nil
}

// DialSSHPty connects to target over SSH, requests an interactive PTY session, and returns a port.TerminalSession.
func DialSSHPty(ctx context.Context, target device.Target, cols, rows int) (port.TerminalSession, error) {
	if target.Host == "" {
		return nil, fmt.Errorf("genericssh: target host is empty")
	}

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

	timeout := target.Timeout
	if timeout <= 0 {
		timeout = 7 * time.Second
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	addr := net.JoinHostPort(target.Host, sshPort)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("genericssh: dial ssh %s@%s: %w", user, addr, err)
	}

	session, err := client.NewSession()
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("genericssh: create ssh session: %w", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if cols <= 0 {
		cols = 120
	}
	if rows <= 0 {
		rows = 35
	}

	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		_ = session.Close()
		_ = client.Close()
		return nil, fmt.Errorf("genericssh: request pty: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		_ = session.Close()
		_ = client.Close()
		return nil, fmt.Errorf("genericssh: stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		_ = session.Close()
		_ = client.Close()
		return nil, fmt.Errorf("genericssh: stdout pipe: %w", err)
	}

	if err := session.Shell(); err != nil {
		_ = session.Close()
		_ = client.Close()
		return nil, fmt.Errorf("genericssh: start shell: %w", err)
	}

	return &SSHPtySession{
		client:  client,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
	}, nil
}
