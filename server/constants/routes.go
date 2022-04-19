package constants

const (
	Status               = "/status"
	GetStatusByEmail     = Status + "/{email}"
	PublishStatusChanged = Status + "/publish"
	Websocket            = "/ws"
)
