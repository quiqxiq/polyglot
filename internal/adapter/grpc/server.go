package grpc

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/usecase/business"
)

// Server encapsulates the gRPC server instance.
type Server struct {
	grpcServer   *grpc.Server
	deviceServer *DeviceServer
	listener     net.Listener
}

// NewServer constructs a new gRPC Server listening on address (e.g., ":50051").
func NewServer(addr string, uc *business.ManageDeviceUseCase, streamH *ws.DeviceStreamHandler) (*Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on gRPC addr %s: %w", addr, err)
	}

	grpcSrv := grpc.NewServer()
	devSrv := NewDeviceServer(uc, streamH)

	devicepb.RegisterDeviceServiceServer(grpcSrv, devSrv)

	return &Server{
		grpcServer:   grpcSrv,
		deviceServer: devSrv,
		listener:     lis,
	}, nil
}

// Start begins serving gRPC requests in a background goroutine.
func (s *Server) Start() {
	go func() {
		log.Printf("[gRPC Adapter] Server listening on %s", s.listener.Addr().String())
		if err := s.grpcServer.Serve(s.listener); err != nil {
			log.Printf("[gRPC Adapter Error] Serve stopped: %v", err)
		}
	}()
}

// Stop gracefully stops the gRPC server.
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
	log.Printf("[gRPC Adapter] Server stopped gracefully")
}
