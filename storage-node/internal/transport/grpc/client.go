package grpc

import (
	"context"
	"time"

	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct{
	Address string
}

func (cl *Client)AppendValue(ctx context.Context, key, value string) (bool, error){
	conn, err := grpc.NewClient(cl.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	log := logger.GetLogger(ctx)
	if err != nil {
		log.Error(ctx, "grpc Client Dial Failed")
		return false, err
	}
	defer conn.Close()

	c := desc.NewNodeServiceClient(conn)

	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := desc.AppendValueRequest{
		Key: key,
		Value: value,
	}
	resp, err := c.AppendValue(timeout, &req)
	
	if err != nil {
		log.Error(ctx, "grpc Request Failed")
		return false, err
	}
	return resp.Success, nil
}