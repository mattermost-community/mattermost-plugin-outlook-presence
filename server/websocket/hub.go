package websocket

import (
	"fmt"

	"github.com/Brightscout/mattermost-plugin-outlook-presence/server/serializer"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

type Hub struct {
	Register        chan *Client
	Unregister      chan *Client
	Clients         map[string]*Client
	Broadcast       chan *serializer.UserStatus
	ConnectionIndex int
}

func NewHub() *Hub {
	return &Hub{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan *serializer.UserStatus, 4096),
	}
}

func (h *Hub) Start(api plugin.API) {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.ConnID] = client
			api.LogInfo(fmt.Sprintf("Client added to hub %d. Size of connection pool: %d", h.ConnectionIndex, len(h.Clients)))
		case client := <-h.Unregister:
			delete(h.Clients, client.ConnID)
			api.LogInfo(fmt.Sprintf("Client removed from hub %d. Size of connection pool: %d", h.ConnectionIndex, len(h.Clients)))
		case statusChangedEvent := <-h.Broadcast:
			if len(h.Clients) == 0 {
				api.LogInfo(fmt.Sprintf("No clients connected to hub %d, status %s", h.ConnectionIndex, statusChangedEvent.Status))
				break
			}
			api.LogInfo(fmt.Sprintf("Sending message to all clients in pool of hub %d, status %s", h.ConnectionIndex, statusChangedEvent.Status))
			for _, client := range h.Clients {
				if err := client.Conn.WriteJSON(statusChangedEvent); err != nil {
					api.LogError("error in broadcasting the status changed event.", "Error", err.Error())
				}
			}
		}
	}
}

func (h *Hub) Stop() {
	close(h.Broadcast)
	close(h.Register)
	close(h.Unregister)
	h.Clients = make(map[string]*Client)
}
