// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: calendar.proto

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

// RecurringPeriod
type RecurringPeriod int32

const (
	RecurringPeriod_RecurringNone    RecurringPeriod = 0
	RecurringPeriod_RecurringDaily   RecurringPeriod = 1
	RecurringPeriod_RecurringWeekly  RecurringPeriod = 2
	RecurringPeriod_RecurringMonthly RecurringPeriod = 3
	RecurringPeriod_RecurringYearly  RecurringPeriod = 4
)

// Enum value maps for RecurringPeriod.
var (
	RecurringPeriod_name = map[int32]string{
		0: "RecurringNone",
		1: "RecurringDaily",
		2: "RecurringWeekly",
		3: "RecurringMonthly",
		4: "RecurringYearly",
	}
	RecurringPeriod_value = map[string]int32{
		"RecurringNone":    0,
		"RecurringDaily":   1,
		"RecurringWeekly":  2,
		"RecurringMonthly": 3,
		"RecurringYearly":  4,
	}
)

func (x RecurringPeriod) Enum() *RecurringPeriod {
	p := new(RecurringPeriod)
	*p = x
	return p
}

func (x RecurringPeriod) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RecurringPeriod) Descriptor() protoreflect.EnumDescriptor {
	return file_calendar_proto_enumTypes[0].Descriptor()
}

func (RecurringPeriod) Type() protoreflect.EnumType {
	return &file_calendar_proto_enumTypes[0]
}

func (x RecurringPeriod) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RecurringPeriod.Descriptor instead.
func (RecurringPeriod) EnumDescriptor() ([]byte, []int) {
	return file_calendar_proto_rawDescGZIP(), []int{0}
}

// CalendarEditPolicy
type CalendarEditPolicy int32

const (
	CalendarEditPolicy_CalendarEditOne       CalendarEditPolicy = 0
	CalendarEditPolicy_CalendarEditFollowing CalendarEditPolicy = 1
	CalendarEditPolicy_CalendarEditAll       CalendarEditPolicy = 2
)

// Enum value maps for CalendarEditPolicy.
var (
	CalendarEditPolicy_name = map[int32]string{
		0: "CalendarEditOne",
		1: "CalendarEditFollowing",
		2: "CalendarEditAll",
	}
	CalendarEditPolicy_value = map[string]int32{
		"CalendarEditOne":       0,
		"CalendarEditFollowing": 1,
		"CalendarEditAll":       2,
	}
)

func (x CalendarEditPolicy) Enum() *CalendarEditPolicy {
	p := new(CalendarEditPolicy)
	*p = x
	return p
}

func (x CalendarEditPolicy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CalendarEditPolicy) Descriptor() protoreflect.EnumDescriptor {
	return file_calendar_proto_enumTypes[1].Descriptor()
}

func (CalendarEditPolicy) Type() protoreflect.EnumType {
	return &file_calendar_proto_enumTypes[1]
}

func (x CalendarEditPolicy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CalendarEditPolicy.Descriptor instead.
func (CalendarEditPolicy) EnumDescriptor() ([]byte, []int) {
	return file_calendar_proto_rawDescGZIP(), []int{1}
}

// CalendarGetEvents
// @Function
// @Return: CalendarEventInstances
type CalendarGetEvents struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	From   int64 `protobuf:"varint,1,opt,name=From,proto3" json:"From,omitempty"`
	To     int64 `protobuf:"varint,2,opt,name=To,proto3" json:"To,omitempty"`
	Filter int32 `protobuf:"varint,3,opt,name=Filter,proto3" json:"Filter,omitempty"`
}

func (x *CalendarGetEvents) Reset() {
	*x = CalendarGetEvents{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CalendarGetEvents) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CalendarGetEvents) ProtoMessage() {}

func (x *CalendarGetEvents) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CalendarGetEvents.ProtoReflect.Descriptor instead.
func (*CalendarGetEvents) Descriptor() ([]byte, []int) {
	return file_calendar_proto_rawDescGZIP(), []int{0}
}

func (x *CalendarGetEvents) GetFrom() int64 {
	if x != nil {
		return x.From
	}
	return 0
}

func (x *CalendarGetEvents) GetTo() int64 {
	if x != nil {
		return x.To
	}
	return 0
}

func (x *CalendarGetEvents) GetFilter() int32 {
	if x != nil {
		return x.Filter
	}
	return 0
}

// CalendarSetEvent
// @Function
// @Return: CalendarEventDescriptor
type CalendarSetEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name       string          `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Date       int64           `protobuf:"varint,2,opt,name=Date,proto3" json:"Date,omitempty"`
	StartRange int64           `protobuf:"varint,3,opt,name=StartRange,proto3" json:"StartRange,omitempty"`
	Duration   int64           `protobuf:"varint,4,opt,name=Duration,proto3" json:"Duration,omitempty"`
	Recurring  bool            `protobuf:"varint,5,opt,name=Recurring,proto3" json:"Recurring,omitempty"`
	Period     RecurringPeriod `protobuf:"varint,6,opt,name=Period,proto3,enum=msg.RecurringPeriod" json:"Period,omitempty"`
	AllDay     bool            `protobuf:"varint,7,opt,name=AllDay,proto3" json:"AllDay,omitempty"`
	Team       bool            `protobuf:"varint,8,opt,name=Team,proto3" json:"Team,omitempty"`
	Global     bool            `protobuf:"varint,9,opt,name=Global,proto3" json:"Global,omitempty"`
}

func (x *CalendarSetEvent) Reset() {
	*x = CalendarSetEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CalendarSetEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CalendarSetEvent) ProtoMessage() {}

func (x *CalendarSetEvent) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CalendarSetEvent.ProtoReflect.Descriptor instead.
func (*CalendarSetEvent) Descriptor() ([]byte, []int) {
	return file_calendar_proto_rawDescGZIP(), []int{1}
}

func (x *CalendarSetEvent) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CalendarSetEvent) GetDate() int64 {
	if x != nil {
		return x.Date
	}
	return 0
}

func (x *CalendarSetEvent) GetStartRange() int64 {
	if x != nil {
		return x.StartRange
	}
	return 0
}

func (x *CalendarSetEvent) GetDuration() int64 {
	if x != nil {
		return x.Duration
	}
	return 0
}

func (x *CalendarSetEvent) GetRecurring() bool {
	if x != nil {
		return x.Recurring
	}
	return false
}

func (x *CalendarSetEvent) GetPeriod() RecurringPeriod {
	if x != nil {
		return x.Period
	}
	return RecurringPeriod_RecurringNone
}

func (x *CalendarSetEvent) GetAllDay() bool {
	if x != nil {
		return x.AllDay
	}
	return false
}

func (x *CalendarSetEvent) GetTeam() bool {
	if x != nil {
		return x.Team
	}
	return false
}

func (x *CalendarSetEvent) GetGlobal() bool {
	if x != nil {
		return x.Global
	}
	return false
}

// CalendarEditEvent
// @Function
// @Return: CalendarEvent
type CalendarEditEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventID    int64              `protobuf:"varint,1,opt,name=EventID,proto3" json:"EventID,omitempty"`
	Name       string             `protobuf:"bytes,2,opt,name=Name,proto3" json:"Name,omitempty"`
	Date       int64              `protobuf:"varint,3,opt,name=Date,proto3" json:"Date,omitempty"`
	StartRange int64              `protobuf:"varint,4,opt,name=StartRange,proto3" json:"StartRange,omitempty"`
	Duration   int64              `protobuf:"varint,5,opt,name=Duration,proto3" json:"Duration,omitempty"`
	Recurring  bool               `protobuf:"varint,6,opt,name=Recurring,proto3" json:"Recurring,omitempty"`
	Period     RecurringPeriod    `protobuf:"varint,7,opt,name=Period,proto3,enum=msg.RecurringPeriod" json:"Period,omitempty"`
	AllDay     bool               `protobuf:"varint,8,opt,name=AllDay,proto3" json:"AllDay,omitempty"`
	Policy     CalendarEditPolicy `protobuf:"varint,9,opt,name=Policy,proto3,enum=msg.CalendarEditPolicy" json:"Policy,omitempty"`
}

func (x *CalendarEditEvent) Reset() {
	*x = CalendarEditEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CalendarEditEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CalendarEditEvent) ProtoMessage() {}

func (x *CalendarEditEvent) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CalendarEditEvent.ProtoReflect.Descriptor instead.
func (*CalendarEditEvent) Descriptor() ([]byte, []int) {
	return file_calendar_proto_rawDescGZIP(), []int{2}
}

func (x *CalendarEditEvent) GetEventID() int64 {
	if x != nil {
		return x.EventID
	}
	return 0
}

func (x *CalendarEditEvent) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CalendarEditEvent) GetDate() int64 {
	if x != nil {
		return x.Date
	}
	return 0
}

func (x *CalendarEditEvent) GetStartRange() int64 {
	if x != nil {
		return x.StartRange
	}
	return 0
}

func (x *CalendarEditEvent) GetDuration() int64 {
	if x != nil {
		return x.Duration
	}
	return 0
}

func (x *CalendarEditEvent) GetRecurring() bool {
	if x != nil {
		return x.Recurring
	}
	return false
}

func (x *CalendarEditEvent) GetPeriod() RecurringPeriod {
	if x != nil {
		return x.Period
	}
	return RecurringPeriod_RecurringNone
}

func (x *CalendarEditEvent) GetAllDay() bool {
	if x != nil {
		return x.AllDay
	}
	return false
}

func (x *CalendarEditEvent) GetPolicy() CalendarEditPolicy {
	if x != nil {
		return x.Policy
	}
	return CalendarEditPolicy_CalendarEditOne
}

// CalendarRemoveEvent
// @Function
// @Return: Bool
type CalendarRemoveEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventID int64 `protobuf:"varint,1,opt,name=EventID,proto3" json:"EventID,omitempty"`
}

func (x *CalendarRemoveEvent) Reset() {
	*x = CalendarRemoveEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CalendarRemoveEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CalendarRemoveEvent) ProtoMessage() {}

func (x *CalendarRemoveEvent) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CalendarRemoveEvent.ProtoReflect.Descriptor instead.
func (*CalendarRemoveEvent) Descriptor() ([]byte, []int) {
	return file_calendar_proto_rawDescGZIP(), []int{3}
}

func (x *CalendarRemoveEvent) GetEventID() int64 {
	if x != nil {
		return x.EventID
	}
	return 0
}

// CalendarEvent
type CalendarEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID        int64           `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Name      string          `protobuf:"bytes,2,opt,name=Name,proto3" json:"Name,omitempty"`
	Recurring bool            `protobuf:"varint,3,opt,name=Recurring,proto3" json:"Recurring,omitempty"`
	Period    RecurringPeriod `protobuf:"varint,4,opt,name=Period,proto3,enum=msg.RecurringPeriod" json:"Period,omitempty"`
	AllDay    bool            `protobuf:"varint,5,opt,name=AllDay,proto3" json:"AllDay,omitempty"`
}

func (x *CalendarEvent) Reset() {
	*x = CalendarEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CalendarEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CalendarEvent) ProtoMessage() {}

func (x *CalendarEvent) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CalendarEvent.ProtoReflect.Descriptor instead.
func (*CalendarEvent) Descriptor() ([]byte, []int) {
	return file_calendar_proto_rawDescGZIP(), []int{4}
}

func (x *CalendarEvent) GetID() int64 {
	if x != nil {
		return x.ID
	}
	return 0
}

func (x *CalendarEvent) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CalendarEvent) GetRecurring() bool {
	if x != nil {
		return x.Recurring
	}
	return false
}

func (x *CalendarEvent) GetPeriod() RecurringPeriod {
	if x != nil {
		return x.Period
	}
	return RecurringPeriod_RecurringNone
}

func (x *CalendarEvent) GetAllDay() bool {
	if x != nil {
		return x.AllDay
	}
	return false
}

// CalendarEventInstance
type CalendarEventInstance struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID      int64  `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	EventID int64  `protobuf:"varint,2,opt,name=EventID,proto3" json:"EventID,omitempty"`
	Start   int64  `protobuf:"varint,3,opt,name=Start,proto3" json:"Start,omitempty"`
	End     int64  `protobuf:"varint,4,opt,name=End,proto3" json:"End,omitempty"`
	Colour  string `protobuf:"bytes,5,opt,name=Colour,proto3" json:"Colour,omitempty"`
}

func (x *CalendarEventInstance) Reset() {
	*x = CalendarEventInstance{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CalendarEventInstance) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CalendarEventInstance) ProtoMessage() {}

func (x *CalendarEventInstance) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CalendarEventInstance.ProtoReflect.Descriptor instead.
func (*CalendarEventInstance) Descriptor() ([]byte, []int) {
	return file_calendar_proto_rawDescGZIP(), []int{5}
}

func (x *CalendarEventInstance) GetID() int64 {
	if x != nil {
		return x.ID
	}
	return 0
}

func (x *CalendarEventInstance) GetEventID() int64 {
	if x != nil {
		return x.EventID
	}
	return 0
}

func (x *CalendarEventInstance) GetStart() int64 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *CalendarEventInstance) GetEnd() int64 {
	if x != nil {
		return x.End
	}
	return 0
}

func (x *CalendarEventInstance) GetColour() string {
	if x != nil {
		return x.Colour
	}
	return ""
}

var File_calendar_proto protoreflect.FileDescriptor

var file_calendar_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x63, 0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x03, 0x6d, 0x73, 0x67, 0x22, 0x4f, 0x0a, 0x11, 0x43, 0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61,
	0x72, 0x47, 0x65, 0x74, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x46, 0x72,
	0x6f, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x46, 0x72, 0x6f, 0x6d, 0x12, 0x0e,
	0x0a, 0x02, 0x54, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x54, 0x6f, 0x12, 0x16,
	0x0a, 0x06, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06,
	0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x22, 0x86, 0x02, 0x0a, 0x10, 0x43, 0x61, 0x6c, 0x65, 0x6e,
	0x64, 0x61, 0x72, 0x53, 0x65, 0x74, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x44, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x44,
	0x61, 0x74, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x61, 0x6e, 0x67,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x61,
	0x6e, 0x67, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x1c, 0x0a, 0x09, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x09, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x2c, 0x0a,
	0x06, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e,
	0x6d, 0x73, 0x67, 0x2e, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e, 0x67, 0x50, 0x65, 0x72,
	0x69, 0x6f, 0x64, 0x52, 0x06, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x41,
	0x6c, 0x6c, 0x44, 0x61, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x41, 0x6c, 0x6c,
	0x44, 0x61, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x65, 0x61, 0x6d, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x04, 0x54, 0x65, 0x61, 0x6d, 0x12, 0x16, 0x0a, 0x06, 0x47, 0x6c, 0x6f, 0x62, 0x61,
	0x6c, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x47, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x22,
	0xa6, 0x02, 0x0a, 0x11, 0x43, 0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x45, 0x64, 0x69, 0x74,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x12,
	0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x44, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x04, 0x44, 0x61, 0x74, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x53, 0x74, 0x61, 0x72, 0x74,
	0x52, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x53, 0x74, 0x61,
	0x72, 0x74, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x44, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x44, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e, 0x67,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e,
	0x67, 0x12, 0x2c, 0x0a, 0x06, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x14, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e,
	0x67, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x52, 0x06, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12,
	0x16, 0x0a, 0x06, 0x41, 0x6c, 0x6c, 0x44, 0x61, 0x79, 0x18, 0x08, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x06, 0x41, 0x6c, 0x6c, 0x44, 0x61, 0x79, 0x12, 0x2f, 0x0a, 0x06, 0x50, 0x6f, 0x6c, 0x69, 0x63,
	0x79, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x43, 0x61,
	0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x45, 0x64, 0x69, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79,
	0x52, 0x06, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x22, 0x2f, 0x0a, 0x13, 0x43, 0x61, 0x6c, 0x65,
	0x6e, 0x64, 0x61, 0x72, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12,
	0x18, 0x0a, 0x07, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x07, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x22, 0x97, 0x01, 0x0a, 0x0d, 0x43, 0x61,
	0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x49,
	0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x1c, 0x0a, 0x09, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x09, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x2c, 0x0a,
	0x06, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e,
	0x6d, 0x73, 0x67, 0x2e, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e, 0x67, 0x50, 0x65, 0x72,
	0x69, 0x6f, 0x64, 0x52, 0x06, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x41,
	0x6c, 0x6c, 0x44, 0x61, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x41, 0x6c, 0x6c,
	0x44, 0x61, 0x79, 0x22, 0x81, 0x01, 0x0a, 0x15, 0x43, 0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x0e, 0x0a,
	0x02, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x49, 0x44, 0x12, 0x18, 0x0a,
	0x07, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x12, 0x14, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x72, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x53, 0x74, 0x61, 0x72, 0x74, 0x12, 0x10, 0x0a,
	0x03, 0x45, 0x6e, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x45, 0x6e, 0x64, 0x12,
	0x16, 0x0a, 0x06, 0x43, 0x6f, 0x6c, 0x6f, 0x75, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x43, 0x6f, 0x6c, 0x6f, 0x75, 0x72, 0x2a, 0x78, 0x0a, 0x0f, 0x52, 0x65, 0x63, 0x75, 0x72,
	0x72, 0x69, 0x6e, 0x67, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x11, 0x0a, 0x0d, 0x52, 0x65,
	0x63, 0x75, 0x72, 0x72, 0x69, 0x6e, 0x67, 0x4e, 0x6f, 0x6e, 0x65, 0x10, 0x00, 0x12, 0x12, 0x0a,
	0x0e, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e, 0x67, 0x44, 0x61, 0x69, 0x6c, 0x79, 0x10,
	0x01, 0x12, 0x13, 0x0a, 0x0f, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e, 0x67, 0x57, 0x65,
	0x65, 0x6b, 0x6c, 0x79, 0x10, 0x02, 0x12, 0x14, 0x0a, 0x10, 0x52, 0x65, 0x63, 0x75, 0x72, 0x72,
	0x69, 0x6e, 0x67, 0x4d, 0x6f, 0x6e, 0x74, 0x68, 0x6c, 0x79, 0x10, 0x03, 0x12, 0x13, 0x0a, 0x0f,
	0x52, 0x65, 0x63, 0x75, 0x72, 0x72, 0x69, 0x6e, 0x67, 0x59, 0x65, 0x61, 0x72, 0x6c, 0x79, 0x10,
	0x04, 0x2a, 0x59, 0x0a, 0x12, 0x43, 0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x45, 0x64, 0x69,
	0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x13, 0x0a, 0x0f, 0x43, 0x61, 0x6c, 0x65, 0x6e,
	0x64, 0x61, 0x72, 0x45, 0x64, 0x69, 0x74, 0x4f, 0x6e, 0x65, 0x10, 0x00, 0x12, 0x19, 0x0a, 0x15,
	0x43, 0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x45, 0x64, 0x69, 0x74, 0x46, 0x6f, 0x6c, 0x6c,
	0x6f, 0x77, 0x69, 0x6e, 0x67, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f, 0x43, 0x61, 0x6c, 0x65, 0x6e,
	0x64, 0x61, 0x72, 0x45, 0x64, 0x69, 0x74, 0x41, 0x6c, 0x6c, 0x10, 0x02, 0x42, 0x08, 0x5a, 0x06,
	0x2e, 0x2f, 0x3b, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_calendar_proto_rawDescOnce sync.Once
	file_calendar_proto_rawDescData = file_calendar_proto_rawDesc
)

func file_calendar_proto_rawDescGZIP() []byte {
	file_calendar_proto_rawDescOnce.Do(func() {
		file_calendar_proto_rawDescData = protoimpl.X.CompressGZIP(file_calendar_proto_rawDescData)
	})
	return file_calendar_proto_rawDescData
}

var file_calendar_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_calendar_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_calendar_proto_goTypes = []interface{}{
	(RecurringPeriod)(0),          // 0: msg.RecurringPeriod
	(CalendarEditPolicy)(0),       // 1: msg.CalendarEditPolicy
	(*CalendarGetEvents)(nil),     // 2: msg.CalendarGetEvents
	(*CalendarSetEvent)(nil),      // 3: msg.CalendarSetEvent
	(*CalendarEditEvent)(nil),     // 4: msg.CalendarEditEvent
	(*CalendarRemoveEvent)(nil),   // 5: msg.CalendarRemoveEvent
	(*CalendarEvent)(nil),         // 6: msg.CalendarEvent
	(*CalendarEventInstance)(nil), // 7: msg.CalendarEventInstance
}
var file_calendar_proto_depIdxs = []int32{
	0, // 0: msg.CalendarSetEvent.Period:type_name -> msg.RecurringPeriod
	0, // 1: msg.CalendarEditEvent.Period:type_name -> msg.RecurringPeriod
	1, // 2: msg.CalendarEditEvent.Policy:type_name -> msg.CalendarEditPolicy
	0, // 3: msg.CalendarEvent.Period:type_name -> msg.RecurringPeriod
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_calendar_proto_init() }
func file_calendar_proto_init() {
	if File_calendar_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_calendar_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CalendarGetEvents); i {
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
		file_calendar_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CalendarSetEvent); i {
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
		file_calendar_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CalendarEditEvent); i {
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
		file_calendar_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CalendarRemoveEvent); i {
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
		file_calendar_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CalendarEvent); i {
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
		file_calendar_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CalendarEventInstance); i {
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
			RawDescriptor: file_calendar_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_calendar_proto_goTypes,
		DependencyIndexes: file_calendar_proto_depIdxs,
		EnumInfos:         file_calendar_proto_enumTypes,
		MessageInfos:      file_calendar_proto_msgTypes,
	}.Build()
	File_calendar_proto = out.File
	file_calendar_proto_rawDesc = nil
	file_calendar_proto_goTypes = nil
	file_calendar_proto_depIdxs = nil
}
