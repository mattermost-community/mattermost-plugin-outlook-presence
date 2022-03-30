package websocket

import (
	"fmt"

	"github.com/Brightscout/mattermost-plugin-outlook-presence/server/serializer"
)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan *serializer.StatusChangedEvent
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *serializer.StatusChangedEvent),
	}
}

func (p *Pool) Start() {
	for {
		select {
		case client := <-p.Register:
			p.Clients[client] = true
			fmt.Printf("Size of connection pool: %d", len(p.Clients))
		case client := <-p.Unregister:
			delete(p.Clients, client)
			fmt.Printf("Size of connection pool: %d", len(p.Clients))
		case statusChangedEvent := <-p.Broadcast:
			fmt.Println("Sending message to all clients in pool")
			for client, _ := range p.Clients {
				if err := client.Conn.WriteJSON(statusChangedEvent); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}
