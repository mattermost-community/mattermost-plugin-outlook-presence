package main

import "github.com/Brightscout/mattermost-plugin-outlook-presence/server/websocket"

func (p *Plugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	// Initialize the router and websocket pool
	p.router = p.InitAPI()
	pool := websocket.NewPool()
	go pool.Start(p.API)
	p.wsPool = pool
	return nil
}
