package devicepb

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/status"
)

type jsonCodec struct{}

func (jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return "json"
}

func init() {
	encoding.RegisterCodec(jsonCodec{})
}

// Device represents a network router/device inventory record in Protobuf/gRPC format.
type Device struct {
	Id             string   `json:"id"`
	TenantId       string   `json:"tenant_id"`
	Name           string   `json:"name"`
	Vendor         string   `json:"vendor"`
	DriverType     string   `json:"driver_type"`
	Host           string   `json:"host"`
	Port           int32    `json:"port"`
	SshPort        int32    `json:"ssh_port"`
	TimeoutMs      int32    `json:"timeout_ms"`
	PollIntervalMs int32    `json:"poll_interval_ms"`
	Tags           []string `json:"tags"`
	Enabled        bool     `json:"enabled"`
}

type ListDevicesRequest struct{}

type ListDevicesResponse struct {
	Devices []*Device `json:"devices"`
}

type GetDeviceRequest struct {
	Id string `json:"id"`
}

type GetDeviceResponse struct {
	Device *Device `json:"device"`
}

type UpdateDeviceRequest struct {
	Device   *Device `json:"device"`
	Username string  `json:"username"`
	Password string  `json:"password"`
}

type UpdateDeviceResponse struct {
	Device  *Device `json:"device"`
	Message string  `json:"message"`
}

type DeleteDeviceRequest struct {
	Id string `json:"id"`
}

type DeleteDeviceResponse struct {
	Message string `json:"message"`
}

type TestDeviceConnectionRequest struct {
	Id string `json:"id"`
}

type TestDeviceConnectionResponse struct {
	DeviceId  string `json:"device_id"`
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Uptime    string `json:"uptime"`
	Version   string `json:"version"`
	BoardName string `json:"board_name"`
	Identity  string `json:"identity"`
	Message   string `json:"message"`
	Success   bool   `json:"success"`
}

type StreamDeviceStatusRequest struct {
	Id string `json:"id"`
}

type DeviceTestMetrics struct {
	DeviceId  string `json:"device_id"`
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Uptime    string `json:"uptime"`
	Version   string `json:"version"`
	BoardName string `json:"board_name"`
	Identity  string `json:"identity"`
	Message   string `json:"message"`
}

type DeviceStatusFrame struct {
	Device *Device            `json:"device"`
	Test   *DeviceTestMetrics `json:"test"`
}

type TerminalFrame struct {
	DeviceId   string `json:"device_id"`
	InputData  []byte `json:"input_data"`
	OutputData []byte `json:"output_data"`
	Cols       int32  `json:"cols"`
	Rows       int32  `json:"rows"`
}

// DeviceServiceServer is the server API for DeviceService service.
type DeviceServiceServer interface {
	ListDevices(context.Context, *ListDevicesRequest) (*ListDevicesResponse, error)
	GetDevice(context.Context, *GetDeviceRequest) (*GetDeviceResponse, error)
	UpdateDevice(context.Context, *UpdateDeviceRequest) (*UpdateDeviceResponse, error)
	StreamDeviceStatus(*StreamDeviceStatusRequest, DeviceService_StreamDeviceStatusServer) error
}

type DeviceService_StreamDeviceStatusServer interface {
	Send(*DeviceStatusFrame) error
	grpc.ServerStream
}

type deviceServiceStreamDeviceStatusServer struct {
	grpc.ServerStream
}

func (x *deviceServiceStreamDeviceStatusServer) Send(m *DeviceStatusFrame) error {
	return x.ServerStream.SendMsg(m)
}

// UnimplementedDeviceServiceServer can be embedded to have forward compatible implementations.
type UnimplementedDeviceServiceServer struct{}

func (UnimplementedDeviceServiceServer) ListDevices(context.Context, *ListDevicesRequest) (*ListDevicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDevices not implemented")
}

func (UnimplementedDeviceServiceServer) GetDevice(context.Context, *GetDeviceRequest) (*GetDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDevice not implemented")
}

func (UnimplementedDeviceServiceServer) UpdateDevice(context.Context, *UpdateDeviceRequest) (*UpdateDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDevice not implemented")
}

func (UnimplementedDeviceServiceServer) StreamDeviceStatus(*StreamDeviceStatusRequest, DeviceService_StreamDeviceStatusServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamDeviceStatus not implemented")
}

// RegisterDeviceServiceServer registers a service implementation with a gRPC server.
func RegisterDeviceServiceServer(s *grpc.Server, srv DeviceServiceServer) {
	s.RegisterService(&DeviceService_ServiceDesc, srv)
}

var DeviceService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "polyglot.v1.DeviceService",
	HandlerType: (*DeviceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListDevices",
			Handler:    _DeviceService_ListDevices_Handler,
		},
		{
			MethodName: "GetDevice",
			Handler:    _DeviceService_GetDevice_Handler,
		},
		{
			MethodName: "UpdateDevice",
			Handler:    _DeviceService_UpdateDevice_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamDeviceStatus",
			Handler:       _DeviceService_StreamDeviceStatus_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "api/proto/v1/device.proto",
}

func _DeviceService_ListDevices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListDevicesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceServiceServer).ListDevices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/polyglot.v1.DeviceService/ListDevices",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceServiceServer).ListDevices(ctx, req.(*ListDevicesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceService_GetDevice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDeviceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceServiceServer).GetDevice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/polyglot.v1.DeviceService/GetDevice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceServiceServer).GetDevice(ctx, req.(*GetDeviceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceService_UpdateDevice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateDeviceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceServiceServer).UpdateDevice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/polyglot.v1.DeviceService/UpdateDevice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceServiceServer).UpdateDevice(ctx, req.(*UpdateDeviceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceService_StreamDeviceStatus_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamDeviceStatusRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DeviceServiceServer).StreamDeviceStatus(m, &deviceServiceStreamDeviceStatusServer{stream})
}

// DeviceServiceClient is the client API for DeviceService service.
type DeviceServiceClient interface {
	ListDevices(ctx context.Context, in *ListDevicesRequest, opts ...grpc.CallOption) (*ListDevicesResponse, error)
	GetDevice(ctx context.Context, in *GetDeviceRequest, opts ...grpc.CallOption) (*GetDeviceResponse, error)
	UpdateDevice(ctx context.Context, in *UpdateDeviceRequest, opts ...grpc.CallOption) (*UpdateDeviceResponse, error)
	StreamDeviceStatus(ctx context.Context, in *StreamDeviceStatusRequest, opts ...grpc.CallOption) (DeviceService_StreamDeviceStatusClient, error)
}

type deviceServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDeviceServiceClient(cc grpc.ClientConnInterface) DeviceServiceClient {
	return &deviceServiceClient{cc}
}

func (c *deviceServiceClient) ListDevices(ctx context.Context, in *ListDevicesRequest, opts ...grpc.CallOption) (*ListDevicesResponse, error) {
	out := new(ListDevicesResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.DeviceService/ListDevices", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceServiceClient) GetDevice(ctx context.Context, in *GetDeviceRequest, opts ...grpc.CallOption) (*GetDeviceResponse, error) {
	out := new(GetDeviceResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.DeviceService/GetDevice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceServiceClient) UpdateDevice(ctx context.Context, in *UpdateDeviceRequest, opts ...grpc.CallOption) (*UpdateDeviceResponse, error) {
	out := new(UpdateDeviceResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.DeviceService/UpdateDevice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type DeviceService_StreamDeviceStatusClient interface {
	Recv() (*DeviceStatusFrame, error)
	grpc.ClientStream
}

type deviceServiceStreamDeviceStatusClient struct {
	grpc.ClientStream
}

func (x *deviceServiceStreamDeviceStatusClient) Recv() (*DeviceStatusFrame, error) {
	m := new(DeviceStatusFrame)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *deviceServiceClient) StreamDeviceStatus(ctx context.Context, in *StreamDeviceStatusRequest, opts ...grpc.CallOption) (DeviceService_StreamDeviceStatusClient, error) {
	stream, err := c.cc.NewStream(ctx, &DeviceService_ServiceDesc.Streams[0], "/polyglot.v1.DeviceService/StreamDeviceStatus", opts...)
	if err != nil {
		return nil, err
	}
	x := &deviceServiceStreamDeviceStatusClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}
