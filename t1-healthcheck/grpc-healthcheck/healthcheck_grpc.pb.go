// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package lbos_t1_dummy

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// HealthcheckWorkerClient is the client API for HealthcheckWorker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HealthcheckWorkerClient interface {
	IsHttpAdvancedCheckOk(ctx context.Context, in *HttpAdvancedData, opts ...grpc.CallOption) (*IsOk, error)
	IsHttpsCheckOk(ctx context.Context, in *HttpsData, opts ...grpc.CallOption) (*IsOk, error)
	IsIcmpCheckOk(ctx context.Context, in *IcmpData, opts ...grpc.CallOption) (*IsOk, error)
	IsTcpCheckOk(ctx context.Context, in *TcpData, opts ...grpc.CallOption) (*IsOk, error)
}

type healthcheckWorkerClient struct {
	cc grpc.ClientConnInterface
}

func NewHealthcheckWorkerClient(cc grpc.ClientConnInterface) HealthcheckWorkerClient {
	return &healthcheckWorkerClient{cc}
}

func (c *healthcheckWorkerClient) IsHttpAdvancedCheckOk(ctx context.Context, in *HttpAdvancedData, opts ...grpc.CallOption) (*IsOk, error) {
	out := new(IsOk)
	err := c.cc.Invoke(ctx, "/lbos.t1.dummy.HealthcheckWorker/IsHttpAdvancedCheckOk", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *healthcheckWorkerClient) IsHttpsCheckOk(ctx context.Context, in *HttpsData, opts ...grpc.CallOption) (*IsOk, error) {
	out := new(IsOk)
	err := c.cc.Invoke(ctx, "/lbos.t1.dummy.HealthcheckWorker/IsHttpsCheckOk", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *healthcheckWorkerClient) IsIcmpCheckOk(ctx context.Context, in *IcmpData, opts ...grpc.CallOption) (*IsOk, error) {
	out := new(IsOk)
	err := c.cc.Invoke(ctx, "/lbos.t1.dummy.HealthcheckWorker/IsIcmpCheckOk", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *healthcheckWorkerClient) IsTcpCheckOk(ctx context.Context, in *TcpData, opts ...grpc.CallOption) (*IsOk, error) {
	out := new(IsOk)
	err := c.cc.Invoke(ctx, "/lbos.t1.dummy.HealthcheckWorker/IsTcpCheckOk", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HealthcheckWorkerServer is the server API for HealthcheckWorker service.
// All implementations must embed UnimplementedHealthcheckWorkerServer
// for forward compatibility
type HealthcheckWorkerServer interface {
	IsHttpAdvancedCheckOk(context.Context, *HttpAdvancedData) (*IsOk, error)
	IsHttpsCheckOk(context.Context, *HttpsData) (*IsOk, error)
	IsIcmpCheckOk(context.Context, *IcmpData) (*IsOk, error)
	IsTcpCheckOk(context.Context, *TcpData) (*IsOk, error)
	mustEmbedUnimplementedHealthcheckWorkerServer()
}

// UnimplementedHealthcheckWorkerServer must be embedded to have forward compatible implementations.
type UnimplementedHealthcheckWorkerServer struct {
}

func (UnimplementedHealthcheckWorkerServer) IsHttpAdvancedCheckOk(context.Context, *HttpAdvancedData) (*IsOk, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsHttpAdvancedCheckOk not implemented")
}
func (UnimplementedHealthcheckWorkerServer) IsHttpsCheckOk(context.Context, *HttpsData) (*IsOk, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsHttpsCheckOk not implemented")
}
func (UnimplementedHealthcheckWorkerServer) IsIcmpCheckOk(context.Context, *IcmpData) (*IsOk, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsIcmpCheckOk not implemented")
}
func (UnimplementedHealthcheckWorkerServer) IsTcpCheckOk(context.Context, *TcpData) (*IsOk, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsTcpCheckOk not implemented")
}
func (UnimplementedHealthcheckWorkerServer) mustEmbedUnimplementedHealthcheckWorkerServer() {}

// UnsafeHealthcheckWorkerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HealthcheckWorkerServer will
// result in compilation errors.
type UnsafeHealthcheckWorkerServer interface {
	mustEmbedUnimplementedHealthcheckWorkerServer()
}

func RegisterHealthcheckWorkerServer(s *grpc.Server, srv HealthcheckWorkerServer) {
	s.RegisterService(&_HealthcheckWorker_serviceDesc, srv)
}

func _HealthcheckWorker_IsHttpAdvancedCheckOk_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HttpAdvancedData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HealthcheckWorkerServer).IsHttpAdvancedCheckOk(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.dummy.HealthcheckWorker/IsHttpAdvancedCheckOk",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HealthcheckWorkerServer).IsHttpAdvancedCheckOk(ctx, req.(*HttpAdvancedData))
	}
	return interceptor(ctx, in, info, handler)
}

func _HealthcheckWorker_IsHttpsCheckOk_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HttpsData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HealthcheckWorkerServer).IsHttpsCheckOk(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.dummy.HealthcheckWorker/IsHttpsCheckOk",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HealthcheckWorkerServer).IsHttpsCheckOk(ctx, req.(*HttpsData))
	}
	return interceptor(ctx, in, info, handler)
}

func _HealthcheckWorker_IsIcmpCheckOk_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IcmpData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HealthcheckWorkerServer).IsIcmpCheckOk(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.dummy.HealthcheckWorker/IsIcmpCheckOk",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HealthcheckWorkerServer).IsIcmpCheckOk(ctx, req.(*IcmpData))
	}
	return interceptor(ctx, in, info, handler)
}

func _HealthcheckWorker_IsTcpCheckOk_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TcpData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HealthcheckWorkerServer).IsTcpCheckOk(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.dummy.HealthcheckWorker/IsTcpCheckOk",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HealthcheckWorkerServer).IsTcpCheckOk(ctx, req.(*TcpData))
	}
	return interceptor(ctx, in, info, handler)
}

var _HealthcheckWorker_serviceDesc = grpc.ServiceDesc{
	ServiceName: "lbos.t1.dummy.HealthcheckWorker",
	HandlerType: (*HealthcheckWorkerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "IsHttpAdvancedCheckOk",
			Handler:    _HealthcheckWorker_IsHttpAdvancedCheckOk_Handler,
		},
		{
			MethodName: "IsHttpsCheckOk",
			Handler:    _HealthcheckWorker_IsHttpsCheckOk_Handler,
		},
		{
			MethodName: "IsIcmpCheckOk",
			Handler:    _HealthcheckWorker_IsIcmpCheckOk_Handler,
		},
		{
			MethodName: "IsTcpCheckOk",
			Handler:    _HealthcheckWorker_IsTcpCheckOk_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "healthcheck.proto",
}
