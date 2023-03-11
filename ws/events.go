package ws

const (
	ev_PING            = "PING"
	EV_UPDATE_MESSAGES = "MESSAGES"
	EV_NEW_MESSAGE     = "MESSAGE"
)

type Event struct {
	Event string `json:"event"`
	Data  any    `json:"data,omitempty"`
}

type EventData struct {
	ID string `json:"id"`
}

type EventDataMerge struct {
	From string `json:"from"`
	To   string `json:"to"`
}
