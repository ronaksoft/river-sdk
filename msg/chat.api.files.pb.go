// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: chat.api.files.proto

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

var FileType_name = map[int32]string{
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

var FileType_value = map[string]int32{
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

func (x FileType) Enum() *FileType {
	p := new(FileType)
	*p = x
	return p
}

func (x FileType) String() string {
	return proto.EnumName(FileType_name, int32(x))
}

func (x *FileType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(FileType_value, data, "FileType")
	if err != nil {
		return err
	}
	*x = FileType(value)
	return nil
}

func (FileType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_05ced85bdf32a77a, []int{0}
}

// FileSavePart
// @Function
// @Return: Bool
type FileSavePart struct {
	FileID     int64  `protobuf:"varint,1,req,name=FileID" json:"FileID"`
	PartID     int32  `protobuf:"varint,2,req,name=PartID" json:"PartID"`
	TotalParts int32  `protobuf:"varint,3,req,name=TotalParts" json:"TotalParts"`
	Bytes      []byte `protobuf:"bytes,4,req,name=Bytes" json:"Bytes"`
}

func (m *FileSavePart) Reset()         { *m = FileSavePart{} }
func (m *FileSavePart) String() string { return proto.CompactTextString(m) }
func (*FileSavePart) ProtoMessage()    {}
func (*FileSavePart) Descriptor() ([]byte, []int) {
	return fileDescriptor_05ced85bdf32a77a, []int{0}
}
func (m *FileSavePart) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FileSavePart) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FileSavePart.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FileSavePart) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileSavePart.Merge(m, src)
}
func (m *FileSavePart) XXX_Size() int {
	return m.Size()
}
func (m *FileSavePart) XXX_DiscardUnknown() {
	xxx_messageInfo_FileSavePart.DiscardUnknown(m)
}

var xxx_messageInfo_FileSavePart proto.InternalMessageInfo

func (m *FileSavePart) GetFileID() int64 {
	if m != nil {
		return m.FileID
	}
	return 0
}

func (m *FileSavePart) GetPartID() int32 {
	if m != nil {
		return m.PartID
	}
	return 0
}

func (m *FileSavePart) GetTotalParts() int32 {
	if m != nil {
		return m.TotalParts
	}
	return 0
}

func (m *FileSavePart) GetBytes() []byte {
	if m != nil {
		return m.Bytes
	}
	return nil
}

// FileGetPart
// @Function
// @Return: File
type FileGet struct {
	Location *InputFileLocation `protobuf:"bytes,1,req,name=Location" json:"Location,omitempty"`
	Offset   int32              `protobuf:"varint,2,req,name=Offset" json:"Offset"`
	Limit    int32              `protobuf:"varint,3,req,name=Limit" json:"Limit"`
}

func (m *FileGet) Reset()         { *m = FileGet{} }
func (m *FileGet) String() string { return proto.CompactTextString(m) }
func (*FileGet) ProtoMessage()    {}
func (*FileGet) Descriptor() ([]byte, []int) {
	return fileDescriptor_05ced85bdf32a77a, []int{1}
}
func (m *FileGet) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FileGet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FileGet.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FileGet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileGet.Merge(m, src)
}
func (m *FileGet) XXX_Size() int {
	return m.Size()
}
func (m *FileGet) XXX_DiscardUnknown() {
	xxx_messageInfo_FileGet.DiscardUnknown(m)
}

var xxx_messageInfo_FileGet proto.InternalMessageInfo

func (m *FileGet) GetLocation() *InputFileLocation {
	if m != nil {
		return m.Location
	}
	return nil
}

func (m *FileGet) GetOffset() int32 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *FileGet) GetLimit() int32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

// File
type File struct {
	Type         FileType `protobuf:"varint,1,req,name=Type,enum=msg.FileType" json:"Type"`
	ModifiedTime int64    `protobuf:"varint,2,req,name=ModifiedTime" json:"ModifiedTime"`
	Bytes        []byte   `protobuf:"bytes,4,req,name=Bytes" json:"Bytes"`
}

func (m *File) Reset()         { *m = File{} }
func (m *File) String() string { return proto.CompactTextString(m) }
func (*File) ProtoMessage()    {}
func (*File) Descriptor() ([]byte, []int) {
	return fileDescriptor_05ced85bdf32a77a, []int{2}
}
func (m *File) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *File) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_File.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *File) XXX_Merge(src proto.Message) {
	xxx_messageInfo_File.Merge(m, src)
}
func (m *File) XXX_Size() int {
	return m.Size()
}
func (m *File) XXX_DiscardUnknown() {
	xxx_messageInfo_File.DiscardUnknown(m)
}

var xxx_messageInfo_File proto.InternalMessageInfo

func (m *File) GetType() FileType {
	if m != nil {
		return m.Type
	}
	return FileType_FileTypeUnknown
}

func (m *File) GetModifiedTime() int64 {
	if m != nil {
		return m.ModifiedTime
	}
	return 0
}

func (m *File) GetBytes() []byte {
	if m != nil {
		return m.Bytes
	}
	return nil
}

func init() {
	proto.RegisterEnum("msg.FileType", FileType_name, FileType_value)
	proto.RegisterType((*FileSavePart)(nil), "msg.FileSavePart")
	proto.RegisterType((*FileGet)(nil), "msg.FileGet")
	proto.RegisterType((*File)(nil), "msg.File")
}

func init() { proto.RegisterFile("chat.api.files.proto", fileDescriptor_05ced85bdf32a77a) }

var fileDescriptor_05ced85bdf32a77a = []byte{
	// 387 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0xdd, 0x6a, 0xe2, 0x40,
	0x18, 0x86, 0x33, 0x49, 0x8c, 0x32, 0xba, 0xeb, 0x30, 0xfb, 0x43, 0x90, 0x25, 0x2b, 0xb2, 0xb0,
	0x61, 0x0f, 0xc2, 0x62, 0x7b, 0x05, 0x41, 0x2a, 0x16, 0xa5, 0x62, 0x2d, 0x3d, 0x4e, 0x75, 0x92,
	0x0e, 0x4d, 0x32, 0x83, 0x19, 0x15, 0x6f, 0xa2, 0xf4, 0x46, 0x7a, 0x1f, 0x1e, 0x7a, 0xd8, 0xa3,
	0x52, 0xf4, 0x46, 0xca, 0xe4, 0xa7, 0x24, 0x27, 0x3d, 0xcb, 0x3c, 0xef, 0x33, 0xdf, 0xfb, 0x25,
	0x81, 0xdf, 0x17, 0xf7, 0x9e, 0x70, 0x3c, 0x4e, 0x1d, 0x9f, 0x86, 0x24, 0x71, 0xf8, 0x8a, 0x09,
	0x86, 0xb5, 0x28, 0x09, 0x3a, 0x3f, 0xd2, 0x68, 0xc1, 0x56, 0xc4, 0x11, 0x3b, 0x5e, 0x64, 0xbd,
	0x47, 0x00, 0x5b, 0x17, 0x34, 0x24, 0xd7, 0xde, 0x86, 0x4c, 0xbd, 0x95, 0xc0, 0x16, 0x34, 0xe4,
	0x79, 0x34, 0x30, 0x41, 0x57, 0xb5, 0x35, 0xd7, 0xd8, 0xbf, 0xfe, 0x56, 0xfe, 0x83, 0x59, 0x4e,
	0xf1, 0x2f, 0x68, 0x48, 0x6f, 0x34, 0x30, 0xd5, 0xae, 0x6a, 0xd7, 0x5c, 0x5d, 0xe6, 0xb3, 0x9c,
	0xe1, 0x3f, 0x10, 0xce, 0x99, 0xf0, 0x42, 0x79, 0x4c, 0x4c, 0xad, 0x64, 0x94, 0x38, 0xee, 0xc0,
	0x9a, 0xbb, 0x13, 0x24, 0x31, 0xf5, 0xae, 0x6a, 0xb7, 0x72, 0x21, 0x43, 0xbd, 0x2d, 0xac, 0xcb,
	0xa6, 0x21, 0x11, 0xb8, 0x0f, 0x1b, 0x63, 0xb6, 0xf0, 0x04, 0x65, 0x71, 0xba, 0x4c, 0xb3, 0xff,
	0xd3, 0x89, 0x92, 0xc0, 0x19, 0xc5, 0x7c, 0x2d, 0xa4, 0x54, 0xa4, 0xb3, 0x0f, 0x4f, 0xae, 0x77,
	0xe5, 0xfb, 0x09, 0x11, 0xd5, 0xf5, 0x32, 0x26, 0x8b, 0xc7, 0x34, 0xa2, 0xa2, 0xb2, 0x59, 0x86,
	0x7a, 0x6b, 0xa8, 0xcb, 0x99, 0xf8, 0x2f, 0xd4, 0xe7, 0x3b, 0x4e, 0xd2, 0xc6, 0xaf, 0xfd, 0x2f,
	0x69, 0xa3, 0x0c, 0x24, 0xcc, 0x6f, 0xa4, 0x02, 0xb6, 0x61, 0x6b, 0xc2, 0x96, 0xd4, 0xa7, 0x64,
	0x39, 0xa7, 0x11, 0x49, 0x0b, 0xb5, 0xdc, 0xa8, 0x24, 0x9f, 0xbd, 0xef, 0xbf, 0x67, 0x00, 0x1b,
	0xc5, 0x78, 0xfc, 0x0d, 0xb6, 0x8b, 0xe7, 0x9b, 0xf8, 0x21, 0x66, 0xdb, 0x18, 0x29, 0x65, 0x28,
	0x3f, 0x1f, 0xf5, 0x42, 0x04, 0x30, 0xca, 0x7e, 0x9b, 0x84, 0x97, 0x9c, 0x04, 0x48, 0xc5, 0x6d,
	0xd8, 0x2c, 0xc8, 0x90, 0xfa, 0x48, 0x2b, 0x83, 0x69, 0x1c, 0x20, 0xbd, 0x7c, 0xe7, 0x96, 0xdc,
	0x71, 0x54, 0x2b, 0x2b, 0x13, 0x7e, 0x86, 0x8c, 0x2a, 0x38, 0x47, 0xf5, 0x0a, 0x60, 0x1b, 0xd4,
	0x70, 0xcd, 0xfd, 0xd1, 0x02, 0x87, 0xa3, 0x05, 0xde, 0x8e, 0x16, 0x78, 0x3a, 0x59, 0xca, 0xe1,
	0x64, 0x29, 0x2f, 0x27, 0x4b, 0x79, 0x0f, 0x00, 0x00, 0xff, 0xff, 0x05, 0x0e, 0x28, 0x99, 0x7d,
	0x02, 0x00, 0x00,
}

func (m *FileSavePart) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FileSavePart) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintChatApiFiles(dAtA, i, uint64(m.FileID))
	dAtA[i] = 0x10
	i++
	i = encodeVarintChatApiFiles(dAtA, i, uint64(m.PartID))
	dAtA[i] = 0x18
	i++
	i = encodeVarintChatApiFiles(dAtA, i, uint64(m.TotalParts))
	if m.Bytes != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintChatApiFiles(dAtA, i, uint64(len(m.Bytes)))
		i += copy(dAtA[i:], m.Bytes)
	}
	return i, nil
}

func (m *FileGet) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FileGet) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Location == nil {
		return 0, github_com_gogo_protobuf_proto.NewRequiredNotSetError("Location")
	} else {
		dAtA[i] = 0xa
		i++
		i = encodeVarintChatApiFiles(dAtA, i, uint64(m.Location.Size()))
		n1, err := m.Location.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	dAtA[i] = 0x10
	i++
	i = encodeVarintChatApiFiles(dAtA, i, uint64(m.Offset))
	dAtA[i] = 0x18
	i++
	i = encodeVarintChatApiFiles(dAtA, i, uint64(m.Limit))
	return i, nil
}

func (m *File) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *File) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintChatApiFiles(dAtA, i, uint64(m.Type))
	dAtA[i] = 0x10
	i++
	i = encodeVarintChatApiFiles(dAtA, i, uint64(m.ModifiedTime))
	if m.Bytes != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintChatApiFiles(dAtA, i, uint64(len(m.Bytes)))
		i += copy(dAtA[i:], m.Bytes)
	}
	return i, nil
}

func encodeVarintChatApiFiles(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *FileSavePart) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovChatApiFiles(uint64(m.FileID))
	n += 1 + sovChatApiFiles(uint64(m.PartID))
	n += 1 + sovChatApiFiles(uint64(m.TotalParts))
	if m.Bytes != nil {
		l = len(m.Bytes)
		n += 1 + l + sovChatApiFiles(uint64(l))
	}
	return n
}

func (m *FileGet) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Location != nil {
		l = m.Location.Size()
		n += 1 + l + sovChatApiFiles(uint64(l))
	}
	n += 1 + sovChatApiFiles(uint64(m.Offset))
	n += 1 + sovChatApiFiles(uint64(m.Limit))
	return n
}

func (m *File) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovChatApiFiles(uint64(m.Type))
	n += 1 + sovChatApiFiles(uint64(m.ModifiedTime))
	if m.Bytes != nil {
		l = len(m.Bytes)
		n += 1 + l + sovChatApiFiles(uint64(l))
	}
	return n
}

func sovChatApiFiles(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozChatApiFiles(x uint64) (n int) {
	return sovChatApiFiles(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *FileSavePart) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChatApiFiles
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
			return fmt.Errorf("proto: FileSavePart: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FileSavePart: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FileID", wireType)
			}
			m.FileID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatApiFiles
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.FileID |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PartID", wireType)
			}
			m.PartID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatApiFiles
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PartID |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000002)
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalParts", wireType)
			}
			m.TotalParts = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatApiFiles
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TotalParts |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000004)
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bytes", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatApiFiles
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthChatApiFiles
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Bytes = append(m.Bytes[:0], dAtA[iNdEx:postIndex]...)
			if m.Bytes == nil {
				m.Bytes = []byte{}
			}
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000008)
		default:
			iNdEx = preIndex
			skippy, err := skipChatApiFiles(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChatApiFiles
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("FileID")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("PartID")
	}
	if hasFields[0]&uint64(0x00000004) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("TotalParts")
	}
	if hasFields[0]&uint64(0x00000008) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Bytes")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *FileGet) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChatApiFiles
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
			return fmt.Errorf("proto: FileGet: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FileGet: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Location", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatApiFiles
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
				return ErrInvalidLengthChatApiFiles
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Location == nil {
				m.Location = &InputFileLocation{}
			}
			if err := m.Location.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Offset", wireType)
			}
			m.Offset = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatApiFiles
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Offset |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000002)
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Limit", wireType)
			}
			m.Limit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatApiFiles
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Limit |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000004)
		default:
			iNdEx = preIndex
			skippy, err := skipChatApiFiles(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChatApiFiles
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Location")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Offset")
	}
	if hasFields[0]&uint64(0x00000004) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Limit")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *File) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChatApiFiles
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
			return fmt.Errorf("proto: File: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: File: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatApiFiles
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= (FileType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ModifiedTime", wireType)
			}
			m.ModifiedTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatApiFiles
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ModifiedTime |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000002)
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bytes", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChatApiFiles
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthChatApiFiles
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Bytes = append(m.Bytes[:0], dAtA[iNdEx:postIndex]...)
			if m.Bytes == nil {
				m.Bytes = []byte{}
			}
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000004)
		default:
			iNdEx = preIndex
			skippy, err := skipChatApiFiles(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChatApiFiles
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Type")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("ModifiedTime")
	}
	if hasFields[0]&uint64(0x00000004) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Bytes")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipChatApiFiles(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowChatApiFiles
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
					return 0, ErrIntOverflowChatApiFiles
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
					return 0, ErrIntOverflowChatApiFiles
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
				return 0, ErrInvalidLengthChatApiFiles
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowChatApiFiles
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
				next, err := skipChatApiFiles(dAtA[start:])
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
	ErrInvalidLengthChatApiFiles = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowChatApiFiles   = fmt.Errorf("proto: integer overflow")
)