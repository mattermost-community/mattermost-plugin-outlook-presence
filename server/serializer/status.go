package serializer

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mattermost/mattermost-server/v6/model"
)

var validStatus = map[string]bool{
	model.StatusOnline:  true,
	model.StatusAway:    true,
	model.StatusDnd:     true,
	model.StatusOffline: true,
}

type StatusChangedEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func StatusChangedEventFromJson(data io.Reader) (*StatusChangedEvent, error) {
	var s *StatusChangedEvent
	if err := json.NewDecoder(data).Decode(&s); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *StatusChangedEvent) PrePublish() error {
	if !model.IsValidId(s.UserID) {
		return fmt.Errorf("user id is not valid")
	}

	if !validStatus[s.Status] {
		return fmt.Errorf("status is not valid")
	}

	return nil
}
