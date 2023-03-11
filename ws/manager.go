package ws

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// websocket bullshit
var upgrader = websocket.Upgrader{
	HandshakeTimeout: 0,
	ReadBufferSize:   0,
	WriteBufferSize:  0,
	WriteBufferPool:  nil,
	Subprotocols:     []string{},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
	},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type eventManager struct {
	*sync.Mutex
	Connections map[string]*websocket.Conn
}

func (r *eventManager) SendEvent(e Event) {
	b, _ := json.Marshal(e)
	r.SendEventRaw(b)
}

func (r *eventManager) SendEventRaw(b []byte) {
	r.Lock()
	defer r.Unlock()

	for id, conn := range r.Connections {
		if conn == nil {
			delete(r.Connections, id)
			continue
		}

		err := conn.WriteMessage(1, b)
		if err != nil {
			conn.Close()
			delete(r.Connections, id)
			continue
		}
	}
}

func (r *eventManager) Ping() {
	wg := &sync.WaitGroup{}
	r.Lock()
	defer r.Unlock()

	tmpLock := &sync.Mutex{}

	for id, conn := range r.Connections {
		wg.Add(1)

		go func(c *websocket.Conn, id string) {
			c.WritePreparedMessage(send_ev_ping)
			c.SetReadDeadline(time.Now().Add(5 * time.Second))

			_, p, err := c.ReadMessage()

			if err == nil && len(p) == 1 && p[0] == 'P' {
				c.SetReadDeadline(time.Time{})
			} else {
				if c != nil {
					c.WriteControl(websocket.CloseMessage, nil, time.Time{})
					c.Close()
					c = nil
				}

				tmpLock.Lock()
				delete(r.Connections, id)
				tmpLock.Unlock()
			}

			wg.Done()
		}(conn, id)
	}

	wg.Wait()
}

func (r *eventManager) AddConn(id string, conn *websocket.Conn) bool {
	r.Lock()

	defer r.Unlock()

	if _, ok := r.Connections[id]; ok {
		return false
	}

	r.Connections[id] = conn

	return true
}

var send_ev_ping *websocket.PreparedMessage

func init() {
	ping := Event{
		Event: ev_PING,
	}

	b, _ := json.Marshal(ping)
	msg, err := websocket.NewPreparedMessage(1, b)
	if err != nil {
		panic(err)
	}

	send_ev_ping = msg
}

func init() {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			GlobalMgr.Ping()
		}
	}()
}
