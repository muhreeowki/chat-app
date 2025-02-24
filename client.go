package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ClientServer struct {
	listenAddr    string
	store         Storage
	clientManager *ClientManager
	logger        *log.Logger
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
		listenAddr:    listenAddr,
		store:         store,
		clientManager: NewClientManager(store),
		logger:        logger,
	}
}

func (s *ClientServer) Run() error {
	r := gin.Default()

	r.GET("/", s.HandleHome)
	r.GET("/chatroom", s.HandleWSConn)
	r.POST("/messages", s.HandleHome)
	r.Static("/assets", "./assets/")

	// mux := http.NewServeMux()
	// mux.HandleFunc("GET /chat", s.HandleHome)
	// mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("assets"))))
	// mux.HandleFunc("/chatroom", s.HandleWSConn)
	// mux.HandleFunc("POST /messages", s.HandlePostMessages)

	go s.clientManager.Start()

	s.logger.Printf("Mchat Client server is live on: %s\n", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, r)
}

// func (s *ClientServer) HandleWSConn(w http.ResponseWriter, r *http.Request) {
func (s *ClientServer) HandleWSConn(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Printf("failed to establish connection.")
		return
	}
	// tokenString := conn.Request().Header.Get("Sec-WebSocket-Protocol")
	// Check if the connection is authenticated
	// token, err := validateJWT(tokenString)
	// if err != nil {
	// 	s.logger.Printf("unauthenticated connection from: %+v\n", conn.RemoteAddr())
	// 	conn.Close()
	// 	return
	// }
	// _, ok := token.Claims.(*AuthClaims)
	// if !ok {
	// 	s.logger.Printf("claims convertion failed.")
	// }
	// Add the client to clientManager and start the read and write functions
	// client := NewClient(token.Claims.(*AuthClaims).Username, conn, s.clientManager)
	client := NewClient("user", conn, s.clientManager)
	s.logger.Printf("New Connection: %+v", client)

	s.clientManager.registerClient <- client

	go func(client *Client) {
		err := client.read()
		log.Printf("read err: %s", err)
	}(client)
	go func(client *Client) {
		err := client.write()
		log.Printf("write err: %s", err)
	}(client)
}

// func (s *ClientServer) HandlePostMessages(w http.ResponseWriter, r *http.Request) {
func (s *ClientServer) HandlePostMessages(c *gin.Context) {
	err := c.Request.ParseForm()
	if err != nil {
		s.logger.Println(err)
	}
	msg := &Message{
		Payload: c.Request.Form.Get("payload"),
		Sender:  c.Request.Form.Get("sender"),
	}
	if msg.Payload == "" {
		// w.Write([]byte("empty message"))
		c.Error(fmt.Errorf("empty message"))
		return
	}
	if err := s.store.StoreMessage(msg); err != nil {
		s.logger.Println(err)
	}
	if err := ChatMessage(msg).Render(c.Request.Context(), c.Writer); err != nil {
		s.logger.Println(err)
	}
	s.logger.Printf("New MSG: %+v", msg)
}

// func (s *ClientServer) HandleHome(w http.ResponseWriter, r *http.Request) {
func (s *ClientServer) HandleHome(c *gin.Context) {
	messages, err := s.store.GetMessages()
	if err != nil {
		s.logger.Println(err)
	}
	WsChat(messages).Render(c.Request.Context(), c.Writer)
}
