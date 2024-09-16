package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mikhaylov/beebroker/internal/config"
	"mikhaylov/beebroker/internal/store"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type BrokerServer struct {
	store  store.Store
	config config.Config
	server *http.Server
	pool   *ConnPool
}

func NewBrokerServer(config config.Config, store store.Store) *BrokerServer {
	srv := &BrokerServer{
		config: config,
		store:  store,
		server: &http.Server{},
	}

	srv.server.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	srv.server.Handler = srv
	return srv
}

func (s BrokerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path, "/")

	queue := params[2]

	switch r.Method {
	case http.MethodPut:
		s.putMessage(w, r, queue)
	case http.MethodGet:
		s.getMessage(w, r, queue)
	}
}

func (s *BrokerServer) Open() error {
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}

	s.pool = NewPool(listener, 20)

	go s.server.Serve(s.pool)

	return nil
}

func (s *BrokerServer) Close(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = s.pool.Close()
	return err
}

func (s *BrokerServer) getMessage(w http.ResponseWriter, r *http.Request, queueName string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate;")
	w.Header().Set("pragma", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	message, err := s.store.Receive(queueName)
	if err != nil {
		if errors.Is(err, store.ErrTopicNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	timeoutParam := r.URL.Query().Get("timeout")
	timeout, err := strconv.Atoi(timeoutParam)
	if err != nil {
		timeout = s.config.Timeout
	}

	for {
		select {
		case msg := <-message:
			json.NewEncoder(w).Encode(msg)
			return
		case <-time.After(time.Duration(timeout) * time.Second):
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
}

func (s *BrokerServer) putMessage(w http.ResponseWriter, r *http.Request, queueName string) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)

	p := store.Payload{}
	err := dec.Decode(&p)
	if err != nil || p.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.store.Send(queueName, p)
	w.WriteHeader(http.StatusOK)
}
