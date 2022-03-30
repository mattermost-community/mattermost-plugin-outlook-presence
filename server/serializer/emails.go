package serializer

import (
	"encoding/json"
	"io"
)

type Emails []string

func EmailsFromJson(data io.Reader) *Emails {
	var e *Emails
	_ = json.NewDecoder(data).Decode(&e)
	return e
}
