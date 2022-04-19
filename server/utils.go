package main

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v6/model"
)

func (p *Plugin) writeError(w http.ResponseWriter, errorMessage string, statusCode int) {
	p.API.LogError(errorMessage)
	http.Error(w, errorMessage, statusCode)
}

func writeStatusOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	m := map[string]string{
		model.STATUS: model.StatusOk,
	}
	_, _ = w.Write([]byte(model.MapToJSON(m)))
}
