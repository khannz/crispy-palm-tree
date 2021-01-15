// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package lbos_t1_ipvs

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// IPVSGetWokerClient is the client API for IPVSGetWoker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IPVSGetWokerClient interface {
	NewIPVSService(ctx context.Context, in *PbGetIPVSServiceData, opts ...grpc.CallOption) (*EmptyGetIPVSData, error)
	AddIPVSApplicationServersForService(ctx context.Context, in *PbGetIPVSServiceData, opts ...grpc.CallOption) (*EmptyGetIPVSData, error)
	RemoveIPVSService(ctx context.Context, in *PbGetIPVSServiceData, opts ...grpc.CallOption) (*EmptyGetIPVSData, error)
	RemoveIPVSApplicationServersFromService(ctx context.Context, in *PbGetIPVSServiceData, opts ...grpc.CallOption) (*EmptyGetIPVSData, error)
	GetIPVSRuntime(ctx context.Context, in *EmptyGetIPVSData, opts ...grpc.CallOption) (*PbGetIPVSRawServicesData, error)
}

type iPVSGetWokerClient struct {
	cc grpc.ClientConnInterface
}

func NewIPVSGetWokerClient(cc grpc.ClientConnInterface) IPVSGetWokerClient {
	return &iPVSGetWokerClient{cc}
}

func (c *iPVSGetWokerClient) NewIPVSService(ctx context.Context, in *PbGetIPVSServiceData, opts ...grpc.CallOption) (*EmptyGetIPVSData, error) {
	out := new(EmptyGetIPVSData)
	err := c.cc.Invoke(ctx, "/lbos.t1.ipvs.IPVSGetWoker/NewIPVSService", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iPVSGetWokerClient) AddIPVSApplicationServersForService(ctx context.Context, in *PbGetIPVSServiceData, opts ...grpc.CallOption) (*EmptyGetIPVSData, error) {
	out := new(EmptyGetIPVSData)
	err := c.cc.Invoke(ctx, "/lbos.t1.ipvs.IPVSGetWoker/AddIPVSApplicationServersForService", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iPVSGetWokerClient) RemoveIPVSService(ctx context.Context, in *PbGetIPVSServiceData, opts ...grpc.CallOption) (*EmptyGetIPVSData, error) {
	out := new(EmptyGetIPVSData)
	err := c.cc.Invoke(ctx, "/lbos.t1.ipvs.IPVSGetWoker/RemoveIPVSService", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iPVSGetWokerClient) RemoveIPVSApplicationServersFromService(ctx context.Context, in *PbGetIPVSServiceData, opts ...grpc.CallOption) (*EmptyGetIPVSData, error) {
	out := new(EmptyGetIPVSData)
	err := c.cc.Invoke(ctx, "/lbos.t1.ipvs.IPVSGetWoker/RemoveIPVSApplicationServersFromService", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iPVSGetWokerClient) GetIPVSRuntime(ctx context.Context, in *EmptyGetIPVSData, opts ...grpc.CallOption) (*PbGetIPVSRawServicesData, error) {
	out := new(PbGetIPVSRawServicesData)
	err := c.cc.Invoke(ctx, "/lbos.t1.ipvs.IPVSGetWoker/GetIPVSRuntime", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IPVSGetWokerServer is the server API for IPVSGetWoker service.
// All implementations must embed UnimplementedIPVSGetWokerServer
// for forward compatibility
type IPVSGetWokerServer interface {
	NewIPVSService(context.Context, *PbGetIPVSServiceData) (*EmptyGetIPVSData, error)
	AddIPVSApplicationServersForService(context.Context, *PbGetIPVSServiceData) (*EmptyGetIPVSData, error)
	RemoveIPVSService(context.Context, *PbGetIPVSServiceData) (*EmptyGetIPVSData, error)
	RemoveIPVSApplicationServersFromService(context.Context, *PbGetIPVSServiceData) (*EmptyGetIPVSData, error)
	GetIPVSRuntime(context.Context, *EmptyGetIPVSData) (*PbGetIPVSRawServicesData, error)
	mustEmbedUnimplementedIPVSGetWokerServer()
}

// UnimplementedIPVSGetWokerServer must be embedded to have forward compatible implementations.
type UnimplementedIPVSGetWokerServer struct {
}

func (UnimplementedIPVSGetWokerServer) NewIPVSService(context.Context, *PbGetIPVSServiceData) (*EmptyGetIPVSData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewIPVSService not implemented")
}
func (UnimplementedIPVSGetWokerServer) AddIPVSApplicationServersForService(context.Context, *PbGetIPVSServiceData) (*EmptyGetIPVSData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddIPVSApplicationServersForService not implemented")
}
func (UnimplementedIPVSGetWokerServer) RemoveIPVSService(context.Context, *PbGetIPVSServiceData) (*EmptyGetIPVSData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveIPVSService not implemented")
}
func (UnimplementedIPVSGetWokerServer) RemoveIPVSApplicationServersFromService(context.Context, *PbGetIPVSServiceData) (*EmptyGetIPVSData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveIPVSApplicationServersFromService not implemented")
}
func (UnimplementedIPVSGetWokerServer) GetIPVSRuntime(context.Context, *EmptyGetIPVSData) (*PbGetIPVSRawServicesData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIPVSRuntime not implemented")
}
func (UnimplementedIPVSGetWokerServer) mustEmbedUnimplementedIPVSGetWokerServer() {}

// UnsafeIPVSGetWokerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IPVSGetWokerServer will
// result in compilation errors.
type UnsafeIPVSGetWokerServer interface {
	mustEmbedUnimplementedIPVSGetWokerServer()
}

func RegisterIPVSGetWokerServer(s *grpc.Server, srv IPVSGetWokerServer) {
	s.RegisterService(&_IPVSGetWoker_serviceDesc, srv)
}

func _IPVSGetWoker_NewIPVSService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PbGetIPVSServiceData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IPVSGetWokerServer).NewIPVSService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.ipvs.IPVSGetWoker/NewIPVSService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IPVSGetWokerServer).NewIPVSService(ctx, req.(*PbGetIPVSServiceData))
	}
	return interceptor(ctx, in, info, handler)
}

func _IPVSGetWoker_AddIPVSApplicationServersForService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PbGetIPVSServiceData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IPVSGetWokerServer).AddIPVSApplicationServersForService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.ipvs.IPVSGetWoker/AddIPVSApplicationServersForService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IPVSGetWokerServer).AddIPVSApplicationServersForService(ctx, req.(*PbGetIPVSServiceData))
	}
	return interceptor(ctx, in, info, handler)
}

func _IPVSGetWoker_RemoveIPVSService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PbGetIPVSServiceData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IPVSGetWokerServer).RemoveIPVSService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.ipvs.IPVSGetWoker/RemoveIPVSService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IPVSGetWokerServer).RemoveIPVSService(ctx, req.(*PbGetIPVSServiceData))
	}
	return interceptor(ctx, in, info, handler)
}

func _IPVSGetWoker_RemoveIPVSApplicationServersFromService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PbGetIPVSServiceData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IPVSGetWokerServer).RemoveIPVSApplicationServersFromService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.ipvs.IPVSGetWoker/RemoveIPVSApplicationServersFromService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IPVSGetWokerServer).RemoveIPVSApplicationServersFromService(ctx, req.(*PbGetIPVSServiceData))
	}
	return interceptor(ctx, in, info, handler)
}

func _IPVSGetWoker_GetIPVSRuntime_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyGetIPVSData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IPVSGetWokerServer).GetIPVSRuntime(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.ipvs.IPVSGetWoker/GetIPVSRuntime",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IPVSGetWokerServer).GetIPVSRuntime(ctx, req.(*EmptyGetIPVSData))
	}
	return interceptor(ctx, in, info, handler)
}

var _IPVSGetWoker_serviceDesc = grpc.ServiceDesc{
	ServiceName: "lbos.t1.ipvs.IPVSGetWoker",
	HandlerType: (*IPVSGetWokerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewIPVSService",
			Handler:    _IPVSGetWoker_NewIPVSService_Handler,
		},
		{
			MethodName: "AddIPVSApplicationServersForService",
			Handler:    _IPVSGetWoker_AddIPVSApplicationServersForService_Handler,
		},
		{
			MethodName: "RemoveIPVSService",
			Handler:    _IPVSGetWoker_RemoveIPVSService_Handler,
		},
		{
			MethodName: "RemoveIPVSApplicationServersFromService",
			Handler:    _IPVSGetWoker_RemoveIPVSApplicationServersFromService_Handler,
		},
		{
			MethodName: "GetIPVSRuntime",
			Handler:    _IPVSGetWoker_GetIPVSRuntime_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "t1-ipvs.proto",
}
