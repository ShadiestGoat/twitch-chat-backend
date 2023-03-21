package main

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/ShadiestGoat/twitch-chat-backend/config"
	"github.com/ShadiestGoat/twitch-chat-backend/db"
	"github.com/ShadiestGoat/twitch-chat-backend/log"
	"github.com/ShadiestGoat/twitch-chat-backend/ws"
	"github.com/ShadiestGoat/twitch-chat-backend/wsutils"
	"github.com/gempir/go-twitch-irc/v4"
)

func main() {
	config.Init()

	db.ParseContent(`first **bold** *ital* ~~strike~~ ***bold ital*** ![image name](https://example.com) [**link** name](https://example.com) ~~*strike ital* **bold strike**~~ ~~**asdsad a~~**`, nil)
	panic("A")

	cleanCB := make(chan bool, 1)

	go db.DBClearLoop(cleanCB)

	go func() {
		for {
			<-cleanCB
			wsutils.UpdateMessages()
		}
	}()

	db.Init()

	c := twitch.NewAnonymousClient()

	// Message Deletion
	c.OnClearMessage(func(e twitch.ClearMessage) {
		db.DeleteMessage(e.TargetMsgID)
		wsutils.UpdateMessages()
	})

	// Ban/Timeout
	c.OnClearChatMessage(func(e twitch.ClearChatMessage) {
		db.DeleteMessagesByAuthor(e.TargetUserID)
		wsutils.UpdateMessages()
	})

	// New Message
	c.OnPrivateMessage(func(e twitch.PrivateMessage) {
		msg := db.ProcessMessage(&e)

		ws.GlobalMgr.SendEvent(ws.Event{
			Event: ws.EV_NEW_MESSAGE,
			Data:  msg,
		})
	})

	c.OnSelfJoinMessage(func(e twitch.UserJoinMessage) {
		log.Success("Listening to twitch chat!")
	})

	c.Join(config.TWITCH_CHANNEL_NAME)

	go func() {
		err := c.Connect()
		if err != nil {
			panic(err)
		}
		log.Debug("Disconnected!")
	}()

	defer c.Disconnect()

	r := router()

	go func() {
		log.Success("Starting server on 0.0.0.0:%s", config.PORT)

		err := http.ListenAndServe(":"+config.PORT, r)
		if err != nil {
			panic(err)
		}
	}()

	cancel := make(chan os.Signal, 1)

	signal.Notify(cancel, os.Interrupt)

	<-cancel

	log.Success("bye :(")
}
