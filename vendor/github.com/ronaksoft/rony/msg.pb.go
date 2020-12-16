// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: msg.proto

package rony

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

// MessageEnvelope
// This type of message will be used to contain another ProtoBuffer Message inside
type MessageEnvelope struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Constructor int64       `protobuf:"varint,1,opt,name=Constructor,proto3" json:"Constructor,omitempty"`
	RequestID   uint64      `protobuf:"fixed64,2,opt,name=RequestID,proto3" json:"RequestID,omitempty"`
	Message     []byte      `protobuf:"bytes,4,opt,name=Message,proto3" json:"Message,omitempty"`
	Auth        []byte      `protobuf:"bytes,8,opt,name=Auth,proto3" json:"Auth,omitempty"`
	Header      []*KeyValue `protobuf:"bytes,10,rep,name=Header,proto3" json:"Header,omitempty"`
}

func (x *MessageEnvelope) Reset() {
	*x = MessageEnvelope{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageEnvelope) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageEnvelope) ProtoMessage() {}

func (x *MessageEnvelope) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageEnvelope.ProtoReflect.Descriptor instead.
func (*MessageEnvelope) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{0}
}

func (x *MessageEnvelope) GetConstructor() int64 {
	if x != nil {
		return x.Constructor
	}
	return 0
}

func (x *MessageEnvelope) GetRequestID() uint64 {
	if x != nil {
		return x.RequestID
	}
	return 0
}

func (x *MessageEnvelope) GetMessage() []byte {
	if x != nil {
		return x.Message
	}
	return nil
}

func (x *MessageEnvelope) GetAuth() []byte {
	if x != nil {
		return x.Auth
	}
	return nil
}

func (x *MessageEnvelope) GetHeader() []*KeyValue {
	if x != nil {
		return x.Header
	}
	return nil
}

// MessageContainer
// This type of message will be used to send multi messages inside a single container message
type MessageContainer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Length    int32              `protobuf:"varint,1,opt,name=Length,proto3" json:"Length,omitempty"`
	Envelopes []*MessageEnvelope `protobuf:"bytes,2,rep,name=Envelopes,proto3" json:"Envelopes,omitempty"`
}

func (x *MessageContainer) Reset() {
	*x = MessageContainer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageContainer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageContainer) ProtoMessage() {}

func (x *MessageContainer) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageContainer.ProtoReflect.Descriptor instead.
func (*MessageContainer) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{1}
}

func (x *MessageContainer) GetLength() int32 {
	if x != nil {
		return x.Length
	}
	return 0
}

func (x *MessageContainer) GetEnvelopes() []*MessageEnvelope {
	if x != nil {
		return x.Envelopes
	}
	return nil
}

// Error
type Error struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code               string   `protobuf:"bytes,1,opt,name=Code,proto3" json:"Code,omitempty"`
	Items              string   `protobuf:"bytes,2,opt,name=Items,proto3" json:"Items,omitempty"`
	Template           string   `protobuf:"bytes,3,opt,name=Template,proto3" json:"Template,omitempty"`
	TemplateItems      []string `protobuf:"bytes,4,rep,name=TemplateItems,proto3" json:"TemplateItems,omitempty"`
	LocalTemplate      string   `protobuf:"bytes,5,opt,name=LocalTemplate,proto3" json:"LocalTemplate,omitempty"`
	LocalTemplateItems []string `protobuf:"bytes,6,rep,name=LocalTemplateItems,proto3" json:"LocalTemplateItems,omitempty"`
}

func (x *Error) Reset() {
	*x = Error{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Error) ProtoMessage() {}

func (x *Error) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Error.ProtoReflect.Descriptor instead.
func (*Error) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{2}
}

func (x *Error) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *Error) GetItems() string {
	if x != nil {
		return x.Items
	}
	return ""
}

func (x *Error) GetTemplate() string {
	if x != nil {
		return x.Template
	}
	return ""
}

func (x *Error) GetTemplateItems() []string {
	if x != nil {
		return x.TemplateItems
	}
	return nil
}

func (x *Error) GetLocalTemplate() string {
	if x != nil {
		return x.LocalTemplate
	}
	return ""
}

func (x *Error) GetLocalTemplateItems() []string {
	if x != nil {
		return x.LocalTemplateItems
	}
	return nil
}

// Redirect
type Redirect struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LeaderHostPort []string `protobuf:"bytes,1,rep,name=LeaderHostPort,proto3" json:"LeaderHostPort,omitempty"`
	HostPorts      []string `protobuf:"bytes,2,rep,name=HostPorts,proto3" json:"HostPorts,omitempty"`
	ServerID       string   `protobuf:"bytes,3,opt,name=ServerID,proto3" json:"ServerID,omitempty"`
	WaitInSec      uint32   `protobuf:"varint,4,opt,name=WaitInSec,proto3" json:"WaitInSec,omitempty"`
}

func (x *Redirect) Reset() {
	*x = Redirect{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Redirect) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Redirect) ProtoMessage() {}

func (x *Redirect) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Redirect.ProtoReflect.Descriptor instead.
func (*Redirect) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{3}
}

func (x *Redirect) GetLeaderHostPort() []string {
	if x != nil {
		return x.LeaderHostPort
	}
	return nil
}

func (x *Redirect) GetHostPorts() []string {
	if x != nil {
		return x.HostPorts
	}
	return nil
}

func (x *Redirect) GetServerID() string {
	if x != nil {
		return x.ServerID
	}
	return ""
}

func (x *Redirect) GetWaitInSec() uint32 {
	if x != nil {
		return x.WaitInSec
	}
	return 0
}

// KeyValue
type KeyValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"`
}

func (x *KeyValue) Reset() {
	*x = KeyValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeyValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeyValue) ProtoMessage() {}

func (x *KeyValue) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KeyValue.ProtoReflect.Descriptor instead.
func (*KeyValue) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{4}
}

func (x *KeyValue) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *KeyValue) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_msg_proto protoreflect.FileDescriptor

var file_msg_proto_rawDesc = []byte{
	0x0a, 0x09, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x72, 0x6f, 0x6e,
	0x79, 0x22, 0xa7, 0x01, 0x0a, 0x0f, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x45, 0x6e, 0x76,
	0x65, 0x6c, 0x6f, 0x70, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x43, 0x6f, 0x6e, 0x73, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x43, 0x6f, 0x6e, 0x73,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x06, 0x52, 0x09, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x41, 0x75, 0x74, 0x68, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x41,
	0x75, 0x74, 0x68, 0x12, 0x26, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x0a, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x72, 0x6f, 0x6e, 0x79, 0x2e, 0x4b, 0x65, 0x79, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x52, 0x06, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x22, 0x5f, 0x0a, 0x10, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x12,
	0x16, 0x0a, 0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x12, 0x33, 0x0a, 0x09, 0x45, 0x6e, 0x76, 0x65, 0x6c,
	0x6f, 0x70, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x72, 0x6f, 0x6e,
	0x79, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x45, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70,
	0x65, 0x52, 0x09, 0x45, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x73, 0x22, 0xc9, 0x01, 0x0a,
	0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x49, 0x74,
	0x65, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x49, 0x74, 0x65, 0x6d, 0x73,
	0x12, 0x1a, 0x0a, 0x08, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x12, 0x24, 0x0a, 0x0d,
	0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x04, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x0d, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65,
	0x6d, 0x73, 0x12, 0x24, 0x0a, 0x0d, 0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x54, 0x65, 0x6d, 0x70, 0x6c,
	0x61, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x4c, 0x6f, 0x63, 0x61, 0x6c,
	0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x12, 0x2e, 0x0a, 0x12, 0x4c, 0x6f, 0x63, 0x61,
	0x6c, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x06,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x12, 0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x54, 0x65, 0x6d, 0x70, 0x6c,
	0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x22, 0x8a, 0x01, 0x0a, 0x08, 0x52, 0x65, 0x64,
	0x69, 0x72, 0x65, 0x63, 0x74, 0x12, 0x26, 0x0a, 0x0e, 0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x48,
	0x6f, 0x73, 0x74, 0x50, 0x6f, 0x72, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0e, 0x4c,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x48, 0x6f, 0x73, 0x74, 0x50, 0x6f, 0x72, 0x74, 0x12, 0x1c, 0x0a,
	0x09, 0x48, 0x6f, 0x73, 0x74, 0x50, 0x6f, 0x72, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x09, 0x48, 0x6f, 0x73, 0x74, 0x50, 0x6f, 0x72, 0x74, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x44, 0x12, 0x1c, 0x0a, 0x09, 0x57, 0x61, 0x69, 0x74, 0x49,
	0x6e, 0x53, 0x65, 0x63, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x57, 0x61, 0x69, 0x74,
	0x49, 0x6e, 0x53, 0x65, 0x63, 0x22, 0x32, 0x0a, 0x08, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x4b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x1b, 0x5a, 0x19, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x6f, 0x6e, 0x61, 0x6b, 0x73, 0x6f, 0x66,
	0x74, 0x2f, 0x72, 0x6f, 0x6e, 0x79, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_msg_proto_rawDescOnce sync.Once
	file_msg_proto_rawDescData = file_msg_proto_rawDesc
)

func file_msg_proto_rawDescGZIP() []byte {
	file_msg_proto_rawDescOnce.Do(func() {
		file_msg_proto_rawDescData = protoimpl.X.CompressGZIP(file_msg_proto_rawDescData)
	})
	return file_msg_proto_rawDescData
}

var file_msg_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_msg_proto_goTypes = []interface{}{
	(*MessageEnvelope)(nil),  // 0: rony.MessageEnvelope
	(*MessageContainer)(nil), // 1: rony.MessageContainer
	(*Error)(nil),            // 2: rony.Error
	(*Redirect)(nil),         // 3: rony.Redirect
	(*KeyValue)(nil),         // 4: rony.KeyValue
}
var file_msg_proto_depIdxs = []int32{
	4, // 0: rony.MessageEnvelope.Header:type_name -> rony.KeyValue
	0, // 1: rony.MessageContainer.Envelopes:type_name -> rony.MessageEnvelope
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_msg_proto_init() }
func file_msg_proto_init() {
	if File_msg_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_msg_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageEnvelope); i {
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
		file_msg_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageContainer); i {
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
		file_msg_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Error); i {
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
		file_msg_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Redirect); i {
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
		file_msg_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KeyValue); i {
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
			RawDescriptor: file_msg_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_msg_proto_goTypes,
		DependencyIndexes: file_msg_proto_depIdxs,
		MessageInfos:      file_msg_proto_msgTypes,
	}.Build()
	File_msg_proto = out.File
	file_msg_proto_rawDesc = nil
	file_msg_proto_goTypes = nil
	file_msg_proto_depIdxs = nil
}
