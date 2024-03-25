// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.3
// source: protos/ensemble-service.proto

package protos

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// EnsembleOperatorClient is the client API for EnsembleOperator service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EnsembleOperatorClient interface {
	RequestStatus(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*Response, error)
	RequestAction(ctx context.Context, in *ActionRequest, opts ...grpc.CallOption) (*Response, error)
}

type ensembleOperatorClient struct {
	cc grpc.ClientConnInterface
}

func NewEnsembleOperatorClient(cc grpc.ClientConnInterface) EnsembleOperatorClient {
	return &ensembleOperatorClient{cc}
}

func (c *ensembleOperatorClient) RequestStatus(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/convergedcomputing.org.grpc.v1.EnsembleOperator/RequestStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ensembleOperatorClient) RequestAction(ctx context.Context, in *ActionRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/convergedcomputing.org.grpc.v1.EnsembleOperator/RequestAction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EnsembleOperatorServer is the server API for EnsembleOperator service.
// All implementations must embed UnimplementedEnsembleOperatorServer
// for forward compatibility
type EnsembleOperatorServer interface {
	RequestStatus(context.Context, *StatusRequest) (*Response, error)
	RequestAction(context.Context, *ActionRequest) (*Response, error)
	mustEmbedUnimplementedEnsembleOperatorServer()
}

// UnimplementedEnsembleOperatorServer must be embedded to have forward compatible implementations.
type UnimplementedEnsembleOperatorServer struct {
}

func (UnimplementedEnsembleOperatorServer) RequestStatus(context.Context, *StatusRequest) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestStatus not implemented")
}
func (UnimplementedEnsembleOperatorServer) RequestAction(context.Context, *ActionRequest) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestAction not implemented")
}
func (UnimplementedEnsembleOperatorServer) mustEmbedUnimplementedEnsembleOperatorServer() {}

// UnsafeEnsembleOperatorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EnsembleOperatorServer will
// result in compilation errors.
type UnsafeEnsembleOperatorServer interface {
	mustEmbedUnimplementedEnsembleOperatorServer()
}

func RegisterEnsembleOperatorServer(s grpc.ServiceRegistrar, srv EnsembleOperatorServer) {
	s.RegisterService(&EnsembleOperator_ServiceDesc, srv)
}

func _EnsembleOperator_RequestStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EnsembleOperatorServer).RequestStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/convergedcomputing.org.grpc.v1.EnsembleOperator/RequestStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EnsembleOperatorServer).RequestStatus(ctx, req.(*StatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EnsembleOperator_RequestAction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ActionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EnsembleOperatorServer).RequestAction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/convergedcomputing.org.grpc.v1.EnsembleOperator/RequestAction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EnsembleOperatorServer).RequestAction(ctx, req.(*ActionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// EnsembleOperator_ServiceDesc is the grpc.ServiceDesc for EnsembleOperator service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EnsembleOperator_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "convergedcomputing.org.grpc.v1.EnsembleOperator",
	HandlerType: (*EnsembleOperatorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestStatus",
			Handler:    _EnsembleOperator_RequestStatus_Handler,
		},
		{
			MethodName: "RequestAction",
			Handler:    _EnsembleOperator_RequestAction_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protos/ensemble-service.proto",
}
