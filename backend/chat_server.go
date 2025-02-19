package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

type Message struct {
	Sender    string    `json:"sender,omitempty"`
	Recipient string    `json:"recipient,omitempty"`
	Payload   string    `json:"payload,omitempty"`
	Datetime  time.Time `json:"datetime,omitempty"`
}

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
		c.conn.Close()
		c.manager.unregisterClient <- c
	}()
	for {
		buf := make([]byte, 1024)
		n, err := c.conn.Read(buf)
		if err != nil {
			return err
		}
		msg := &Message{Sender: c.username, Payload: string(buf[:n])}
		c.manager.broadcast <- msg
	}
}

func (c *Client) write() error {
	defer func() {
		c.conn.Close()
		c.manager.unregisterClient <- c
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				_, err := c.conn.Write(msg)
				if err != nil {
					return err
				}
				return fmt.Errorf("failed to write message from [%s]: %s", c.conn.RemoteAddr(), msg)
			}
			_, err := c.conn.Write(msg)
			if err != nil {
				return err
			}
		}
	}
}

type ClientManager struct {
	clients          map[*Client]bool
	broadcast        chan *Message
	registerClient   chan *Client
	unregisterClient chan *Client
	store            Storage
	logger           *log.Logger
}

func NewClientManager(store Storage) *ClientManager {
	return &ClientManager{
		clients:          make(map[*Client]bool),
		broadcast:        make(chan *Message),
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
			manager.clients[client] = true
			msg, _ := json.Marshal(&Message{Payload: fmt.Sprintf("/A new socket [%s] has connected.", client.conn.RemoteAddr())})
			manager.Send(msg, client)
			manager.logger.Printf("/A new socket [%s] has connected.", client.conn.RemoteAddr())
		case client := <-manager.unregisterClient:
			if _, ok := manager.clients[client]; ok {
				close(client.send)
				delete(manager.clients, client)
				msg, _ := json.Marshal(&Message{Payload: fmt.Sprintf("/A socket [%s] has disconnected.", client.conn.RemoteAddr())})
				manager.Send(msg, client)
				manager.logger.Printf("/A socket [%s] has disconnected.", client.conn.RemoteAddr())
			}
		case msg := <-manager.broadcast:
			jsonMsg, _ := json.Marshal(msg)
			for client := range manager.clients {
				select {
				case client.send <- jsonMsg:
				default:
					close(client.send)
					delete(manager.clients, client)
				}
			}
			// Store the message
			manager.store.StoreMessage(msg)
			manager.logger.Printf("/A new message was broadcasted: %s", msg)
		}
	}
}

func (manager *ClientManager) Send(msg []byte, from *Client) {
	for client := range manager.clients {
		if client != from {
			client.send <- msg
		}
	}
	manager.logger.Printf("/A new message was broadcasted: %s", msg)
}

type ChatServer struct {
	listenAddr    string
	clientManager *ClientManager
	logger        *log.Logger
}

func NewChatServer(listenAddr string, store Storage) *ChatServer {
	return &ChatServer{
		clientManager: NewClientManager(store),
		listenAddr:    listenAddr,
		logger:        log.New(os.Stdout, "[chat-server] ", log.LstdFlags),
	}
}

func (s *ChatServer) Run() error {
	router := http.NewServeMux()
	router.Handle("/", websocket.Handler(s.HandleWSConn))

	go s.clientManager.Start()

	s.logger.Printf("Mchat CHAT server is live on: %s\n", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, router)
}

func (s *ChatServer) HandleWSConn(conn *websocket.Conn) {
	// Check if the connection is authenticated
	tokenString := conn.Request().Header.Get("Sec-WebSocket-Protocol")
	token, err := validateJWT(tokenString)
	_, ok := token.Claims.(AuthClaims)
	if err != nil || !ok {
		s.logger.Printf("unauthenticated connection from: %+v\n", conn.RemoteAddr())
		conn.Close()
		return
	}
	client := NewClient(token.Claims.(AuthClaims).Username, conn, s.clientManager)
	s.clientManager.registerClient <- client

	go client.read()
	go client.write()
}
