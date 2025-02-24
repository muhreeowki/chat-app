package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
		msg := &Message{}
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
			// msg, _ := json.Marshal(&Message{
			// 	Payload: "A new socket connected.",
			// })
			// manager.Send(msg, client)

		case client := <-manager.unregisterClient:
			if _, ok := manager.clients[client]; ok {
				manager.logger.Printf("/Socket [%s] disconnected.", client.conn.RemoteAddr())
				close(client.send)
				delete(manager.clients, client)
				// msg, _ := json.Marshal(&Message{
				// 	Payload: "A socket disconnected.",
				// })
				// manager.Send(msg, client)
			}

		case msg := <-manager.broadcast:
			buf := new(bytes.Buffer)
			WsChatMessage(msg).Render(context.Background(), buf)
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
