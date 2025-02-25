package main

import (
	"log"

	_ "github.com/lib/pq"
)

// TODO: ** Add the following features **
// 1. Chat Rooms that users can join and leave freely (kinda like house party)
// 2. Chat Commands to do various things within the chat room
// 3. Add CLI Client
// 3. User Authentication & Authorization
//    - Handle guests and limiting guest usage

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
