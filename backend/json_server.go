package main

import (
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"
)

type JSONServer struct {
	logger     *log.Logger
	store      Storage
	listenAddr string
}

func NewJSONServer(listenAddr string, store Storage) *JSONServer {
	return &JSONServer{
		logger:     log.New(os.Stdout, "[json-server] ", log.LstdFlags),
		store:      store,
		listenAddr: listenAddr,
	}
}

func (s *JSONServer) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /messages", s.withJWTAuth(s.makeHttpHandler(s.HandleGetMessages)))
	mux.HandleFunc("GET /users", s.withJWTAuth(s.makeHttpHandler(s.HandleGetUsers)))
	mux.HandleFunc("POST /signup", s.makeHttpHandler(s.HandleSignUp))
	mux.HandleFunc("POST /login", s.makeHttpHandler(s.HandleLogin))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://chatclient:3000"},
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	handler := c.Handler(mux)

	s.logger.Printf("Mchat JSON server is live on: %s\n", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, handler)
}

func (s *JSONServer) HandleGetMessages(w http.ResponseWriter, r *http.Request) *JSONServerError {
	messages, err := s.store.GetMessages()
	if err != nil {
		return &JSONServerError{
			code:  500,
			error: err.Error(),
		}
	}
	WriteJSON(w, http.StatusOK, messages)
	s.logger.Printf("retrieved %d user messages", len(messages))
	return nil
}

func (s *JSONServer) HandleSignUp(w http.ResponseWriter, r *http.Request) *JSONServerError {
	reqData, err := UnmarshalUserJSON(r)
	if err != nil {
		s.logger.Printf("user unmarshal error: %s\n", err)
		return &JSONServerError{
			code:  http.StatusBadRequest,
			error: "invalid request data",
		}
	}
	hashedPass, err := HashPassword(reqData.Password)
	if err != nil {
		return &JSONServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}
	reqData.Password = hashedPass
	// Create User
	if err := s.store.CreateUser(reqData); err != nil {
		return &JSONServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}
	usr := &UserJSONResponse{
		Id:       reqData.Id,
		Username: reqData.Username,
	}
	usr.Token, err = createJWT(reqData)
	if err != nil {
		return &JSONServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}
	// Return the user object and the jwt token
	s.logger.Printf("created user with username: %s\n", usr.Username)
	WriteJSON(w, http.StatusCreated, usr)
	return nil
}

func (s *JSONServer) HandleLogin(w http.ResponseWriter, r *http.Request) *JSONServerError {
	reqData, err := UnmarshalUserJSON(r)
	if err != nil {
		s.logger.Printf("user unmarshal error: %s\n", err)
		return &JSONServerError{
			code:  500,
			error: "invalid request data",
		}
	}
	// Check if user exists
	dbUsr, err := s.store.GetUser(reqData.Username)
	if err != nil {
		s.logger.Printf("user unmarshal error: %s\n", err)
		return &JSONServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}
	// Validate Password
	valid := VerifyPassword(reqData.Password, dbUsr.Password)
	if !valid {
		return &JSONServerError{
			code:  http.StatusUnauthorized,
			error: "invalid credentials",
		}
	}

	usr := &UserJSONResponse{
		Username: dbUsr.Username,
		Id:       dbUsr.Id,
	}
	usr.Token, err = createJWT(dbUsr)
	if err != nil {
		s.logger.Printf("jwt create err: %s\n", err)
		return &JSONServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}

	s.logger.Printf("succesfully logged in: %s\n", usr.Username)
	WriteJSON(w, http.StatusOK, usr)
	return nil
}

func (s *JSONServer) HandleGetUsers(w http.ResponseWriter, r *http.Request) *JSONServerError {
	usrs, err := s.store.GetUsers()
	if err != nil {
		s.logger.Printf("get users err: %s\n", err)
		return &JSONServerError{
			code:  500,
			error: err.Error(),
		}
	}
	WriteJSON(w, http.StatusOK, usrs)
	return nil
}

func (s *JSONServer) makeHttpHandler(serverFunc JSONServerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := serverFunc(w, r); err != nil {
			WriteJSON(w, err.code, err.Error())
			s.logger.Printf("handler err: %s\n", err)
			return
		}
	}
}

func (s *JSONServer) withJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		_, err := validateJWT(tokenString)
		if err != nil {
			s.logger.Printf("jwt auth err: %s\n", err)
			WriteJSON(w, http.StatusUnauthorized, JSONServerError{
				code:  http.StatusUnauthorized,
				error: "invalid token",
			})
			return
		}

		handlerFunc(w, r)
	}
}
