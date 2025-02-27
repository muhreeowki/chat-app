package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/muhreeowki/mchat/templates"
)

type Storage interface {
	StoreMessage(*templates.Message) error
	GetMessages() ([]*templates.Message, error)
	CreateUser(*User) error
	GetUser(string) (*User, error)
	GetUsers() ([]*User, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := os.Getenv("DB_CONN_STR")
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
	if err := s.initUsersTable(); err != nil {
		log.Println("initialized users table")
		return fmt.Errorf("users table init error: %s", err)
	}
	if err := s.initMessagesTable(); err != nil {
		log.Println("initialized messages table")
		return fmt.Errorf("message table init error: %s", err)
	}
	return nil
}

func (s *PostgresStore) initUsersTable() error {
	createUserTableQuery := `CREATE TABLE IF NOT EXISTS users (
    id SERIAL NOT NULL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    pass TEXT NOT NULL
  )`
	_, err := s.db.Exec(createUserTableQuery)
	return err
}

func (s *PostgresStore) initMessagesTable() error {
	createMessageTableQuery := `CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    payload TEXT NOT NULL,
    sender TEXT NOT NULL,
    recipient TEXT,
    datetime TIMESTAMP DEFAULT NOW()
  )`
	_, err := s.db.Exec(createMessageTableQuery)
	return err
}

func (s *PostgresStore) CreateUser(usr *User) error {
	query := `INSERT INTO users (username, pass) VALUES ($1, $2) RETURNING id`
	row := s.db.QueryRow(query, usr.Username, usr.Password)
	if err := row.Scan(&usr.Id); err != nil {
		return fmt.Errorf("failed to create user")
	}
	return nil
}

func (s *PostgresStore) GetUser(username string) (*User, error) {
	query := `SELECT id, username, pass FROM users WHERE username=$1`
	row := s.db.QueryRow(query, username)
	usr := new(User)
	if err := row.Scan(&usr.Id, &usr.Username, &usr.Password); err != nil {
		return nil, fmt.Errorf("failed to get user")
	}
	return usr, nil
}

func (s *PostgresStore) GetUsers() ([]*User, error) {
	query := `SELECT id, username FROM users`
	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("get users error: %s\n", err.Error())
		return nil, fmt.Errorf("failed to get users")
	}
	usrs := []*User{}
	for rows.Next() {
		usr := new(User)
		if err := rows.Scan(&usr.Id, &usr.Username); err != nil {
			log.Printf("get users error: %s\n", err)
			continue
		}
		usrs = append(usrs, usr)
	}
	return usrs, nil
}

func (s *PostgresStore) StoreMessage(msg *templates.Message) error {
	query := `INSERT INTO messages (payload, sender, recipient, datetime) VALUES ($1, $2, $3, $4) RETURNING payload, sender, recipient, datetime`
	row := s.db.QueryRow(query, msg.Payload, msg.Sender, msg.Recipient, msg.Datetime)
	respMsg := new(templates.Message)
	if err := row.Scan(&respMsg.Payload, &respMsg.Sender, &respMsg.Recipient, &respMsg.Datetime); err != nil {
		return fmt.Errorf("failed to create new message: %s", err.Error())
	}
	return nil
}

func (s *PostgresStore) GetMessages() ([]*templates.Message, error) {
	query := `SELECT payload, sender, datetime FROM messages`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages")
	}
	messages := []*templates.Message{}
	for rows.Next() {
		msg := new(templates.Message)
		if err := rows.Scan(&msg.Payload, &msg.Sender, &msg.Datetime); err != nil {
			fmt.Printf("get messages error: %s\n", err)
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (s *PostgresStore) Drop() {
	query := `DROP TABLE IF EXISTS messages, users`
	_, err := s.db.Exec(query)
	if err != nil {
		log.Printf("db drop error: %s\n", err)
	}
	log.Println("dropped message and user tables")
}
