package ws

import (
	"fmt"
	"sync"
	"time"
)

var lockIDGen = &sync.Mutex{}

func genID() string {
	lockIDGen.Lock()
	defer lockIDGen.Unlock()

	v := fmt.Sprint(time.Now().UnixNano())
	time.Sleep(time.Nanosecond)
	return v
}
