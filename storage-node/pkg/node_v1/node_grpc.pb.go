package node_v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	NodeService_StartElection_FullMethodName  = "/node_v1.NodeService/StartElection"
	NodeService_AnnounceLeader_FullMethodName = "/node_v1.NodeService/AnnounceLeader"
	NodeService_Prepare_FullMethodName        = "/node_v1.NodeService/Prepare"
	NodeService_SubmitValue_FullMethodName    = "/node_v1.NodeService/SubmitValue"
	NodeService_AppendValue_FullMethodName    = "/node_v1.NodeService/AppendValue"
	NodeService_AcceptEntries_FullMethodName  = "/node_v1.NodeService/AcceptEntries"
	NodeService_CommitEntries_FullMethodName  = "/node_v1.NodeService/CommitEntries"
)

type NodeServiceClient interface {
	StartElection(ctx context.Context, in *StartElectionRequest, opts ...grpc.CallOption) (*StartElectionResponse, error)
	AnnounceLeader(ctx context.Context, in *AnnounceLeaderRequest, opts ...grpc.CallOption) (*AnnounceLeaderResponse, error)
	Prepare(ctx context.Context, in *PrepareRequest, opts ...grpc.CallOption) (*PrepareResponse, error)
	SubmitValue(ctx context.Context, in *AppendValueRequest, opts ...grpc.CallOption) (*AppendValueResponse, error)
	AppendValue(ctx context.Context, in *AppendValueRequest, opts ...grpc.CallOption) (*AppendValueResponse, error)
	AcceptEntries(ctx context.Context, in *AcceptEntriesRequest, opts ...grpc.CallOption) (*JournalResponse, error)
	CommitEntries(ctx context.Context, in *CommitEntriesRequest, opts ...grpc.CallOption) (*JournalResponse, error)
}

type nodeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNodeServiceClient(cc grpc.ClientConnInterface) NodeServiceClient {
	return &nodeServiceClient{cc}
}

func (c *nodeServiceClient) StartElection(ctx context.Context, in *StartElectionRequest, opts ...grpc.CallOption) (*StartElectionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StartElectionResponse)
	err := c.cc.Invoke(ctx, NodeService_StartElection_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) AnnounceLeader(ctx context.Context, in *AnnounceLeaderRequest, opts ...grpc.CallOption) (*AnnounceLeaderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AnnounceLeaderResponse)
	err := c.cc.Invoke(ctx, NodeService_AnnounceLeader_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) Prepare(ctx context.Context, in *PrepareRequest, opts ...grpc.CallOption) (*PrepareResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PrepareResponse)
	err := c.cc.Invoke(ctx, NodeService_Prepare_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) SubmitValue(ctx context.Context, in *AppendValueRequest, opts ...grpc.CallOption) (*AppendValueResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AppendValueResponse)
	err := c.cc.Invoke(ctx, NodeService_SubmitValue_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) AppendValue(ctx context.Context, in *AppendValueRequest, opts ...grpc.CallOption) (*AppendValueResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AppendValueResponse)
	err := c.cc.Invoke(ctx, NodeService_AppendValue_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) AcceptEntries(ctx context.Context, in *AcceptEntriesRequest, opts ...grpc.CallOption) (*JournalResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(JournalResponse)
	err := c.cc.Invoke(ctx, NodeService_AcceptEntries_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) CommitEntries(ctx context.Context, in *CommitEntriesRequest, opts ...grpc.CallOption) (*JournalResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(JournalResponse)
	err := c.cc.Invoke(ctx, NodeService_CommitEntries_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type NodeServiceServer interface {
	StartElection(context.Context, *StartElectionRequest) (*StartElectionResponse, error)
	AnnounceLeader(context.Context, *AnnounceLeaderRequest) (*AnnounceLeaderResponse, error)
	Prepare(context.Context, *PrepareRequest) (*PrepareResponse, error)
	SubmitValue(context.Context, *AppendValueRequest) (*AppendValueResponse, error)
	AppendValue(context.Context, *AppendValueRequest) (*AppendValueResponse, error)
	AcceptEntries(context.Context, *AcceptEntriesRequest) (*JournalResponse, error)
	CommitEntries(context.Context, *CommitEntriesRequest) (*JournalResponse, error)
	mustEmbedUnimplementedNodeServiceServer()
}

type UnimplementedNodeServiceServer struct{}

func (UnimplementedNodeServiceServer) StartElection(context.Context, *StartElectionRequest) (*StartElectionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartElection not implemented")
}
func (UnimplementedNodeServiceServer) AnnounceLeader(context.Context, *AnnounceLeaderRequest) (*AnnounceLeaderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AnnounceLeader not implemented")
}
func (UnimplementedNodeServiceServer) Prepare(context.Context, *PrepareRequest) (*PrepareResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Prepare not implemented")
}
func (UnimplementedNodeServiceServer) SubmitValue(context.Context, *AppendValueRequest) (*AppendValueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitValue not implemented")
}
func (UnimplementedNodeServiceServer) AppendValue(context.Context, *AppendValueRequest) (*AppendValueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppendValue not implemented")
}
func (UnimplementedNodeServiceServer) AcceptEntries(context.Context, *AcceptEntriesRequest) (*JournalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AcceptEntries not implemented")
}
func (UnimplementedNodeServiceServer) CommitEntries(context.Context, *CommitEntriesRequest) (*JournalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommitEntries not implemented")
}
func (UnimplementedNodeServiceServer) mustEmbedUnimplementedNodeServiceServer() {}
func (UnimplementedNodeServiceServer) testEmbeddedByValue()                     {}

type UnsafeNodeServiceServer interface {
	mustEmbedUnimplementedNodeServiceServer()
}

func RegisterNodeServiceServer(s grpc.ServiceRegistrar, srv NodeServiceServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&NodeService_ServiceDesc, srv)
}

func _NodeService_StartElection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartElectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).StartElection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NodeService_StartElection_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).StartElection(ctx, req.(*StartElectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_AnnounceLeader_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AnnounceLeaderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).AnnounceLeader(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NodeService_AnnounceLeader_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).AnnounceLeader(ctx, req.(*AnnounceLeaderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_Prepare_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PrepareRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).Prepare(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NodeService_Prepare_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).Prepare(ctx, req.(*PrepareRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_SubmitValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppendValueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).SubmitValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NodeService_SubmitValue_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).SubmitValue(ctx, req.(*AppendValueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_AppendValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppendValueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).AppendValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NodeService_AppendValue_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).AppendValue(ctx, req.(*AppendValueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_AcceptEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AcceptEntriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).AcceptEntries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NodeService_AcceptEntries_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).AcceptEntries(ctx, req.(*AcceptEntriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_CommitEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitEntriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).CommitEntries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NodeService_CommitEntries_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).CommitEntries(ctx, req.(*CommitEntriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var NodeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "node_v1.NodeService",
	HandlerType: (*NodeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "StartElection",
			Handler:    _NodeService_StartElection_Handler,
		},
		{
			MethodName: "AnnounceLeader",
			Handler:    _NodeService_AnnounceLeader_Handler,
		},
		{
			MethodName: "Prepare",
			Handler:    _NodeService_Prepare_Handler,
		},
		{
			MethodName: "SubmitValue",
			Handler:    _NodeService_SubmitValue_Handler,
		},
		{
			MethodName: "AppendValue",
			Handler:    _NodeService_AppendValue_Handler,
		},
		{
			MethodName: "AcceptEntries",
			Handler:    _NodeService_AcceptEntries_Handler,
		},
		{
			MethodName: "CommitEntries",
			Handler:    _NodeService_CommitEntries_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "node.proto",
}
