package main

import (
	"log"
	"sync"

	_ "github.com/lib/pq"
)

func main() {
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

	chatServer := NewChatServer(":3000", store)
	restServer := NewJSONRESTServer(":8080", store)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		if err := chatServer.Run(); err != nil {
			log.Fatalf("failed to run chatServer: %s\n", err)
		}
		wg.Done()
	}()

	go func() {
		if err := restServer.Run(); err != nil {
			log.Fatalf("failed to run restServer: %s\n", err)
		}
		wg.Done()
	}()

	wg.Wait()
}
