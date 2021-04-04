// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: dev.proto

package msg

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

// EchoWithDelay
type EchoWithDelay struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DelayInSeconds int32 `protobuf:"varint,1,opt,name=DelayInSeconds,proto3" json:"DelayInSeconds,omitempty"`
}

func (x *EchoWithDelay) Reset() {
	*x = EchoWithDelay{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dev_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EchoWithDelay) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EchoWithDelay) ProtoMessage() {}

func (x *EchoWithDelay) ProtoReflect() protoreflect.Message {
	mi := &file_dev_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EchoWithDelay.ProtoReflect.Descriptor instead.
func (*EchoWithDelay) Descriptor() ([]byte, []int) {
	return file_dev_proto_rawDescGZIP(), []int{0}
}

func (x *EchoWithDelay) GetDelayInSeconds() int32 {
	if x != nil {
		return x.DelayInSeconds
	}
	return 0
}

type TestRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Payload []byte `protobuf:"bytes,1,opt,name=Payload,proto3" json:"Payload,omitempty"`
	Hash    []byte `protobuf:"bytes,2,opt,name=Hash,proto3" json:"Hash,omitempty"`
}

func (x *TestRequest) Reset() {
	*x = TestRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dev_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestRequest) ProtoMessage() {}

func (x *TestRequest) ProtoReflect() protoreflect.Message {
	mi := &file_dev_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestRequest.ProtoReflect.Descriptor instead.
func (*TestRequest) Descriptor() ([]byte, []int) {
	return file_dev_proto_rawDescGZIP(), []int{1}
}

func (x *TestRequest) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *TestRequest) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

type TestResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hash []byte `protobuf:"bytes,2,opt,name=Hash,proto3" json:"Hash,omitempty"`
}

func (x *TestResponse) Reset() {
	*x = TestResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dev_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestResponse) ProtoMessage() {}

func (x *TestResponse) ProtoReflect() protoreflect.Message {
	mi := &file_dev_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestResponse.ProtoReflect.Descriptor instead.
func (*TestResponse) Descriptor() ([]byte, []int) {
	return file_dev_proto_rawDescGZIP(), []int{2}
}

func (x *TestResponse) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

type TestRequestWithString struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Payload string `protobuf:"bytes,1,opt,name=Payload,proto3" json:"Payload,omitempty"`
	Hash    string `protobuf:"bytes,2,opt,name=Hash,proto3" json:"Hash,omitempty"`
}

func (x *TestRequestWithString) Reset() {
	*x = TestRequestWithString{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dev_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestRequestWithString) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestRequestWithString) ProtoMessage() {}

func (x *TestRequestWithString) ProtoReflect() protoreflect.Message {
	mi := &file_dev_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestRequestWithString.ProtoReflect.Descriptor instead.
func (*TestRequestWithString) Descriptor() ([]byte, []int) {
	return file_dev_proto_rawDescGZIP(), []int{3}
}

func (x *TestRequestWithString) GetPayload() string {
	if x != nil {
		return x.Payload
	}
	return ""
}

func (x *TestRequestWithString) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

type TestResponseWithString struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hash []byte `protobuf:"bytes,1,opt,name=Hash,proto3" json:"Hash,omitempty"`
}

func (x *TestResponseWithString) Reset() {
	*x = TestResponseWithString{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dev_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestResponseWithString) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestResponseWithString) ProtoMessage() {}

func (x *TestResponseWithString) ProtoReflect() protoreflect.Message {
	mi := &file_dev_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestResponseWithString.ProtoReflect.Descriptor instead.
func (*TestResponseWithString) Descriptor() ([]byte, []int) {
	return file_dev_proto_rawDescGZIP(), []int{4}
}

func (x *TestResponseWithString) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

var File_dev_proto protoreflect.FileDescriptor

var file_dev_proto_rawDesc = []byte{
	0x0a, 0x09, 0x64, 0x65, 0x76, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x6d, 0x73, 0x67,
	0x22, 0x37, 0x0a, 0x0d, 0x45, 0x63, 0x68, 0x6f, 0x57, 0x69, 0x74, 0x68, 0x44, 0x65, 0x6c, 0x61,
	0x79, 0x12, 0x26, 0x0a, 0x0e, 0x44, 0x65, 0x6c, 0x61, 0x79, 0x49, 0x6e, 0x53, 0x65, 0x63, 0x6f,
	0x6e, 0x64, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x44, 0x65, 0x6c, 0x61, 0x79,
	0x49, 0x6e, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x22, 0x3b, 0x0a, 0x0b, 0x54, 0x65, 0x73,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x50, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x48, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x04, 0x48, 0x61, 0x73, 0x68, 0x22, 0x22, 0x0a, 0x0c, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x48, 0x61, 0x73, 0x68, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x48, 0x61, 0x73, 0x68, 0x22, 0x45, 0x0a, 0x15, 0x54, 0x65,
	0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x57, 0x69, 0x74, 0x68, 0x53, 0x74, 0x72,
	0x69, 0x6e, 0x67, 0x12, 0x18, 0x0a, 0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x48, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x48, 0x61, 0x73,
	0x68, 0x22, 0x2c, 0x0a, 0x16, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x57, 0x69, 0x74, 0x68, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x48,
	0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x48, 0x61, 0x73, 0x68, 0x42,
	0x08, 0x5a, 0x06, 0x2e, 0x2f, 0x3b, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_dev_proto_rawDescOnce sync.Once
	file_dev_proto_rawDescData = file_dev_proto_rawDesc
)

func file_dev_proto_rawDescGZIP() []byte {
	file_dev_proto_rawDescOnce.Do(func() {
		file_dev_proto_rawDescData = protoimpl.X.CompressGZIP(file_dev_proto_rawDescData)
	})
	return file_dev_proto_rawDescData
}

var file_dev_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_dev_proto_goTypes = []interface{}{
	(*EchoWithDelay)(nil),          // 0: msg.EchoWithDelay
	(*TestRequest)(nil),            // 1: msg.TestRequest
	(*TestResponse)(nil),           // 2: msg.TestResponse
	(*TestRequestWithString)(nil),  // 3: msg.TestRequestWithString
	(*TestResponseWithString)(nil), // 4: msg.TestResponseWithString
}
var file_dev_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_dev_proto_init() }
func file_dev_proto_init() {
	if File_dev_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_dev_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EchoWithDelay); i {
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
		file_dev_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestRequest); i {
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
		file_dev_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestResponse); i {
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
		file_dev_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestRequestWithString); i {
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
		file_dev_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestResponseWithString); i {
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
			RawDescriptor: file_dev_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_dev_proto_goTypes,
		DependencyIndexes: file_dev_proto_depIdxs,
		MessageInfos:      file_dev_proto_msgTypes,
	}.Build()
	File_dev_proto = out.File
	file_dev_proto_rawDesc = nil
	file_dev_proto_goTypes = nil
	file_dev_proto_depIdxs = nil
}
