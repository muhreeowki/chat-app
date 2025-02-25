package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/muhreeowki/mchat/templates"
)

type Client struct {
	id       string
	username string
	conn     *websocket.Conn
	manager  *ClientManager
	send     chan []byte
}

func NewClient(usrname string, conn *websocket.Conn, manager *ClientManager) *Client {
	return &Client{
		id:       uuid.NewString(),
		username: usrname,
		conn:     conn,
		manager:  manager,
		send:     make(chan []byte),
	}
}

func (c *Client) read() error {
	defer func() {
		c.manager.unregisterClient <- c
		c.conn.Close()
	}()
	for {
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			c.manager.unregisterClient <- c
			c.conn.Close()
			return err
		}
		msg := &templates.Message{}
		if err := json.NewDecoder(bytes.NewReader(msgBytes)).Decode(msg); err != nil {
			return err
		}
		c.manager.broadcast <- msg
		log.Printf("successfully read from connection (%s): %s", c.conn.RemoteAddr(), msg)
	}
}

func (c *Client) write() error {
	defer func() {
		c.conn.Close()
	}()
	for msg := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return fmt.Errorf("failed to write message from [%s]: %s", c.conn.RemoteAddr(), err)
		}
		log.Printf("successfully wrote to connection (%s): %s", c.conn.RemoteAddr(), msg)
	}
	return nil
}

type ClientManager struct {
	clients          map[*Client]bool
	broadcast        chan *templates.Message
	registerClient   chan *Client
	unregisterClient chan *Client
	store            Storage
	logger           *log.Logger
}

func NewClientManager(store Storage) *ClientManager {
	return &ClientManager{
		clients:          make(map[*Client]bool),
		broadcast:        make(chan *templates.Message),
		registerClient:   make(chan *Client),
		unregisterClient: make(chan *Client),
		store:            store,
		logger:           log.New(os.Stdout, "[client-manager] ", log.LstdFlags),
	}
}

func (manager *ClientManager) Start() {
	for {
		select {
		case client := <-manager.registerClient:
			manager.logger.Printf("/Socket [%s] connected.", client.conn.RemoteAddr())
			manager.clients[client] = true
			// msg, _ := json.Marshal(&templates.Message{
			// 	Payload: "A new socket connected.",
			// })
			// manager.Send(msg, client)

		case client := <-manager.unregisterClient:
			if _, ok := manager.clients[client]; ok {
				manager.logger.Printf("/Socket [%s] disconnected.", client.conn.RemoteAddr())
				close(client.send)
				delete(manager.clients, client)
				// msg, _ := json.Marshal(&templates.Message{
				// 	Payload: "A socket disconnected.",
				// })
				// manager.Send(msg, client)
			}

		case msg := <-manager.broadcast:
			buf := new(bytes.Buffer)
			templates.WsChatMessage(msg).Render(context.Background(), buf)
			for client := range manager.clients {
				select {
				case client.send <- buf.Bytes():
				default:
					close(client.send)
					delete(manager.clients, client)
				}
			}
			// Store the message
			err := manager.store.StoreMessage(msg)
			if err != nil {
				manager.logger.Printf("error storing message: %s", err)
			}
			manager.logger.Printf("broadcasted message: %s", msg)
		}
	}
}

func (manager *ClientManager) Send(msg []byte, from *Client) {
	for client := range manager.clients {
		if client != from {
			client.send <- msg
		}
	}
	manager.logger.Printf("broadcasted message: %s", msg)
}

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
	err := store.StoreMessage(&templates.Message{
		Sender:   "jake",
		Payload:  "hey guys im jake the human",
		Datetime: time.Now(),
	})
	if err != nil {
		logger.Println(err)
	}

	err = store.StoreMessage(&templates.Message{
		Sender:   "bob",
		Payload:  "hey jake im bob. The martian.",
		Datetime: time.Now().Add(time.Minute * 5),
	})
	if err != nil {
		logger.Println(err)
	}

	err = store.StoreMessage(&templates.Message{
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
	// TODO: USER AUTHENTICATION & AUTHORIZATION
	// tokenString := conn.Request().Header.Get("Sec-WebSocket-Protocol")
	// token, err := validateJWT(tokenString)
	// if err != nil {
	// 	s.logger.Printf("unauthenticated connection from: %+v\n", conn.RemoteAddr())
	// 	conn.Close()
	// 	return
	// }
	// _, ok := token.Claims.(*AuthClaims)
	// if !ok {
	// 	s.logger.Printf("claims convertion failed.")
	// 	conn.Close()
	// 	return
	// }
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
	msg := &templates.Message{
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
	if err := templates.ChatMessage(msg).Render(c.Request.Context(), c.Writer); err != nil {
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
	templates.WsChat(messages).Render(c.Request.Context(), c.Writer)
}
