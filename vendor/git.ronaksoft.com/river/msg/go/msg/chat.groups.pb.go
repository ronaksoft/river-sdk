// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: chat.groups.proto

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

// GroupsCreate
// @Function
// @Return: Bool
type GroupsCreate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Users    []*InputUser `protobuf:"bytes,1,rep,name=Users,proto3" json:"Users,omitempty"`
	Title    string       `protobuf:"bytes,2,opt,name=Title,proto3" json:"Title,omitempty"`
	RandomID int64        `protobuf:"varint,3,opt,name=RandomID,proto3" json:"RandomID,omitempty"`
}

func (x *GroupsCreate) Reset() {
	*x = GroupsCreate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsCreate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsCreate) ProtoMessage() {}

func (x *GroupsCreate) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsCreate.ProtoReflect.Descriptor instead.
func (*GroupsCreate) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{0}
}

func (x *GroupsCreate) GetUsers() []*InputUser {
	if x != nil {
		return x.Users
	}
	return nil
}

func (x *GroupsCreate) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *GroupsCreate) GetRandomID() int64 {
	if x != nil {
		return x.RandomID
	}
	return 0
}

// GroupsAddUser
// @Function
// @Return: Bool
type GroupsAddUser struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupID      int64      `protobuf:"varint,1,opt,name=GroupID,proto3" json:"GroupID,omitempty"`
	User         *InputUser `protobuf:"bytes,2,opt,name=User,proto3" json:"User,omitempty"`
	ForwardLimit int32      `protobuf:"varint,3,opt,name=ForwardLimit,proto3" json:"ForwardLimit,omitempty"`
}

func (x *GroupsAddUser) Reset() {
	*x = GroupsAddUser{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsAddUser) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsAddUser) ProtoMessage() {}

func (x *GroupsAddUser) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsAddUser.ProtoReflect.Descriptor instead.
func (*GroupsAddUser) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{1}
}

func (x *GroupsAddUser) GetGroupID() int64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

func (x *GroupsAddUser) GetUser() *InputUser {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *GroupsAddUser) GetForwardLimit() int32 {
	if x != nil {
		return x.ForwardLimit
	}
	return 0
}

// GroupsEditTitle
// @Function
// @Return: Bool
type GroupsEditTitle struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupID int64  `protobuf:"varint,1,opt,name=GroupID,proto3" json:"GroupID,omitempty"`
	Title   string `protobuf:"bytes,2,opt,name=Title,proto3" json:"Title,omitempty"`
}

func (x *GroupsEditTitle) Reset() {
	*x = GroupsEditTitle{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsEditTitle) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsEditTitle) ProtoMessage() {}

func (x *GroupsEditTitle) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsEditTitle.ProtoReflect.Descriptor instead.
func (*GroupsEditTitle) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{2}
}

func (x *GroupsEditTitle) GetGroupID() int64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

func (x *GroupsEditTitle) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

// GroupsDeleteUser
// @Function
// @Return: Bool
type GroupsDeleteUser struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupID int64      `protobuf:"varint,1,opt,name=GroupID,proto3" json:"GroupID,omitempty"`
	User    *InputUser `protobuf:"bytes,2,opt,name=User,proto3" json:"User,omitempty"`
}

func (x *GroupsDeleteUser) Reset() {
	*x = GroupsDeleteUser{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsDeleteUser) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsDeleteUser) ProtoMessage() {}

func (x *GroupsDeleteUser) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsDeleteUser.ProtoReflect.Descriptor instead.
func (*GroupsDeleteUser) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{3}
}

func (x *GroupsDeleteUser) GetGroupID() int64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

func (x *GroupsDeleteUser) GetUser() *InputUser {
	if x != nil {
		return x.User
	}
	return nil
}

// GroupsGetFull
// @Function
// @Return: GroupFull
type GroupsGetFull struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupID int64 `protobuf:"varint,1,opt,name=GroupID,proto3" json:"GroupID,omitempty"`
}

func (x *GroupsGetFull) Reset() {
	*x = GroupsGetFull{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsGetFull) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsGetFull) ProtoMessage() {}

func (x *GroupsGetFull) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsGetFull.ProtoReflect.Descriptor instead.
func (*GroupsGetFull) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{4}
}

func (x *GroupsGetFull) GetGroupID() int64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

// GroupsToggleAdmins
// @Function
// @Return: Bool
type GroupsToggleAdmins struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupID      int64 `protobuf:"varint,1,opt,name=GroupID,proto3" json:"GroupID,omitempty"`
	AdminEnabled bool  `protobuf:"varint,2,opt,name=AdminEnabled,proto3" json:"AdminEnabled,omitempty"`
}

func (x *GroupsToggleAdmins) Reset() {
	*x = GroupsToggleAdmins{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsToggleAdmins) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsToggleAdmins) ProtoMessage() {}

func (x *GroupsToggleAdmins) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsToggleAdmins.ProtoReflect.Descriptor instead.
func (*GroupsToggleAdmins) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{5}
}

func (x *GroupsToggleAdmins) GetGroupID() int64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

func (x *GroupsToggleAdmins) GetAdminEnabled() bool {
	if x != nil {
		return x.AdminEnabled
	}
	return false
}

// GroupsUpdateAdmin
// @Function
// @Return: Bool
type GroupsUpdateAdmin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupID int64      `protobuf:"varint,1,opt,name=GroupID,proto3" json:"GroupID,omitempty"`
	User    *InputUser `protobuf:"bytes,2,opt,name=User,proto3" json:"User,omitempty"`
	Admin   bool       `protobuf:"varint,3,opt,name=Admin,proto3" json:"Admin,omitempty"`
}

func (x *GroupsUpdateAdmin) Reset() {
	*x = GroupsUpdateAdmin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsUpdateAdmin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsUpdateAdmin) ProtoMessage() {}

func (x *GroupsUpdateAdmin) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsUpdateAdmin.ProtoReflect.Descriptor instead.
func (*GroupsUpdateAdmin) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{6}
}

func (x *GroupsUpdateAdmin) GetGroupID() int64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

func (x *GroupsUpdateAdmin) GetUser() *InputUser {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *GroupsUpdateAdmin) GetAdmin() bool {
	if x != nil {
		return x.Admin
	}
	return false
}

// GroupsUploadPhoto
// @Function
// @Return: Bool / GroupPhoto
type GroupsUploadPhoto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupID      int64      `protobuf:"varint,1,opt,name=GroupID,proto3" json:"GroupID,omitempty"`
	File         *InputFile `protobuf:"bytes,2,opt,name=File,proto3" json:"File,omitempty"`
	ReturnObject bool       `protobuf:"varint,3,opt,name=ReturnObject,proto3" json:"ReturnObject,omitempty"`
}

func (x *GroupsUploadPhoto) Reset() {
	*x = GroupsUploadPhoto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsUploadPhoto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsUploadPhoto) ProtoMessage() {}

func (x *GroupsUploadPhoto) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsUploadPhoto.ProtoReflect.Descriptor instead.
func (*GroupsUploadPhoto) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{7}
}

func (x *GroupsUploadPhoto) GetGroupID() int64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

func (x *GroupsUploadPhoto) GetFile() *InputFile {
	if x != nil {
		return x.File
	}
	return nil
}

func (x *GroupsUploadPhoto) GetReturnObject() bool {
	if x != nil {
		return x.ReturnObject
	}
	return false
}

// GroupsRemovePhoto
// @Function
// @Return: Bool
type GroupsRemovePhoto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupID int64 `protobuf:"varint,1,opt,name=GroupID,proto3" json:"GroupID,omitempty"`
	PhotoID int64 `protobuf:"varint,2,opt,name=PhotoID,proto3" json:"PhotoID,omitempty"`
}

func (x *GroupsRemovePhoto) Reset() {
	*x = GroupsRemovePhoto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsRemovePhoto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsRemovePhoto) ProtoMessage() {}

func (x *GroupsRemovePhoto) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsRemovePhoto.ProtoReflect.Descriptor instead.
func (*GroupsRemovePhoto) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{8}
}

func (x *GroupsRemovePhoto) GetGroupID() int64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

func (x *GroupsRemovePhoto) GetPhotoID() int64 {
	if x != nil {
		return x.PhotoID
	}
	return 0
}

// GroupsUpdatePhoto
// @Function
// @Return: Bool
type GroupsUpdatePhoto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PhotoID int64 `protobuf:"varint,1,opt,name=PhotoID,proto3" json:"PhotoID,omitempty"`
	GroupID int64 `protobuf:"varint,2,opt,name=GroupID,proto3" json:"GroupID,omitempty"`
}

func (x *GroupsUpdatePhoto) Reset() {
	*x = GroupsUpdatePhoto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsUpdatePhoto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsUpdatePhoto) ProtoMessage() {}

func (x *GroupsUpdatePhoto) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsUpdatePhoto.ProtoReflect.Descriptor instead.
func (*GroupsUpdatePhoto) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{9}
}

func (x *GroupsUpdatePhoto) GetPhotoID() int64 {
	if x != nil {
		return x.PhotoID
	}
	return 0
}

func (x *GroupsUpdatePhoto) GetGroupID() int64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

// GroupsGetHistoryStats
// @Function
// @Return: GroupsHistoryStats
type GroupsGetReadHistoryStats struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupID int64 `protobuf:"varint,1,opt,name=GroupID,proto3" json:"GroupID,omitempty"`
}

func (x *GroupsGetReadHistoryStats) Reset() {
	*x = GroupsGetReadHistoryStats{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsGetReadHistoryStats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsGetReadHistoryStats) ProtoMessage() {}

func (x *GroupsGetReadHistoryStats) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsGetReadHistoryStats.ProtoReflect.Descriptor instead.
func (*GroupsGetReadHistoryStats) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{10}
}

func (x *GroupsGetReadHistoryStats) GetGroupID() int64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

// GroupsHistoryStats
type GroupsHistoryStats struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Stats []*ReadHistoryStat `protobuf:"bytes,1,rep,name=Stats,proto3" json:"Stats,omitempty"`
	Users []*User            `protobuf:"bytes,2,rep,name=Users,proto3" json:"Users,omitempty"`
	Empty bool               `protobuf:"varint,3,opt,name=Empty,proto3" json:"Empty,omitempty"`
}

func (x *GroupsHistoryStats) Reset() {
	*x = GroupsHistoryStats{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupsHistoryStats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupsHistoryStats) ProtoMessage() {}

func (x *GroupsHistoryStats) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupsHistoryStats.ProtoReflect.Descriptor instead.
func (*GroupsHistoryStats) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{11}
}

func (x *GroupsHistoryStats) GetStats() []*ReadHistoryStat {
	if x != nil {
		return x.Stats
	}
	return nil
}

func (x *GroupsHistoryStats) GetUsers() []*User {
	if x != nil {
		return x.Users
	}
	return nil
}

func (x *GroupsHistoryStats) GetEmpty() bool {
	if x != nil {
		return x.Empty
	}
	return false
}

// ReadHistoryStat
type ReadHistoryStat struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserID    int64 `protobuf:"varint,1,opt,name=UserID,proto3" json:"UserID,omitempty"`
	MessageID int64 `protobuf:"varint,2,opt,name=MessageID,proto3" json:"MessageID,omitempty"`
}

func (x *ReadHistoryStat) Reset() {
	*x = ReadHistoryStat{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chat_groups_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadHistoryStat) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadHistoryStat) ProtoMessage() {}

func (x *ReadHistoryStat) ProtoReflect() protoreflect.Message {
	mi := &file_chat_groups_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadHistoryStat.ProtoReflect.Descriptor instead.
func (*ReadHistoryStat) Descriptor() ([]byte, []int) {
	return file_chat_groups_proto_rawDescGZIP(), []int{12}
}

func (x *ReadHistoryStat) GetUserID() int64 {
	if x != nil {
		return x.UserID
	}
	return 0
}

func (x *ReadHistoryStat) GetMessageID() int64 {
	if x != nil {
		return x.MessageID
	}
	return 0
}

var File_chat_groups_proto protoreflect.FileDescriptor

var file_chat_groups_proto_rawDesc = []byte{
	0x0a, 0x11, 0x63, 0x68, 0x61, 0x74, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x03, 0x6d, 0x73, 0x67, 0x1a, 0x10, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x66, 0x0a, 0x0c, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x73, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x24, 0x0a, 0x05, 0x55, 0x73,
	0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x6d, 0x73, 0x67, 0x2e,
	0x49, 0x6e, 0x70, 0x75, 0x74, 0x55, 0x73, 0x65, 0x72, 0x52, 0x05, 0x55, 0x73, 0x65, 0x72, 0x73,
	0x12, 0x14, 0x0a, 0x05, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d,
	0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d,
	0x49, 0x44, 0x22, 0x75, 0x0a, 0x0d, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x41, 0x64, 0x64, 0x55,
	0x73, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x30, 0x01, 0x52, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49,
	0x44, 0x12, 0x22, 0x0a, 0x04, 0x55, 0x73, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0e, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x55, 0x73, 0x65, 0x72, 0x52,
	0x04, 0x55, 0x73, 0x65, 0x72, 0x12, 0x22, 0x0a, 0x0c, 0x46, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64,
	0x4c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x46, 0x6f, 0x72,
	0x77, 0x61, 0x72, 0x64, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x22, 0x45, 0x0a, 0x0f, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x73, 0x45, 0x64, 0x69, 0x74, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x1c, 0x0a, 0x07,
	0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x30,
	0x01, 0x52, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x12, 0x14, 0x0a, 0x05, 0x54, 0x69,
	0x74, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x54, 0x69, 0x74, 0x6c, 0x65,
	0x22, 0x54, 0x0a, 0x10, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x55, 0x73, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x30, 0x01, 0x52, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x49, 0x44, 0x12, 0x22, 0x0a, 0x04, 0x55, 0x73, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0e, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x55, 0x73, 0x65, 0x72,
	0x52, 0x04, 0x55, 0x73, 0x65, 0x72, 0x22, 0x2d, 0x0a, 0x0d, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73,
	0x47, 0x65, 0x74, 0x46, 0x75, 0x6c, 0x6c, 0x12, 0x1c, 0x0a, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x30, 0x01, 0x52, 0x07, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x49, 0x44, 0x22, 0x56, 0x0a, 0x12, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x54,
	0x6f, 0x67, 0x67, 0x6c, 0x65, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x73, 0x12, 0x1c, 0x0a, 0x07, 0x47,
	0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x30, 0x01,
	0x52, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x12, 0x22, 0x0a, 0x0c, 0x41, 0x64, 0x6d,
	0x69, 0x6e, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x0c, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x22, 0x6b, 0x0a,
	0x11, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x41, 0x64, 0x6d,
	0x69, 0x6e, 0x12, 0x1c, 0x0a, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x42, 0x02, 0x30, 0x01, 0x52, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44,
	0x12, 0x22, 0x0a, 0x04, 0x55, 0x73, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x55, 0x73, 0x65, 0x72, 0x52, 0x04,
	0x55, 0x73, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x05, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x22, 0x79, 0x0a, 0x11, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x73, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x50, 0x68, 0x6f, 0x74, 0x6f, 0x12,
	0x1c, 0x0a, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x42, 0x02, 0x30, 0x01, 0x52, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x12, 0x22, 0x0a,
	0x04, 0x46, 0x69, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x6d, 0x73,
	0x67, 0x2e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x04, 0x46, 0x69, 0x6c,
	0x65, 0x12, 0x22, 0x0a, 0x0c, 0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x4f, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x4f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x22, 0x4f, 0x0a, 0x11, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x52,
	0x65, 0x6d, 0x6f, 0x76, 0x65, 0x50, 0x68, 0x6f, 0x74, 0x6f, 0x12, 0x1c, 0x0a, 0x07, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x30, 0x01, 0x52,
	0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x12, 0x1c, 0x0a, 0x07, 0x50, 0x68, 0x6f, 0x74,
	0x6f, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x30, 0x01, 0x52, 0x07, 0x50,
	0x68, 0x6f, 0x74, 0x6f, 0x49, 0x44, 0x22, 0x4f, 0x0a, 0x11, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x68, 0x6f, 0x74, 0x6f, 0x12, 0x1c, 0x0a, 0x07, 0x50,
	0x68, 0x6f, 0x74, 0x6f, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x30, 0x01,
	0x52, 0x07, 0x50, 0x68, 0x6f, 0x74, 0x6f, 0x49, 0x44, 0x12, 0x1c, 0x0a, 0x07, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x30, 0x01, 0x52, 0x07,
	0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x22, 0x39, 0x0a, 0x19, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x73, 0x47, 0x65, 0x74, 0x52, 0x65, 0x61, 0x64, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x53,
	0x74, 0x61, 0x74, 0x73, 0x12, 0x1c, 0x0a, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x30, 0x01, 0x52, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x49, 0x44, 0x22, 0x77, 0x0a, 0x12, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x48, 0x69, 0x73, 0x74,
	0x6f, 0x72, 0x79, 0x53, 0x74, 0x61, 0x74, 0x73, 0x12, 0x2a, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x52, 0x65,
	0x61, 0x64, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x53, 0x74, 0x61, 0x74, 0x52, 0x05, 0x53,
	0x74, 0x61, 0x74, 0x73, 0x12, 0x1f, 0x0a, 0x05, 0x55, 0x73, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x52, 0x05,
	0x55, 0x73, 0x65, 0x72, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x4b, 0x0a, 0x0f, 0x52,
	0x65, 0x61, 0x64, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x53, 0x74, 0x61, 0x74, 0x12, 0x1a,
	0x0a, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x02,
	0x30, 0x01, 0x52, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x12, 0x1c, 0x0a, 0x09, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x42, 0x08, 0x5a, 0x06, 0x2e, 0x2f, 0x3b, 0x6d,
	0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_chat_groups_proto_rawDescOnce sync.Once
	file_chat_groups_proto_rawDescData = file_chat_groups_proto_rawDesc
)

func file_chat_groups_proto_rawDescGZIP() []byte {
	file_chat_groups_proto_rawDescOnce.Do(func() {
		file_chat_groups_proto_rawDescData = protoimpl.X.CompressGZIP(file_chat_groups_proto_rawDescData)
	})
	return file_chat_groups_proto_rawDescData
}

var file_chat_groups_proto_msgTypes = make([]protoimpl.MessageInfo, 13)
var file_chat_groups_proto_goTypes = []interface{}{
	(*GroupsCreate)(nil),              // 0: msg.GroupsCreate
	(*GroupsAddUser)(nil),             // 1: msg.GroupsAddUser
	(*GroupsEditTitle)(nil),           // 2: msg.GroupsEditTitle
	(*GroupsDeleteUser)(nil),          // 3: msg.GroupsDeleteUser
	(*GroupsGetFull)(nil),             // 4: msg.GroupsGetFull
	(*GroupsToggleAdmins)(nil),        // 5: msg.GroupsToggleAdmins
	(*GroupsUpdateAdmin)(nil),         // 6: msg.GroupsUpdateAdmin
	(*GroupsUploadPhoto)(nil),         // 7: msg.GroupsUploadPhoto
	(*GroupsRemovePhoto)(nil),         // 8: msg.GroupsRemovePhoto
	(*GroupsUpdatePhoto)(nil),         // 9: msg.GroupsUpdatePhoto
	(*GroupsGetReadHistoryStats)(nil), // 10: msg.GroupsGetReadHistoryStats
	(*GroupsHistoryStats)(nil),        // 11: msg.GroupsHistoryStats
	(*ReadHistoryStat)(nil),           // 12: msg.ReadHistoryStat
	(*InputUser)(nil),                 // 13: msg.InputUser
	(*InputFile)(nil),                 // 14: msg.InputFile
	(*User)(nil),                      // 15: msg.User
}
var file_chat_groups_proto_depIdxs = []int32{
	13, // 0: msg.GroupsCreate.Users:type_name -> msg.InputUser
	13, // 1: msg.GroupsAddUser.User:type_name -> msg.InputUser
	13, // 2: msg.GroupsDeleteUser.User:type_name -> msg.InputUser
	13, // 3: msg.GroupsUpdateAdmin.User:type_name -> msg.InputUser
	14, // 4: msg.GroupsUploadPhoto.File:type_name -> msg.InputFile
	12, // 5: msg.GroupsHistoryStats.Stats:type_name -> msg.ReadHistoryStat
	15, // 6: msg.GroupsHistoryStats.Users:type_name -> msg.User
	7,  // [7:7] is the sub-list for method output_type
	7,  // [7:7] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_chat_groups_proto_init() }
func file_chat_groups_proto_init() {
	if File_chat_groups_proto != nil {
		return
	}
	file_core_types_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_chat_groups_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsCreate); i {
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
		file_chat_groups_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsAddUser); i {
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
		file_chat_groups_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsEditTitle); i {
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
		file_chat_groups_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsDeleteUser); i {
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
		file_chat_groups_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsGetFull); i {
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
		file_chat_groups_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsToggleAdmins); i {
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
		file_chat_groups_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsUpdateAdmin); i {
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
		file_chat_groups_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsUploadPhoto); i {
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
		file_chat_groups_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsRemovePhoto); i {
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
		file_chat_groups_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsUpdatePhoto); i {
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
		file_chat_groups_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsGetReadHistoryStats); i {
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
		file_chat_groups_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupsHistoryStats); i {
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
		file_chat_groups_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadHistoryStat); i {
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
			RawDescriptor: file_chat_groups_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   13,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_chat_groups_proto_goTypes,
		DependencyIndexes: file_chat_groups_proto_depIdxs,
		MessageInfos:      file_chat_groups_proto_msgTypes,
	}.Build()
	File_chat_groups_proto = out.File
	file_chat_groups_proto_rawDesc = nil
	file_chat_groups_proto_goTypes = nil
	file_chat_groups_proto_depIdxs = nil
}
