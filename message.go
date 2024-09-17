package main

import (
	"encoding/json"
	"time"
)

type Message struct {
	Datetime time.Time
	Payload  string
	From     string
}

func UnmarshalMessage(data []byte) (*Message, error) {
	msg := new(Message)
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}
