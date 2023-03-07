package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/ShadiestGoat/twitch-chat-backend/config"
	"github.com/ShadiestGoat/twitch-chat-backend/db"
	"github.com/ShadiestGoat/twitch-chat-backend/ws"
	"github.com/ShadiestGoat/twitch-chat-backend/wsutils"
	"github.com/gempir/go-twitch-irc/v4"
)

func main() {
	config.Init()
	
	cleanCB := make(chan bool, 1)
	
	go db.DBClearLoop(cleanCB)

	go func() {
		for {
			<- cleanCB
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
		fmt.Println("Ready & Listening!")
	})

	c.Join(config.TWITCH_CHANNEL_NAME)

	go func() {
		err := c.Connect()
		if err != nil {
			panic(err)
		}
		fmt.Println("disconnected!")
	}()

	defer c.Disconnect()

	r := router()

	go func() {
		err := http.ListenAndServe(":" + config.PORT, r)
		if err != nil {
			panic(err)
		}
	}()

	cancel := make(chan os.Signal, 1)

	signal.Notify(cancel, os.Interrupt)

	<-cancel

	fmt.Println("bye :(")
}
