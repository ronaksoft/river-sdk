// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: chat.core.message.actions.proto

package msg

import (
	fmt "fmt"
	github_com_gogo_protobuf_proto "github.com/gogo/protobuf/proto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// MessageActionGroupAddUser
type MessageActionGroupAddUser struct {
	UserIDs []int64 `protobuf:"varint,1,rep,name=UserIDs" json:"UserIDs,omitempty"`
}

func (m *MessageActionGroupAddUser) Reset()         { *m = MessageActionGroupAddUser{} }
func (m *MessageActionGroupAddUser) String() string { return proto.CompactTextString(m) }
func (*MessageActionGroupAddUser) ProtoMessage()    {}
func (*MessageActionGroupAddUser) Descriptor() ([]byte, []int) {
	return fileDescriptor_e134edebab6f8250, []int{0}
}
func (m *MessageActionGroupAddUser) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MessageActionGroupAddUser) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MessageActionGroupAddUser.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MessageActionGroupAddUser) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MessageActionGroupAddUser.Merge(m, src)
}
func (m *MessageActionGroupAddUser) XXX_Size() int {
	return m.Size()
}
func (m *MessageActionGroupAddUser) XXX_DiscardUnknown() {
	xxx_messageInfo_MessageActionGroupAddUser.DiscardUnknown(m)
}

var xxx_messageInfo_MessageActionGroupAddUser proto.InternalMessageInfo

func (m *MessageActionGroupAddUser) GetUserIDs() []int64 {
	if m != nil {
		return m.UserIDs
	}
	return nil
}

// MessageActionGroupDeleteUser
type MessageActionGroupDeleteUser struct {
	UserIDs []int64 `protobuf:"varint,1,rep,name=UserIDs" json:"UserIDs,omitempty"`
}

func (m *MessageActionGroupDeleteUser) Reset()         { *m = MessageActionGroupDeleteUser{} }
func (m *MessageActionGroupDeleteUser) String() string { return proto.CompactTextString(m) }
func (*MessageActionGroupDeleteUser) ProtoMessage()    {}
func (*MessageActionGroupDeleteUser) Descriptor() ([]byte, []int) {
	return fileDescriptor_e134edebab6f8250, []int{1}
}
func (m *MessageActionGroupDeleteUser) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MessageActionGroupDeleteUser) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MessageActionGroupDeleteUser.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MessageActionGroupDeleteUser) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MessageActionGroupDeleteUser.Merge(m, src)
}
func (m *MessageActionGroupDeleteUser) XXX_Size() int {
	return m.Size()
}
func (m *MessageActionGroupDeleteUser) XXX_DiscardUnknown() {
	xxx_messageInfo_MessageActionGroupDeleteUser.DiscardUnknown(m)
}

var xxx_messageInfo_MessageActionGroupDeleteUser proto.InternalMessageInfo

func (m *MessageActionGroupDeleteUser) GetUserIDs() []int64 {
	if m != nil {
		return m.UserIDs
	}
	return nil
}

// MessageActionGroupCreated
type MessageActionGroupCreated struct {
	GroupTitle string  `protobuf:"bytes,1,req,name=GroupTitle" json:"GroupTitle"`
	UserIDs    []int64 `protobuf:"varint,2,rep,name=UserIDs" json:"UserIDs,omitempty"`
}

func (m *MessageActionGroupCreated) Reset()         { *m = MessageActionGroupCreated{} }
func (m *MessageActionGroupCreated) String() string { return proto.CompactTextString(m) }
func (*MessageActionGroupCreated) ProtoMessage()    {}
func (*MessageActionGroupCreated) Descriptor() ([]byte, []int) {
	return fileDescriptor_e134edebab6f8250, []int{2}
}
func (m *MessageActionGroupCreated) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MessageActionGroupCreated) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MessageActionGroupCreated.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MessageActionGroupCreated) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MessageActionGroupCreated.Merge(m, src)
}
func (m *MessageActionGroupCreated) XXX_Size() int {
	return m.Size()
}
func (m *MessageActionGroupCreated) XXX_DiscardUnknown() {
	xxx_messageInfo_MessageActionGroupCreated.DiscardUnknown(m)
}

var xxx_messageInfo_MessageActionGroupCreated proto.InternalMessageInfo

func (m *MessageActionGroupCreated) GetGroupTitle() string {
	if m != nil {
		return m.GroupTitle
	}
	return ""
}

func (m *MessageActionGroupCreated) GetUserIDs() []int64 {
	if m != nil {
		return m.UserIDs
	}
	return nil
}

// MessageActionGroupTitleChanged
type MessageActionGroupTitleChanged struct {
	GroupTitle string `protobuf:"bytes,1,req,name=GroupTitle" json:"GroupTitle"`
}

func (m *MessageActionGroupTitleChanged) Reset()         { *m = MessageActionGroupTitleChanged{} }
func (m *MessageActionGroupTitleChanged) String() string { return proto.CompactTextString(m) }
func (*MessageActionGroupTitleChanged) ProtoMessage()    {}
func (*MessageActionGroupTitleChanged) Descriptor() ([]byte, []int) {
	return fileDescriptor_e134edebab6f8250, []int{3}
}
func (m *MessageActionGroupTitleChanged) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MessageActionGroupTitleChanged) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MessageActionGroupTitleChanged.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MessageActionGroupTitleChanged) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MessageActionGroupTitleChanged.Merge(m, src)
}
func (m *MessageActionGroupTitleChanged) XXX_Size() int {
	return m.Size()
}
func (m *MessageActionGroupTitleChanged) XXX_DiscardUnknown() {
	xxx_messageInfo_MessageActionGroupTitleChanged.DiscardUnknown(m)
}

var xxx_messageInfo_MessageActionGroupTitleChanged proto.InternalMessageInfo

func (m *MessageActionGroupTitleChanged) GetGroupTitle() string {
	if m != nil {
		return m.GroupTitle
	}
	return ""
}

type MessageActionGroupPhotoChanged struct {
	Photo *GroupPhoto `protobuf:"bytes,1,opt,name=Photo" json:"Photo,omitempty"`
}

func (m *MessageActionGroupPhotoChanged) Reset()         { *m = MessageActionGroupPhotoChanged{} }
func (m *MessageActionGroupPhotoChanged) String() string { return proto.CompactTextString(m) }
func (*MessageActionGroupPhotoChanged) ProtoMessage()    {}
func (*MessageActionGroupPhotoChanged) Descriptor() ([]byte, []int) {
	return fileDescriptor_e134edebab6f8250, []int{4}
}
func (m *MessageActionGroupPhotoChanged) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MessageActionGroupPhotoChanged) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MessageActionGroupPhotoChanged.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MessageActionGroupPhotoChanged) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MessageActionGroupPhotoChanged.Merge(m, src)
}
func (m *MessageActionGroupPhotoChanged) XXX_Size() int {
	return m.Size()
}
func (m *MessageActionGroupPhotoChanged) XXX_DiscardUnknown() {
	xxx_messageInfo_MessageActionGroupPhotoChanged.DiscardUnknown(m)
}

var xxx_messageInfo_MessageActionGroupPhotoChanged proto.InternalMessageInfo

func (m *MessageActionGroupPhotoChanged) GetPhoto() *GroupPhoto {
	if m != nil {
		return m.Photo
	}
	return nil
}

// MessageActionClearHistory
type MessageActionClearHistory struct {
	MaxID  int64 `protobuf:"varint,1,req,name=MaxID" json:"MaxID"`
	Delete bool  `protobuf:"varint,2,req,name=Delete" json:"Delete"`
}

func (m *MessageActionClearHistory) Reset()         { *m = MessageActionClearHistory{} }
func (m *MessageActionClearHistory) String() string { return proto.CompactTextString(m) }
func (*MessageActionClearHistory) ProtoMessage()    {}
func (*MessageActionClearHistory) Descriptor() ([]byte, []int) {
	return fileDescriptor_e134edebab6f8250, []int{5}
}
func (m *MessageActionClearHistory) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MessageActionClearHistory) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MessageActionClearHistory.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MessageActionClearHistory) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MessageActionClearHistory.Merge(m, src)
}
func (m *MessageActionClearHistory) XXX_Size() int {
	return m.Size()
}
func (m *MessageActionClearHistory) XXX_DiscardUnknown() {
	xxx_messageInfo_MessageActionClearHistory.DiscardUnknown(m)
}

var xxx_messageInfo_MessageActionClearHistory proto.InternalMessageInfo

func (m *MessageActionClearHistory) GetMaxID() int64 {
	if m != nil {
		return m.MaxID
	}
	return 0
}

func (m *MessageActionClearHistory) GetDelete() bool {
	if m != nil {
		return m.Delete
	}
	return false
}

// MessageActionContactRegistered
type MessageActionContactRegistered struct {
}

func (m *MessageActionContactRegistered) Reset()         { *m = MessageActionContactRegistered{} }
func (m *MessageActionContactRegistered) String() string { return proto.CompactTextString(m) }
func (*MessageActionContactRegistered) ProtoMessage()    {}
func (*MessageActionContactRegistered) Descriptor() ([]byte, []int) {
	return fileDescriptor_e134edebab6f8250, []int{6}
}
func (m *MessageActionContactRegistered) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MessageActionContactRegistered) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MessageActionContactRegistered.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MessageActionContactRegistered) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MessageActionContactRegistered.Merge(m, src)
}
func (m *MessageActionContactRegistered) XXX_Size() int {
	return m.Size()
}
func (m *MessageActionContactRegistered) XXX_DiscardUnknown() {
	xxx_messageInfo_MessageActionContactRegistered.DiscardUnknown(m)
}

var xxx_messageInfo_MessageActionContactRegistered proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MessageActionGroupAddUser)(nil), "msg.MessageActionGroupAddUser")
	proto.RegisterType((*MessageActionGroupDeleteUser)(nil), "msg.MessageActionGroupDeleteUser")
	proto.RegisterType((*MessageActionGroupCreated)(nil), "msg.MessageActionGroupCreated")
	proto.RegisterType((*MessageActionGroupTitleChanged)(nil), "msg.MessageActionGroupTitleChanged")
	proto.RegisterType((*MessageActionGroupPhotoChanged)(nil), "msg.MessageActionGroupPhotoChanged")
	proto.RegisterType((*MessageActionClearHistory)(nil), "msg.MessageActionClearHistory")
	proto.RegisterType((*MessageActionContactRegistered)(nil), "msg.MessageActionContactRegistered")
}

func init() { proto.RegisterFile("chat.core.message.actions.proto", fileDescriptor_e134edebab6f8250) }

var fileDescriptor_e134edebab6f8250 = []byte{
	// 314 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0x4f, 0x4b, 0xc3, 0x40,
	0x10, 0xc5, 0xb3, 0x89, 0xf5, 0xcf, 0x7a, 0x10, 0x02, 0x42, 0x2c, 0x65, 0x1b, 0x82, 0x42, 0x4e,
	0x8b, 0x78, 0x13, 0xbc, 0xb4, 0x29, 0xd6, 0x1e, 0x0a, 0x12, 0xec, 0x59, 0x96, 0x64, 0x48, 0x03,
	0x6d, 0x37, 0xec, 0x8e, 0x60, 0xbf, 0x85, 0x1f, 0xab, 0xc7, 0x1e, 0x3d, 0x89, 0x24, 0x5f, 0x44,
	0x4c, 0x52, 0x5a, 0x9a, 0x1e, 0x7a, 0x0a, 0xf3, 0xde, 0xbc, 0xf9, 0xbd, 0x2c, 0xed, 0x46, 0x53,
	0x81, 0x3c, 0x92, 0x0a, 0xf8, 0x1c, 0xb4, 0x16, 0x09, 0x70, 0x11, 0x61, 0x2a, 0x17, 0x9a, 0x67,
	0x4a, 0xa2, 0xb4, 0xad, 0xb9, 0x4e, 0xda, 0xd7, 0xdb, 0x2d, 0x5c, 0x66, 0x50, 0x7b, 0xde, 0x23,
	0xbd, 0x19, 0x57, 0xa1, 0x5e, 0x99, 0x19, 0x2a, 0xf9, 0x91, 0xf5, 0xe2, 0x78, 0xa2, 0x41, 0xd9,
	0x1d, 0x7a, 0xf6, 0xff, 0x1d, 0x0d, 0xb4, 0x43, 0x5c, 0xcb, 0xb7, 0xfa, 0xe6, 0x3d, 0x09, 0x37,
	0x92, 0xf7, 0x44, 0x3b, 0xcd, 0xe8, 0x00, 0x66, 0x80, 0x70, 0x44, 0xfa, 0xfd, 0x10, 0x38, 0x50,
	0x20, 0x10, 0x62, 0xfb, 0x96, 0xd2, 0x72, 0x7e, 0x4b, 0x71, 0x06, 0x0e, 0x71, 0x4d, 0xff, 0xa2,
	0x7f, 0xb2, 0xfa, 0xe9, 0x1a, 0xe1, 0x8e, 0xbe, 0x0b, 0x30, 0x9b, 0x80, 0x67, 0xca, 0x9a, 0x80,
	0x32, 0x18, 0x4c, 0xc5, 0x22, 0x39, 0x96, 0xe2, 0x0d, 0x0f, 0xdd, 0x79, 0x9d, 0x4a, 0x94, 0x9b,
	0x3b, 0x77, 0xb4, 0x55, 0xce, 0x0e, 0x71, 0x89, 0x7f, 0xf9, 0x70, 0xc5, 0xe7, 0x3a, 0xe1, 0xdb,
	0xb5, 0xb0, 0x72, 0xbd, 0xc9, 0xde, 0x1f, 0x07, 0x33, 0x10, 0xea, 0x25, 0xd5, 0x28, 0xd5, 0xd2,
	0x6e, 0xd3, 0xd6, 0x58, 0x7c, 0x8e, 0x06, 0x65, 0x0d, 0xab, 0xae, 0x51, 0x49, 0x76, 0x87, 0x9e,
	0x56, 0xcf, 0xea, 0x98, 0xae, 0xe9, 0x9f, 0xd7, 0x66, 0xad, 0x79, 0xee, 0x5e, 0xbf, 0x40, 0x2e,
	0x50, 0x44, 0x18, 0x42, 0x92, 0x6a, 0x04, 0x05, 0x71, 0xdf, 0x59, 0xe5, 0x8c, 0xac, 0x73, 0x46,
	0x7e, 0x73, 0x46, 0xbe, 0x0a, 0x66, 0xac, 0x0b, 0x66, 0x7c, 0x17, 0xcc, 0xf8, 0x0b, 0x00, 0x00,
	0xff, 0xff, 0xfb, 0xf4, 0xfb, 0xf9, 0x3b, 0x02, 0x00, 0x00,
}

func (m *MessageActionGroupAddUser) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MessageActionGroupAddUser) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.UserIDs) > 0 {
		for _, num := range m.UserIDs {
			dAtA[i] = 0x8
			i++
			i = encodeVarintChatCoreMessageActions(dAtA, i, uint64(num))
		}
	}
	return i, nil
}

func (m *MessageActionGroupDeleteUser) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MessageActionGroupDeleteUser) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.UserIDs) > 0 {
		for _, num := range m.UserIDs {
			dAtA[i] = 0x8
			i++
			i = encodeVarintChatCoreMessageActions(dAtA, i, uint64(num))
		}
	}
	return i, nil
}

func (m *MessageActionGroupCreated) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MessageActionGroupCreated) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintChatCoreMessageActions(dAtA, i, uint64(len(m.GroupTitle)))
	i += copy(dAtA[i:], m.GroupTitle)
	if len(m.UserIDs) > 0 {
		for _, num := range m.UserIDs {
			dAtA[i] = 0x10
			i++
			i = encodeVarintChatCoreMessageActions(dAtA, i, uint64(num))
		}
	}
	return i, nil
}

func (m *MessageActionGroupTitleChanged) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MessageActionGroupTitleChanged) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintChatCoreMessageActions(dAtA, i, uint64(len(m.GroupTitle)))
	i += copy(dAtA[i:], m.GroupTitle)
	return i, nil
}

func (m *MessageActionGroupPhotoChanged) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MessageActionGroupPhotoChanged) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Photo != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintChatCoreMessageActions(dAtA, i, uint64(m.Photo.Size()))
		n1, err := m.Photo.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	return i, nil
}

func (m *MessageActionClearHistory) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MessageActionClearHistory) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintChatCoreMessageActions(dAtA, i, uint64(m.MaxID))
	dAtA[i] = 0x10
	i++
	if m.Delete {
		dAtA[i] = 1
	} else {
		dAtA[i] = 0
	}
	i++
	return i, nil
}

func (m *MessageActionContactRegistered) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MessageActionContactRegistered) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func encodeVarintChatCoreMessageActions(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *MessageActionGroupAddUser) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.UserIDs) > 0 {
		for _, e := range m.UserIDs {
			n += 1 + sovChatCoreMessageActions(uint64(e))
		}
	}
	return n
}

func (m *MessageActionGroupDeleteUser) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.UserIDs) > 0 {
		for _, e := range m.UserIDs {
			n += 1 + sovChatCoreMessageActions(uint64(e))
		}
	}
	return n
}

func (m *MessageActionGroupCreated) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.GroupTitle)
	n += 1 + l + sovChatCoreMessageActions(uint64(l))
	if len(m.UserIDs) > 0 {
		for _, e := range m.UserIDs {
			n += 1 + sovChatCoreMessageActions(uint64(e))
		}
	}
	return n
}

func (m *MessageActionGroupTitleChanged) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.GroupTitle)
	n += 1 + l + sovChatCoreMessageActions(uint64(l))
	return n
}

func (m *MessageActionGroupPhotoChanged) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Photo != nil {
		l = m.Photo.Size()
		n += 1 + l + sovChatCoreMessageActions(uint64(l))
	}
	return n
}

func (m *MessageActionClearHistory) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovChatCoreMessageActions(uint64(m.MaxID))
	n += 2
	return n
}

func (m *MessageActionContactRegistered) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovChatCoreMessageActions(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozChatCoreMessageActions(x uint64) (n int) {
	return sovChatCoreMessageActions(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MessageActionGroupAddUser) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChatCoreMessageActions
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MessageActionGroupAddUser: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MessageActionGroupAddUser: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType == 0 {
				var v int64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowChatCoreMessageActions
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= (int64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.UserIDs = append(m.UserIDs, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowChatCoreMessageActions
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthChatCoreMessageActions
				}
				postIndex := iNdEx + packedLen
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.UserIDs) == 0 {
					m.UserIDs = make([]int64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v int64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowChatCoreMessageActions
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= (int64(b) & 0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.UserIDs = append(m.UserIDs, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field UserIDs", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipChatCoreMessageActions(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChatCoreMessageActions
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MessageActionGroupDeleteUser) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChatCoreMessageActions
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MessageActionGroupDeleteUser: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MessageActionGroupDeleteUser: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType == 0 {
				var v int64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowChatCoreMessageActions
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= (int64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.UserIDs = append(m.UserIDs, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowChatCoreMessageActions
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthChatCoreMessageActions
				}
				postIndex := iNdEx + packedLen
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.UserIDs) == 0 {
					m.UserIDs = make([]int64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v int64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowChatCoreMessageActions
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= (int64(b) & 0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.UserIDs = append(m.UserIDs, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field UserIDs", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipChatCoreMessageActions(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChatCoreMessageActions
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MessageActionGroupCreated) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChatCoreMessageActions
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MessageActionGroupCreated: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MessageActionGroupCreated: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GroupTitle", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatCoreMessageActions
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthChatCoreMessageActions
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.GroupTitle = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType == 0 {
				var v int64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowChatCoreMessageActions
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= (int64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.UserIDs = append(m.UserIDs, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowChatCoreMessageActions
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthChatCoreMessageActions
				}
				postIndex := iNdEx + packedLen
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.UserIDs) == 0 {
					m.UserIDs = make([]int64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v int64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowChatCoreMessageActions
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= (int64(b) & 0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.UserIDs = append(m.UserIDs, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field UserIDs", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipChatCoreMessageActions(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChatCoreMessageActions
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("GroupTitle")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MessageActionGroupTitleChanged) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChatCoreMessageActions
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MessageActionGroupTitleChanged: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MessageActionGroupTitleChanged: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GroupTitle", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatCoreMessageActions
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthChatCoreMessageActions
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.GroupTitle = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000001)
		default:
			iNdEx = preIndex
			skippy, err := skipChatCoreMessageActions(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChatCoreMessageActions
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("GroupTitle")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MessageActionGroupPhotoChanged) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChatCoreMessageActions
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MessageActionGroupPhotoChanged: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MessageActionGroupPhotoChanged: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Photo", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatCoreMessageActions
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthChatCoreMessageActions
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Photo == nil {
				m.Photo = &GroupPhoto{}
			}
			if err := m.Photo.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipChatCoreMessageActions(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChatCoreMessageActions
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MessageActionClearHistory) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChatCoreMessageActions
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MessageActionClearHistory: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MessageActionClearHistory: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxID", wireType)
			}
			m.MaxID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatCoreMessageActions
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxID |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Delete", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatCoreMessageActions
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Delete = bool(v != 0)
			hasFields[0] |= uint64(0x00000002)
		default:
			iNdEx = preIndex
			skippy, err := skipChatCoreMessageActions(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChatCoreMessageActions
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("MaxID")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Delete")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MessageActionContactRegistered) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChatCoreMessageActions
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MessageActionContactRegistered: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MessageActionContactRegistered: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipChatCoreMessageActions(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChatCoreMessageActions
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipChatCoreMessageActions(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowChatCoreMessageActions
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowChatCoreMessageActions
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowChatCoreMessageActions
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthChatCoreMessageActions
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowChatCoreMessageActions
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipChatCoreMessageActions(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthChatCoreMessageActions = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowChatCoreMessageActions   = fmt.Errorf("proto: integer overflow")
)