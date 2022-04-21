package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime/debug"

	"github.com/Brightscout/mattermost-plugin-outlook-presence/server/constants"
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
	s.HandleFunc(constants.PublishStatusChanged, p.PublishStatusChanged).Methods(http.MethodPost)
	s.HandleFunc(constants.GetStatusesForAllUsers, p.handleAuthRequired(p.GetStatusesForAllUsers)).Methods(http.MethodGet)
	s.HandleFunc(constants.Websocket, p.handleAuthRequired(p.serveWebSocket))

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
		p.writeError(w, fmt.Sprintf("error in creating websocket connection. Error: %s", err.Error()), http.StatusInternalServerError)
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
	statusChangedEvent, err := serializer.UserStatusFromJSON(r.Body)
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

func (p *Plugin) GetStatusesForAllUsers(w http.ResponseWriter, r *http.Request) {
	page, err := parseIntParamFromURL(r.URL, constants.Page, constants.DefaultPage)
	if err != nil {
		p.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	users, userErr := p.API.GetUsers(&model.UserGetOptions{
		Active:  true,
		Page:    page,
		PerPage: p.getConfiguration().PerPageStatuses,
	})
	if userErr != nil {
		p.writeError(w, fmt.Sprintf("failed to get users. Error: %s", userErr.Error()), userErr.StatusCode)
		return
	}

	userStatusArr := make([]*serializer.UserStatus, len(users))
	userIds := make([]string, len(users))
	statusMap := make(map[string]*serializer.UserStatus)
	for index, user := range users {
		userIds[index] = user.Id
		userStatus := serializer.UserStatus{
			UserID: user.Id,
			Email:  user.Email,
			Status: model.StatusOffline,
		}
		statusMap[user.Id] = &userStatus
	}

	statusArr, statusErr := p.API.GetUserStatusesByIds(userIds)
	if statusErr != nil {
		p.writeError(w, fmt.Sprintf("error in getting statuses. Error: %s", statusErr.Error()), statusErr.StatusCode)
		return
	}

	for index, status := range statusArr {
		statusMap[status.UserId].Status = status.Status
		userStatusArr[index] = statusMap[status.UserId]
	}

	w.Header().Set("Content-Type", "application/json")
	response, respErr := json.Marshal(userStatusArr)
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
