package devicepb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MikhmonDashboardSummary struct {
	TotalHotspotUsers int32   `json:"total_hotspot_users"`
	TotalActiveUsers  int32   `json:"total_active_users"`
	TodayIncome       float64 `json:"today_income"`
	Uptime            string  `json:"uptime"`
}

type MikhmonProfile struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	SharedUsers  string  `json:"shared_users"`
	RateLimit    string  `json:"rate_limit"`
	ModeExpire   string  `json:"mode_expire"`
	Validity     string  `json:"validity"`
	Price        float64 `json:"price"`
	SellingPrice float64 `json:"selling_price"`
	LockUser     string  `json:"lock_user"`
	ParentQueue  string  `json:"parent_queue"`
	Comment      string  `json:"comment"`
}

type MikhmonUser struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Password    string `json:"password"`
	Profile     string `json:"profile"`
	LimitUptime string `json:"limit_uptime"`
	LimitBytes  string `json:"limit_bytes"`
	Uptime      string `json:"uptime"`
	BytesIn     string `json:"bytes_in"`
	BytesOut    string `json:"bytes_out"`
	Comment     string `json:"comment"`
	Disabled    bool   `json:"disabled"`
}

type MikhmonActiveSession struct {
	Id         string `json:"id"`
	Server     string `json:"server"`
	User       string `json:"user"`
	Address    string `json:"address"`
	MacAddress string `json:"mac_address"`
	Uptime     string `json:"uptime"`
	BytesIn    string `json:"bytes_in"`
	BytesOut   string `json:"bytes_out"`
	Comment    string `json:"comment"`
}

type MikhmonDHCPLease struct {
	Id         string `json:"id"`
	Address    string `json:"address"`
	MacAddress string `json:"mac_address"`
	HostName   string `json:"host_name"`
	Status     string `json:"status"`
	Blocked    bool   `json:"blocked"`
	Comment    string `json:"comment"`
}

type GetMikhmonDashboardRequest struct {
	DeviceId string `json:"device_id"`
}

type GetMikhmonDashboardResponse struct {
	Summary *MikhmonDashboardSummary `json:"summary"`
}

type ListMikhmonProfilesRequest struct {
	DeviceId string `json:"device_id"`
}

type ListMikhmonProfilesResponse struct {
	Profiles []*MikhmonProfile `json:"profiles"`
}

type ListMikhmonUsersRequest struct {
	DeviceId string `json:"device_id"`
	Profile  string `json:"profile"`
}

type ListMikhmonUsersResponse struct {
	Users []*MikhmonUser `json:"users"`
}

type ListMikhmonActiveSessionsRequest struct {
	DeviceId string `json:"device_id"`
}

type ListMikhmonActiveSessionsResponse struct {
	Sessions []*MikhmonActiveSession `json:"sessions"`
}

type KickMikhmonSessionRequest struct {
	DeviceId string `json:"device_id"`
	RosId    string `json:"ros_id"`
}

type KickMikhmonSessionResponse struct {
	Message string `json:"message"`
}

type ListMikhmonDHCPLeasesRequest struct {
	DeviceId  string `json:"device_id"`
	MacFilter string `json:"mac_filter"`
}

type ListMikhmonDHCPLeasesResponse struct {
	Leases []*MikhmonDHCPLease `json:"leases"`
}

type BlockDHCPLeaseRequest struct {
	DeviceId string `json:"device_id"`
	RosId    string `json:"ros_id"`
	Blocked  bool   `json:"blocked"`
	Comment  string `json:"comment"`
}

type BlockDHCPLeaseResponse struct {
	Message string `json:"message"`
}

type GenerateVouchersRequest struct {
	DeviceId     string `json:"device_id"`
	Profile      string `json:"profile"`
	Count        int32  `json:"count"`
	UserType     string `json:"user_type"`
	UserLength   int32  `json:"user_length"`
	Prefix       string `json:"prefix"`
	CharacterSet string `json:"character_set"`
}

type GenerateVouchersResponse struct {
	Vouchers []*MikhmonUser `json:"vouchers"`
	Message  string         `json:"message"`
}

type StreamTrafficRequest struct {
	DeviceId  string `json:"device_id"`
	Interface string `json:"interface"`
}

type TrafficStreamData struct {
	DeviceId      string `json:"device_id"`
	Interface     string `json:"interface"`
	RxBps         int64  `json:"rx_bps"`
	TxBps         int64  `json:"tx_bps"`
	TimestampUnix int64  `json:"timestamp_unix"`
}

type StreamResourceRequest struct {
	DeviceId string `json:"device_id"`
}

type ResourceStreamData struct {
	DeviceId      string `json:"device_id"`
	CpuLoad       int32  `json:"cpu_load"`
	FreeMemory    string `json:"free_memory"`
	Uptime        string `json:"uptime"`
	TimestampUnix int64  `json:"timestamp_unix"`
}

type StreamActiveSessionsRequest struct {
	DeviceId   string `json:"device_id"`
	UserFilter string `json:"user_filter"`
}

type ActiveSessionsStreamData struct {
	DeviceId      string                  `json:"device_id"`
	Sessions      []*MikhmonActiveSession `json:"sessions"`
	TimestampUnix int64                   `json:"timestamp_unix"`
}

type MikhmonServiceServer interface {
	GetDashboard(context.Context, *GetMikhmonDashboardRequest) (*GetMikhmonDashboardResponse, error)
	ListProfiles(context.Context, *ListMikhmonProfilesRequest) (*ListMikhmonProfilesResponse, error)
	ListUsers(context.Context, *ListMikhmonUsersRequest) (*ListMikhmonUsersResponse, error)
	ListActiveSessions(context.Context, *ListMikhmonActiveSessionsRequest) (*ListMikhmonActiveSessionsResponse, error)
	KickActiveSession(context.Context, *KickMikhmonSessionRequest) (*KickMikhmonSessionResponse, error)
	ListDHCPLeases(context.Context, *ListMikhmonDHCPLeasesRequest) (*ListMikhmonDHCPLeasesResponse, error)
	BlockDHCPLease(context.Context, *BlockDHCPLeaseRequest) (*BlockDHCPLeaseResponse, error)
	GenerateVouchers(context.Context, *GenerateVouchersRequest) (*GenerateVouchersResponse, error)
}

type UnimplementedMikhmonServiceServer struct{}

func (UnimplementedMikhmonServiceServer) GetDashboard(context.Context, *GetMikhmonDashboardRequest) (*GetMikhmonDashboardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDashboard not implemented")
}

func (UnimplementedMikhmonServiceServer) ListProfiles(context.Context, *ListMikhmonProfilesRequest) (*ListMikhmonProfilesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListProfiles not implemented")
}

func (UnimplementedMikhmonServiceServer) ListUsers(context.Context, *ListMikhmonUsersRequest) (*ListMikhmonUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUsers not implemented")
}

func (UnimplementedMikhmonServiceServer) ListActiveSessions(context.Context, *ListMikhmonActiveSessionsRequest) (*ListMikhmonActiveSessionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListActiveSessions not implemented")
}

func (UnimplementedMikhmonServiceServer) KickActiveSession(context.Context, *KickMikhmonSessionRequest) (*KickMikhmonSessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method KickActiveSession not implemented")
}

func (UnimplementedMikhmonServiceServer) ListDHCPLeases(context.Context, *ListMikhmonDHCPLeasesRequest) (*ListMikhmonDHCPLeasesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDHCPLeases not implemented")
}

func (UnimplementedMikhmonServiceServer) BlockDHCPLease(context.Context, *BlockDHCPLeaseRequest) (*BlockDHCPLeaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BlockDHCPLease not implemented")
}

func (UnimplementedMikhmonServiceServer) GenerateVouchers(context.Context, *GenerateVouchersRequest) (*GenerateVouchersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateVouchers not implemented")
}

func RegisterMikhmonServiceServer(s *grpc.Server, srv MikhmonServiceServer) {
	s.RegisterService(&MikhmonService_ServiceDesc, srv)
}

var MikhmonService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "polyglot.v1.MikhmonService",
	HandlerType: (*MikhmonServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "api/proto/v1/mikhmon.proto",
}
