package main

import (
	"log"

	_ "github.com/lib/pq"
)

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err.Error())
	}
	// store.Drop()
	// return

	if err := store.Init(); err != nil {
		log.Fatal(err.Error())
	}

	clientServer := NewClientServer(":3000", store)
	if err := clientServer.Run(); err != nil {
		log.Fatal(err.Error())
	}
}
