package main

import (
	"fmt"
	"io"
	"net"
	"net/http"

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
		fmt.Printf("dropping connection from: %+v, err: %+v\n", conn, err)
		conn.Close()
		s.conns[conn] = false
	}()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Printf("read error: %s\n", err)
			continue
		}

		msg := buf[:n]
		fmt.Printf("message: %s\n", string(msg))

		s.broadcast(msg)
	}
}

func (s *ChatServer) broadcast(msg []byte) {
	for conn, connected := range s.conns {
		if connected {
			go func(conn net.Conn, msg []byte) {
				if _, err := conn.Write(msg); err != nil {
					fmt.Printf("read error: %s\n", err)
					return
				}
			}(conn, msg)
		}
	}
}
