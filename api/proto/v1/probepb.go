package devicepb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PingTelemetry struct {
	ProbeId       string `json:"probe_id"`
	TargetIp      string `json:"target_ip"`
	LatencyMs     int64  `json:"latency_ms"`
	IsAlive       bool   `json:"is_alive"`
	TimestampUnix int64  `json:"timestamp_unix"`
}

type ProbeControlCommand struct {
	CommandId string `json:"command_id"`
	Action    string `json:"action"`
	Target    string `json:"target"`
}

type ProbeStatusRequest struct {
	ProbeId       string `json:"probe_id"`
	Version       string `json:"version"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

type ProbeStatusResponse struct {
	Acknowledged   bool  `json:"acknowledged"`
	ServerTimeUnix int64 `json:"server_time_unix"`
}

type ProbeServiceClient interface {
	ReportStatus(ctx context.Context, in *ProbeStatusRequest, opts ...grpc.CallOption) (*ProbeStatusResponse, error)
	StreamTelemetry(ctx context.Context, opts ...grpc.CallOption) (ProbeService_StreamTelemetryClient, error)
}

type probeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewProbeServiceClient(cc grpc.ClientConnInterface) ProbeServiceClient {
	return &probeServiceClient{cc}
}

func (c *probeServiceClient) ReportStatus(ctx context.Context, in *ProbeStatusRequest, opts ...grpc.CallOption) (*ProbeStatusResponse, error) {
	out := new(ProbeStatusResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.ProbeService/ReportStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *probeServiceClient) StreamTelemetry(ctx context.Context, opts ...grpc.CallOption) (ProbeService_StreamTelemetryClient, error) {
	stream, err := c.cc.NewStream(ctx, &ProbeService_ServiceDesc.Streams[0], "/polyglot.v1.ProbeService/StreamTelemetry", opts...)
	if err != nil {
		return nil, err
	}
	x := &probeServiceStreamTelemetryClient{stream}
	return x, nil
}

type ProbeService_StreamTelemetryClient interface {
	Send(*PingTelemetry) error
	Recv() (*ProbeControlCommand, error)
	grpc.ClientStream
}

type probeServiceStreamTelemetryClient struct {
	grpc.ClientStream
}

func (x *probeServiceStreamTelemetryClient) Send(m *PingTelemetry) error {
	return x.ClientStream.SendMsg(m)
}

func (x *probeServiceStreamTelemetryClient) Recv() (*ProbeControlCommand, error) {
	m := new(ProbeControlCommand)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

type ProbeServiceServer interface {
	ReportStatus(context.Context, *ProbeStatusRequest) (*ProbeStatusResponse, error)
	StreamTelemetry(ProbeService_StreamTelemetryServer) error
	mustEmbedUnimplementedProbeServiceServer()
}

type UnimplementedProbeServiceServer struct{}

func (UnimplementedProbeServiceServer) ReportStatus(context.Context, *ProbeStatusRequest) (*ProbeStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportStatus not implemented")
}

func (UnimplementedProbeServiceServer) StreamTelemetry(ProbeService_StreamTelemetryServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamTelemetry not implemented")
}

func (UnimplementedProbeServiceServer) mustEmbedUnimplementedProbeServiceServer() {}

type ProbeService_StreamTelemetryServer interface {
	Send(*ProbeControlCommand) error
	Recv() (*PingTelemetry, error)
	grpc.ServerStream
}

type probeServiceStreamTelemetryServer struct {
	grpc.ServerStream
}

func (x *probeServiceStreamTelemetryServer) Send(m *ProbeControlCommand) error {
	return x.ServerStream.SendMsg(m)
}

func (x *probeServiceStreamTelemetryServer) Recv() (*PingTelemetry, error) {
	m := new(PingTelemetry)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func RegisterProbeServiceServer(s grpc.ServiceRegistrar, srv ProbeServiceServer) {
	s.RegisterService(&ProbeService_ServiceDesc, srv)
}

var ProbeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "polyglot.v1.ProbeService",
	HandlerType: (*ProbeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReportStatus",
			Handler:    _ProbeService_ReportStatus_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamTelemetry",
			Handler:       _ProbeService_StreamTelemetry_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "api/proto/v1/probe.proto",
}

func _ProbeService_ReportStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProbeStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProbeServiceServer).ReportStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/polyglot.v1.ProbeService/ReportStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProbeServiceServer).ReportStatus(ctx, req.(*ProbeStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProbeService_StreamTelemetry_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ProbeServiceServer).StreamTelemetry(&probeServiceStreamTelemetryServer{stream})
}
