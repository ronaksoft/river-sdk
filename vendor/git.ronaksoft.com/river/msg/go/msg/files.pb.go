// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: files.proto

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

// FileType
type FileType int32

const (
	FileType_FileTypeUnknown FileType = 0
	FileType_FileTypePartial FileType = 1
	FileType_FileTypeJpeg    FileType = 2
	FileType_FileTypeGif     FileType = 3
	FileType_FileTypePng     FileType = 4
	FileType_FileTypeWebp    FileType = 5
	FileType_FileTypeMp3     FileType = 6
	FileType_FileTypeMp4     FileType = 7
	FileType_FileTypeMov     FileType = 8
)

// Enum value maps for FileType.
var (
	FileType_name = map[int32]string{
		0: "FileTypeUnknown",
		1: "FileTypePartial",
		2: "FileTypeJpeg",
		3: "FileTypeGif",
		4: "FileTypePng",
		5: "FileTypeWebp",
		6: "FileTypeMp3",
		7: "FileTypeMp4",
		8: "FileTypeMov",
	}
	FileType_value = map[string]int32{
		"FileTypeUnknown": 0,
		"FileTypePartial": 1,
		"FileTypeJpeg":    2,
		"FileTypeGif":     3,
		"FileTypePng":     4,
		"FileTypeWebp":    5,
		"FileTypeMp3":     6,
		"FileTypeMp4":     7,
		"FileTypeMov":     8,
	}
)

func (x FileType) Enum() *FileType {
	p := new(FileType)
	*p = x
	return p
}

func (x FileType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (FileType) Descriptor() protoreflect.EnumDescriptor {
	return file_files_proto_enumTypes[0].Descriptor()
}

func (FileType) Type() protoreflect.EnumType {
	return &file_files_proto_enumTypes[0]
}

func (x FileType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use FileType.Descriptor instead.
func (FileType) EnumDescriptor() ([]byte, []int) {
	return file_files_proto_rawDescGZIP(), []int{0}
}

// FileSavePart
// @Function
// @Return: Bool
type FileSavePart struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FileID     int64  `protobuf:"varint,1,opt,name=FileID,proto3" json:"FileID,omitempty"`
	PartID     int32  `protobuf:"varint,2,opt,name=PartID,proto3" json:"PartID,omitempty"`
	TotalParts int32  `protobuf:"varint,3,opt,name=TotalParts,proto3" json:"TotalParts,omitempty"`
	Bytes      []byte `protobuf:"bytes,4,opt,name=Bytes,proto3" json:"Bytes,omitempty"`
}

func (x *FileSavePart) Reset() {
	*x = FileSavePart{}
	if protoimpl.UnsafeEnabled {
		mi := &file_files_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileSavePart) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileSavePart) ProtoMessage() {}

func (x *FileSavePart) ProtoReflect() protoreflect.Message {
	mi := &file_files_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileSavePart.ProtoReflect.Descriptor instead.
func (*FileSavePart) Descriptor() ([]byte, []int) {
	return file_files_proto_rawDescGZIP(), []int{0}
}

func (x *FileSavePart) GetFileID() int64 {
	if x != nil {
		return x.FileID
	}
	return 0
}

func (x *FileSavePart) GetPartID() int32 {
	if x != nil {
		return x.PartID
	}
	return 0
}

func (x *FileSavePart) GetTotalParts() int32 {
	if x != nil {
		return x.TotalParts
	}
	return 0
}

func (x *FileSavePart) GetBytes() []byte {
	if x != nil {
		return x.Bytes
	}
	return nil
}

// FileGetPart
// @Function
// @Return: File
type FileGet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Location *InputFileLocation `protobuf:"bytes,1,opt,name=Location,proto3" json:"Location,omitempty"`
	Offset   int32              `protobuf:"varint,2,opt,name=Offset,proto3" json:"Offset,omitempty"`
	Limit    int32              `protobuf:"varint,3,opt,name=Limit,proto3" json:"Limit,omitempty"`
}

func (x *FileGet) Reset() {
	*x = FileGet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_files_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileGet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileGet) ProtoMessage() {}

func (x *FileGet) ProtoReflect() protoreflect.Message {
	mi := &file_files_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileGet.ProtoReflect.Descriptor instead.
func (*FileGet) Descriptor() ([]byte, []int) {
	return file_files_proto_rawDescGZIP(), []int{1}
}

func (x *FileGet) GetLocation() *InputFileLocation {
	if x != nil {
		return x.Location
	}
	return nil
}

func (x *FileGet) GetOffset() int32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *FileGet) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

// FileGetBySha256
// @Function
// @Return: FileLocation
type FileGetBySha256 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sha256   []byte `protobuf:"bytes,1,opt,name=Sha256,proto3" json:"Sha256,omitempty"`
	FileSize int32  `protobuf:"varint,2,opt,name=FileSize,proto3" json:"FileSize,omitempty"`
}

func (x *FileGetBySha256) Reset() {
	*x = FileGetBySha256{}
	if protoimpl.UnsafeEnabled {
		mi := &file_files_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileGetBySha256) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileGetBySha256) ProtoMessage() {}

func (x *FileGetBySha256) ProtoReflect() protoreflect.Message {
	mi := &file_files_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileGetBySha256.ProtoReflect.Descriptor instead.
func (*FileGetBySha256) Descriptor() ([]byte, []int) {
	return file_files_proto_rawDescGZIP(), []int{2}
}

func (x *FileGetBySha256) GetSha256() []byte {
	if x != nil {
		return x.Sha256
	}
	return nil
}

func (x *FileGetBySha256) GetFileSize() int32 {
	if x != nil {
		return x.FileSize
	}
	return 0
}

// File
type File struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type         FileType `protobuf:"varint,1,opt,name=Type,proto3,enum=msg.FileType" json:"Type,omitempty"`
	ModifiedTime int64    `protobuf:"varint,2,opt,name=ModifiedTime,proto3" json:"ModifiedTime,omitempty"`
	Bytes        []byte   `protobuf:"bytes,4,opt,name=Bytes,proto3" json:"Bytes,omitempty"`
	MD5Hash      string   `protobuf:"bytes,5,opt,name=MD5Hash,proto3" json:"MD5Hash,omitempty"`
}

func (x *File) Reset() {
	*x = File{}
	if protoimpl.UnsafeEnabled {
		mi := &file_files_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *File) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*File) ProtoMessage() {}

func (x *File) ProtoReflect() protoreflect.Message {
	mi := &file_files_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use File.ProtoReflect.Descriptor instead.
func (*File) Descriptor() ([]byte, []int) {
	return file_files_proto_rawDescGZIP(), []int{3}
}

func (x *File) GetType() FileType {
	if x != nil {
		return x.Type
	}
	return FileType_FileTypeUnknown
}

func (x *File) GetModifiedTime() int64 {
	if x != nil {
		return x.ModifiedTime
	}
	return 0
}

func (x *File) GetBytes() []byte {
	if x != nil {
		return x.Bytes
	}
	return nil
}

func (x *File) GetMD5Hash() string {
	if x != nil {
		return x.MD5Hash
	}
	return ""
}

var File_files_proto protoreflect.FileDescriptor

var file_files_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x6d,
	0x73, 0x67, 0x1a, 0x10, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x78, 0x0a, 0x0c, 0x46, 0x69, 0x6c, 0x65, 0x53, 0x61, 0x76, 0x65,
	0x50, 0x61, 0x72, 0x74, 0x12, 0x1a, 0x0a, 0x06, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x44, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x30, 0x01, 0x52, 0x06, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x44,
	0x12, 0x16, 0x0a, 0x06, 0x50, 0x61, 0x72, 0x74, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x06, 0x50, 0x61, 0x72, 0x74, 0x49, 0x44, 0x12, 0x1e, 0x0a, 0x0a, 0x54, 0x6f, 0x74, 0x61,
	0x6c, 0x50, 0x61, 0x72, 0x74, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x54, 0x6f,
	0x74, 0x61, 0x6c, 0x50, 0x61, 0x72, 0x74, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x42, 0x79, 0x74, 0x65,
	0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x42, 0x79, 0x74, 0x65, 0x73, 0x22, 0x6b,
	0x0a, 0x07, 0x46, 0x69, 0x6c, 0x65, 0x47, 0x65, 0x74, 0x12, 0x32, 0x0a, 0x08, 0x4c, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x6d, 0x73,
	0x67, 0x2e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x4c, 0x6f, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x08, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a,
	0x06, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x4f,
	0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x22, 0x45, 0x0a, 0x0f, 0x46,
	0x69, 0x6c, 0x65, 0x47, 0x65, 0x74, 0x42, 0x79, 0x53, 0x68, 0x61, 0x32, 0x35, 0x36, 0x12, 0x16,
	0x0a, 0x06, 0x53, 0x68, 0x61, 0x32, 0x35, 0x36, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06,
	0x53, 0x68, 0x61, 0x32, 0x35, 0x36, 0x12, 0x1a, 0x0a, 0x08, 0x46, 0x69, 0x6c, 0x65, 0x53, 0x69,
	0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x46, 0x69, 0x6c, 0x65, 0x53, 0x69,
	0x7a, 0x65, 0x22, 0x7d, 0x0a, 0x04, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x21, 0x0a, 0x04, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x46,
	0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x22, 0x0a,
	0x0c, 0x4d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x0c, 0x4d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x54, 0x69, 0x6d,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x42, 0x79, 0x74, 0x65, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x05, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x4d, 0x44, 0x35, 0x48, 0x61,
	0x73, 0x68, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x4d, 0x44, 0x35, 0x48, 0x61, 0x73,
	0x68, 0x2a, 0xad, 0x01, 0x0a, 0x08, 0x46, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x13,
	0x0a, 0x0f, 0x46, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x55, 0x6e, 0x6b, 0x6e, 0x6f, 0x77,
	0x6e, 0x10, 0x00, 0x12, 0x13, 0x0a, 0x0f, 0x46, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x50,
	0x61, 0x72, 0x74, 0x69, 0x61, 0x6c, 0x10, 0x01, 0x12, 0x10, 0x0a, 0x0c, 0x46, 0x69, 0x6c, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x4a, 0x70, 0x65, 0x67, 0x10, 0x02, 0x12, 0x0f, 0x0a, 0x0b, 0x46, 0x69,
	0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x47, 0x69, 0x66, 0x10, 0x03, 0x12, 0x0f, 0x0a, 0x0b, 0x46,
	0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x50, 0x6e, 0x67, 0x10, 0x04, 0x12, 0x10, 0x0a, 0x0c,
	0x46, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x57, 0x65, 0x62, 0x70, 0x10, 0x05, 0x12, 0x0f,
	0x0a, 0x0b, 0x46, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x4d, 0x70, 0x33, 0x10, 0x06, 0x12,
	0x0f, 0x0a, 0x0b, 0x46, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x4d, 0x70, 0x34, 0x10, 0x07,
	0x12, 0x0f, 0x0a, 0x0b, 0x46, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x4d, 0x6f, 0x76, 0x10,
	0x08, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x3b, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_files_proto_rawDescOnce sync.Once
	file_files_proto_rawDescData = file_files_proto_rawDesc
)

func file_files_proto_rawDescGZIP() []byte {
	file_files_proto_rawDescOnce.Do(func() {
		file_files_proto_rawDescData = protoimpl.X.CompressGZIP(file_files_proto_rawDescData)
	})
	return file_files_proto_rawDescData
}

var file_files_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_files_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_files_proto_goTypes = []interface{}{
	(FileType)(0),             // 0: msg.FileType
	(*FileSavePart)(nil),      // 1: msg.FileSavePart
	(*FileGet)(nil),           // 2: msg.FileGet
	(*FileGetBySha256)(nil),   // 3: msg.FileGetBySha256
	(*File)(nil),              // 4: msg.File
	(*InputFileLocation)(nil), // 5: msg.InputFileLocation
}
var file_files_proto_depIdxs = []int32{
	5, // 0: msg.FileGet.Location:type_name -> msg.InputFileLocation
	0, // 1: msg.File.Type:type_name -> msg.FileType
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_files_proto_init() }
func file_files_proto_init() {
	if File_files_proto != nil {
		return
	}
	file_core_types_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_files_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileSavePart); i {
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
		file_files_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileGet); i {
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
		file_files_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileGetBySha256); i {
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
		file_files_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*File); i {
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
			RawDescriptor: file_files_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_files_proto_goTypes,
		DependencyIndexes: file_files_proto_depIdxs,
		EnumInfos:         file_files_proto_enumTypes,
		MessageInfos:      file_files_proto_msgTypes,
	}.Build()
	File_files_proto = out.File
	file_files_proto_rawDesc = nil
	file_files_proto_goTypes = nil
	file_files_proto_depIdxs = nil
}
