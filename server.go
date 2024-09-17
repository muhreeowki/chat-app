package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type ChatServer struct {
	listenAddr string
	conns      map[net.Conn]bool
}

func NewChatServer(listenAddr string) *ChatServer {
	return &ChatServer{
		listenAddr: listenAddr,
		conns:      make(map[net.Conn]bool),
	}
}

func (s *ChatServer) Run() error {
	http.Handle("/", websocket.Handler(s.HandleWSConn))
	fmt.Printf("Mchat running on %s\n", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, nil)
}

func (s *ChatServer) HandleWSConn(conn *websocket.Conn) {
	fmt.Printf("incomming connection from: %+v\n", conn.RemoteAddr())

	s.conns[conn] = true

	s.ReadLoop(conn)
}

func (s *ChatServer) ReadLoop(conn *websocket.Conn) {
	var err error
	defer func() {
		fmt.Printf("dropping connection from: %+v, err: %+v\n", conn.RemoteAddr(), err)
		conn.Close()
		delete(s.conns, conn)
	}()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Printf("read error: %s\n", err)
			return
		}

		msg, err := UnmarshalMessage(buf[:n])
		if err != nil {
			fmt.Printf("unmarshal error: %s\n", err)
			return
		}
		msg.Datetime = time.Now()

		fmt.Printf("message: %+v\n", msg)
		s.broadcast(msg)
	}
}

func (s *ChatServer) broadcast(msg *Message) {
	for conn := range s.conns {
		go func(conn net.Conn, msg *Message) {
			if _, err := fmt.Fprintf(conn, "%s | %s: %s", msg.Datetime, msg.From, msg.Payload); err != nil {
				fmt.Printf("broadcast write error: %s\n", err)
				return
			}
		}(conn, msg)
	}
}
