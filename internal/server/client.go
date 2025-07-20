package server

import "github.com/gorilla/websocket"

type Client struct {
	Conn   *websocket.Conn
	Send   chan Message
	Name   string
	Room   *Room
	IsHost bool
}

func (c *Client) readPump() {
	defer func() { c.Room.unregister <- c; c.Conn.Close() }()
	for {
		var msg Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}
		msg.Sender = c.Name
		c.Room.broadcast <- msg
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		c.Conn.WriteJSON(msg)
	}
}
