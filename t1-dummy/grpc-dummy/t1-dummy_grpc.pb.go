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

// DummyGetWorkerClient is the client API for DummyGetWorker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DummyGetWorkerClient interface {
	AddToDummy(ctx context.Context, in *IpData, opts ...grpc.CallOption) (*EmptyGetDummyData, error)
	RemoveFromDummy(ctx context.Context, in *IpData, opts ...grpc.CallOption) (*EmptyGetDummyData, error)
	GetDummyRuntime(ctx context.Context, in *EmptyGetDummyData, opts ...grpc.CallOption) (*GetDummyRuntimeData, error)
}

type dummyGetWorkerClient struct {
	cc grpc.ClientConnInterface
}

func NewDummyGetWorkerClient(cc grpc.ClientConnInterface) DummyGetWorkerClient {
	return &dummyGetWorkerClient{cc}
}

func (c *dummyGetWorkerClient) AddToDummy(ctx context.Context, in *IpData, opts ...grpc.CallOption) (*EmptyGetDummyData, error) {
	out := new(EmptyGetDummyData)
	err := c.cc.Invoke(ctx, "/lbos.t1.dummy.DummyGetWorker/AddToDummy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dummyGetWorkerClient) RemoveFromDummy(ctx context.Context, in *IpData, opts ...grpc.CallOption) (*EmptyGetDummyData, error) {
	out := new(EmptyGetDummyData)
	err := c.cc.Invoke(ctx, "/lbos.t1.dummy.DummyGetWorker/RemoveFromDummy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dummyGetWorkerClient) GetDummyRuntime(ctx context.Context, in *EmptyGetDummyData, opts ...grpc.CallOption) (*GetDummyRuntimeData, error) {
	out := new(GetDummyRuntimeData)
	err := c.cc.Invoke(ctx, "/lbos.t1.dummy.DummyGetWorker/GetDummyRuntime", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DummyGetWorkerServer is the server API for DummyGetWorker service.
// All implementations must embed UnimplementedDummyGetWorkerServer
// for forward compatibility
type DummyGetWorkerServer interface {
	AddToDummy(context.Context, *IpData) (*EmptyGetDummyData, error)
	RemoveFromDummy(context.Context, *IpData) (*EmptyGetDummyData, error)
	GetDummyRuntime(context.Context, *EmptyGetDummyData) (*GetDummyRuntimeData, error)
	mustEmbedUnimplementedDummyGetWorkerServer()
}

// UnimplementedDummyGetWorkerServer must be embedded to have forward compatible implementations.
type UnimplementedDummyGetWorkerServer struct {
}

func (UnimplementedDummyGetWorkerServer) AddToDummy(context.Context, *IpData) (*EmptyGetDummyData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddToDummy not implemented")
}
func (UnimplementedDummyGetWorkerServer) RemoveFromDummy(context.Context, *IpData) (*EmptyGetDummyData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveFromDummy not implemented")
}
func (UnimplementedDummyGetWorkerServer) GetDummyRuntime(context.Context, *EmptyGetDummyData) (*GetDummyRuntimeData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDummyRuntime not implemented")
}
func (UnimplementedDummyGetWorkerServer) mustEmbedUnimplementedDummyGetWorkerServer() {}

// UnsafeDummyGetWorkerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DummyGetWorkerServer will
// result in compilation errors.
type UnsafeDummyGetWorkerServer interface {
	mustEmbedUnimplementedDummyGetWorkerServer()
}

func RegisterDummyGetWorkerServer(s grpc.ServiceRegistrar, srv DummyGetWorkerServer) {
	s.RegisterService(&DummyGetWorker_ServiceDesc, srv)
}

func _DummyGetWorker_AddToDummy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IpData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DummyGetWorkerServer).AddToDummy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.dummy.DummyGetWorker/AddToDummy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DummyGetWorkerServer).AddToDummy(ctx, req.(*IpData))
	}
	return interceptor(ctx, in, info, handler)
}

func _DummyGetWorker_RemoveFromDummy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IpData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DummyGetWorkerServer).RemoveFromDummy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.dummy.DummyGetWorker/RemoveFromDummy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DummyGetWorkerServer).RemoveFromDummy(ctx, req.(*IpData))
	}
	return interceptor(ctx, in, info, handler)
}

func _DummyGetWorker_GetDummyRuntime_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyGetDummyData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DummyGetWorkerServer).GetDummyRuntime(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lbos.t1.dummy.DummyGetWorker/GetDummyRuntime",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DummyGetWorkerServer).GetDummyRuntime(ctx, req.(*EmptyGetDummyData))
	}
	return interceptor(ctx, in, info, handler)
}

// DummyGetWorker_ServiceDesc is the grpc.ServiceDesc for DummyGetWorker service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DummyGetWorker_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "lbos.t1.dummy.DummyGetWorker",
	HandlerType: (*DummyGetWorkerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddToDummy",
			Handler:    _DummyGetWorker_AddToDummy_Handler,
		},
		{
			MethodName: "RemoveFromDummy",
			Handler:    _DummyGetWorker_RemoveFromDummy_Handler,
		},
		{
			MethodName: "GetDummyRuntime",
			Handler:    _DummyGetWorker_GetDummyRuntime_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "t1-dummy.proto",
}
