package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime/debug"

	"github.com/Brightscout/mattermost-plugin-outlook-presence/server/serializer"
	"github.com/Brightscout/mattermost-plugin-outlook-presence/server/websocket"
	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v6/model"
)

// InitAPI initializes the REST API
func (p *Plugin) InitAPI() *mux.Router {
	r := mux.NewRouter()
	r.Use(p.withRecovery)

	p.handleStaticFiles(r)
	s := r.PathPrefix("/api/v1").Subrouter()

	// Add the custom plugin routes here
	s.HandleFunc("/status/publish", p.PublishStatusChanged).Methods(http.MethodPost)
	s.HandleFunc("/status/{email}", p.handleAuthRequired(p.GetStatusByEmail)).Methods(http.MethodGet)
	// TODO: TODO: Remove the GetStatusesByEmails API as it is unnecessary
	s.HandleFunc("/statuses", p.handleAuthRequired(p.GetStatusesByEmails)).Methods(http.MethodPost)
	s.HandleFunc("/ws", p.handleAuthRequired(p.serveWebSocket))

	// 404 handler
	r.Handle("{anything:.*}", http.NotFoundHandler())
	return r
}

// handleAuthRequired verifies if provided request is performed by an authorized source.
func (p *Plugin) handleAuthRequired(handleFunc func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if status, err := verifyHTTPSecret(p.getConfiguration().Secret, r.FormValue("secret")); err != nil {
			p.writeError(w, fmt.Sprintf("Invalid Secret. Error: %s", err.Error()), status)
			return
		}

		handleFunc(w, r)
	}
}

func (p *Plugin) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	connection, err := websocket.CreateConnection(w, r)
	if err != nil {
		p.API.LogError("error in creating websocket connection", "Error", err.Error())
		return
	}

	client := &websocket.Client{
		Conn: connection,
		Pool: p.wsPool,
	}

	p.wsPool.Register <- client
	client.Read(p.API)
}

func (p *Plugin) PublishStatusChanged(w http.ResponseWriter, r *http.Request) {
	statusChangedEvent, err := serializer.StatusChangedEventFromJson(r.Body)
	if err != nil {
		p.writeError(w, fmt.Sprintf("error in deserializing the request body. Error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if err = statusChangedEvent.PrePublish(); err != nil {
		p.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, userErr := p.API.GetUser(statusChangedEvent.UserID)
	if userErr != nil {
		p.writeError(w, fmt.Sprintf("Unable to get user by id %s. Error: %s", statusChangedEvent.UserID, userErr.Error()), userErr.StatusCode)
		return
	}

	statusChangedEvent.Email = user.Email
	p.wsPool.Broadcast <- statusChangedEvent
	writeStatusOK(w)
}

func (p *Plugin) GetStatusByEmail(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	email := params["email"]
	if !model.IsValidEmail(email) {
		p.writeError(w, fmt.Sprintf("email %s is not valid", email), http.StatusBadRequest)
		return
	}

	user, userErr := p.API.GetUserByEmail(email)
	if userErr != nil {
		p.writeError(w, fmt.Sprintf("Unable to get user with email %s. Error: %s", email, userErr.Error()), userErr.StatusCode)
		return
	}

	status, err := p.API.GetUserStatus(user.Id)
	if err != nil {
		p.writeError(w, fmt.Sprintf("Unable to get user's status. Id: %s. Error: %s", user.Id, err.Error()), err.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response, respErr := status.ToJSON()
	if respErr != nil {
		p.writeError(w, fmt.Sprintf("Unable to convert user's status to JSON. Error: %s", respErr.Error()), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(response); err != nil {
		p.writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (p *Plugin) GetStatusesByEmails(w http.ResponseWriter, r *http.Request) {
	var emails []string
	if err := json.NewDecoder(r.Body).Decode(&emails); err != nil {
		p.writeError(w, fmt.Sprintf("error in deserializing the request body. Error: %s", err.Error()), http.StatusBadRequest)
		return
	}
	userIds := make([]string, len(emails))
	for _, email := range emails {
		if !model.IsValidEmail(email) {
			p.writeError(w, fmt.Sprintf("email %s is not valid", email), http.StatusBadRequest)
			return
		}

		user, userErr := p.API.GetUserByEmail(email)
		if userErr != nil {
			p.writeError(w, fmt.Sprintf("Unable to get user with email %s. Error: %s", email, userErr.Error()), userErr.StatusCode)
			return
		}

		userIds = append(userIds, user.Id)
	}

	statuses, err := p.API.GetUserStatusesByIds(userIds)
	if err != nil {
		p.writeError(w, fmt.Sprintf("Unable to get users' statuses. Error: %s", err.Error()), err.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response, respErr := json.Marshal(statuses)
	if respErr != nil {
		p.writeError(w, fmt.Sprintf("Unable to convert users' statuses to JSON. Error: %s", respErr.Error()), http.StatusInternalServerError)
		return
	}

	if _, wErr := w.Write(response); wErr != nil {
		p.writeError(w, wErr.Error(), http.StatusInternalServerError)
	}
}

// handleStaticFiles handles the static files under the assets directory.
func (p *Plugin) handleStaticFiles(r *mux.Router) {
	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		p.API.LogWarn("Failed to get bundle path.", "Error", err.Error())
		return
	}

	// This will serve static files from the 'assets' directory under '/static/<filename>'
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(bundlePath, "assets")))))
}

// withRecovery allows recovery from panics
func (p *Plugin) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if x := recover(); x != nil {
				p.API.LogError("Recovered from a panic",
					"url", r.URL.String(),
					"error", x,
					"stack", string(debug.Stack()))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
