package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

var GlobalMgr = &eventManager{
	Mutex:       &sync.Mutex{},
	Connections: map[string]*websocket.Conn{},
}
