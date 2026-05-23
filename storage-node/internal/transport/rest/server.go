package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/models"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
)

type Repository interface{
	Add(key, value string) error
	Get(key string) (string, error)
}

type Server struct {
	port int
	serv *http.Server

	repo Repository
	ctx context.Context
}

func New(c context.Context, p int, repository Repository) (*Server, error){
	return &Server{
		port: p,
		serv: nil,
		repo: repository,
		ctx: c,
	}, nil
}

func (s *Server) Start(ctx context.Context, 
	sendToReplicas func(func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request)) error {
	log := logger.GetLogger(ctx)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/value/{key}", s.getValueByKey)
	mux.HandleFunc("POST /api/v1/addValue", sendToReplicas(s.addValue))
	s.serv = &http.Server{
		Addr: 		fmt.Sprintf(":%d", s.port), 
		Handler: 	mux,
	}
	log.Info(ctx, fmt.Sprintf("Server Started on %d", s.port))
	err := s.serv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
        return err
    }
    return nil
}

func (s *Server) getValueByKey(w http.ResponseWriter, r *http.Request){
	key := r.PathValue("key")
	log := logger.GetLogger(s.ctx)
	log.Info(s.ctx, fmt.Sprintf("new Get value by key request %v", string(key)))
	value, err := s.repo.Get(key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Error(s.ctx, "response repo error")
		w.Write([]byte("Key not found"))
		return
	}
	resp := models.RestMessage{
		Key: key,
		Value: value,
	}
	encodedresp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		log.Error(s.ctx, "response json error")
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(encodedresp)
}

func (s *Server) addValue(w http.ResponseWriter, r *http.Request){
	request, err := io.ReadAll(r.Body)
	log := logger.GetLogger(s.ctx)
	log.Info(s.ctx, fmt.Sprintf("new request %v", string(request)))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Error(s.ctx, "response failed")
		w.Write([]byte("Wrong Request Format"))
		return
	}
	var restRequest models.RestMessage
	err = json.Unmarshal(request, &restRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Error(s.ctx, "response failed")
		w.Write([]byte("Wrong Request Format"))
		return
	}
	err = s.repo.Add(restRequest.Key, restRequest.Value)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		log.Error(s.ctx, "response failed")
		w.Write([]byte("Operation Failed"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) Stop() error {
	return s.serv.Shutdown(s.ctx)
}