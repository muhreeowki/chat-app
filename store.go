package main

import (
	"database/sql"
	"fmt"
)

type Storage interface {
	GetMessages() ([]*Message, error)
	CreateMessage(*Message) error
	GetUser(string) (*User, error)
	GetUsers() ([]*UserJSONResponse, error)
	CreateUser(*User) error
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
	if err := s.initUsersTable(); err != nil {
		return fmt.Errorf("users table init error: %s", err)
	}
	if err := s.initMessagesTable(); err != nil {
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
    id SERIAL NOT NULL PRIMARY KEY,
    payload TEXT NOT NULL,
    sender TEXT NOT NULL,
    datetime TIMESTAMP DEFAULT NOW()
  )`
	_, err := s.db.Exec(createMessageTableQuery)
	return err
}

func (s *PostgresStore) CreateUser(usr *User) error {
	query := `INSERT INTO users (username, pass) VALUES ($1, $2) RETURNING id`
	row := s.db.QueryRow(query, usr.Username, usr.Password)
	fmt.Println("usr pass:", usr.Password)
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

func (s *PostgresStore) GetUsers() ([]*UserJSONResponse, error) {
	query := `SELECT id, username FROM users`
	rows, err := s.db.Query(query)
	if err != nil {
		fmt.Printf("GetUsers error: %s\n", err.Error())
		return nil, fmt.Errorf("failed to get users")
	}
	usrs := []*UserJSONResponse{}
	for rows.Next() {
		usr := new(UserJSONResponse)
		if err := rows.Scan(&usr.Id, &usr.Username); err != nil {
			fmt.Printf("get messages error: %s\n", err)
			continue
		}
		usrs = append(usrs, usr)
	}
	return usrs, nil
}

func (s *PostgresStore) CreateMessage(msg *Message) error {
	query := `INSERT INTO messages (payload, sender, datetime) VALUES ($1, $2, $3) RETURNING payload, sender, datetime`
	row := s.db.QueryRow(query, msg.Payload, msg.Sender, msg.Datetime)
	respMsg := new(Message)
	if err := row.Scan(&respMsg.Payload, &respMsg.Sender, &respMsg.Datetime); err != nil {
		return fmt.Errorf("failed to creat new message")
	}
	return nil
}

func (s *PostgresStore) GetMessages() ([]*Message, error) {
	query := `SELECT payload, sender, datetime FROM messages`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages")
	}
	messages := []*Message{}
	for rows.Next() {
		msg := new(Message)
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
		fmt.Printf("db drop error: %s\n", err)
	}
}
