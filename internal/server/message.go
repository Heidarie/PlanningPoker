package server

import "encoding/json"

type Message struct {
	Type    string          `json:"type"`
	Sender  string          `json:"sender,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}
