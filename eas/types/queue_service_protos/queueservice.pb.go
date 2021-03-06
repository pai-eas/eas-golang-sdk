// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.19.1
// source: pkg/streaming/types/queueservice.proto

package queue_service_protos

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

type DataFrameProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Index   uint64            `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	Tags    map[string]string `protobuf:"bytes,2,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Data    []byte            `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	Message string            `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *DataFrameProto) Reset() {
	*x = DataFrameProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_streaming_types_queueservice_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataFrameProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataFrameProto) ProtoMessage() {}

func (x *DataFrameProto) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_streaming_types_queueservice_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataFrameProto.ProtoReflect.Descriptor instead.
func (*DataFrameProto) Descriptor() ([]byte, []int) {
	return file_pkg_streaming_types_queueservice_proto_rawDescGZIP(), []int{0}
}

func (x *DataFrameProto) GetIndex() uint64 {
	if x != nil {
		return x.Index
	}
	return 0
}

func (x *DataFrameProto) GetTags() map[string]string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *DataFrameProto) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *DataFrameProto) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type DataFrameListProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Index []*DataFrameProto `protobuf:"bytes,1,rep,name=index,proto3" json:"index,omitempty"`
}

func (x *DataFrameListProto) Reset() {
	*x = DataFrameListProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_streaming_types_queueservice_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataFrameListProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataFrameListProto) ProtoMessage() {}

func (x *DataFrameListProto) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_streaming_types_queueservice_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataFrameListProto.ProtoReflect.Descriptor instead.
func (*DataFrameListProto) Descriptor() ([]byte, []int) {
	return file_pkg_streaming_types_queueservice_proto_rawDescGZIP(), []int{1}
}

func (x *DataFrameListProto) GetIndex() []*DataFrameProto {
	if x != nil {
		return x.Index
	}
	return nil
}

type AttributesProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Attributes map[string]string `protobuf:"bytes,1,rep,name=attributes,proto3" json:"attributes,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *AttributesProto) Reset() {
	*x = AttributesProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_streaming_types_queueservice_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AttributesProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AttributesProto) ProtoMessage() {}

func (x *AttributesProto) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_streaming_types_queueservice_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AttributesProto.ProtoReflect.Descriptor instead.
func (*AttributesProto) Descriptor() ([]byte, []int) {
	return file_pkg_streaming_types_queueservice_proto_rawDescGZIP(), []int{2}
}

func (x *AttributesProto) GetAttributes() map[string]string {
	if x != nil {
		return x.Attributes
	}
	return nil
}

var File_pkg_streaming_types_queueservice_proto protoreflect.FileDescriptor

var file_pkg_streaming_types_queueservice_proto_rawDesc = []byte{
	0x0a, 0x26, 0x70, 0x6b, 0x67, 0x2f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x2f,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x71, 0x75, 0x65, 0x75, 0x65, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x69, 0x6e, 0x67, 0x22, 0xc6, 0x01, 0x0a, 0x0e, 0x44, 0x61, 0x74, 0x61, 0x46, 0x72, 0x61, 0x6d,
	0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x37, 0x0a, 0x04,
	0x74, 0x61, 0x67, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x46, 0x72, 0x61, 0x6d, 0x65,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x04, 0x74, 0x61, 0x67, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x1a, 0x37, 0x0a, 0x09, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x45, 0x0a, 0x12,
	0x44, 0x61, 0x74, 0x61, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x4c, 0x69, 0x73, 0x74, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x2f, 0x0a, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x2e, 0x44, 0x61,
	0x74, 0x61, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x05, 0x69, 0x6e,
	0x64, 0x65, 0x78, 0x22, 0x9c, 0x01, 0x0a, 0x0f, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
	0x65, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x4a, 0x0a, 0x0a, 0x61, 0x74, 0x74, 0x72, 0x69,
	0x62, 0x75, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x73, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x2e, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
	0x65, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
	0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
	0x74, 0x65, 0x73, 0x1a, 0x3d, 0x0a, 0x0f, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x42, 0x15, 0x5a, 0x13, 0x70, 0x6b, 0x67, 0x2f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x69, 0x6e, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_pkg_streaming_types_queueservice_proto_rawDescOnce sync.Once
	file_pkg_streaming_types_queueservice_proto_rawDescData = file_pkg_streaming_types_queueservice_proto_rawDesc
)

func file_pkg_streaming_types_queueservice_proto_rawDescGZIP() []byte {
	file_pkg_streaming_types_queueservice_proto_rawDescOnce.Do(func() {
		file_pkg_streaming_types_queueservice_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_streaming_types_queueservice_proto_rawDescData)
	})
	return file_pkg_streaming_types_queueservice_proto_rawDescData
}

var file_pkg_streaming_types_queueservice_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_pkg_streaming_types_queueservice_proto_goTypes = []interface{}{
	(*DataFrameProto)(nil),     // 0: streaming.DataFrameProto
	(*DataFrameListProto)(nil), // 1: streaming.DataFrameListProto
	(*AttributesProto)(nil),    // 2: streaming.AttributesProto
	nil,                        // 3: streaming.DataFrameProto.TagsEntry
	nil,                        // 4: streaming.AttributesProto.AttributesEntry
}
var file_pkg_streaming_types_queueservice_proto_depIdxs = []int32{
	3, // 0: streaming.DataFrameProto.tags:type_name -> streaming.DataFrameProto.TagsEntry
	0, // 1: streaming.DataFrameListProto.index:type_name -> streaming.DataFrameProto
	4, // 2: streaming.AttributesProto.attributes:type_name -> streaming.AttributesProto.AttributesEntry
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_pkg_streaming_types_queueservice_proto_init() }
func file_pkg_streaming_types_queueservice_proto_init() {
	if File_pkg_streaming_types_queueservice_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_streaming_types_queueservice_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataFrameProto); i {
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
		file_pkg_streaming_types_queueservice_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataFrameListProto); i {
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
		file_pkg_streaming_types_queueservice_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AttributesProto); i {
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
			RawDescriptor: file_pkg_streaming_types_queueservice_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_streaming_types_queueservice_proto_goTypes,
		DependencyIndexes: file_pkg_streaming_types_queueservice_proto_depIdxs,
		MessageInfos:      file_pkg_streaming_types_queueservice_proto_msgTypes,
	}.Build()
	File_pkg_streaming_types_queueservice_proto = out.File
	file_pkg_streaming_types_queueservice_proto_rawDesc = nil
	file_pkg_streaming_types_queueservice_proto_goTypes = nil
	file_pkg_streaming_types_queueservice_proto_depIdxs = nil
}
