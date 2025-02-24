package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
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
	buf := make([]byte, 1024)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			c.manager.unregisterClient <- c
			c.conn.Close()
			return err
		}
		msg := &Message{Sender: c.username, Payload: string(buf[:n])}
		c.manager.broadcast <- msg
		log.Printf("successfully read from connection (%s): %s", c.conn.RemoteAddr(), msg)
	}
}

func (c *Client) write() error {
	defer func() {
		c.conn.Close()
	}()
	for msg, ok := <-c.send; ok; {
		if !ok {
			err := c.conn.WriteClose(websocket.CloseFrame)
			if err != nil {
				return err
			}
			return fmt.Errorf("failed to write message from [%s]: %s", c.conn.RemoteAddr(), msg)
		}
		_, err := c.conn.Write(msg)
		if err != nil {
			return err
		}
		log.Printf("successfully wrote to connection (%s): %s", c.conn.RemoteAddr(), msg)
	}
	return nil
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
			manager.logger.Printf("/Socket [%s] connected.", client.conn.RemoteAddr())
			manager.clients[client] = true
			msg, _ := json.Marshal(&Message{
				Payload: "A new socket connected.",
			})
			manager.Send(msg, client)

		case client := <-manager.unregisterClient:
			if _, ok := manager.clients[client]; ok {
				manager.logger.Printf("/Socket [%s] disconnected.", client.conn.RemoteAddr())
				// close(client.send)
				// delete(manager.clients, client)
				msg, _ := json.Marshal(&Message{
					Payload: "A socket disconnected.",
				})
				manager.Send(msg, client)
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

	go func(client *Client) {
		err := client.read()
		log.Printf("read err: %s", err)
	}(client)
	go func(client *Client) {
		err := client.write()
		log.Printf("write err: %s", err)
	}(client)
}
