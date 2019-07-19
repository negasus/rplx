// Code generated by protoc-gen-go. DO NOT EDIT.
// source: message.proto

package rplx

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type SyncNodeValue struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Value                int64    `protobuf:"varint,2,opt,name=value,proto3" json:"value,omitempty"`
	Stamp                int64    `protobuf:"varint,3,opt,name=stamp,proto3" json:"stamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SyncNodeValue) Reset()         { *m = SyncNodeValue{} }
func (m *SyncNodeValue) String() string { return proto.CompactTextString(m) }
func (*SyncNodeValue) ProtoMessage()    {}
func (*SyncNodeValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_ccdf0675d32df01a, []int{0}
}
func (m *SyncNodeValue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SyncNodeValue.Unmarshal(m, b)
}
func (m *SyncNodeValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SyncNodeValue.Marshal(b, m, deterministic)
}
func (dst *SyncNodeValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyncNodeValue.Merge(dst, src)
}
func (m *SyncNodeValue) XXX_Size() int {
	return xxx_messageInfo_SyncNodeValue.Size(m)
}
func (m *SyncNodeValue) XXX_DiscardUnknown() {
	xxx_messageInfo_SyncNodeValue.DiscardUnknown(m)
}

var xxx_messageInfo_SyncNodeValue proto.InternalMessageInfo

func (m *SyncNodeValue) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *SyncNodeValue) GetValue() int64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *SyncNodeValue) GetStamp() int64 {
	if m != nil {
		return m.Stamp
	}
	return 0
}

type SyncVariable struct {
	// map key - nodeID
	NodesValues          map[string]*SyncNodeValue `protobuf:"bytes,1,rep,name=NodesValues,proto3" json:"NodesValues,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=v,proto3"`
	TTL                  int64                     `protobuf:"varint,2,opt,name=TTL,proto3" json:"TTL,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *SyncVariable) Reset()         { *m = SyncVariable{} }
func (m *SyncVariable) String() string { return proto.CompactTextString(m) }
func (*SyncVariable) ProtoMessage()    {}
func (*SyncVariable) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_ccdf0675d32df01a, []int{1}
}
func (m *SyncVariable) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SyncVariable.Unmarshal(m, b)
}
func (m *SyncVariable) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SyncVariable.Marshal(b, m, deterministic)
}
func (dst *SyncVariable) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyncVariable.Merge(dst, src)
}
func (m *SyncVariable) XXX_Size() int {
	return xxx_messageInfo_SyncVariable.Size(m)
}
func (m *SyncVariable) XXX_DiscardUnknown() {
	xxx_messageInfo_SyncVariable.DiscardUnknown(m)
}

var xxx_messageInfo_SyncVariable proto.InternalMessageInfo

func (m *SyncVariable) GetNodesValues() map[string]*SyncNodeValue {
	if m != nil {
		return m.NodesValues
	}
	return nil
}

func (m *SyncVariable) GetTTL() int64 {
	if m != nil {
		return m.TTL
	}
	return 0
}

type SyncRequest struct {
	NodeID string `protobuf:"bytes,1,opt,name=NodeID,proto3" json:"NodeID,omitempty"`
	// map key - variable name
	Variables            map[string]*SyncVariable `protobuf:"bytes,2,rep,name=Variables,proto3" json:"Variables,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=v,proto3"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *SyncRequest) Reset()         { *m = SyncRequest{} }
func (m *SyncRequest) String() string { return proto.CompactTextString(m) }
func (*SyncRequest) ProtoMessage()    {}
func (*SyncRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_ccdf0675d32df01a, []int{2}
}
func (m *SyncRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SyncRequest.Unmarshal(m, b)
}
func (m *SyncRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SyncRequest.Marshal(b, m, deterministic)
}
func (dst *SyncRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyncRequest.Merge(dst, src)
}
func (m *SyncRequest) XXX_Size() int {
	return xxx_messageInfo_SyncRequest.Size(m)
}
func (m *SyncRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SyncRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SyncRequest proto.InternalMessageInfo

func (m *SyncRequest) GetNodeID() string {
	if m != nil {
		return m.NodeID
	}
	return ""
}

func (m *SyncRequest) GetVariables() map[string]*SyncVariable {
	if m != nil {
		return m.Variables
	}
	return nil
}

type SyncResponse struct {
	Code                 int64    `protobuf:"varint,1,opt,name=Code,proto3" json:"Code,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SyncResponse) Reset()         { *m = SyncResponse{} }
func (m *SyncResponse) String() string { return proto.CompactTextString(m) }
func (*SyncResponse) ProtoMessage()    {}
func (*SyncResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_ccdf0675d32df01a, []int{3}
}
func (m *SyncResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SyncResponse.Unmarshal(m, b)
}
func (m *SyncResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SyncResponse.Marshal(b, m, deterministic)
}
func (dst *SyncResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyncResponse.Merge(dst, src)
}
func (m *SyncResponse) XXX_Size() int {
	return xxx_messageInfo_SyncResponse.Size(m)
}
func (m *SyncResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SyncResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SyncResponse proto.InternalMessageInfo

func (m *SyncResponse) GetCode() int64 {
	if m != nil {
		return m.Code
	}
	return 0
}

func init() {
	proto.RegisterType((*SyncNodeValue)(nil), "rplx.SyncNodeValue")
	proto.RegisterType((*SyncVariable)(nil), "rplx.SyncVariable")
	proto.RegisterMapType((map[string]*SyncNodeValue)(nil), "rplx.SyncVariable.NodesValuesEntry")
	proto.RegisterType((*SyncRequest)(nil), "rplx.SyncRequest")
	proto.RegisterMapType((map[string]*SyncVariable)(nil), "rplx.SyncRequest.VariablesEntry")
	proto.RegisterType((*SyncResponse)(nil), "rplx.SyncResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ReplicatorClient is the client API for Replicator service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ReplicatorClient interface {
	Sync(ctx context.Context, in *SyncRequest, opts ...grpc.CallOption) (*SyncResponse, error)
}

type replicatorClient struct {
	cc *grpc.ClientConn
}

func NewReplicatorClient(cc *grpc.ClientConn) ReplicatorClient {
	return &replicatorClient{cc}
}

func (c *replicatorClient) Sync(ctx context.Context, in *SyncRequest, opts ...grpc.CallOption) (*SyncResponse, error) {
	out := new(SyncResponse)
	err := c.cc.Invoke(ctx, "/rplx.Replicator/Sync", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReplicatorServer is the server API for Replicator service.
type ReplicatorServer interface {
	Sync(context.Context, *SyncRequest) (*SyncResponse, error)
}

func RegisterReplicatorServer(s *grpc.Server, srv ReplicatorServer) {
	s.RegisterService(&_Replicator_serviceDesc, srv)
}

func _Replicator_Sync_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SyncRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicatorServer).Sync(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rplx.Replicator/Sync",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicatorServer).Sync(ctx, req.(*SyncRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Replicator_serviceDesc = grpc.ServiceDesc{
	ServiceName: "rplx.Replicator",
	HandlerType: (*ReplicatorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Sync",
			Handler:    _Replicator_Sync_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "message.proto",
}

func init() { proto.RegisterFile("message.proto", fileDescriptor_message_ccdf0675d32df01a) }

var fileDescriptor_message_ccdf0675d32df01a = []byte{
	// 320 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0xcf, 0x4e, 0xc2, 0x40,
	0x10, 0xc6, 0xdd, 0x16, 0x48, 0x98, 0x0a, 0xc1, 0xd1, 0x98, 0x86, 0x53, 0x53, 0x2f, 0xf5, 0x52,
	0x13, 0xbc, 0x18, 0x13, 0xbd, 0x08, 0x07, 0xa2, 0x31, 0x66, 0x21, 0xdc, 0x17, 0x98, 0x18, 0x62,
	0x61, 0x6b, 0xb7, 0x18, 0x79, 0x32, 0x2f, 0x3e, 0x9c, 0xd9, 0xdd, 0xda, 0x16, 0xf1, 0x36, 0x7f,
	0xbe, 0x6f, 0xf6, 0x37, 0x93, 0x85, 0xce, 0x9a, 0x94, 0x12, 0xaf, 0x14, 0xa7, 0x99, 0xcc, 0x25,
	0x36, 0xb2, 0x34, 0xf9, 0x0c, 0x1f, 0xa1, 0x33, 0xd9, 0x6d, 0x16, 0xcf, 0x72, 0x49, 0x33, 0x91,
	0x6c, 0x09, 0xbb, 0xe0, 0x8c, 0x87, 0x3e, 0x0b, 0x58, 0xd4, 0xe6, 0xce, 0x78, 0x88, 0x67, 0xd0,
	0x34, 0x0d, 0xdf, 0x09, 0x58, 0xe4, 0x72, 0x9b, 0xe8, 0xea, 0x24, 0x17, 0xeb, 0xd4, 0x77, 0x6d,
	0xd5, 0x24, 0xe1, 0x37, 0x83, 0x63, 0x3d, 0x6d, 0x26, 0xb2, 0x95, 0x98, 0x27, 0x84, 0x23, 0xf0,
	0xf4, 0x64, 0x65, 0x4c, 0xca, 0x67, 0x81, 0x1b, 0x79, 0x83, 0x8b, 0x58, 0xbf, 0x1c, 0xd7, 0x85,
	0x71, 0x4d, 0x35, 0xda, 0xe4, 0xd9, 0x8e, 0xd7, 0x7d, 0xd8, 0x03, 0x77, 0x3a, 0x7d, 0x2a, 0x08,
	0x74, 0xd8, 0x9f, 0x40, 0xef, 0xaf, 0x45, 0xab, 0xde, 0x68, 0x57, 0xa0, 0xeb, 0x10, 0x2f, 0xa1,
	0xf9, 0x51, 0xb2, 0x7b, 0x83, 0xd3, 0xea, 0xe1, 0x72, 0x5f, 0x6e, 0x15, 0xb7, 0xce, 0x0d, 0x0b,
	0xbf, 0x18, 0x78, 0xba, 0xc9, 0xe9, 0x7d, 0x4b, 0x2a, 0xc7, 0x73, 0x68, 0x69, 0x5d, 0x79, 0x8e,
	0x22, 0xc3, 0x7b, 0x68, 0xff, 0x82, 0x2b, 0xdf, 0x31, 0x3b, 0x05, 0xd5, 0xe8, 0xc2, 0x1d, 0x97,
	0x12, 0xbb, 0x50, 0x65, 0xe9, 0xbf, 0x40, 0x77, 0xbf, 0xf9, 0x0f, 0x7a, 0xb4, 0x8f, 0x8e, 0x87,
	0x37, 0xab, 0x93, 0x87, 0xf6, 0xee, 0x9c, 0x54, 0x2a, 0x37, 0x8a, 0x10, 0xa1, 0xf1, 0x20, 0x97,
	0x64, 0x06, 0xba, 0xdc, 0xc4, 0x83, 0x3b, 0x00, 0x4e, 0x69, 0xb2, 0x5a, 0x88, 0x5c, 0x66, 0x78,
	0x05, 0x0d, 0xed, 0xc0, 0x93, 0x03, 0xf0, 0x3e, 0xd6, 0x4b, 0x76, 0x60, 0x78, 0x34, 0x6f, 0x99,
	0x5f, 0x73, 0xfd, 0x13, 0x00, 0x00, 0xff, 0xff, 0x7f, 0xee, 0x5c, 0xab, 0x46, 0x02, 0x00, 0x00,
}
