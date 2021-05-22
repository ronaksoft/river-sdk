// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.8
// source: msg.proto

package rony

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

// RedirectReason
type RedirectReason int32

const (
	RedirectReason_ReplicaSetSession RedirectReason = 0
	RedirectReason_ReplicaSetRequest RedirectReason = 1
	RedirectReason_Reserved1         RedirectReason = 3
	RedirectReason_Reserved2         RedirectReason = 4
	RedirectReason_Reserved3         RedirectReason = 5
	RedirectReason_Reserved4         RedirectReason = 6
)

// Enum value maps for RedirectReason.
var (
	RedirectReason_name = map[int32]string{
		0: "ReplicaSetSession",
		1: "ReplicaSetRequest",
		3: "Reserved1",
		4: "Reserved2",
		5: "Reserved3",
		6: "Reserved4",
	}
	RedirectReason_value = map[string]int32{
		"ReplicaSetSession": 0,
		"ReplicaSetRequest": 1,
		"Reserved1":         3,
		"Reserved2":         4,
		"Reserved3":         5,
		"Reserved4":         6,
	}
)

func (x RedirectReason) Enum() *RedirectReason {
	p := new(RedirectReason)
	*p = x
	return p
}

func (x RedirectReason) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RedirectReason) Descriptor() protoreflect.EnumDescriptor {
	return file_msg_proto_enumTypes[0].Descriptor()
}

func (RedirectReason) Type() protoreflect.EnumType {
	return &file_msg_proto_enumTypes[0]
}

func (x RedirectReason) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RedirectReason.Descriptor instead.
func (RedirectReason) EnumDescriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{0}
}

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
		mi := &file_msg_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeyValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeyValue) ProtoMessage() {}

func (x *KeyValue) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use KeyValue.ProtoReflect.Descriptor instead.
func (*KeyValue) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{1}
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
		mi := &file_msg_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageContainer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageContainer) ProtoMessage() {}

func (x *MessageContainer) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use MessageContainer.ProtoReflect.Descriptor instead.
func (*MessageContainer) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{2}
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

	Code        string `protobuf:"bytes,1,opt,name=Code,proto3" json:"Code,omitempty"`
	Items       string `protobuf:"bytes,2,opt,name=Items,proto3" json:"Items,omitempty"`
	Description string `protobuf:"bytes,3,opt,name=Description,proto3" json:"Description,omitempty"`
}

func (x *Error) Reset() {
	*x = Error{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Error) ProtoMessage() {}

func (x *Error) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use Error.ProtoReflect.Descriptor instead.
func (*Error) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{3}
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

func (x *Error) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

// Redirect
type Redirect struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Reason    RedirectReason `protobuf:"varint,100,opt,name=Reason,proto3,enum=rony.RedirectReason" json:"Reason,omitempty"`
	Edges     []*Edge        `protobuf:"bytes,2,rep,name=Edges,proto3" json:"Edges,omitempty"`
	WaitInSec uint32         `protobuf:"varint,3,opt,name=WaitInSec,proto3" json:"WaitInSec,omitempty"`
}

func (x *Redirect) Reset() {
	*x = Redirect{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Redirect) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Redirect) ProtoMessage() {}

func (x *Redirect) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use Redirect.ProtoReflect.Descriptor instead.
func (*Redirect) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{4}
}

func (x *Redirect) GetReason() RedirectReason {
	if x != nil {
		return x.Reason
	}
	return RedirectReason_ReplicaSetSession
}

func (x *Redirect) GetEdges() []*Edge {
	if x != nil {
		return x.Edges
	}
	return nil
}

func (x *Redirect) GetWaitInSec() uint32 {
	if x != nil {
		return x.WaitInSec
	}
	return 0
}

// Edge
type Edge struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReplicaSet uint64   `protobuf:"varint,1,opt,name=ReplicaSet,proto3" json:"ReplicaSet,omitempty"`
	ServerID   string   `protobuf:"bytes,2,opt,name=ServerID,proto3" json:"ServerID,omitempty"`
	HostPorts  []string `protobuf:"bytes,3,rep,name=HostPorts,proto3" json:"HostPorts,omitempty"`
}

func (x *Edge) Reset() {
	*x = Edge{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Edge) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Edge) ProtoMessage() {}

func (x *Edge) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Edge.ProtoReflect.Descriptor instead.
func (*Edge) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{5}
}

func (x *Edge) GetReplicaSet() uint64 {
	if x != nil {
		return x.ReplicaSet
	}
	return 0
}

func (x *Edge) GetServerID() string {
	if x != nil {
		return x.ServerID
	}
	return ""
}

func (x *Edge) GetHostPorts() []string {
	if x != nil {
		return x.HostPorts
	}
	return nil
}

// Edges
type Edges struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Nodes []*Edge `protobuf:"bytes,1,rep,name=Nodes,proto3" json:"Nodes,omitempty"`
}

func (x *Edges) Reset() {
	*x = Edges{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Edges) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Edges) ProtoMessage() {}

func (x *Edges) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Edges.ProtoReflect.Descriptor instead.
func (*Edges) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{6}
}

func (x *Edges) GetNodes() []*Edge {
	if x != nil {
		return x.Nodes
	}
	return nil
}

// GetNodes
// @Function
// @Return: Edges
type GetNodes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReplicaSet []uint64 `protobuf:"varint,1,rep,packed,name=ReplicaSet,proto3" json:"ReplicaSet,omitempty"`
}

func (x *GetNodes) Reset() {
	*x = GetNodes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetNodes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNodes) ProtoMessage() {}

func (x *GetNodes) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNodes.ProtoReflect.Descriptor instead.
func (*GetNodes) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{7}
}

func (x *GetNodes) GetReplicaSet() []uint64 {
	if x != nil {
		return x.ReplicaSet
	}
	return nil
}

// GetNodes
// @Function
// @Return: Edges
type GetAllNodes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetAllNodes) Reset() {
	*x = GetAllNodes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetAllNodes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAllNodes) ProtoMessage() {}

func (x *GetAllNodes) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAllNodes.ProtoReflect.Descriptor instead.
func (*GetAllNodes) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{8}
}

// HttpBody used by REST proxies to fill the output.
type HttpBody struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ContentType string      `protobuf:"bytes,1,opt,name=ContentType,proto3" json:"ContentType,omitempty"`
	Header      []*KeyValue `protobuf:"bytes,2,rep,name=Header,proto3" json:"Header,omitempty"`
	Body        []byte      `protobuf:"bytes,3,opt,name=Body,proto3" json:"Body,omitempty"`
}

func (x *HttpBody) Reset() {
	*x = HttpBody{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HttpBody) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HttpBody) ProtoMessage() {}

func (x *HttpBody) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HttpBody.ProtoReflect.Descriptor instead.
func (*HttpBody) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{9}
}

func (x *HttpBody) GetContentType() string {
	if x != nil {
		return x.ContentType
	}
	return ""
}

func (x *HttpBody) GetHeader() []*KeyValue {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *HttpBody) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
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
	0x6c, 0x75, 0x65, 0x52, 0x06, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x22, 0x32, 0x0a, 0x08, 0x4b,
	0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x4b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x4b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22,
	0x5f, 0x0a, 0x10, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x69,
	0x6e, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x12, 0x33, 0x0a, 0x09, 0x45,
	0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15,
	0x2e, 0x72, 0x6f, 0x6e, 0x79, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x45, 0x6e, 0x76,
	0x65, 0x6c, 0x6f, 0x70, 0x65, 0x52, 0x09, 0x45, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x73,
	0x22, 0x53, 0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x43, 0x6f, 0x64,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x14, 0x0a,
	0x05, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x49, 0x74,
	0x65, 0x6d, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x78, 0x0a, 0x08, 0x52, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63,
	0x74, 0x12, 0x2c, 0x0a, 0x06, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x64, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x14, 0x2e, 0x72, 0x6f, 0x6e, 0x79, 0x2e, 0x52, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63,
	0x74, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x52, 0x06, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12,
	0x20, 0x0a, 0x05, 0x45, 0x64, 0x67, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0a,
	0x2e, 0x72, 0x6f, 0x6e, 0x79, 0x2e, 0x45, 0x64, 0x67, 0x65, 0x52, 0x05, 0x45, 0x64, 0x67, 0x65,
	0x73, 0x12, 0x1c, 0x0a, 0x09, 0x57, 0x61, 0x69, 0x74, 0x49, 0x6e, 0x53, 0x65, 0x63, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x57, 0x61, 0x69, 0x74, 0x49, 0x6e, 0x53, 0x65, 0x63, 0x22,
	0x60, 0x0a, 0x04, 0x45, 0x64, 0x67, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x52, 0x65, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x53, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x52, 0x65, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x53, 0x65, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x53, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x53, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x49, 0x44, 0x12, 0x1c, 0x0a, 0x09, 0x48, 0x6f, 0x73, 0x74, 0x50, 0x6f, 0x72, 0x74, 0x73,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x48, 0x6f, 0x73, 0x74, 0x50, 0x6f, 0x72, 0x74,
	0x73, 0x22, 0x29, 0x0a, 0x05, 0x45, 0x64, 0x67, 0x65, 0x73, 0x12, 0x20, 0x0a, 0x05, 0x4e, 0x6f,
	0x64, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x72, 0x6f, 0x6e, 0x79,
	0x2e, 0x45, 0x64, 0x67, 0x65, 0x52, 0x05, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x22, 0x2a, 0x0a, 0x08,
	0x47, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x52, 0x65, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x53, 0x65, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x04, 0x52, 0x0a, 0x52, 0x65,
	0x70, 0x6c, 0x69, 0x63, 0x61, 0x53, 0x65, 0x74, 0x22, 0x0d, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x41,
	0x6c, 0x6c, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x22, 0x68, 0x0a, 0x08, 0x48, 0x74, 0x74, 0x70, 0x42,
	0x6f, 0x64, 0x79, 0x12, 0x20, 0x0a, 0x0b, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x26, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x72, 0x6f, 0x6e, 0x79, 0x2e, 0x4b, 0x65, 0x79,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x06, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x12, 0x0a,
	0x04, 0x42, 0x6f, 0x64, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x42, 0x6f, 0x64,
	0x79, 0x2a, 0x7a, 0x0a, 0x0e, 0x52, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x52, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x12, 0x15, 0x0a, 0x11, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x53, 0x65,
	0x74, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x52, 0x65,
	0x70, 0x6c, 0x69, 0x63, 0x61, 0x53, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x10,
	0x01, 0x12, 0x0d, 0x0a, 0x09, 0x52, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x31, 0x10, 0x03,
	0x12, 0x0d, 0x0a, 0x09, 0x52, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x32, 0x10, 0x04, 0x12,
	0x0d, 0x0a, 0x09, 0x52, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x33, 0x10, 0x05, 0x12, 0x0d,
	0x0a, 0x09, 0x52, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x34, 0x10, 0x06, 0x42, 0x1b, 0x5a,
	0x19, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x6f, 0x6e, 0x61,
	0x6b, 0x73, 0x6f, 0x66, 0x74, 0x2f, 0x72, 0x6f, 0x6e, 0x79, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
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

var file_msg_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_msg_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_msg_proto_goTypes = []interface{}{
	(RedirectReason)(0),      // 0: rony.RedirectReason
	(*MessageEnvelope)(nil),  // 1: rony.MessageEnvelope
	(*KeyValue)(nil),         // 2: rony.KeyValue
	(*MessageContainer)(nil), // 3: rony.MessageContainer
	(*Error)(nil),            // 4: rony.Error
	(*Redirect)(nil),         // 5: rony.Redirect
	(*Edge)(nil),             // 6: rony.Edge
	(*Edges)(nil),            // 7: rony.Edges
	(*GetNodes)(nil),         // 8: rony.GetNodes
	(*GetAllNodes)(nil),      // 9: rony.GetAllNodes
	(*HttpBody)(nil),         // 10: rony.HttpBody
}
var file_msg_proto_depIdxs = []int32{
	2, // 0: rony.MessageEnvelope.Header:type_name -> rony.KeyValue
	1, // 1: rony.MessageContainer.Envelopes:type_name -> rony.MessageEnvelope
	0, // 2: rony.Redirect.Reason:type_name -> rony.RedirectReason
	6, // 3: rony.Redirect.Edges:type_name -> rony.Edge
	6, // 4: rony.Edges.Nodes:type_name -> rony.Edge
	2, // 5: rony.HttpBody.Header:type_name -> rony.KeyValue
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
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
		file_msg_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_msg_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
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
		file_msg_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
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
		file_msg_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Edge); i {
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
		file_msg_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Edges); i {
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
		file_msg_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetNodes); i {
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
		file_msg_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetAllNodes); i {
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
		file_msg_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HttpBody); i {
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
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_msg_proto_goTypes,
		DependencyIndexes: file_msg_proto_depIdxs,
		EnumInfos:         file_msg_proto_enumTypes,
		MessageInfos:      file_msg_proto_msgTypes,
	}.Build()
	File_msg_proto = out.File
	file_msg_proto_rawDesc = nil
	file_msg_proto_goTypes = nil
	file_msg_proto_depIdxs = nil
}
