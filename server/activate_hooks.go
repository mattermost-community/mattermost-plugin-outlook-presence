package main

import (
	"hash/maphash"
	"runtime"

	"github.com/Brightscout/mattermost-plugin-outlook-presence/server/serializer"
	"github.com/Brightscout/mattermost-plugin-outlook-presence/server/websocket"
)

func (p *Plugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	// Initialize the router and websocket hubs
	p.router = p.InitAPI()
	p.hashSeed = maphash.MakeSeed()
	go p.StartHubs()
	return nil
}

func (p *Plugin) OnDeactivate() error {
	p.StopHubs()
	return nil
}

func (p *Plugin) StartHubs() {
	// Total number of hubs is twice the number of CPUs.
	numberOfHubs := runtime.NumCPU() * 2
	p.API.LogInfo("Starting websocket hubs", "No. of hubs", numberOfHubs)
	hubs := make([]*websocket.Hub, numberOfHubs)

	for i := 0; i < numberOfHubs; i++ {
		hubs[i] = websocket.NewHub()
		hubs[i].ConnectionIndex = i
		go hubs[i].Start(p.API)
	}

	p.wsHubs = hubs
}

// StopHubs stops all the hubs.
func (p *Plugin) StopHubs() {
	p.API.LogInfo("Stopping websocket hub connections")

	for _, hub := range p.wsHubs {
		hub.Stop()
	}
}

func (p *Plugin) BroadCastHubs(event *serializer.UserStatus) {
	for _, hub := range p.wsHubs {
		hub.Broadcast <- event
	}
}

func (p *Plugin) RegisterHub(client *websocket.Client) {
	hub := p.GetHubForConnId(client.ConnID)
	client.Hub = hub
	if hub != nil {
		hub.Register <- client
	}
}

// GetHubForConnId returns the hub for a given conn id.
func (p *Plugin) GetHubForConnId(connID string) *websocket.Hub {
	var hash maphash.Hash
	hash.SetSeed(p.hashSeed)
	hash.Write([]byte(connID))
	index := hash.Sum64() % uint64(len(p.wsHubs))

	return p.wsHubs[int(index)]
}
