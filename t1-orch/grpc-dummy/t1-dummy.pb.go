// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: t1-dummy.proto

package lbos_t1_dummy

import (
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

type IpData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ip string `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Id string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *IpData) Reset() {
	*x = IpData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_t1_dummy_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IpData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IpData) ProtoMessage() {}

func (x *IpData) ProtoReflect() protoreflect.Message {
	mi := &file_t1_dummy_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IpData.ProtoReflect.Descriptor instead.
func (*IpData) Descriptor() ([]byte, []int) {
	return file_t1_dummy_proto_rawDescGZIP(), []int{0}
}

func (x *IpData) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *IpData) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetDummyRuntimeData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Services map[string]int32 `protobuf:"bytes,1,rep,name=services,proto3" json:"services,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	Id       string           `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetDummyRuntimeData) Reset() {
	*x = GetDummyRuntimeData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_t1_dummy_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDummyRuntimeData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDummyRuntimeData) ProtoMessage() {}

func (x *GetDummyRuntimeData) ProtoReflect() protoreflect.Message {
	mi := &file_t1_dummy_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDummyRuntimeData.ProtoReflect.Descriptor instead.
func (*GetDummyRuntimeData) Descriptor() ([]byte, []int) {
	return file_t1_dummy_proto_rawDescGZIP(), []int{1}
}

func (x *GetDummyRuntimeData) GetServices() map[string]int32 {
	if x != nil {
		return x.Services
	}
	return nil
}

func (x *GetDummyRuntimeData) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type EmptyGetDummyData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *EmptyGetDummyData) Reset() {
	*x = EmptyGetDummyData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_t1_dummy_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EmptyGetDummyData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmptyGetDummyData) ProtoMessage() {}

func (x *EmptyGetDummyData) ProtoReflect() protoreflect.Message {
	mi := &file_t1_dummy_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmptyGetDummyData.ProtoReflect.Descriptor instead.
func (*EmptyGetDummyData) Descriptor() ([]byte, []int) {
	return file_t1_dummy_proto_rawDescGZIP(), []int{2}
}

func (x *EmptyGetDummyData) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

var File_t1_dummy_proto protoreflect.FileDescriptor

var file_t1_dummy_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x74, 0x31, 0x2d, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x0d, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74, 0x31, 0x2e, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x22,
	0x28, 0x0a, 0x06, 0x49, 0x70, 0x44, 0x61, 0x74, 0x61, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0xb0, 0x01, 0x0a, 0x13, 0x47, 0x65,
	0x74, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x44, 0x61, 0x74,
	0x61, 0x12, 0x4c, 0x0a, 0x08, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74, 0x31, 0x2e, 0x64, 0x75,
	0x6d, 0x6d, 0x79, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x52, 0x75, 0x6e, 0x74,
	0x69, 0x6d, 0x65, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x1a,
	0x3b, 0x0a, 0x0d, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x23, 0x0a, 0x11,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x47, 0x65, 0x74, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x44, 0x61, 0x74,
	0x61, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x32, 0x82, 0x02, 0x0a, 0x0e, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x47, 0x65, 0x74, 0x57, 0x6f,
	0x72, 0x6b, 0x65, 0x72, 0x12, 0x47, 0x0a, 0x0a, 0x41, 0x64, 0x64, 0x54, 0x6f, 0x44, 0x75, 0x6d,
	0x6d, 0x79, 0x12, 0x15, 0x2e, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74, 0x31, 0x2e, 0x64, 0x75, 0x6d,
	0x6d, 0x79, 0x2e, 0x49, 0x70, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x20, 0x2e, 0x6c, 0x62, 0x6f, 0x73,
	0x2e, 0x74, 0x31, 0x2e, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x47,
	0x65, 0x74, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x44, 0x61, 0x74, 0x61, 0x22, 0x00, 0x12, 0x4c, 0x0a,
	0x0f, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x46, 0x72, 0x6f, 0x6d, 0x44, 0x75, 0x6d, 0x6d, 0x79,
	0x12, 0x15, 0x2e, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74, 0x31, 0x2e, 0x64, 0x75, 0x6d, 0x6d, 0x79,
	0x2e, 0x49, 0x70, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x20, 0x2e, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74,
	0x31, 0x2e, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x47, 0x65, 0x74,
	0x44, 0x75, 0x6d, 0x6d, 0x79, 0x44, 0x61, 0x74, 0x61, 0x22, 0x00, 0x12, 0x59, 0x0a, 0x0f, 0x47,
	0x65, 0x74, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x20,
	0x2e, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74, 0x31, 0x2e, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x2e, 0x45,
	0x6d, 0x70, 0x74, 0x79, 0x47, 0x65, 0x74, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x44, 0x61, 0x74, 0x61,
	0x1a, 0x22, 0x2e, 0x6c, 0x62, 0x6f, 0x73, 0x2e, 0x74, 0x31, 0x2e, 0x64, 0x75, 0x6d, 0x6d, 0x79,
	0x2e, 0x47, 0x65, 0x74, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65,
	0x44, 0x61, 0x74, 0x61, 0x22, 0x00, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_t1_dummy_proto_rawDescOnce sync.Once
	file_t1_dummy_proto_rawDescData = file_t1_dummy_proto_rawDesc
)

func file_t1_dummy_proto_rawDescGZIP() []byte {
	file_t1_dummy_proto_rawDescOnce.Do(func() {
		file_t1_dummy_proto_rawDescData = protoimpl.X.CompressGZIP(file_t1_dummy_proto_rawDescData)
	})
	return file_t1_dummy_proto_rawDescData
}

var file_t1_dummy_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_t1_dummy_proto_goTypes = []interface{}{
	(*IpData)(nil),              // 0: lbos.t1.dummy.IpData
	(*GetDummyRuntimeData)(nil), // 1: lbos.t1.dummy.GetDummyRuntimeData
	(*EmptyGetDummyData)(nil),   // 2: lbos.t1.dummy.EmptyGetDummyData
	nil,                         // 3: lbos.t1.dummy.GetDummyRuntimeData.ServicesEntry
}
var file_t1_dummy_proto_depIdxs = []int32{
	3, // 0: lbos.t1.dummy.GetDummyRuntimeData.services:type_name -> lbos.t1.dummy.GetDummyRuntimeData.ServicesEntry
	0, // 1: lbos.t1.dummy.DummyGetWorker.AddToDummy:input_type -> lbos.t1.dummy.IpData
	0, // 2: lbos.t1.dummy.DummyGetWorker.RemoveFromDummy:input_type -> lbos.t1.dummy.IpData
	2, // 3: lbos.t1.dummy.DummyGetWorker.GetDummyRuntime:input_type -> lbos.t1.dummy.EmptyGetDummyData
	2, // 4: lbos.t1.dummy.DummyGetWorker.AddToDummy:output_type -> lbos.t1.dummy.EmptyGetDummyData
	2, // 5: lbos.t1.dummy.DummyGetWorker.RemoveFromDummy:output_type -> lbos.t1.dummy.EmptyGetDummyData
	1, // 6: lbos.t1.dummy.DummyGetWorker.GetDummyRuntime:output_type -> lbos.t1.dummy.GetDummyRuntimeData
	4, // [4:7] is the sub-list for method output_type
	1, // [1:4] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_t1_dummy_proto_init() }
func file_t1_dummy_proto_init() {
	if File_t1_dummy_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_t1_dummy_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IpData); i {
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
		file_t1_dummy_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDummyRuntimeData); i {
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
		file_t1_dummy_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EmptyGetDummyData); i {
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
			RawDescriptor: file_t1_dummy_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_t1_dummy_proto_goTypes,
		DependencyIndexes: file_t1_dummy_proto_depIdxs,
		MessageInfos:      file_t1_dummy_proto_msgTypes,
	}.Build()
	File_t1_dummy_proto = out.File
	file_t1_dummy_proto_rawDesc = nil
	file_t1_dummy_proto_goTypes = nil
	file_t1_dummy_proto_depIdxs = nil
}
