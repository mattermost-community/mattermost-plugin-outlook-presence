package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mattermost/mattermost-server/v6/model"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  model.SocketMaxMessageSizeKb,
	WriteBufferSize: model.SocketMaxMessageSizeKb,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func CreateConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return ws, nil
}
