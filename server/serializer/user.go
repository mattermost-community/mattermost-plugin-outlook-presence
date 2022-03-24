package serializer

import (
	"encoding/json"
	"io"
)

type Users struct {
	Emails []string `json:"emails"`
}

func UsersFromJson(data io.Reader) *Users {
	var u *Users
	_ = json.NewDecoder(data).Decode(&u)
	return u
}
