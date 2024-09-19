package main

import (
	"log"
	"sync"

	_ "github.com/lib/pq"
)

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// store.Drop()
	// return

	err = store.Init()
	if err != nil {
		log.Fatalf(err.Error())
	}

	chatServer := NewChatServer(":4000", store)
	restServer := NewJSONServer(":8080", store)

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
