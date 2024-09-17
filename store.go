package main

import (
	"database/sql"
	"fmt"
)

type Storage interface {
	GetMessages() ([]Message, error)
	CreateMessage(*Message) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=postgres password=mchat sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	query := `CREATE TABLE IF NOT EXISTS messages (
    id SERIAL NOT NULL PRIMARY KEY,
    payload TEXT NOT NULL,
    sender TEXT NOT NULL,
    datetime TIMESTAMP DEFAULT NOW() 
  )`
	_, err := s.db.Exec(query)
	if err != nil {
		fmt.Printf("db init error: %s\n", err)
		return fmt.Errorf("db init error")
	}
	return err
}

func (s *PostgresStore) CreateMessage(msg *Message) error {
	query := `INSERT INTO messages (payload, sender, datetime) VALUES ($1, $2, $3) RETURNING payload, sender, datetime`
	row := s.db.QueryRow(query, msg.Payload, msg.Sender, msg.Datetime)
	respMsg := new(Message)
	if err := row.Scan(&respMsg.Payload, &respMsg.Sender, &respMsg.Datetime); err != nil {
		fmt.Printf("create message error: %s\n", err)
		return fmt.Errorf("create message init error")
	}
	fmt.Printf("created message: %+v\n", respMsg)
	return nil
}

func (s *PostgresStore) GetMessages() ([]Message, error) {
	return []Message{}, nil
}

func (s *PostgresStore) Drop() {
	query := `DROP TABLE messages`
	_, err := s.db.Exec(query)
	if err != nil {
		fmt.Printf("db drop error: %s\n", err)
	}
}
