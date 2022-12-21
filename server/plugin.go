package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-outlook-presence/server/constants"
	"github.com/mattermost/mattermost-plugin-outlook-presence/server/serializer"
	"github.com/mattermost/mattermost-plugin-outlook-presence/server/websocket"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
	router        *mux.Router
	wsPool        *websocket.Pool
}

// ServeHTTP handles HTTP requests
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.API.LogDebug("New plugin request:", "Host", r.Host, "RequestURI", r.RequestURI, "Method", r.Method)
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) OnPluginClusterEvent(c *plugin.Context, ev model.PluginClusterEvent) {
	if ev.Id != constants.ClusterEvent {
		return
	}

	var event *serializer.UserStatus
	if err := json.Unmarshal(ev.Data, &event); err != nil {
		p.API.LogDebug("Error in unmarshaling the cluster event data", "Error", err.Error())
		return
	}

	// Broadcast the event for all the clusters
	p.BroadcastEvent(event)
}

func (p *Plugin) RegisterClient(client *websocket.Client) {
	p.wsPool.Register <- client
}

func (p *Plugin) BroadcastEvent(event *serializer.UserStatus) {
	p.wsPool.Broadcast <- event
}
