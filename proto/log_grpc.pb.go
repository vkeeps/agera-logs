// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.2
// source: proto/log.proto

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
	LogService_SendLog_FullMethodName = "/log.LogService/SendLog"
)

// LogServiceClient is the client API for LogService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LogServiceClient interface {
	SendLog(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (*LogResponse, error)
}

type logServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewLogServiceClient(cc grpc.ClientConnInterface) LogServiceClient {
	return &logServiceClient{cc}
}

func (c *logServiceClient) SendLog(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (*LogResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LogResponse)
	err := c.cc.Invoke(ctx, LogService_SendLog_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LogServiceServer is the server API for LogService service.
// All implementations must embed UnimplementedLogServiceServer
// for forward compatibility.
type LogServiceServer interface {
	SendLog(context.Context, *LogRequest) (*LogResponse, error)
	mustEmbedUnimplementedLogServiceServer()
}

// UnimplementedLogServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedLogServiceServer struct{}

func (UnimplementedLogServiceServer) SendLog(context.Context, *LogRequest) (*LogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendLog not implemented")
}
func (UnimplementedLogServiceServer) mustEmbedUnimplementedLogServiceServer() {}
func (UnimplementedLogServiceServer) testEmbeddedByValue()                    {}

// UnsafeLogServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LogServiceServer will
// result in compilation errors.
type UnsafeLogServiceServer interface {
	mustEmbedUnimplementedLogServiceServer()
}

func RegisterLogServiceServer(s grpc.ServiceRegistrar, srv LogServiceServer) {
	// If the following call pancis, it indicates UnimplementedLogServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&LogService_ServiceDesc, srv)
}

func _LogService_SendLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogServiceServer).SendLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LogService_SendLog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogServiceServer).SendLog(ctx, req.(*LogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// LogService_ServiceDesc is the grpc.ServiceDesc for LogService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LogService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "log.LogService",
	HandlerType: (*LogServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendLog",
			Handler:    _LogService_SendLog_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/log.proto",
}
