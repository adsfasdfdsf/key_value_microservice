package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"io"
	"fmt"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/models"
)

type Repository interface{
	Add(key, value string) error
	Get(key string) (string, error)
}

type Server struct {
	port int
	mux *http.ServeMux
	repo Repository
	ctx context.Context
}

func New(c context.Context, p int, repository Repository) (*Server, error){
	return &Server{
		port: p,
		mux: http.NewServeMux(),
		repo: repository,
		ctx: c,
	}, nil
}

func (s *Server) Start() error {
	s.mux.HandleFunc("GET /api/v1/value/{key}", s.getValueByKey)
	s.mux.HandleFunc("POST /api/v1/addValue", s.addValue)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.mux)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) getValueByKey(w http.ResponseWriter, r *http.Request){
	key := r.PathValue("id")
	value, ok := s.repo.Get(key)
	if ok != nil {
		w.WriteHeader(http.StatusNotFound)
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
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(encodedresp)
}

func (s *Server) addValue(w http.ResponseWriter, r *http.Request){
	request, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong Request Format"))
		return
	}
	var restRequest models.RestMessage
	err = json.Unmarshal(request, restRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong Request Format"))
		return
	}
	err = s.repo.Add(restRequest.Key, restRequest.Value)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Operation Failed"))
		return
	}
	w.WriteHeader(http.StatusOK)
}