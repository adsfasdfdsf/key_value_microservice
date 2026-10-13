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

func (cl *Client) AppendValue(ctx context.Context, user, key, value string) (bool, error) {
	conn, err := grpc.NewClient(cl.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		User:  user,
		Key:   key,
		Value: value,
	}
	resp, err := c.AppendValue(timeout, &req)

	if err != nil {
		log.Error(ctx, "grpc Request Failed")
		return false, err
	}
	return resp.Success, nil
}

func (cl *Client) SubmitValue(ctx context.Context, user, key, value string) error {
	conn, err := grpc.NewClient(cl.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
