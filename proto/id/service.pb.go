// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service.proto

package id

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Status int32

const (
	Status_Success Status = 0
	Status_Error   Status = 1
)

var Status_name = map[int32]string{
	0: "Success",
	1: "Error",
}

var Status_value = map[string]int32{
	"Success": 0,
	"Error":   1,
}

func (x Status) String() string {
	return proto.EnumName(Status_name, int32(x))
}

func (Status) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{0}
}

type Generate32BitRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Generate32BitRequest) Reset()         { *m = Generate32BitRequest{} }
func (m *Generate32BitRequest) String() string { return proto.CompactTextString(m) }
func (*Generate32BitRequest) ProtoMessage()    {}
func (*Generate32BitRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{0}
}

func (m *Generate32BitRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Generate32BitRequest.Unmarshal(m, b)
}
func (m *Generate32BitRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Generate32BitRequest.Marshal(b, m, deterministic)
}
func (m *Generate32BitRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Generate32BitRequest.Merge(m, src)
}
func (m *Generate32BitRequest) XXX_Size() int {
	return xxx_messageInfo_Generate32BitRequest.Size(m)
}
func (m *Generate32BitRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_Generate32BitRequest.DiscardUnknown(m)
}

var xxx_messageInfo_Generate32BitRequest proto.InternalMessageInfo

type Generate32BitReply struct {
	Status               Status   `protobuf:"varint,1,opt,name=status,proto3,enum=id.Status" json:"status,omitempty"`
	Id                   int32    `protobuf:"varint,2,opt,name=id,proto3" json:"id,omitempty"`
	Source               string   `protobuf:"bytes,3,opt,name=source,proto3" json:"source,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Generate32BitReply) Reset()         { *m = Generate32BitReply{} }
func (m *Generate32BitReply) String() string { return proto.CompactTextString(m) }
func (*Generate32BitReply) ProtoMessage()    {}
func (*Generate32BitReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{1}
}

func (m *Generate32BitReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Generate32BitReply.Unmarshal(m, b)
}
func (m *Generate32BitReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Generate32BitReply.Marshal(b, m, deterministic)
}
func (m *Generate32BitReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Generate32BitReply.Merge(m, src)
}
func (m *Generate32BitReply) XXX_Size() int {
	return xxx_messageInfo_Generate32BitReply.Size(m)
}
func (m *Generate32BitReply) XXX_DiscardUnknown() {
	xxx_messageInfo_Generate32BitReply.DiscardUnknown(m)
}

var xxx_messageInfo_Generate32BitReply proto.InternalMessageInfo

func (m *Generate32BitReply) GetStatus() Status {
	if m != nil {
		return m.Status
	}
	return Status_Success
}

func (m *Generate32BitReply) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Generate32BitReply) GetSource() string {
	if m != nil {
		return m.Source
	}
	return ""
}

type GetSourceRequest struct {
	Id                   int32    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetSourceRequest) Reset()         { *m = GetSourceRequest{} }
func (m *GetSourceRequest) String() string { return proto.CompactTextString(m) }
func (*GetSourceRequest) ProtoMessage()    {}
func (*GetSourceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{2}
}

func (m *GetSourceRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetSourceRequest.Unmarshal(m, b)
}
func (m *GetSourceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetSourceRequest.Marshal(b, m, deterministic)
}
func (m *GetSourceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetSourceRequest.Merge(m, src)
}
func (m *GetSourceRequest) XXX_Size() int {
	return xxx_messageInfo_GetSourceRequest.Size(m)
}
func (m *GetSourceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetSourceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetSourceRequest proto.InternalMessageInfo

func (m *GetSourceRequest) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

type GetSourceReply struct {
	Status               Status   `protobuf:"varint,1,opt,name=status,proto3,enum=id.Status" json:"status,omitempty"`
	Source               string   `protobuf:"bytes,2,opt,name=source,proto3" json:"source,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetSourceReply) Reset()         { *m = GetSourceReply{} }
func (m *GetSourceReply) String() string { return proto.CompactTextString(m) }
func (*GetSourceReply) ProtoMessage()    {}
func (*GetSourceReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{3}
}

func (m *GetSourceReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetSourceReply.Unmarshal(m, b)
}
func (m *GetSourceReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetSourceReply.Marshal(b, m, deterministic)
}
func (m *GetSourceReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetSourceReply.Merge(m, src)
}
func (m *GetSourceReply) XXX_Size() int {
	return xxx_messageInfo_GetSourceReply.Size(m)
}
func (m *GetSourceReply) XXX_DiscardUnknown() {
	xxx_messageInfo_GetSourceReply.DiscardUnknown(m)
}

var xxx_messageInfo_GetSourceReply proto.InternalMessageInfo

func (m *GetSourceReply) GetStatus() Status {
	if m != nil {
		return m.Status
	}
	return Status_Success
}

func (m *GetSourceReply) GetSource() string {
	if m != nil {
		return m.Source
	}
	return ""
}

func init() {
	proto.RegisterEnum("id.Status", Status_name, Status_value)
	proto.RegisterType((*Generate32BitRequest)(nil), "id.Generate32BitRequest")
	proto.RegisterType((*Generate32BitReply)(nil), "id.Generate32BitReply")
	proto.RegisterType((*GetSourceRequest)(nil), "id.GetSourceRequest")
	proto.RegisterType((*GetSourceReply)(nil), "id.GetSourceReply")
}

func init() { proto.RegisterFile("service.proto", fileDescriptor_a0b84a42fa06f626) }

var fileDescriptor_a0b84a42fa06f626 = []byte{
	// 249 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x91, 0x3f, 0x6b, 0xc3, 0x30,
	0x10, 0xc5, 0x23, 0x95, 0x38, 0xf8, 0x4a, 0x8c, 0x39, 0x82, 0x31, 0x99, 0x8c, 0x26, 0xd3, 0xc1,
	0x83, 0x33, 0x74, 0x6f, 0x29, 0x59, 0x3a, 0xd9, 0x5f, 0xa0, 0xa9, 0x75, 0x83, 0x20, 0x54, 0xae,
	0xfe, 0x14, 0xf2, 0x09, 0xfa, 0xb5, 0x4b, 0x65, 0xb5, 0x4d, 0x82, 0x87, 0x8e, 0xf7, 0xee, 0x3d,
	0x7e, 0xba, 0x27, 0x58, 0x5b, 0x32, 0x1f, 0x6a, 0xa0, 0x66, 0x34, 0xda, 0x69, 0xe4, 0x4a, 0x8a,
	0x02, 0x36, 0x7b, 0x7a, 0x23, 0x73, 0x70, 0xb4, 0x6b, 0x1f, 0x94, 0xeb, 0xe8, 0xdd, 0x93, 0x75,
	0xe2, 0x05, 0xf0, 0x4a, 0x1f, 0x8f, 0x27, 0x14, 0x90, 0x58, 0x77, 0x70, 0xde, 0x96, 0xac, 0x62,
	0x75, 0xd6, 0x42, 0xa3, 0x64, 0xd3, 0x07, 0xa5, 0x8b, 0x1b, 0xcc, 0x80, 0x2b, 0x59, 0xf2, 0x8a,
	0xd5, 0xcb, 0x8e, 0x2b, 0x89, 0x05, 0x24, 0x56, 0x7b, 0x33, 0x50, 0x79, 0x53, 0xb1, 0x3a, 0xed,
	0xe2, 0x24, 0x04, 0xe4, 0x7b, 0x72, 0x7d, 0x18, 0x22, 0x35, 0x66, 0xd9, 0x4f, 0x56, 0x3c, 0x43,
	0x76, 0xe6, 0xf9, 0xef, 0x0b, 0xfe, 0x88, 0xfc, 0x9c, 0x78, 0x57, 0x41, 0x32, 0x39, 0xf1, 0x16,
	0x56, 0xbd, 0x1f, 0x06, 0xb2, 0x36, 0x5f, 0x60, 0x0a, 0xcb, 0x27, 0x63, 0xb4, 0xc9, 0x59, 0xfb,
	0xc9, 0x60, 0xd5, 0x4f, 0x1d, 0xe1, 0x23, 0xac, 0x2f, 0x1a, 0xc0, 0xf2, 0x1b, 0x35, 0x57, 0xd6,
	0xb6, 0x98, 0xd9, 0x8c, 0xc7, 0x93, 0x58, 0xe0, 0x3d, 0xa4, 0xbf, 0x07, 0xe0, 0x66, 0xb2, 0x5d,
	0xde, 0xbc, 0xc5, 0x2b, 0x35, 0x04, 0x5f, 0x93, 0xf0, 0x45, 0xbb, 0xaf, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xa0, 0xa6, 0xfd, 0x3c, 0xb3, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ServiceClient is the client API for Service service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ServiceClient interface {
	Generate32Bit(ctx context.Context, in *Generate32BitRequest, opts ...grpc.CallOption) (*Generate32BitReply, error)
	GetSource(ctx context.Context, in *GetSourceRequest, opts ...grpc.CallOption) (*GetSourceReply, error)
}

type serviceClient struct {
	cc *grpc.ClientConn
}

func NewServiceClient(cc *grpc.ClientConn) ServiceClient {
	return &serviceClient{cc}
}

func (c *serviceClient) Generate32Bit(ctx context.Context, in *Generate32BitRequest, opts ...grpc.CallOption) (*Generate32BitReply, error) {
	out := new(Generate32BitReply)
	err := c.cc.Invoke(ctx, "/id.Service/Generate32Bit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) GetSource(ctx context.Context, in *GetSourceRequest, opts ...grpc.CallOption) (*GetSourceReply, error) {
	out := new(GetSourceReply)
	err := c.cc.Invoke(ctx, "/id.Service/GetSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServiceServer is the server API for Service service.
type ServiceServer interface {
	Generate32Bit(context.Context, *Generate32BitRequest) (*Generate32BitReply, error)
	GetSource(context.Context, *GetSourceRequest) (*GetSourceReply, error)
}

// UnimplementedServiceServer can be embedded to have forward compatible implementations.
type UnimplementedServiceServer struct {
}

func (*UnimplementedServiceServer) Generate32Bit(ctx context.Context, req *Generate32BitRequest) (*Generate32BitReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Generate32Bit not implemented")
}
func (*UnimplementedServiceServer) GetSource(ctx context.Context, req *GetSourceRequest) (*GetSourceReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSource not implemented")
}

func RegisterServiceServer(s *grpc.Server, srv ServiceServer) {
	s.RegisterService(&_Service_serviceDesc, srv)
}

func _Service_Generate32Bit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Generate32BitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).Generate32Bit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/id.Service/Generate32Bit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).Generate32Bit(ctx, req.(*Generate32BitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_GetSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).GetSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/id.Service/GetSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).GetSource(ctx, req.(*GetSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Service_serviceDesc = grpc.ServiceDesc{
	ServiceName: "id.Service",
	HandlerType: (*ServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Generate32Bit",
			Handler:    _Service_Generate32Bit_Handler,
		},
		{
			MethodName: "GetSource",
			Handler:    _Service_GetSource_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service.proto",
}
