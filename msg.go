package main

import (
	"encoding/json"
	"fmt"
)

// Message is a standard format to be sent and received
// Command can be PROPOSE, PAY or SETTLE
// Assumes nodes are stateful and keep track honestly
// Can be expanded to have signatures
type Message struct {
	HostID string `json:"host"`
	PeerID string `json:"peer"`
	Type   string `json:"type"`
	Amount uint64 `json:"amt"`
}

func serialize(msg *Message) []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return b
}

func parseRawBytes(b []byte) Message {
	var m Message
	err := json.Unmarshal(b, &m)
	if err != nil {
		fmt.Println(err)
		return Message{}
	}
	return m
}
