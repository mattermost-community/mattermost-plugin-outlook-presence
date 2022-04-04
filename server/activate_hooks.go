package main

import "github.com/Brightscout/mattermost-plugin-outlook-presence/server/websocket"

func (p *Plugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	// Initialize DB service
	pool := websocket.NewPool()
	go pool.Start(p.API)
	p.wsPool = pool
	p.router = p.InitAPI()
	return nil
}
