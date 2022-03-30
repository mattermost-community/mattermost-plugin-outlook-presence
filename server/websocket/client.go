package websocket

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Pool *Pool
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		messageType, content, err := c.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if err = c.Conn.WriteMessage(messageType, content); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}
