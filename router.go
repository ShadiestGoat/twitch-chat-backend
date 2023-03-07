package main

import (
	"net/http"

	"github.com/ShadiestGoat/twitch-chat-backend/ws"
	"github.com/go-chi/chi/v5"
)

func router() http.Handler {
	r := chi.NewMux()

	r.HandleFunc("/ws", ws.SocketHandler)

	return r
}
