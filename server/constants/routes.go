package constants

const (
	status                 = "/status"
	GetStatusByEmail       = status + "/{email}"
	PublishStatusChanged   = status + "/publish"
	GetStatusesForAllUsers = status
	Websocket              = "/ws"
)
