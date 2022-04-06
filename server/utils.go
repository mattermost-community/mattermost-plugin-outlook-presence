package main

import "net/http"

func (p *Plugin) writeError(w http.ResponseWriter, errorMessage string, statusCode int) {
	p.API.LogError(errorMessage)
	http.Error(w, errorMessage, statusCode)
}
