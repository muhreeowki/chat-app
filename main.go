package main

import "log"

func main() {
	listenAddr := ":3000"

	server := NewChatServer(listenAddr)
	if err := server.Run(); err != nil {
		log.Fatalf("error occured: %+v\n", err)
	}
}
