// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package lbos_t1_orch

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// SendRuntimeClient is the client API for SendRuntime service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SendRuntimeClient interface {
	DummyRuntime(ctx context.Context, in *DummyRuntimeData, opts ...grpc.CallOption) (*EmptyDummyData, error)
	IPVSRuntime(ctx context.Context, in *PbIPVSRawServicesData, opts ...grpc.CallOption) (*EmptyIPVSData, error)
	IpRuleRuntime(ctx context.Context, in *IpRuleRuntimeData, opts ...grpc.CallOption) (*EmptyIpRuleData, error)
	TunnelRuntime(ctx context.Context, in *TunnelRuntimeData, opts ...grpc.CallOption) (*EmptyTunnelData, error)
	SendRouteRuntime(ctx context.Context, in *SendRouteRuntimeData, opts ...grpc.CallOption) (*EmptySendRouteData, error)
}

type sendRuntimeClient struct {
	cc grpc.ClientConnInterface
}

func NewSendRuntimeClient(cc grpc.ClientConnInterface) SendRuntimeClient {
	return &sendRuntimeClient{cc}
}

func (c *sendRuntimeClient) DummyRuntime(ctx context.Context, in *DummyRuntimeData, opts ...grpc.CallOption) (*EmptyDummyData, error) {
	out := new(EmptyDummyData)
	err := c.cc.Invoke(ctx, "/lbos.t1.orch.SendRuntime/DummyRuntime", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sendRuntimeClient) IPVSRuntime(ctx context.Context, in *PbIPVSRawServicesData, opts ...grpc.CallOption) (*EmptyIPVSData, error) {
	out := new(EmptyIPVSData)
	err := c.cc.Invoke(ctx, "/lbos.t1.orch.SendRuntime/IPVSRuntime", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sendRuntimeClient) IpRuleRuntime(ctx context.Context, in *IpRuleRuntimeData, opts ...grpc.CallOption) (*EmptyIpRuleData, error) {
	out := new(EmptyIpRuleData)
	err := c.cc.Invoke(ctx, "/lbos.t1.orch.SendRuntime/IpRuleRuntime", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sendRuntimeClient) TunnelRuntime(ctx context.Context, in *TunnelRuntimeData, opts ...grpc.CallOption) (*EmptyTunnelData, error) {
	out := new(EmptyTunnelData)
	err := c.cc.Invoke(ctx, "/lbos.t1.orch.SendRuntime/TunnelRuntime", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sendRuntimeClient) SendRouteRuntime(ctx context.Context, in *SendRouteRuntimeData, opts ...grpc.CallOption) (*EmptySendRouteData, error) {
	out := new(EmptySendRouteData)
	err := c.cc.Invoke(ctx, "/lbos.t1.orch.SendRuntime/SendRouteRuntime", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SendRuntimeServer is the server API for SendRuntime service.
// All implementations must embed UnimplementedSendRuntimeServer
// for forward compatibility
type SendRuntimeServer interface {
	DummyRuntime(context.Context, *DummyRuntimeData) (*EmptyDummyData, error)
	IPVSRuntime(context.Context, *PbIPVSRawServicesData) (*EmptyIPVSData, error)
	IpRuleRuntime(context.Context, *IpRuleRuntimeData) (*EmptyIpRuleData, error)
	TunnelRuntime(context.Context, *TunnelRuntimeData) (*EmptyTunnelData, error)
	SendRouteRuntime(context.Context, *SendRouteRuntimeData) (*EmptySendRouteData, error)
	mustEmbedUnimplementedSendRuntimeServer()
}

// UnimplementedSendRuntimeServer must be embedded to have forward compatible implementations.
type UnimplementedSendRuntimeServer struct {
}

func (UnimplementedSendRuntimeServer) DummyRuntime(context.Context, *DummyRuntimeData) (*EmptyDummyData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DummyRuntime not implemented")
}
func (UnimplementedSendRuntimeServer) IPVSRuntime(context.Context, *PbIPVSRawServicesData) (*EmptyIPVSData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IPVSRuntime not implemented")
}
func (UnimplementedSendRuntimeServer) IpRuleRuntime(context.Context, *IpRuleRuntimeData) (*EmptyIpRuleData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IpRuleRuntime not implemented")
}
func (UnimplementedSendRuntimeServer) TunnelRuntime(context.Context, *TunnelRuntimeData) (*EmptyTunnelData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TunnelRuntime not implemented")
}
func (UnimplementedSendRuntimeServer) SendRouteRuntime(context.Context, *SendRouteRuntimeData) (*EmptySendRouteData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendRouteRuntime not implemented")
}
func (UnimplementedSendRuntimeServer) mustEmbedUnimplementedSendRuntimeServer() {}

// UnsafeSendRuntimeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SendRuntimeServer will
// result in compilation errors.
type UnsafeSendRuntimeServer interface {
	mustEmbedUnimplementedSendRuntimeServer()
}

func RegisterSendRuntimeServer(s grpc.ServiceRegistrar, srv SendRuntimeServer) {
	s.RegisterService(&SendRuntime_ServiceDesc, srv)
}

func _SendRuntime_DummyRuntime_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DummyRuntimeData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SendRuntimeServer).DummyRuntime(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.orch.SendRuntime/DummyRuntime",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SendRuntimeServer).DummyRuntime(ctx, req.(*DummyRuntimeData))
	}
	return interceptor(ctx, in, info, handler)
}

func _SendRuntime_IPVSRuntime_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PbIPVSRawServicesData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SendRuntimeServer).IPVSRuntime(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.orch.SendRuntime/IPVSRuntime",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SendRuntimeServer).IPVSRuntime(ctx, req.(*PbIPVSRawServicesData))
	}
	return interceptor(ctx, in, info, handler)
}

func _SendRuntime_IpRuleRuntime_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IpRuleRuntimeData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SendRuntimeServer).IpRuleRuntime(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.orch.SendRuntime/IpRuleRuntime",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SendRuntimeServer).IpRuleRuntime(ctx, req.(*IpRuleRuntimeData))
	}
	return interceptor(ctx, in, info, handler)
}

func _SendRuntime_TunnelRuntime_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TunnelRuntimeData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SendRuntimeServer).TunnelRuntime(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.orch.SendRuntime/TunnelRuntime",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SendRuntimeServer).TunnelRuntime(ctx, req.(*TunnelRuntimeData))
	}
	return interceptor(ctx, in, info, handler)
}

func _SendRuntime_SendRouteRuntime_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendRouteRuntimeData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SendRuntimeServer).SendRouteRuntime(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.orch.SendRuntime/SendRouteRuntime",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SendRuntimeServer).SendRouteRuntime(ctx, req.(*SendRouteRuntimeData))
	}
	return interceptor(ctx, in, info, handler)
}

// SendRuntime_ServiceDesc is the grpc.ServiceDesc for SendRuntime service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SendRuntime_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "lbos.t1.orch.SendRuntime",
	HandlerType: (*SendRuntimeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DummyRuntime",
			Handler:    _SendRuntime_DummyRuntime_Handler,
		},
		{
			MethodName: "IPVSRuntime",
			Handler:    _SendRuntime_IPVSRuntime_Handler,
		},
		{
			MethodName: "IpRuleRuntime",
			Handler:    _SendRuntime_IpRuleRuntime_Handler,
		},
		{
			MethodName: "TunnelRuntime",
			Handler:    _SendRuntime_TunnelRuntime_Handler,
		},
		{
			MethodName: "SendRouteRuntime",
			Handler:    _SendRuntime_SendRouteRuntime_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "t1-orch.proto",
}
