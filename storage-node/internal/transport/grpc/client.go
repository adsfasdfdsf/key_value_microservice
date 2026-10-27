package grpc

import (
	"context"
	"time"

	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	Address string
}

func (cl *Client) AppendValue(ctx context.Context, user, key, value string, ballot *desc.Ballot) (bool, error) {
	conn, err := grpc.NewClient(cl.Address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(32<<20), grpc.MaxCallSendMsgSize(32<<20)))
	log := logger.GetLogger(ctx)
	if err != nil {
		log.Error(ctx, "grpc Client Dial Failed")
		return false, err
	}
	defer conn.Close()

	c := desc.NewNodeServiceClient(conn)

	timeout, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	req := desc.AppendValueRequest{
		Ballot: ballot,
		User:   user,
		Key:    key,
		Value:  value,
	}
	resp, err := c.AppendValue(timeout, &req)

	if err != nil {
		log.Error(ctx, "grpc Request Failed")
		return false, err
	}
	return resp.Success, nil
}

func (cl *Client) SubmitValue(ctx context.Context, user, key, value string) error {
	conn, err := grpc.NewClient(cl.Address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(32<<20), grpc.MaxCallSendMsgSize(32<<20)))
	if err != nil {
		return err
	}
	defer conn.Close()
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	response, err := desc.NewNodeServiceClient(conn).SubmitValue(timeout, &desc.AppendValueRequest{User: user, Key: key, Value: value})
	if err != nil {
		return err
	}
	if !response.Success {
		return status.Error(codes.Internal, "leader did not accept write")
	}
	return nil
}

func (cl *Client) Prepare(ctx context.Context, ballot *desc.Ballot) (*desc.PrepareResponse, error) {
	conn, err := grpc.NewClient(cl.Address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(32<<20), grpc.MaxCallSendMsgSize(32<<20)))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return desc.NewNodeServiceClient(conn).Prepare(ctx, &desc.PrepareRequest{Ballot: ballot})
}

func (cl *Client) AnnounceLeader(ctx context.Context, ballot *desc.Ballot, address string, committed ...uint64) (*desc.AnnounceLeaderResponse, error) {
	conn, err := grpc.NewClient(cl.Address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(32<<20), grpc.MaxCallSendMsgSize(32<<20)))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	index := uint64(0)
	if len(committed) > 0 {
		index = committed[0]
	}
	return desc.NewNodeServiceClient(conn).AnnounceLeader(ctx, &desc.AnnounceLeaderRequest{Ballot: ballot, LeaderGrpcAddress: address, CommittedIndex: index})
}

func (cl *Client) StartElection(ctx context.Context) (*desc.StartElectionResponse, error) {
	conn, err := grpc.NewClient(cl.Address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(32<<20), grpc.MaxCallSendMsgSize(32<<20)))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	return desc.NewNodeServiceClient(conn).StartElection(ctx, &desc.StartElectionRequest{})
}

func (cl *Client) AcceptEntries(ctx context.Context, req *desc.AcceptEntriesRequest) (*desc.JournalResponse, error) {
	conn, err := grpc.NewClient(cl.Address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(32<<20), grpc.MaxCallSendMsgSize(32<<20)))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return desc.NewNodeServiceClient(conn).AcceptEntries(ctx, req)
}
func (cl *Client) CommitEntries(ctx context.Context, req *desc.CommitEntriesRequest) (*desc.JournalResponse, error) {
	conn, err := grpc.NewClient(cl.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return desc.NewNodeServiceClient(conn).CommitEntries(ctx, req)
}
