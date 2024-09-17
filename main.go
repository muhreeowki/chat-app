package main

import (
	"log"

	_ "github.com/lib/pq"
)

func main() {
	listenAddr := ":3000"
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatalf("error occured: %+v\n", err)
	}

	// store.Drop()
	// return

	store.Init()
	if err != nil {
		log.Fatalf("error occured: %+v\n", err)
	}

	server := NewChatServer(listenAddr, store)

	if err := server.Run(); err != nil {
		log.Fatalf("error occured: %+v\n", err)
	}
}
