package main

import (
	"log"
	"sync"

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

	wg := new(sync.WaitGroup)
	wg.Add(2)

	chatServer := NewChatServer(":4000", store)
	go func() {
		if err := chatServer.Run(); err != nil {
			log.Fatalf("failed to run chatServer: %s\n", err)
		}
		wg.Done()
	}()

	clientServer := NewClientServer(":3000", store)
	go func() {
		if err := clientServer.Run(); err != nil {
			log.Fatalf("failed to run clientServer: %s\n", err)
		}
		wg.Done()
	}()
	wg.Wait()
}
