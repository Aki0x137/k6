// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.3
// source: kontext.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	KontextKV_Get_FullMethodName = "/kontext.KontextKV/Get"
	KontextKV_Set_FullMethodName = "/kontext.KontextKV/Set"
)

// KontextKVClient is the client API for KontextKV service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KontextKVClient interface {
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	Set(ctx context.Context, in *SetRequest, opts ...grpc.CallOption) (*SetResponse, error)
}

type kontextKVClient struct {
	cc grpc.ClientConnInterface
}

func NewKontextKVClient(cc grpc.ClientConnInterface) KontextKVClient {
	return &kontextKVClient{cc}
}

func (c *kontextKVClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, KontextKV_Get_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kontextKVClient) Set(ctx context.Context, in *SetRequest, opts ...grpc.CallOption) (*SetResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SetResponse)
	err := c.cc.Invoke(ctx, KontextKV_Set_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KontextKVServer is the server API for KontextKV service.
// All implementations must embed UnimplementedKontextKVServer
// for forward compatibility.
type KontextKVServer interface {
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Set(context.Context, *SetRequest) (*SetResponse, error)
	mustEmbedUnimplementedKontextKVServer()
}

// UnimplementedKontextKVServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedKontextKVServer struct{}

func (UnimplementedKontextKVServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedKontextKVServer) Set(context.Context, *SetRequest) (*SetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (UnimplementedKontextKVServer) mustEmbedUnimplementedKontextKVServer() {}
func (UnimplementedKontextKVServer) testEmbeddedByValue()                   {}

// UnsafeKontextKVServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KontextKVServer will
// result in compilation errors.
type UnsafeKontextKVServer interface {
	mustEmbedUnimplementedKontextKVServer()
}

func RegisterKontextKVServer(s grpc.ServiceRegistrar, srv KontextKVServer) {
	// If the following call pancis, it indicates UnimplementedKontextKVServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&KontextKV_ServiceDesc, srv)
}

func _KontextKV_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KontextKVServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KontextKV_Get_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KontextKVServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KontextKV_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KontextKVServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KontextKV_Set_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KontextKVServer).Set(ctx, req.(*SetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// KontextKV_ServiceDesc is the grpc.ServiceDesc for KontextKV service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var KontextKV_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kontext.KontextKV",
	HandlerType: (*KontextKVServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _KontextKV_Get_Handler,
		},
		{
			MethodName: "Set",
			Handler:    _KontextKV_Set_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kontext.proto",
}
