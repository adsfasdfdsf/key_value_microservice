package grpc

import (
	"context"
	"fmt"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"

	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
	"golang.org/x/sync/errgroup"
)

type Multiclient struct {
	clients []Client
}

func NewMulticlient() *Multiclient {
	return &Multiclient{
		clients: make([]Client, 0),
	}
}

func (mc *Multiclient) AddClient(cl Client) {
	mc.clients = append(mc.clients, cl)
}

func (mc *Multiclient) AppendValue(ctx context.Context, user, key, value string, ballot *desc.Ballot) error {
	eg := errgroup.Group{}
	log := logger.GetLogger(ctx)
	for _, i := range mc.clients {
		eg.Go(func() error {
			success, err := i.AppendValue(ctx, user, key, value, ballot)
			if err != nil || success == false {
				log.Error(ctx, fmt.Sprintf("Append value via grpc on server %v failed", i.Address))
			}
			return err
		})
	}
	return eg.Wait()
}
