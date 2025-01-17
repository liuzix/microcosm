// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: error.proto

package pb

import (
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	proto "github.com/gogo/protobuf/proto"
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

type ErrorCode int32

const (
	ErrorCode_None ErrorCode = 0
	// This master is not leader right now.
	ErrorCode_MasterNotLeader ErrorCode = 1
	// Executor has been removed so it can't be recognized.
	ErrorCode_UnknownExecutor ErrorCode = 2
	// no enough resource can be used.
	ErrorCode_NotEnoughResource ErrorCode = 3
	// submit subjob failed
	ErrorCode_SubJobSubmitFailed ErrorCode = 4
	// TombstoneExecuto
	ErrorCode_TombstoneExecutor ErrorCode = 5
	//
	ErrorCode_SubJobBuildFailed ErrorCode = 6
	ErrorCode_UnknownError      ErrorCode = 10001
)

var ErrorCode_name = map[int32]string{
	0:     "None",
	1:     "MasterNotLeader",
	2:     "UnknownExecutor",
	3:     "NotEnoughResource",
	4:     "SubJobSubmitFailed",
	5:     "TombstoneExecutor",
	6:     "SubJobBuildFailed",
	10001: "UnknownError",
}

var ErrorCode_value = map[string]int32{
	"None":               0,
	"MasterNotLeader":    1,
	"UnknownExecutor":    2,
	"NotEnoughResource":  3,
	"SubJobSubmitFailed": 4,
	"TombstoneExecutor":  5,
	"SubJobBuildFailed":  6,
	"UnknownError":       10001,
}

func (x ErrorCode) String() string {
	return proto.EnumName(ErrorCode_name, int32(x))
}

func (ErrorCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_0579b252106fcf4a, []int{0}
}

type Error struct {
	Code    ErrorCode `protobuf:"varint,1,opt,name=code,proto3,enum=pb.ErrorCode" json:"code,omitempty"`
	Message string    `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (m *Error) Reset()         { *m = Error{} }
func (m *Error) String() string { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()    {}
func (*Error) Descriptor() ([]byte, []int) {
	return fileDescriptor_0579b252106fcf4a, []int{0}
}
func (m *Error) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Error) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Error.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Error) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Error.Merge(m, src)
}
func (m *Error) XXX_Size() int {
	return m.Size()
}
func (m *Error) XXX_DiscardUnknown() {
	xxx_messageInfo_Error.DiscardUnknown(m)
}

var xxx_messageInfo_Error proto.InternalMessageInfo

func (m *Error) GetCode() ErrorCode {
	if m != nil {
		return m.Code
	}
	return ErrorCode_None
}

func (m *Error) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterEnum("pb.ErrorCode", ErrorCode_name, ErrorCode_value)
	proto.RegisterType((*Error)(nil), "pb.Error")
}

func init() { proto.RegisterFile("error.proto", fileDescriptor_0579b252106fcf4a) }

var fileDescriptor_0579b252106fcf4a = []byte{
	// 265 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x90, 0x4d, 0x4e, 0xc3, 0x30,
	0x10, 0x85, 0xe3, 0x90, 0x16, 0x6a, 0xfe, 0x5c, 0x23, 0x50, 0x56, 0x56, 0x61, 0x55, 0xb1, 0xc8,
	0x02, 0x6e, 0x50, 0x08, 0x0b, 0x04, 0x59, 0xa4, 0x70, 0x80, 0x38, 0x19, 0x95, 0x88, 0xc6, 0x13,
	0x39, 0xb6, 0xe0, 0x18, 0x70, 0x12, 0xae, 0xc1, 0xb2, 0x4b, 0x96, 0x28, 0xb9, 0x08, 0x72, 0x68,
	0xbb, 0x9c, 0xef, 0x7d, 0x9a, 0x19, 0x3d, 0xba, 0x0f, 0x5a, 0xa3, 0x8e, 0x6a, 0x8d, 0x06, 0xb9,
	0x5f, 0xcb, 0x8b, 0x5b, 0x3a, 0x88, 0x1d, 0xe2, 0xe7, 0x34, 0xc8, 0xb1, 0x80, 0x90, 0x4c, 0xc8,
	0xf4, 0xe8, 0xea, 0x30, 0xaa, 0x65, 0xd4, 0x07, 0x37, 0x58, 0x40, 0xda, 0x47, 0x3c, 0xa4, 0xbb,
	0x15, 0x34, 0x4d, 0xb6, 0x80, 0xd0, 0x9f, 0x90, 0xe9, 0x28, 0xdd, 0x8c, 0x97, 0x5f, 0x84, 0x8e,
	0xb6, 0x36, 0xdf, 0xa3, 0x41, 0x82, 0x0a, 0x98, 0xc7, 0x4f, 0xe8, 0xf1, 0x63, 0xd6, 0x18, 0xd0,
	0x09, 0x9a, 0x07, 0xc8, 0x0a, 0xd0, 0x8c, 0x38, 0xf8, 0xac, 0x5e, 0x15, 0xbe, 0xa9, 0xf8, 0x1d,
	0x72, 0x6b, 0x50, 0x33, 0x9f, 0x9f, 0xd2, 0x71, 0x82, 0x26, 0x56, 0x68, 0x17, 0x2f, 0x29, 0x34,
	0x68, 0x75, 0x0e, 0x6c, 0x87, 0x9f, 0x51, 0x3e, 0xb7, 0xf2, 0x1e, 0xe5, 0xdc, 0xca, 0xaa, 0x34,
	0x77, 0x59, 0xb9, 0x84, 0x82, 0x05, 0x4e, 0x7f, 0xc2, 0x4a, 0x36, 0x06, 0x15, 0x6c, 0xb7, 0x0c,
	0x1c, 0xfe, 0xd7, 0x67, 0xb6, 0x5c, 0x16, 0x6b, 0x7b, 0xc8, 0xc7, 0xf4, 0x60, 0x73, 0xd1, 0x3d,
	0xc9, 0x3e, 0x93, 0x59, 0xf8, 0xdd, 0x0a, 0xb2, 0x6a, 0x05, 0xf9, 0x6d, 0x05, 0xf9, 0xe8, 0x84,
	0xb7, 0xea, 0x84, 0xf7, 0xd3, 0x09, 0x4f, 0x0e, 0xfb, 0x72, 0xae, 0xff, 0x02, 0x00, 0x00, 0xff,
	0xff, 0xdd, 0x09, 0x98, 0x66, 0x2b, 0x01, 0x00, 0x00,
}

func (m *Error) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Error) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Error) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Message) > 0 {
		i -= len(m.Message)
		copy(dAtA[i:], m.Message)
		i = encodeVarintError(dAtA, i, uint64(len(m.Message)))
		i--
		dAtA[i] = 0x12
	}
	if m.Code != 0 {
		i = encodeVarintError(dAtA, i, uint64(m.Code))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintError(dAtA []byte, offset int, v uint64) int {
	offset -= sovError(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Error) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Code != 0 {
		n += 1 + sovError(uint64(m.Code))
	}
	l = len(m.Message)
	if l > 0 {
		n += 1 + l + sovError(uint64(l))
	}
	return n
}

func sovError(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozError(x uint64) (n int) {
	return sovError(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Error) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowError
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
			return fmt.Errorf("proto: Error: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Error: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Code", wireType)
			}
			m.Code = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowError
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Code |= ErrorCode(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Message", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowError
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
				return ErrInvalidLengthError
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthError
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Message = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipError(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthError
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
func skipError(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowError
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
					return 0, ErrIntOverflowError
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
					return 0, ErrIntOverflowError
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
				return 0, ErrInvalidLengthError
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupError
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthError
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthError        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowError          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupError = fmt.Errorf("proto: unexpected end of group")
)
