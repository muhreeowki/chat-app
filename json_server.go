package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JSONServer struct {
	store      Storage
	listenAddr string
}

func NewJSONServer(listenAddr string, store Storage) *JSONServer {
	return &JSONServer{
		store:      store,
		listenAddr: listenAddr,
	}
}

func (s *JSONServer) Run() error {
	router := http.NewServeMux()
	router.HandleFunc("GET /", withJWTAuth(makeHttpHandler(s.HandleGetMessages)))

	fmt.Printf("Mchat JSON server is live on: %s\n", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, router)
}

func (s *JSONServer) HandleGetMessages(w http.ResponseWriter, r *http.Request) *JSONServerError {
	messages, err := s.store.GetMessages()
	if err != nil {
		return &JSONServerError{
			code:  500,
			error: err.Error(),
		}
	}
	writeJSON(w, http.StatusOK, messages)
	return nil
}

func (s *JSONServer) HandleSignUp(w http.ResponseWriter, r *http.Request) *JSONServerError {
	return nil
}

type JSONServerFunc func(w http.ResponseWriter, r *http.Request) *JSONServerError

type JSONServerError struct {
	error string
	code  int
}

func (e *JSONServerError) Error() string {
	return e.error
}

func withJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	fmt.Println("checking JWT Token")

	return func(w http.ResponseWriter, r *http.Request) {
		handlerFunc(w, r)
	}
}

func makeHttpHandler(serverFunc JSONServerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := serverFunc(w, r); err != nil {
			writeJSON(w, err.code, err.Error())
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) error {
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "applictation/json")
	return json.NewEncoder(w).Encode(v)
}
