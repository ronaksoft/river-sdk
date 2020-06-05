// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: team.proto

package msg

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	github_com_gogo_protobuf_proto "github.com/gogo/protobuf/proto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// TeamGet
// @Function
// @Return: Team
type TeamGet struct {
	ID int64 `protobuf:"varint,1,req,name=ID" json:"ID"`
}

func (m *TeamGet) Reset()         { *m = TeamGet{} }
func (m *TeamGet) String() string { return proto.CompactTextString(m) }
func (*TeamGet) ProtoMessage()    {}
func (*TeamGet) Descriptor() ([]byte, []int) {
	return fileDescriptor_8b4e9e93d7b2c6bb, []int{0}
}
func (m *TeamGet) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TeamGet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TeamGet.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TeamGet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TeamGet.Merge(m, src)
}
func (m *TeamGet) XXX_Size() int {
	return m.Size()
}
func (m *TeamGet) XXX_DiscardUnknown() {
	xxx_messageInfo_TeamGet.DiscardUnknown(m)
}

var xxx_messageInfo_TeamGet proto.InternalMessageInfo

func (m *TeamGet) GetID() int64 {
	if m != nil {
		return m.ID
	}
	return 0
}

// TeamCreate
// @Function
// @Return: Team
type TeamCreate struct {
	Name       string `protobuf:"bytes,1,req,name=Name" json:"Name"`
	Capacity   int32  `protobuf:"varint,2,req,name=Capacity" json:"Capacity"`
	ExpireDate int64  `protobuf:"varint,3,req,name=ExpireDate" json:"ExpireDate"`
}

func (m *TeamCreate) Reset()         { *m = TeamCreate{} }
func (m *TeamCreate) String() string { return proto.CompactTextString(m) }
func (*TeamCreate) ProtoMessage()    {}
func (*TeamCreate) Descriptor() ([]byte, []int) {
	return fileDescriptor_8b4e9e93d7b2c6bb, []int{1}
}
func (m *TeamCreate) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TeamCreate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TeamCreate.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TeamCreate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TeamCreate.Merge(m, src)
}
func (m *TeamCreate) XXX_Size() int {
	return m.Size()
}
func (m *TeamCreate) XXX_DiscardUnknown() {
	xxx_messageInfo_TeamCreate.DiscardUnknown(m)
}

var xxx_messageInfo_TeamCreate proto.InternalMessageInfo

func (m *TeamCreate) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *TeamCreate) GetCapacity() int32 {
	if m != nil {
		return m.Capacity
	}
	return 0
}

func (m *TeamCreate) GetExpireDate() int64 {
	if m != nil {
		return m.ExpireDate
	}
	return 0
}

// TeamAddMember
// @Function
// @Return: Bool
type TeamAddMember struct {
	TeamID int64 `protobuf:"varint,1,req,name=TeamID" json:"TeamID"`
	UserID int64 `protobuf:"varint,2,req,name=UserID" json:"UserID"`
}

func (m *TeamAddMember) Reset()         { *m = TeamAddMember{} }
func (m *TeamAddMember) String() string { return proto.CompactTextString(m) }
func (*TeamAddMember) ProtoMessage()    {}
func (*TeamAddMember) Descriptor() ([]byte, []int) {
	return fileDescriptor_8b4e9e93d7b2c6bb, []int{2}
}
func (m *TeamAddMember) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TeamAddMember) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TeamAddMember.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TeamAddMember) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TeamAddMember.Merge(m, src)
}
func (m *TeamAddMember) XXX_Size() int {
	return m.Size()
}
func (m *TeamAddMember) XXX_DiscardUnknown() {
	xxx_messageInfo_TeamAddMember.DiscardUnknown(m)
}

var xxx_messageInfo_TeamAddMember proto.InternalMessageInfo

func (m *TeamAddMember) GetTeamID() int64 {
	if m != nil {
		return m.TeamID
	}
	return 0
}

func (m *TeamAddMember) GetUserID() int64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

// TeamRemoveMember
// @Function
// @Return: Bool
type TeamRemoveMember struct {
	TeamID int64 `protobuf:"varint,1,req,name=TeamID" json:"TeamID"`
	UserID int64 `protobuf:"varint,2,req,name=UserID" json:"UserID"`
}

func (m *TeamRemoveMember) Reset()         { *m = TeamRemoveMember{} }
func (m *TeamRemoveMember) String() string { return proto.CompactTextString(m) }
func (*TeamRemoveMember) ProtoMessage()    {}
func (*TeamRemoveMember) Descriptor() ([]byte, []int) {
	return fileDescriptor_8b4e9e93d7b2c6bb, []int{3}
}
func (m *TeamRemoveMember) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TeamRemoveMember) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TeamRemoveMember.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TeamRemoveMember) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TeamRemoveMember.Merge(m, src)
}
func (m *TeamRemoveMember) XXX_Size() int {
	return m.Size()
}
func (m *TeamRemoveMember) XXX_DiscardUnknown() {
	xxx_messageInfo_TeamRemoveMember.DiscardUnknown(m)
}

var xxx_messageInfo_TeamRemoveMember proto.InternalMessageInfo

func (m *TeamRemoveMember) GetTeamID() int64 {
	if m != nil {
		return m.TeamID
	}
	return 0
}

func (m *TeamRemoveMember) GetUserID() int64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

// TeamLeave
// @Function
// @Return: Bool
type TeamLeave struct {
	TeamID int64 `protobuf:"varint,2,req,name=TeamID" json:"TeamID"`
}

func (m *TeamLeave) Reset()         { *m = TeamLeave{} }
func (m *TeamLeave) String() string { return proto.CompactTextString(m) }
func (*TeamLeave) ProtoMessage()    {}
func (*TeamLeave) Descriptor() ([]byte, []int) {
	return fileDescriptor_8b4e9e93d7b2c6bb, []int{4}
}
func (m *TeamLeave) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TeamLeave) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TeamLeave.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TeamLeave) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TeamLeave.Merge(m, src)
}
func (m *TeamLeave) XXX_Size() int {
	return m.Size()
}
func (m *TeamLeave) XXX_DiscardUnknown() {
	xxx_messageInfo_TeamLeave.DiscardUnknown(m)
}

var xxx_messageInfo_TeamLeave proto.InternalMessageInfo

func (m *TeamLeave) GetTeamID() int64 {
	if m != nil {
		return m.TeamID
	}
	return 0
}

// TeamJoin
// @Function
// @Return: Bool
type TeamJoin struct {
	TeamID int64  `protobuf:"varint,1,req,name=TeamID" json:"TeamID"`
	Token  []byte `protobuf:"bytes,2,req,name=Token" json:"Token"`
}

func (m *TeamJoin) Reset()         { *m = TeamJoin{} }
func (m *TeamJoin) String() string { return proto.CompactTextString(m) }
func (*TeamJoin) ProtoMessage()    {}
func (*TeamJoin) Descriptor() ([]byte, []int) {
	return fileDescriptor_8b4e9e93d7b2c6bb, []int{5}
}
func (m *TeamJoin) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TeamJoin) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TeamJoin.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TeamJoin) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TeamJoin.Merge(m, src)
}
func (m *TeamJoin) XXX_Size() int {
	return m.Size()
}
func (m *TeamJoin) XXX_DiscardUnknown() {
	xxx_messageInfo_TeamJoin.DiscardUnknown(m)
}

var xxx_messageInfo_TeamJoin proto.InternalMessageInfo

func (m *TeamJoin) GetTeamID() int64 {
	if m != nil {
		return m.TeamID
	}
	return 0
}

func (m *TeamJoin) GetToken() []byte {
	if m != nil {
		return m.Token
	}
	return nil
}

// TeamListMembers
// @Function
// @Return: UsersMany
type TeamListMembers struct {
	TeamID int64 `protobuf:"varint,1,req,name=TeamID" json:"TeamID"`
}

func (m *TeamListMembers) Reset()         { *m = TeamListMembers{} }
func (m *TeamListMembers) String() string { return proto.CompactTextString(m) }
func (*TeamListMembers) ProtoMessage()    {}
func (*TeamListMembers) Descriptor() ([]byte, []int) {
	return fileDescriptor_8b4e9e93d7b2c6bb, []int{6}
}
func (m *TeamListMembers) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TeamListMembers) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TeamListMembers.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TeamListMembers) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TeamListMembers.Merge(m, src)
}
func (m *TeamListMembers) XXX_Size() int {
	return m.Size()
}
func (m *TeamListMembers) XXX_DiscardUnknown() {
	xxx_messageInfo_TeamListMembers.DiscardUnknown(m)
}

var xxx_messageInfo_TeamListMembers proto.InternalMessageInfo

func (m *TeamListMembers) GetTeamID() int64 {
	if m != nil {
		return m.TeamID
	}
	return 0
}

func init() {
	proto.RegisterType((*TeamGet)(nil), "msg.TeamGet")
	proto.RegisterType((*TeamCreate)(nil), "msg.TeamCreate")
	proto.RegisterType((*TeamAddMember)(nil), "msg.TeamAddMember")
	proto.RegisterType((*TeamRemoveMember)(nil), "msg.TeamRemoveMember")
	proto.RegisterType((*TeamLeave)(nil), "msg.TeamLeave")
	proto.RegisterType((*TeamJoin)(nil), "msg.TeamJoin")
	proto.RegisterType((*TeamListMembers)(nil), "msg.TeamListMembers")
}

func init() { proto.RegisterFile("team.proto", fileDescriptor_8b4e9e93d7b2c6bb) }

var fileDescriptor_8b4e9e93d7b2c6bb = []byte{
	// 309 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x8e, 0xcf, 0x4e, 0xc2, 0x40,
	0x10, 0xc6, 0xdb, 0x05, 0x14, 0x26, 0x1a, 0xcd, 0xc6, 0x43, 0x43, 0xc8, 0x42, 0x36, 0x1e, 0xf0,
	0x20, 0x3c, 0x83, 0x50, 0x63, 0xf0, 0x0f, 0x07, 0x82, 0x0f, 0xb0, 0xc0, 0x58, 0x1b, 0xb3, 0x2c,
	0x69, 0x17, 0xa2, 0x6f, 0xe1, 0x63, 0x71, 0xec, 0xd1, 0x93, 0xd1, 0xf6, 0x45, 0xcc, 0x6e, 0x1b,
	0xd2, 0x68, 0xa2, 0x17, 0x6f, 0x33, 0xbf, 0xef, 0x9b, 0xef, 0x1b, 0x00, 0x8d, 0x42, 0xf6, 0x56,
	0x91, 0xd2, 0x8a, 0x56, 0x64, 0x1c, 0x34, 0xcf, 0x83, 0x50, 0x3f, 0xae, 0x67, 0xbd, 0xb9, 0x92,
	0xfd, 0x40, 0x05, 0xaa, 0x6f, 0xb5, 0xd9, 0xfa, 0xc1, 0x6e, 0x76, 0xb1, 0x53, 0x7e, 0xc3, 0xdb,
	0xb0, 0x3f, 0x45, 0x21, 0xaf, 0x50, 0xd3, 0x13, 0x20, 0x23, 0xdf, 0x73, 0x3b, 0xa4, 0x5b, 0x19,
	0x54, 0xb7, 0xef, 0x6d, 0x67, 0x42, 0x46, 0x3e, 0x5f, 0x02, 0x18, 0xc3, 0x30, 0x42, 0xa1, 0x91,
	0x7a, 0x50, 0x1d, 0x0b, 0x89, 0xd6, 0xd5, 0x28, 0x5c, 0x96, 0xd0, 0x0e, 0xd4, 0x87, 0x62, 0x25,
	0xe6, 0xa1, 0x7e, 0xf1, 0x48, 0x87, 0x74, 0x6b, 0x85, 0xba, 0xa3, 0xf4, 0x14, 0xe0, 0xf2, 0x79,
	0x15, 0x46, 0xe8, 0x0b, 0x8d, 0x5e, 0xa5, 0xd4, 0x53, 0xe2, 0xfc, 0x06, 0x0e, 0x4d, 0xdf, 0xc5,
	0x62, 0x71, 0x87, 0x72, 0x86, 0x11, 0x6d, 0xc1, 0x9e, 0x01, 0xdf, 0x5e, 0x2b, 0x98, 0x51, 0xef,
	0x63, 0x8c, 0x46, 0xbe, 0x2d, 0xdd, 0xa9, 0x39, 0xe3, 0x63, 0x38, 0x36, 0xbe, 0x09, 0x4a, 0xb5,
	0xc1, 0x7f, 0xc8, 0x3b, 0x83, 0x86, 0xf1, 0xdd, 0xa2, 0xd8, 0x60, 0x29, 0x88, 0xfc, 0x0c, 0xe2,
	0x3e, 0xd4, 0xcd, 0x74, 0xad, 0xc2, 0xe5, 0x1f, 0x95, 0x4d, 0xa8, 0x4d, 0xd5, 0x13, 0x2e, 0x6d,
	0xcc, 0x41, 0x21, 0xe6, 0x88, 0xf7, 0xe1, 0xc8, 0x16, 0x86, 0xb1, 0xce, 0xdf, 0x8f, 0x7f, 0x0f,
	0x1b, 0xb4, 0x92, 0x4f, 0xe6, 0x6c, 0x53, 0xe6, 0x26, 0x29, 0x73, 0x3f, 0x52, 0xe6, 0xbe, 0x66,
	0xcc, 0x49, 0x32, 0xe6, 0xbc, 0x65, 0xcc, 0xf9, 0x0a, 0x00, 0x00, 0xff, 0xff, 0x0f, 0x83, 0x84,
	0x3c, 0x2e, 0x02, 0x00, 0x00,
}

func (m *TeamGet) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TeamGet) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TeamGet) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	i = encodeVarintTeam(dAtA, i, uint64(m.ID))
	i--
	dAtA[i] = 0x8
	return len(dAtA) - i, nil
}

func (m *TeamCreate) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TeamCreate) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TeamCreate) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	i = encodeVarintTeam(dAtA, i, uint64(m.ExpireDate))
	i--
	dAtA[i] = 0x18
	i = encodeVarintTeam(dAtA, i, uint64(m.Capacity))
	i--
	dAtA[i] = 0x10
	i -= len(m.Name)
	copy(dAtA[i:], m.Name)
	i = encodeVarintTeam(dAtA, i, uint64(len(m.Name)))
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *TeamAddMember) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TeamAddMember) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TeamAddMember) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	i = encodeVarintTeam(dAtA, i, uint64(m.UserID))
	i--
	dAtA[i] = 0x10
	i = encodeVarintTeam(dAtA, i, uint64(m.TeamID))
	i--
	dAtA[i] = 0x8
	return len(dAtA) - i, nil
}

func (m *TeamRemoveMember) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TeamRemoveMember) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TeamRemoveMember) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	i = encodeVarintTeam(dAtA, i, uint64(m.UserID))
	i--
	dAtA[i] = 0x10
	i = encodeVarintTeam(dAtA, i, uint64(m.TeamID))
	i--
	dAtA[i] = 0x8
	return len(dAtA) - i, nil
}

func (m *TeamLeave) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TeamLeave) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TeamLeave) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	i = encodeVarintTeam(dAtA, i, uint64(m.TeamID))
	i--
	dAtA[i] = 0x10
	return len(dAtA) - i, nil
}

func (m *TeamJoin) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TeamJoin) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TeamJoin) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Token != nil {
		i -= len(m.Token)
		copy(dAtA[i:], m.Token)
		i = encodeVarintTeam(dAtA, i, uint64(len(m.Token)))
		i--
		dAtA[i] = 0x12
	}
	i = encodeVarintTeam(dAtA, i, uint64(m.TeamID))
	i--
	dAtA[i] = 0x8
	return len(dAtA) - i, nil
}

func (m *TeamListMembers) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TeamListMembers) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TeamListMembers) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	i = encodeVarintTeam(dAtA, i, uint64(m.TeamID))
	i--
	dAtA[i] = 0x8
	return len(dAtA) - i, nil
}

func encodeVarintTeam(dAtA []byte, offset int, v uint64) int {
	offset -= sovTeam(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *TeamGet) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovTeam(uint64(m.ID))
	return n
}

func (m *TeamCreate) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Name)
	n += 1 + l + sovTeam(uint64(l))
	n += 1 + sovTeam(uint64(m.Capacity))
	n += 1 + sovTeam(uint64(m.ExpireDate))
	return n
}

func (m *TeamAddMember) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovTeam(uint64(m.TeamID))
	n += 1 + sovTeam(uint64(m.UserID))
	return n
}

func (m *TeamRemoveMember) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovTeam(uint64(m.TeamID))
	n += 1 + sovTeam(uint64(m.UserID))
	return n
}

func (m *TeamLeave) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovTeam(uint64(m.TeamID))
	return n
}

func (m *TeamJoin) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovTeam(uint64(m.TeamID))
	if m.Token != nil {
		l = len(m.Token)
		n += 1 + l + sovTeam(uint64(l))
	}
	return n
}

func (m *TeamListMembers) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovTeam(uint64(m.TeamID))
	return n
}

func sovTeam(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTeam(x uint64) (n int) {
	return sovTeam(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *TeamGet) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTeam
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TeamGet: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TeamGet: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			m.ID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ID |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		default:
			iNdEx = preIndex
			skippy, err := skipTeam(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("ID")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TeamCreate) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTeam
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TeamCreate: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TeamCreate: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTeam
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTeam
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Capacity", wireType)
			}
			m.Capacity = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Capacity |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000002)
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExpireDate", wireType)
			}
			m.ExpireDate = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExpireDate |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000004)
		default:
			iNdEx = preIndex
			skippy, err := skipTeam(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Name")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Capacity")
	}
	if hasFields[0]&uint64(0x00000004) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("ExpireDate")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TeamAddMember) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTeam
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TeamAddMember: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TeamAddMember: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TeamID", wireType)
			}
			m.TeamID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TeamID |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UserID", wireType)
			}
			m.UserID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UserID |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000002)
		default:
			iNdEx = preIndex
			skippy, err := skipTeam(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("TeamID")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("UserID")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TeamRemoveMember) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTeam
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TeamRemoveMember: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TeamRemoveMember: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TeamID", wireType)
			}
			m.TeamID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TeamID |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UserID", wireType)
			}
			m.UserID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UserID |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000002)
		default:
			iNdEx = preIndex
			skippy, err := skipTeam(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("TeamID")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("UserID")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TeamLeave) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTeam
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TeamLeave: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TeamLeave: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TeamID", wireType)
			}
			m.TeamID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TeamID |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		default:
			iNdEx = preIndex
			skippy, err := skipTeam(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("TeamID")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TeamJoin) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTeam
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TeamJoin: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TeamJoin: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TeamID", wireType)
			}
			m.TeamID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TeamID |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Token", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTeam
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTeam
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Token = append(m.Token[:0], dAtA[iNdEx:postIndex]...)
			if m.Token == nil {
				m.Token = []byte{}
			}
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000002)
		default:
			iNdEx = preIndex
			skippy, err := skipTeam(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("TeamID")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Token")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TeamListMembers) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTeam
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TeamListMembers: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TeamListMembers: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TeamID", wireType)
			}
			m.TeamID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TeamID |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		default:
			iNdEx = preIndex
			skippy, err := skipTeam(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTeam
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("TeamID")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTeam(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTeam
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
					return 0, ErrIntOverflowTeam
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTeam
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
			if length < 0 {
				return 0, ErrInvalidLengthTeam
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTeam
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTeam
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTeam        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTeam          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTeam = fmt.Errorf("proto: unexpected end of group")
)
