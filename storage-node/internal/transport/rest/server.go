package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/interfaces"
	"io"
	"net/http"
	"time"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/models"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
)

type Server struct {
	port int
	serv *http.Server

	repo interfaces.Repository
	ctx  context.Context
}

func New(c context.Context, p int, repository interfaces.Repository) (*Server, error) {
	return &Server{
		port: p,
		serv: &http.Server{Addr: fmt.Sprintf(":%d", p), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, MaxHeaderBytes: 1 << 20},
		repo: repository,
		ctx:  c,
	}, nil
}

func (s *Server) Start(ctx context.Context,
	sendToReplicas func(func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request), health func(http.ResponseWriter, *http.Request), ready func() bool) error {
	log := logger.GetLogger(ctx)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/value/{key}", s.getValueByKey)
	mux.HandleFunc("GET /api/v1/values/{user}", s.getAllValues)
	mux.HandleFunc("POST /api/v1/addValue", sendToReplicas(s.addValue))
	mux.HandleFunc("PUT /api/v1/addValue", sendToReplicas(s.addValue))
	mux.HandleFunc("GET /health/live", health)
	mux.HandleFunc("GET /health/ready", health)
	mux.HandleFunc("GET /api/v1/status", health)
	s.serv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && (r.URL.Path != "/health/live" && r.URL.Path != "/health/ready" && r.URL.Path != "/api/v1/status") && !ready() {
			http.Error(w, "node recovering", http.StatusServiceUnavailable)
			return
		}
		mux.ServeHTTP(w, r)
	})

	log.Info(ctx, fmt.Sprintf("Server Started on %d", s.port))
	err := s.serv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) getValueByKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	user := r.URL.Query().Get("user")
	if user == "" || key == "" {
		http.Error(w, "user and key are required", http.StatusBadRequest)
		return
	}
	log := logger.GetLogger(s.ctx)
	log.Info(s.ctx, fmt.Sprintf("new Get value by key request %v", string(key)))
	value, err := s.repo.Get(r.Context(), user, key)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, interfaces.ErrNotFound) {
			status = http.StatusNotFound
		}
		w.WriteHeader(status)
		log.Warn(s.ctx, "response repo error")
		w.Write([]byte(http.StatusText(status)))
		return
	}
	resp := models.RestMessage{
		User:  user,
		Key:   key,
		Value: value,
	}
	encodedresp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		log.Error(s.ctx, "response json error")
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(encodedresp)
}

func (s *Server) addValue(w http.ResponseWriter, r *http.Request) {
	request, err := io.ReadAll(r.Body)
	log := logger.GetLogger(s.ctx)
	log.Info(s.ctx, fmt.Sprintf("new request %v", string(request)))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Warn(s.ctx, "response failed")
		w.Write([]byte("Wrong Request Format"))
		return
	}
	var restRequest models.RestMessage
	err = json.Unmarshal(request, &restRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Warn(s.ctx, "response failed")
		w.Write([]byte("Wrong Request Format"))
		return
	}
	if restRequest.User == "" || restRequest.Key == "" {
		http.Error(w, "user and key are required", http.StatusBadRequest)
		return
	}
	err = s.repo.Add(r.Context(), restRequest.User, restRequest.Key, restRequest.Value)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		log.Error(s.ctx, "response failed")
		w.Write([]byte("Operation Failed"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getAllValues(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	if user == "" {
		http.Error(w, "user is required", http.StatusBadRequest)
		return
	}
	log := logger.GetLogger(s.ctx)
	log.Info(s.ctx, fmt.Sprintf("new Get all values request %v", string(user)))
	values, err := s.repo.GetAll(r.Context(), user)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, interfaces.ErrNotFound) {
			status = http.StatusNotFound
		}
		w.WriteHeader(status)
		log.Warn(s.ctx, "response repo error")
		w.Write([]byte(http.StatusText(status)))
		return
	}
	resp := models.RestAllValuesMessage{
		User:   user,
		Values: values,
	}
	encodedresp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		log.Error(s.ctx, "response json error")
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(encodedresp)
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.serv.Shutdown(ctx)
	if err != nil {
		_ = s.serv.Close()
	}
	return err
}
