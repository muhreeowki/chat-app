package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type ClientServer struct {
	logger     *log.Logger
	store      Storage
	listenAddr string
}

func NewClientServer(listenAddr string, store Storage) *ClientServer {
	logger := log.New(os.Stdout, "[client-server] ", log.LstdFlags)
	err := store.StoreMessage(&Message{
		Sender:   "jake",
		Payload:  "hey guys im jake the human",
		Datetime: time.Now(),
	})
	if err != nil {
		logger.Println(err)
	}

	err = store.StoreMessage(&Message{
		Sender:   "bob",
		Payload:  "hey jake im bob. The martian.",
		Datetime: time.Now().Add(time.Minute * 5),
	})
	if err != nil {
		logger.Println(err)
	}

	err = store.StoreMessage(&Message{
		Sender:   "jake",
		Payload:  "cool! nice to meet you bob. Wanna play fortnite?",
		Datetime: time.Now().Add(time.Minute * 10),
	})
	if err != nil {
		logger.Println(err)
	}

	return &ClientServer{
		logger:     logger,
		store:      store,
		listenAddr: listenAddr,
	}
}

func (s *ClientServer) Run() error {
	r := gin.Default()
	r.GET("/chatroom", s.HandleHome)
	r.Static("/assets/", "./assets/")
	r.POST("/messages", s.HandlePostMessages)

	s.logger.Printf("Mchat Client server is live on: %s\n", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, r)
}

func (s *ClientServer) HandlePostMessages(c *gin.Context) {
	//	msg := make(map[string]string)
	err := c.Request.ParseForm()
	if err != nil {
		s.logger.Println(err)
	}
	msg := &Message{
		Payload: c.Request.Form.Get("payload"),
		Sender:  c.Request.Form.Get("sender"),
	}
	if msg.Payload == "" {
		c.Error(fmt.Errorf("empty message"))
		return
	}
	s.logger.Printf("New MSG: %+v", msg)
	if err := s.store.StoreMessage(msg); err != nil {
		s.logger.Println(err)
	}
	if err := ChatMessage(msg).Render(c.Request.Context(), c.Writer); err != nil {
		s.logger.Println(err)
	}
}

func (s *ClientServer) HandleHome(c *gin.Context) {
	messages, err := s.store.GetMessages()
	if err != nil {
		s.logger.Println(err)
	}
	Chat(messages).Render(c.Request.Context(), c.Writer)
}

func (s *ClientServer) HandleGetMessages(w http.ResponseWriter, r *http.Request) *ClientServerError {
	messages, err := s.store.GetMessages()
	if err != nil {
		return &ClientServerError{
			code:  500,
			error: err.Error(),
		}
	}
	WriteJSON(w, http.StatusOK, messages)
	s.logger.Printf("retrieved %d user messages", len(messages))
	return nil
}

func (s *ClientServer) HandleSignUp(w http.ResponseWriter, r *http.Request) *ClientServerError {
	reqData, err := UnmarshalUserJSON(r)
	if err != nil {
		s.logger.Printf("user unmarshal error: %s\n", err)
		return &ClientServerError{
			code:  http.StatusBadRequest,
			error: "invalid request data",
		}
	}
	hashedPass, err := HashPassword(reqData.Password)
	if err != nil {
		return &ClientServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}
	reqData.Password = hashedPass
	// Create User
	if err := s.store.CreateUser(reqData); err != nil {
		return &ClientServerError{
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
		return &ClientServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}
	// Return the user object and the jwt token
	s.logger.Printf("created user with username: %s\n", usr.Username)
	WriteJSON(w, http.StatusCreated, usr)
	return nil
}

func (s *ClientServer) HandleLogin(w http.ResponseWriter, r *http.Request) *ClientServerError {
	reqData, err := UnmarshalUserJSON(r)
	if err != nil {
		s.logger.Printf("user unmarshal error: %s\n", err)
		return &ClientServerError{
			code:  500,
			error: "invalid request data",
		}
	}
	// Check if user exists
	dbUsr, err := s.store.GetUser(reqData.Username)
	if err != nil {
		s.logger.Printf("user unmarshal error: %s\n", err)
		return &ClientServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}
	// Validate Password
	valid := VerifyPassword(reqData.Password, dbUsr.Password)
	if !valid {
		return &ClientServerError{
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
		return &ClientServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}

	s.logger.Printf("succesfully logged in: %s\n", usr.Username)
	WriteJSON(w, http.StatusOK, usr)
	return nil
}

func (s *ClientServer) HandleGetUsers(w http.ResponseWriter, r *http.Request) *ClientServerError {
	usrs, err := s.store.GetUsers()
	if err != nil {
		s.logger.Printf("get users err: %s\n", err)
		return &ClientServerError{
			code:  500,
			error: err.Error(),
		}
	}
	WriteJSON(w, http.StatusOK, usrs)
	return nil
}

func (s *ClientServer) makeHttpHandler(serverFunc ClientServerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := serverFunc(w, r); err != nil {
			WriteJSON(w, err.code, err.Error())
			s.logger.Printf("handler err: %s\n", err)
			return
		}
	}
}

func (s *ClientServer) withJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		_, err := validateJWT(tokenString)
		if err != nil {
			s.logger.Printf("jwt auth err: %s\n", err)
			WriteJSON(w, http.StatusUnauthorized, ClientServerError{
				code:  http.StatusUnauthorized,
				error: "invalid token",
			})
			return
		}

		handlerFunc(w, r)
	}
}
