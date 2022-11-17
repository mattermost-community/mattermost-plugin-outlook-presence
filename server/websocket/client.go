package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

type Client struct {
	Conn   *websocket.Conn
	Hub    *Hub
	ConnID string
}

func (c *Client) Read(api plugin.API) {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, content, err := c.Conn.ReadMessage()
		if err != nil {
			api.LogDebug("error in reading the message received through the websocket.", "Error", err.Error())
			return
		}

		api.LogInfo("message received through the websocket.", "Message", string(content))
	}
}
