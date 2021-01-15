// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package lbos_t1_route

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// RouteGetWorkerClient is the client API for RouteGetWorker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RouteGetWorkerClient interface {
	AddRoute(ctx context.Context, in *RouteData, opts ...grpc.CallOption) (*EmptyRouteData, error)
	RemoveRoute(ctx context.Context, in *RouteData, opts ...grpc.CallOption) (*EmptyRouteData, error)
	GetRouteRuntimeConfig(ctx context.Context, in *EmptyRouteData, opts ...grpc.CallOption) (*GetAllRoutesData, error)
}

type routeGetWorkerClient struct {
	cc grpc.ClientConnInterface
}

func NewRouteGetWorkerClient(cc grpc.ClientConnInterface) RouteGetWorkerClient {
	return &routeGetWorkerClient{cc}
}

func (c *routeGetWorkerClient) AddRoute(ctx context.Context, in *RouteData, opts ...grpc.CallOption) (*EmptyRouteData, error) {
	out := new(EmptyRouteData)
	err := c.cc.Invoke(ctx, "/lbos.t1.route.RouteGetWorker/AddRoute", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *routeGetWorkerClient) RemoveRoute(ctx context.Context, in *RouteData, opts ...grpc.CallOption) (*EmptyRouteData, error) {
	out := new(EmptyRouteData)
	err := c.cc.Invoke(ctx, "/lbos.t1.route.RouteGetWorker/RemoveRoute", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *routeGetWorkerClient) GetRouteRuntimeConfig(ctx context.Context, in *EmptyRouteData, opts ...grpc.CallOption) (*GetAllRoutesData, error) {
	out := new(GetAllRoutesData)
	err := c.cc.Invoke(ctx, "/lbos.t1.route.RouteGetWorker/GetRouteRuntimeConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RouteGetWorkerServer is the server API for RouteGetWorker service.
// All implementations must embed UnimplementedRouteGetWorkerServer
// for forward compatibility
type RouteGetWorkerServer interface {
	AddRoute(context.Context, *RouteData) (*EmptyRouteData, error)
	RemoveRoute(context.Context, *RouteData) (*EmptyRouteData, error)
	GetRouteRuntimeConfig(context.Context, *EmptyRouteData) (*GetAllRoutesData, error)
	mustEmbedUnimplementedRouteGetWorkerServer()
}

// UnimplementedRouteGetWorkerServer must be embedded to have forward compatible implementations.
type UnimplementedRouteGetWorkerServer struct {
}

func (UnimplementedRouteGetWorkerServer) AddRoute(context.Context, *RouteData) (*EmptyRouteData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddRoute not implemented")
}
func (UnimplementedRouteGetWorkerServer) RemoveRoute(context.Context, *RouteData) (*EmptyRouteData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveRoute not implemented")
}
func (UnimplementedRouteGetWorkerServer) GetRouteRuntimeConfig(context.Context, *EmptyRouteData) (*GetAllRoutesData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRouteRuntimeConfig not implemented")
}
func (UnimplementedRouteGetWorkerServer) mustEmbedUnimplementedRouteGetWorkerServer() {}

// UnsafeRouteGetWorkerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RouteGetWorkerServer will
// result in compilation errors.
type UnsafeRouteGetWorkerServer interface {
	mustEmbedUnimplementedRouteGetWorkerServer()
}

func RegisterRouteGetWorkerServer(s grpc.ServiceRegistrar, srv RouteGetWorkerServer) {
	s.RegisterService(&RouteGetWorker_ServiceDesc, srv)
}

func _RouteGetWorker_AddRoute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RouteData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RouteGetWorkerServer).AddRoute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.route.RouteGetWorker/AddRoute",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RouteGetWorkerServer).AddRoute(ctx, req.(*RouteData))
	}
	return interceptor(ctx, in, info, handler)
}

func _RouteGetWorker_RemoveRoute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RouteData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RouteGetWorkerServer).RemoveRoute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.route.RouteGetWorker/RemoveRoute",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RouteGetWorkerServer).RemoveRoute(ctx, req.(*RouteData))
	}
	return interceptor(ctx, in, info, handler)
}

func _RouteGetWorker_GetRouteRuntimeConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRouteData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RouteGetWorkerServer).GetRouteRuntimeConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.route.RouteGetWorker/GetRouteRuntimeConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RouteGetWorkerServer).GetRouteRuntimeConfig(ctx, req.(*EmptyRouteData))
	}
	return interceptor(ctx, in, info, handler)
}

// RouteGetWorker_ServiceDesc is the grpc.ServiceDesc for RouteGetWorker service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RouteGetWorker_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "lbos.t1.route.RouteGetWorker",
	HandlerType: (*RouteGetWorkerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddRoute",
			Handler:    _RouteGetWorker_AddRoute_Handler,
		},
		{
			MethodName: "RemoveRoute",
			Handler:    _RouteGetWorker_RemoveRoute_Handler,
		},
		{
			MethodName: "GetRouteRuntimeConfig",
			Handler:    _RouteGetWorker_GetRouteRuntimeConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "t1-route.proto",
}
