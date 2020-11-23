// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.4
// source: t1-orch.proto

package lbos_t1_orch

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type SendDummyRuntimeData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Services map[string]*EmptySendDummyData `protobuf:"bytes,1,rep,name=services,proto3" json:"services,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Id       string                         `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *SendDummyRuntimeData) Reset() {
	*x = SendDummyRuntimeData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_t1_orch_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendDummyRuntimeData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendDummyRuntimeData) ProtoMessage() {}

func (x *SendDummyRuntimeData) ProtoReflect() protoreflect.Message {
	mi := &file_t1_orch_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendDummyRuntimeData.ProtoReflect.Descriptor instead.
func (*SendDummyRuntimeData) Descriptor() ([]byte, []int) {
	return file_t1_orch_proto_rawDescGZIP(), []int{0}
}

func (x *SendDummyRuntimeData) GetServices() map[string]*EmptySendDummyData {
	if x != nil {
		return x.Services
	}
	return nil
}

func (x *SendDummyRuntimeData) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type EmptySendDummyData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *EmptySendDummyData) Reset() {
	*x = EmptySendDummyData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_t1_orch_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EmptySendDummyData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmptySendDummyData) ProtoMessage() {}

func (x *EmptySendDummyData) ProtoReflect() protoreflect.Message {
	mi := &file_t1_orch_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmptySendDummyData.ProtoReflect.Descriptor instead.
func (*EmptySendDummyData) Descriptor() ([]byte, []int) {
	return file_t1_orch_proto_rawDescGZIP(), []int{1}
}

func (x *EmptySendDummyData) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type SendAllRoutesRuntimeData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RouteData []string `protobuf:"bytes,1,rep,name=routeData,proto3" json:"routeData,omitempty"`
	Id        string   `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *SendAllRoutesRuntimeData) Reset() {
	*x = SendAllRoutesRuntimeData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_t1_orch_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendAllRoutesRuntimeData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendAllRoutesRuntimeData) ProtoMessage() {}

func (x *SendAllRoutesRuntimeData) ProtoReflect() protoreflect.Message {
	mi := &file_t1_orch_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendAllRoutesRuntimeData.ProtoReflect.Descriptor instead.
func (*SendAllRoutesRuntimeData) Descriptor() ([]byte, []int) {
	return file_t1_orch_proto_rawDescGZIP(), []int{2}
}

func (x *SendAllRoutesRuntimeData) GetRouteData() []string {
	if x != nil {
		return x.RouteData
	}
	return nil
}

func (x *SendAllRoutesRuntimeData) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type EmptySendRouteData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *EmptySendRouteData) Reset() {
	*x = EmptySendRouteData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_t1_orch_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EmptySendRouteData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmptySendRouteData) ProtoMessage() {}

func (x *EmptySendRouteData) ProtoReflect() protoreflect.Message {
	mi := &file_t1_orch_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmptySendRouteData.ProtoReflect.Descriptor instead.
func (*EmptySendRouteData) Descriptor() ([]byte, []int) {
	return file_t1_orch_proto_rawDescGZIP(), []int{3}
}

func (x *EmptySendRouteData) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

var File_t1_orch_proto protoreflect.FileDescriptor

var file_t1_orch_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x74, 0x31, 0x2d, 0x6f, 0x72, 0x63, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0c, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74, 0x31, 0x2e, 0x6f, 0x72, 0x63, 0x68, 0x22, 0xd3, 0x01,
	0x0a, 0x14, 0x53, 0x65, 0x6e, 0x64, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x52, 0x75, 0x6e, 0x74, 0x69,
	0x6d, 0x65, 0x44, 0x61, 0x74, 0x61, 0x12, 0x4c, 0x0a, 0x08, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x6c, 0x62, 0x6f, 0x73, 0x2e,
	0x74, 0x31, 0x2e, 0x6f, 0x72, 0x63, 0x68, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x44, 0x75, 0x6d, 0x6d,
	0x79, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x1a, 0x5d, 0x0a, 0x0d, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x36, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74, 0x31,
	0x2e, 0x6f, 0x72, 0x63, 0x68, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x53, 0x65, 0x6e, 0x64, 0x44,
	0x75, 0x6d, 0x6d, 0x79, 0x44, 0x61, 0x74, 0x61, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a,
	0x02, 0x38, 0x01, 0x22, 0x24, 0x0a, 0x12, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x53, 0x65, 0x6e, 0x64,
	0x44, 0x75, 0x6d, 0x6d, 0x79, 0x44, 0x61, 0x74, 0x61, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x48, 0x0a, 0x18, 0x53, 0x65, 0x6e,
	0x64, 0x41, 0x6c, 0x6c, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x73, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d,
	0x65, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1c, 0x0a, 0x09, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x44, 0x61,
	0x74, 0x61, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x22, 0x24, 0x0a, 0x12, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x53, 0x65, 0x6e, 0x64,
	0x52, 0x6f, 0x75, 0x74, 0x65, 0x44, 0x61, 0x74, 0x61, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x32, 0x6e, 0x0a, 0x10, 0x53, 0x65, 0x6e,
	0x64, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x5a, 0x0a,
	0x10, 0x53, 0x65, 0x6e, 0x64, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d,
	0x65, 0x12, 0x22, 0x2e, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74, 0x31, 0x2e, 0x6f, 0x72, 0x63, 0x68,
	0x2e, 0x53, 0x65, 0x6e, 0x64, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d,
	0x65, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x20, 0x2e, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74, 0x31, 0x2e,
	0x6f, 0x72, 0x63, 0x68, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x53, 0x65, 0x6e, 0x64, 0x44, 0x75,
	0x6d, 0x6d, 0x79, 0x44, 0x61, 0x74, 0x61, 0x22, 0x00, 0x32, 0x72, 0x0a, 0x10, 0x53, 0x65, 0x6e,
	0x64, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x5e, 0x0a,
	0x10, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d,
	0x65, 0x12, 0x26, 0x2e, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74, 0x31, 0x2e, 0x6f, 0x72, 0x63, 0x68,
	0x2e, 0x53, 0x65, 0x6e, 0x64, 0x41, 0x6c, 0x6c, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x73, 0x52, 0x75,
	0x6e, 0x74, 0x69, 0x6d, 0x65, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x20, 0x2e, 0x6c, 0x62, 0x6f, 0x73,
	0x2e, 0x74, 0x31, 0x2e, 0x6f, 0x72, 0x63, 0x68, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x53, 0x65,
	0x6e, 0x64, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x44, 0x61, 0x74, 0x61, 0x22, 0x00, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_t1_orch_proto_rawDescOnce sync.Once
	file_t1_orch_proto_rawDescData = file_t1_orch_proto_rawDesc
)

func file_t1_orch_proto_rawDescGZIP() []byte {
	file_t1_orch_proto_rawDescOnce.Do(func() {
		file_t1_orch_proto_rawDescData = protoimpl.X.CompressGZIP(file_t1_orch_proto_rawDescData)
	})
	return file_t1_orch_proto_rawDescData
}

var file_t1_orch_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_t1_orch_proto_goTypes = []interface{}{
	(*SendDummyRuntimeData)(nil),     // 0: lbos.t1.orch.SendDummyRuntimeData
	(*EmptySendDummyData)(nil),       // 1: lbos.t1.orch.EmptySendDummyData
	(*SendAllRoutesRuntimeData)(nil), // 2: lbos.t1.orch.SendAllRoutesRuntimeData
	(*EmptySendRouteData)(nil),       // 3: lbos.t1.orch.EmptySendRouteData
	nil,                              // 4: lbos.t1.orch.SendDummyRuntimeData.ServicesEntry
}
var file_t1_orch_proto_depIdxs = []int32{
	4, // 0: lbos.t1.orch.SendDummyRuntimeData.services:type_name -> lbos.t1.orch.SendDummyRuntimeData.ServicesEntry
	1, // 1: lbos.t1.orch.SendDummyRuntimeData.ServicesEntry.value:type_name -> lbos.t1.orch.EmptySendDummyData
	0, // 2: lbos.t1.orch.SendDummyRuntime.SendDummyRuntime:input_type -> lbos.t1.orch.SendDummyRuntimeData
	2, // 3: lbos.t1.orch.SendRouteRuntime.SendRouteRuntime:input_type -> lbos.t1.orch.SendAllRoutesRuntimeData
	1, // 4: lbos.t1.orch.SendDummyRuntime.SendDummyRuntime:output_type -> lbos.t1.orch.EmptySendDummyData
	3, // 5: lbos.t1.orch.SendRouteRuntime.SendRouteRuntime:output_type -> lbos.t1.orch.EmptySendRouteData
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_t1_orch_proto_init() }
func file_t1_orch_proto_init() {
	if File_t1_orch_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_t1_orch_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendDummyRuntimeData); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_t1_orch_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EmptySendDummyData); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_t1_orch_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendAllRoutesRuntimeData); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_t1_orch_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EmptySendRouteData); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_t1_orch_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_t1_orch_proto_goTypes,
		DependencyIndexes: file_t1_orch_proto_depIdxs,
		MessageInfos:      file_t1_orch_proto_msgTypes,
	}.Build()
	File_t1_orch_proto = out.File
	file_t1_orch_proto_rawDesc = nil
	file_t1_orch_proto_goTypes = nil
	file_t1_orch_proto_depIdxs = nil
}
