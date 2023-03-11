package wsutils

import (
	"github.com/ShadiestGoat/twitch-chat-backend/db"
	"github.com/ShadiestGoat/twitch-chat-backend/ws"
)

func UpdateMessages() {
	ws.GlobalMgr.SendEvent(ws.Event{
		Event: ws.EV_UPDATE_MESSAGES,
		Data:  db.FetchMessages(),
	})
}
