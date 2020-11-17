// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.4
// source: t1-orch.proto

package transport

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

type EmptySendIPVSData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *EmptySendIPVSData) Reset() {
	*x = EmptySendIPVSData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_t1_orch_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EmptySendIPVSData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmptySendIPVSData) ProtoMessage() {}

func (x *EmptySendIPVSData) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use EmptySendIPVSData.ProtoReflect.Descriptor instead.
func (*EmptySendIPVSData) Descriptor() ([]byte, []int) {
	return file_t1_orch_proto_rawDescGZIP(), []int{0}
}

func (x *EmptySendIPVSData) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type PbSendIPVSRawServicesData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RawServicesData map[string]*PbSendRawIPVSServiceData `protobuf:"bytes,1,rep,name=rawServicesData,proto3" json:"rawServicesData,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Id              string                               `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *PbSendIPVSRawServicesData) Reset() {
	*x = PbSendIPVSRawServicesData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_t1_orch_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PbSendIPVSRawServicesData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PbSendIPVSRawServicesData) ProtoMessage() {}

func (x *PbSendIPVSRawServicesData) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use PbSendIPVSRawServicesData.ProtoReflect.Descriptor instead.
func (*PbSendIPVSRawServicesData) Descriptor() ([]byte, []int) {
	return file_t1_orch_proto_rawDescGZIP(), []int{1}
}

func (x *PbSendIPVSRawServicesData) GetRawServicesData() map[string]*PbSendRawIPVSServiceData {
	if x != nil {
		return x.RawServicesData
	}
	return nil
}

func (x *PbSendIPVSRawServicesData) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type PbSendRawIPVSServiceData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RawServiceData map[string]uint32 `protobuf:"bytes,1,rep,name=rawServiceData,proto3" json:"rawServiceData,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	Id             string            `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *PbSendRawIPVSServiceData) Reset() {
	*x = PbSendRawIPVSServiceData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_t1_orch_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PbSendRawIPVSServiceData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PbSendRawIPVSServiceData) ProtoMessage() {}

func (x *PbSendRawIPVSServiceData) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use PbSendRawIPVSServiceData.ProtoReflect.Descriptor instead.
func (*PbSendRawIPVSServiceData) Descriptor() ([]byte, []int) {
	return file_t1_orch_proto_rawDescGZIP(), []int{2}
}

func (x *PbSendRawIPVSServiceData) GetRawServiceData() map[string]uint32 {
	if x != nil {
		return x.RawServiceData
	}
	return nil
}

func (x *PbSendRawIPVSServiceData) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

var File_t1_orch_proto protoreflect.FileDescriptor

var file_t1_orch_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x74, 0x31, 0x2d, 0x6f, 0x72, 0x63, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x09, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x23, 0x0a, 0x11, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x53, 0x65, 0x6e, 0x64, 0x49, 0x50, 0x56, 0x53, 0x44, 0x61, 0x74, 0x61, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22,
	0xf9, 0x01, 0x0a, 0x19, 0x50, 0x62, 0x53, 0x65, 0x6e, 0x64, 0x49, 0x50, 0x56, 0x53, 0x52, 0x61,
	0x77, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x44, 0x61, 0x74, 0x61, 0x12, 0x63, 0x0a,
	0x0f, 0x72, 0x61, 0x77, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x44, 0x61, 0x74, 0x61,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x39, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f,
	0x72, 0x74, 0x2e, 0x50, 0x62, 0x53, 0x65, 0x6e, 0x64, 0x49, 0x50, 0x56, 0x53, 0x52, 0x61, 0x77,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x52, 0x61, 0x77,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x0f, 0x72, 0x61, 0x77, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x44, 0x61,
	0x74, 0x61, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x1a, 0x67, 0x0a, 0x14, 0x52, 0x61, 0x77, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x73, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x39, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x50, 0x62, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x61,
	0x77, 0x49, 0x50, 0x56, 0x53, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x44, 0x61, 0x74, 0x61,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xce, 0x01, 0x0a, 0x18,
	0x50, 0x62, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x61, 0x77, 0x49, 0x50, 0x56, 0x53, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x44, 0x61, 0x74, 0x61, 0x12, 0x5f, 0x0a, 0x0e, 0x72, 0x61, 0x77, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x44, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x37, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x50, 0x62, 0x53,
	0x65, 0x6e, 0x64, 0x52, 0x61, 0x77, 0x49, 0x50, 0x56, 0x53, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x52, 0x61, 0x77, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0e, 0x72, 0x61, 0x77, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x44, 0x61, 0x74, 0x61, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x1a, 0x41, 0x0a, 0x13, 0x52, 0x61, 0x77,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x32, 0x6a, 0x0a, 0x0f,
	0x53, 0x65, 0x6e, 0x64, 0x49, 0x50, 0x56, 0x53, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x12,
	0x57, 0x0a, 0x0f, 0x53, 0x65, 0x6e, 0x64, 0x49, 0x50, 0x56, 0x53, 0x52, 0x75, 0x6e, 0x74, 0x69,
	0x6d, 0x65, 0x12, 0x24, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x50,
	0x62, 0x53, 0x65, 0x6e, 0x64, 0x49, 0x50, 0x56, 0x53, 0x52, 0x61, 0x77, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x1c, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x70, 0x6f, 0x72, 0x74, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x53, 0x65, 0x6e, 0x64, 0x49, 0x50,
	0x56, 0x53, 0x44, 0x61, 0x74, 0x61, 0x22, 0x00, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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
	(*EmptySendIPVSData)(nil),         // 0: transport.EmptySendIPVSData
	(*PbSendIPVSRawServicesData)(nil), // 1: transport.PbSendIPVSRawServicesData
	(*PbSendRawIPVSServiceData)(nil),  // 2: transport.PbSendRawIPVSServiceData
	nil,                               // 3: transport.PbSendIPVSRawServicesData.RawServicesDataEntry
	nil,                               // 4: transport.PbSendRawIPVSServiceData.RawServiceDataEntry
}
var file_t1_orch_proto_depIdxs = []int32{
	3, // 0: transport.PbSendIPVSRawServicesData.rawServicesData:type_name -> transport.PbSendIPVSRawServicesData.RawServicesDataEntry
	4, // 1: transport.PbSendRawIPVSServiceData.rawServiceData:type_name -> transport.PbSendRawIPVSServiceData.RawServiceDataEntry
	2, // 2: transport.PbSendIPVSRawServicesData.RawServicesDataEntry.value:type_name -> transport.PbSendRawIPVSServiceData
	1, // 3: transport.SendIPVSRuntime.SendIPVSRuntime:input_type -> transport.PbSendIPVSRawServicesData
	0, // 4: transport.SendIPVSRuntime.SendIPVSRuntime:output_type -> transport.EmptySendIPVSData
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_t1_orch_proto_init() }
func file_t1_orch_proto_init() {
	if File_t1_orch_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_t1_orch_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EmptySendIPVSData); i {
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
			switch v := v.(*PbSendIPVSRawServicesData); i {
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
			switch v := v.(*PbSendRawIPVSServiceData); i {
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
			NumServices:   1,
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
