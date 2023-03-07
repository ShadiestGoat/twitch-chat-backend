package ws

import (
	"fmt"
	"net/http"

	"github.com/ShadiestGoat/twitch-chat-backend/db"
	"github.com/gorilla/websocket"
)

// true -> everything is fine
// false -> close ws & send err
func tryUpgrade(w http.ResponseWriter, r *http.Request) (bool, *websocket.Conn) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("1")
		return false, nil
	}

	success := GlobalMgr.AddConn(genID(), conn)
	if !success {
		fmt.Printf("5")
		return false, conn
	}
	
	msgs := db.FetchMessages()

	err = conn.WriteJSON(Event{
		Event: EV_UPDATE_MESSAGES,
		Data: msgs,
	})
	
	if err != nil {
		panic(err)
	}

	return true, conn
}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	ok, conn := tryUpgrade(w, r)
	
	if !ok && conn != nil {
		conn.WriteJSON(struct {
			Err string `json:"error"`
		}{
			Err: "bad connection",
		})
	}
}
