package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime/debug"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-starter-template/server/serializer"
	"github.com/mattermost/mattermost-server/v6/model"
)

// InitAPI initializes the REST API
func (p *Plugin) InitAPI() *mux.Router {
	r := mux.NewRouter()
	r.Use(p.withRecovery)

	p.handleStaticFiles(r)
	s := r.PathPrefix("/api/v1").Subrouter()

	// TODO: Remove the GetStatusByEmail API as it is unnecessary
	s.HandleFunc("/status/{email}", p.GetStatusByEmail).Methods(http.MethodGet)
	s.HandleFunc("/statuses", p.GetStatusesByEmails).Methods(http.MethodPost)

	// 404 handler
	r.Handle("{anything:.*}", http.NotFoundHandler())
	return r
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
		p.writeError(w, fmt.Sprintf("Unable to get user with email %s. Error: %v", email, userErr.Error()), userErr.StatusCode)
		return
	}

	status, err := p.API.GetUserStatus(user.Id)
	if err != nil {
		p.writeError(w, fmt.Sprintf("Unable to get user's status. Id: %s. Error: %v", user.Id, err.Error()), err.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response, respErr := status.ToJSON()
	if respErr != nil {
		p.writeError(w, fmt.Sprintf("Unable to convert user's status to JSON. Error: %v", respErr.Error()), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(response); err != nil {
		p.writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (p *Plugin) GetStatusesByEmails(w http.ResponseWriter, r *http.Request) {
	users := serializer.UsersFromJson(r.Body)
	userIds := make([]string, len(users.Emails))
	for _, email := range users.Emails {
		if !model.IsValidEmail(email) {
			p.writeError(w, fmt.Sprintf("email %s is not valid", email), http.StatusBadRequest)
			return
		}

		user, userErr := p.API.GetUserByEmail(email)
		if userErr != nil {
			p.writeError(w, fmt.Sprintf("Unable to get user with email %s. Error: %v", email, userErr.Error()), userErr.StatusCode)
			return
		}

		userIds = append(userIds, user.Id)
	}

	statuses, err := p.API.GetUserStatusesByIds(userIds)
	if err != nil {
		p.writeError(w, fmt.Sprintf("Unable to get users' statuses. Error: %v", err.Error()), err.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response, respErr := json.Marshal(statuses)
	if respErr != nil {
		p.writeError(w, fmt.Sprintf("Unable to marshal users' statuses. Error: %v", respErr.Error()), http.StatusInternalServerError)
		return
	}

	if _, wErr := w.Write(response); wErr != nil {
		p.writeError(w, wErr.Error(), http.StatusInternalServerError)
	}
}

func (p *Plugin) writeError(w http.ResponseWriter, errorMessage string, statusCode int) {
	p.API.LogError(errorMessage)
	http.Error(w, errorMessage, statusCode)
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
