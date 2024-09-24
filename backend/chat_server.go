package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

type ChatServer struct {
	conns      map[net.Conn]bool
	logger     *log.Logger
	store      Storage
	listenAddr string
}

func NewChatServer(listenAddr string, store Storage) *ChatServer {
	return &ChatServer{
		conns:      make(map[net.Conn]bool),
		store:      store,
		listenAddr: listenAddr,
		logger:     log.New(os.Stdout, "[chat-server] ", log.LstdFlags),
	}
}

func (s *ChatServer) Run() error {
	router := http.NewServeMux()
	router.Handle("/", websocket.Handler(s.HandleWSConn))

	s.logger.Printf("Mchat CHAT server is live on: %s\n", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, router)
}

func (s *ChatServer) HandleWSConn(conn *websocket.Conn) {
	s.logger.Printf("incomming connection from: %+v\n", conn.RemoteAddr())

	// Check if the connection is authenticated
	tokenString := conn.Request().Header.Get("Sec-WebSocket-Protocol")
	_, err := validateJWT(tokenString)
	if err != nil {
		s.logger.Fatalf("unauthenticated connection from: %+v\n", conn.RemoteAddr())
		conn.Close()
		return
	}

	s.conns[conn] = true

	s.ReadLoop(conn)
}

func (s *ChatServer) ReadLoop(conn *websocket.Conn) {
	var err error
	defer func() {
		s.logger.Printf("dropping connection from: %+v, err: %+v\n", conn.RemoteAddr(), err)
		conn.Close()
		delete(s.conns, conn)
	}()

	buf := make([]byte, 1024)

	for {
		// Read the message from the client
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			s.logger.Printf("ws read error: %s\n", err)
			return
		}
		// Unmarshal the Message
		msg, err := UnmarshalMessage(buf[:n])
		if err != nil {
			s.logger.Printf("message unmarshal error: %s\n", err)
			return
		}
		// Set the message time
		msg.Datetime = time.Now().UTC().Truncate(time.Minute)
		// Create the message in the DB (go routine)
		if err := s.store.CreateMessage(msg); err != nil {
			WriteMessage(conn, nil)
		}
		// Broadcast the message the the other connected clients
		s.logger.Printf("new message from %s: %s\n", msg.Sender, msg.Payload)
		s.broadcast(msg)
	}
}

func (s *ChatServer) broadcast(msg *Message) {
	for conn := range s.conns {
		go func(conn net.Conn, msg *Message) {
			if err := WriteMessage(conn, msg); err != nil {
				s.logger.Printf("broadcast write error: %s\n", err)
				return
			}
		}(conn, msg)
	}
}
