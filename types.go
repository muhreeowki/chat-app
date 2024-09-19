package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type User struct {
	Username string
	Password string
	Id       int
}

type Message struct {
	Datetime time.Time
	Payload  string
	Sender   string
}

func UnmarshalMessage(data []byte) (*Message, error) {
	msg := new(Message)
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func WriteMessage(w io.Writer, msg *Message) (err error) {
	if msg != nil {
		_, err = fmt.Fprintf(w, "[%s] %s: %s", msg.Datetime, msg.Sender, msg.Payload)
	} else {
		_, err = fmt.Fprintf(w, "error occured, try again.")
	}
	return err
}
