package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ClientServerFunc func(w http.ResponseWriter, r *http.Request) *ClientServerError

type ClientServerError struct {
	error string
	code  int
}

func (e *ClientServerError) Error() string {
	return e.error
}

type User struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Password string `json:"password"`
	Id       int    `json:"id"`
}

type UserJSONResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	Id       int    `json:"id"`
}

type Message struct {
	Sender    string    `json:"sender,omitempty"`
	Recipient string    `json:"recipient,omitempty"`
	Payload   string    `json:"payload,omitempty"`
	Datetime  time.Time `json:"datetime,omitempty"`
}

func UnmarshalUserJSON(r *http.Request) (*User, error) {
	usr := new(User)
	if err := json.NewDecoder(r.Body).Decode(usr); err != nil {
		return nil, err
	}
	return usr, nil
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
