package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JSONRESTServer struct {
	store      Storage
	listenAddr string
}

func NewJSONRESTServer(listenAddr string, store Storage) *JSONRESTServer {
	return &JSONRESTServer{
		store:      store,
		listenAddr: listenAddr,
	}
}

func (s *JSONRESTServer) Run() error {
	router := http.NewServeMux()
	router.HandleFunc("GET /", s.HandleGetMessages)

	fmt.Printf("Mchat REST server is live on: %s\n", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, router)
}

func (s *JSONRESTServer) HandleGetMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := s.store.GetMessages()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
}
