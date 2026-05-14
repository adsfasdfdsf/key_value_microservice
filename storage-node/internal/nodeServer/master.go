package nodeServer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/models"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/grpc"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/rest"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
)


type Master struct {
	ctx context.Context
	rest *rest.Server
	clients *grpc.Multiclient
}

func NewMaster(ctx context.Context, rest *rest.Server, clients *grpc.Multiclient) *Master{
	return &Master{
		ctx: ctx,
		rest: rest,
		clients: clients,
	}
}

func (m *Master) Start() error{
	return m.rest.Start(m.ctx, m.sendToReplicas)
}

func (m *Master) Stop() error{
	return m.rest.Stop()
}

func (m *Master) sendToReplicas(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request){
	return func(w http.ResponseWriter, r *http.Request){
		defer r.Body.Close()

		request, err := io.ReadAll(r.Body)
		log := logger.GetLogger(m.ctx)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error(m.ctx, "response failed")
			w.Write([]byte("Wrong Request Format"))
			return
		}

		var restRequest models.RestMessage
		err = json.Unmarshal(request, &restRequest)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error(m.ctx, "response failed")
			w.Write([]byte("Wrong Request Format"))
			return
		}

		err = m.clients.AppendValue(m.ctx, restRequest.Key, restRequest.Value)
		if err != nil {
			log.Error(m.ctx, "Failed to send request to replicas")
		}
		handler(w, r)
	}
}