// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: api.files.proto

package msg

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import github_com_gogo_protobuf_proto "github.com/gogo/protobuf/proto"
import encoding_binary "encoding/binary"

import io "io"

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
	return fileDescriptor_api_files_d19aa5b00a9e9750, []int{0}
}

// FileSavePart
// @Function
// @Returns: Bool
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
	return fileDescriptor_api_files_d19aa5b00a9e9750, []int{0}
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
func (dst *FileSavePart) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileSavePart.Merge(dst, src)
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
// @Returns File
type FileGet struct {
	Location *FileLocation `protobuf:"bytes,1,req,name=Location" json:"Location,omitempty"`
	Offset   int32         `protobuf:"varint,2,req,name=Offset" json:"Offset"`
	Limit    int32         `protobuf:"varint,3,req,name=Limit" json:"Limit"`
}

func (m *FileGet) Reset()         { *m = FileGet{} }
func (m *FileGet) String() string { return proto.CompactTextString(m) }
func (*FileGet) ProtoMessage()    {}
func (*FileGet) Descriptor() ([]byte, []int) {
	return fileDescriptor_api_files_d19aa5b00a9e9750, []int{1}
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
func (dst *FileGet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileGet.Merge(dst, src)
}
func (m *FileGet) XXX_Size() int {
	return m.Size()
}
func (m *FileGet) XXX_DiscardUnknown() {
	xxx_messageInfo_FileGet.DiscardUnknown(m)
}

var xxx_messageInfo_FileGet proto.InternalMessageInfo

func (m *FileGet) GetLocation() *FileLocation {
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

// FileLocation
type FileLocation struct {
	PartitionID int64  `protobuf:"varint,1,req,name=PartitionID" json:"PartitionID"`
	FileID      int64  `protobuf:"varint,2,req,name=FileID" json:"FileID"`
	AccessHash  uint64 `protobuf:"fixed64,3,req,name=AccessHash" json:"AccessHash"`
	TotalParts  int32  `protobuf:"varint,4,req,name=TotalParts" json:"TotalParts"`
}

func (m *FileLocation) Reset()         { *m = FileLocation{} }
func (m *FileLocation) String() string { return proto.CompactTextString(m) }
func (*FileLocation) ProtoMessage()    {}
func (*FileLocation) Descriptor() ([]byte, []int) {
	return fileDescriptor_api_files_d19aa5b00a9e9750, []int{2}
}
func (m *FileLocation) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FileLocation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FileLocation.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *FileLocation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileLocation.Merge(dst, src)
}
func (m *FileLocation) XXX_Size() int {
	return m.Size()
}
func (m *FileLocation) XXX_DiscardUnknown() {
	xxx_messageInfo_FileLocation.DiscardUnknown(m)
}

var xxx_messageInfo_FileLocation proto.InternalMessageInfo

func (m *FileLocation) GetPartitionID() int64 {
	if m != nil {
		return m.PartitionID
	}
	return 0
}

func (m *FileLocation) GetFileID() int64 {
	if m != nil {
		return m.FileID
	}
	return 0
}

func (m *FileLocation) GetAccessHash() uint64 {
	if m != nil {
		return m.AccessHash
	}
	return 0
}

func (m *FileLocation) GetTotalParts() int32 {
	if m != nil {
		return m.TotalParts
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
	return fileDescriptor_api_files_d19aa5b00a9e9750, []int{3}
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
func (dst *File) XXX_Merge(src proto.Message) {
	xxx_messageInfo_File.Merge(dst, src)
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
	proto.RegisterType((*FileSavePart)(nil), "msg.FileSavePart")
	proto.RegisterType((*FileGet)(nil), "msg.FileGet")
	proto.RegisterType((*FileLocation)(nil), "msg.FileLocation")
	proto.RegisterType((*File)(nil), "msg.File")
	proto.RegisterEnum("msg.FileType", FileType_name, FileType_value)
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
	i = encodeVarintApiFiles(dAtA, i, uint64(m.FileID))
	dAtA[i] = 0x10
	i++
	i = encodeVarintApiFiles(dAtA, i, uint64(m.PartID))
	dAtA[i] = 0x18
	i++
	i = encodeVarintApiFiles(dAtA, i, uint64(m.TotalParts))
	if m.Bytes != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintApiFiles(dAtA, i, uint64(len(m.Bytes)))
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
		i = encodeVarintApiFiles(dAtA, i, uint64(m.Location.Size()))
		n1, err := m.Location.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	dAtA[i] = 0x10
	i++
	i = encodeVarintApiFiles(dAtA, i, uint64(m.Offset))
	dAtA[i] = 0x18
	i++
	i = encodeVarintApiFiles(dAtA, i, uint64(m.Limit))
	return i, nil
}

func (m *FileLocation) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FileLocation) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintApiFiles(dAtA, i, uint64(m.PartitionID))
	dAtA[i] = 0x10
	i++
	i = encodeVarintApiFiles(dAtA, i, uint64(m.FileID))
	dAtA[i] = 0x19
	i++
	encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(m.AccessHash))
	i += 8
	dAtA[i] = 0x20
	i++
	i = encodeVarintApiFiles(dAtA, i, uint64(m.TotalParts))
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
	i = encodeVarintApiFiles(dAtA, i, uint64(m.Type))
	dAtA[i] = 0x10
	i++
	i = encodeVarintApiFiles(dAtA, i, uint64(m.ModifiedTime))
	if m.Bytes != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintApiFiles(dAtA, i, uint64(len(m.Bytes)))
		i += copy(dAtA[i:], m.Bytes)
	}
	return i, nil
}

func encodeVarintApiFiles(dAtA []byte, offset int, v uint64) int {
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
	n += 1 + sovApiFiles(uint64(m.FileID))
	n += 1 + sovApiFiles(uint64(m.PartID))
	n += 1 + sovApiFiles(uint64(m.TotalParts))
	if m.Bytes != nil {
		l = len(m.Bytes)
		n += 1 + l + sovApiFiles(uint64(l))
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
		n += 1 + l + sovApiFiles(uint64(l))
	}
	n += 1 + sovApiFiles(uint64(m.Offset))
	n += 1 + sovApiFiles(uint64(m.Limit))
	return n
}

func (m *FileLocation) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovApiFiles(uint64(m.PartitionID))
	n += 1 + sovApiFiles(uint64(m.FileID))
	n += 9
	n += 1 + sovApiFiles(uint64(m.TotalParts))
	return n
}

func (m *File) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovApiFiles(uint64(m.Type))
	n += 1 + sovApiFiles(uint64(m.ModifiedTime))
	if m.Bytes != nil {
		l = len(m.Bytes)
		n += 1 + l + sovApiFiles(uint64(l))
	}
	return n
}

func sovApiFiles(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozApiFiles(x uint64) (n int) {
	return sovApiFiles(uint64((x << 1) ^ uint64((int64(x) >> 63))))
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
				return ErrIntOverflowApiFiles
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
					return ErrIntOverflowApiFiles
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
					return ErrIntOverflowApiFiles
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
					return ErrIntOverflowApiFiles
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
					return ErrIntOverflowApiFiles
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
				return ErrInvalidLengthApiFiles
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
			skippy, err := skipApiFiles(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthApiFiles
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
				return ErrIntOverflowApiFiles
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
					return ErrIntOverflowApiFiles
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
				return ErrInvalidLengthApiFiles
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Location == nil {
				m.Location = &FileLocation{}
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
					return ErrIntOverflowApiFiles
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
					return ErrIntOverflowApiFiles
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
			skippy, err := skipApiFiles(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthApiFiles
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
func (m *FileLocation) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowApiFiles
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
			return fmt.Errorf("proto: FileLocation: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FileLocation: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PartitionID", wireType)
			}
			m.PartitionID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApiFiles
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PartitionID |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FileID", wireType)
			}
			m.FileID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApiFiles
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
			hasFields[0] |= uint64(0x00000002)
		case 3:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccessHash", wireType)
			}
			m.AccessHash = 0
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			m.AccessHash = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			hasFields[0] |= uint64(0x00000004)
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalParts", wireType)
			}
			m.TotalParts = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApiFiles
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
			hasFields[0] |= uint64(0x00000008)
		default:
			iNdEx = preIndex
			skippy, err := skipApiFiles(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthApiFiles
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("PartitionID")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("FileID")
	}
	if hasFields[0]&uint64(0x00000004) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("AccessHash")
	}
	if hasFields[0]&uint64(0x00000008) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("TotalParts")
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
				return ErrIntOverflowApiFiles
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
					return ErrIntOverflowApiFiles
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
					return ErrIntOverflowApiFiles
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
					return ErrIntOverflowApiFiles
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
				return ErrInvalidLengthApiFiles
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
			skippy, err := skipApiFiles(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthApiFiles
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
func skipApiFiles(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowApiFiles
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
					return 0, ErrIntOverflowApiFiles
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
					return 0, ErrIntOverflowApiFiles
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
				return 0, ErrInvalidLengthApiFiles
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowApiFiles
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
				next, err := skipApiFiles(dAtA[start:])
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
	ErrInvalidLengthApiFiles = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowApiFiles   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("api.files.proto", fileDescriptor_api_files_d19aa5b00a9e9750) }

var fileDescriptor_api_files_d19aa5b00a9e9750 = []byte{
	// 412 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0xdf, 0x8a, 0xd3, 0x40,
	0x14, 0xc6, 0x33, 0x49, 0x9a, 0x2d, 0xa7, 0xd5, 0x8e, 0xe3, 0x4d, 0x10, 0x89, 0x25, 0x88, 0x06,
	0xc1, 0x20, 0xab, 0x2f, 0x60, 0x58, 0x5c, 0x57, 0x76, 0x71, 0xa9, 0x15, 0xaf, 0x63, 0x3b, 0x89,
	0x83, 0x49, 0x26, 0x74, 0xc6, 0x4a, 0x5f, 0x42, 0x7c, 0x0a, 0xef, 0x7c, 0x8f, 0x5e, 0xf6, 0xd2,
	0x2b, 0x91, 0xf6, 0x45, 0x64, 0xf2, 0x8f, 0x89, 0x96, 0xbd, 0xcb, 0xf9, 0x7d, 0xdf, 0x9c, 0xf3,
	0xf1, 0x05, 0x26, 0x71, 0xc9, 0xc2, 0x84, 0x65, 0x54, 0x84, 0xe5, 0x8a, 0x4b, 0x4e, 0xac, 0x5c,
	0xa4, 0xfe, 0x37, 0x04, 0xe3, 0x57, 0x2c, 0xa3, 0xef, 0xe2, 0x35, 0xbd, 0x8e, 0x57, 0x92, 0x78,
	0xe0, 0xa8, 0xf9, 0xe2, 0xcc, 0x45, 0x53, 0x33, 0xb0, 0x22, 0x67, 0xfb, 0xfb, 0x81, 0xf1, 0x0c,
	0xcd, 0x1a, 0x4a, 0xee, 0x83, 0xa3, 0x7c, 0x17, 0x67, 0xae, 0x39, 0x35, 0x83, 0x41, 0x64, 0x2b,
	0x7d, 0xd6, 0x30, 0xf2, 0x10, 0x60, 0xce, 0x65, 0x9c, 0xa9, 0x51, 0xb8, 0x96, 0xe6, 0xd0, 0x38,
	0xb9, 0x07, 0x83, 0x68, 0x23, 0xa9, 0x70, 0xed, 0xa9, 0x19, 0x8c, 0x1b, 0x43, 0x8d, 0xfc, 0x15,
	0x9c, 0xa8, 0x4b, 0xe7, 0x54, 0x92, 0xa7, 0x30, 0xbc, 0xe4, 0x8b, 0x58, 0x32, 0x5e, 0x54, 0x61,
	0x46, 0xa7, 0x77, 0xc2, 0x5c, 0xa4, 0xa1, 0xd2, 0x5b, 0x61, 0xd6, 0x59, 0x54, 0xb2, 0xb7, 0x49,
	0x22, 0xa8, 0xec, 0x27, 0xab, 0x99, 0xba, 0x79, 0xc9, 0x72, 0x26, 0x7b, 0xa1, 0x6a, 0xe4, 0xff,
	0x68, 0x4a, 0xe8, 0x56, 0x05, 0x30, 0x52, 0x49, 0x99, 0x1a, 0xfe, 0x6b, 0x42, 0x97, 0xb4, 0xba,
	0xcc, 0xa3, 0x75, 0x3d, 0x02, 0x78, 0xb9, 0x58, 0x50, 0x21, 0x5e, 0xc7, 0xe2, 0x53, 0x75, 0xdb,
	0xe9, 0x3c, 0x9a, 0xf2, 0x4f, 0x71, 0xf6, 0xf1, 0xe2, 0xfc, 0x2f, 0x60, 0xab, 0xbd, 0xe4, 0x31,
	0xd8, 0xf3, 0x4d, 0x49, 0xab, 0x60, 0xb7, 0x4f, 0x6f, 0x75, 0xad, 0x28, 0xd8, 0x3c, 0xab, 0x0c,
	0x24, 0x80, 0xf1, 0x15, 0x5f, 0xb2, 0x84, 0xd1, 0xe5, 0x9c, 0xe5, 0xb4, 0x09, 0x59, 0x3b, 0x7a,
	0xca, 0x4d, 0xff, 0xe4, 0xc9, 0x4f, 0x04, 0xc3, 0x76, 0x3d, 0xb9, 0x0b, 0x93, 0xf6, 0xfb, 0x7d,
	0xf1, 0xb9, 0xe0, 0x5f, 0x0b, 0x6c, 0xe8, 0xb0, 0x6a, 0x27, 0xce, 0x30, 0x22, 0xb8, 0x6e, 0x55,
	0xc1, 0x37, 0x25, 0x4d, 0xb1, 0x49, 0x26, 0x30, 0x6a, 0xc9, 0x39, 0x4b, 0xb0, 0xa5, 0x83, 0xeb,
	0x22, 0xc5, 0xb6, 0xfe, 0xe6, 0x03, 0xfd, 0x58, 0xe2, 0x81, 0x6e, 0xb9, 0x2a, 0x9f, 0x63, 0xa7,
	0x0f, 0x5e, 0xe0, 0x93, 0x1e, 0xe0, 0x6b, 0x3c, 0x8c, 0xdc, 0xed, 0xde, 0x43, 0xbb, 0xbd, 0x87,
	0xfe, 0xec, 0x3d, 0xf4, 0xfd, 0xe0, 0x19, 0xbb, 0x83, 0x67, 0xfc, 0x3a, 0x78, 0xc6, 0xdf, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x71, 0x50, 0x33, 0xaf, 0x05, 0x03, 0x00, 0x00,
}
