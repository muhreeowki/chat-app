package main

type Storage interface {
	GetMessages()
	CreateMessage()
}

type Store struct{}
