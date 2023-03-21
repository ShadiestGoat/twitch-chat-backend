package db

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ShadiestGoat/twitch-chat-backend/config"
	"github.com/ShadiestGoat/twitch-chat-backend/log"
	"github.com/gempir/go-twitch-irc/v4"
)

type emoteDB struct {
	*sync.RWMutex
	// emote name : true
	blacklist map[string]bool

	// name : id
	sevenTV map[string]string
}

type sevenTVEmote struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var emotes = &emoteDB{
	RWMutex:   &sync.RWMutex{},
	blacklist: map[string]bool{},
	sevenTV:   map[string]string{},
}

func Init() {
	for _, e := range config.EMOTE_BLACKLIST {
		emotes.blacklist[e] = true
	}

	go func() {
		for {
			resp, err := http.Get(`https://7tv.io/v3/users/twitch/` + config.TWITCH_CHANNEL_ID)
			if log.ErrorIfErr(err, "fetching emotes") {
				time.Sleep(10 * time.Minute)
				continue
			}
			if resp == nil || resp.Body == nil || resp.StatusCode != 200 {
				status := "???"
				body := "???"
				if resp != nil {
					status = fmt.Sprint(resp.StatusCode)
					if resp.Body != nil {
						b, _ := io.ReadAll(resp.Body)
						body = string(b)
					}
				}

				log.Error("bad emote response:\nStatus: %s\nBody: %s\nResp: %v", status, body, resp)
				time.Sleep(10 * time.Minute)
				continue
			}

			var emotesRaw *struct {
				Set *struct {
					Emotes []*sevenTVEmote `json:"emotes"`
				} `json:"emote_set"`
			}

			err = json.NewDecoder(resp.Body).Decode(&emotesRaw)

			if log.ErrorIfErr(err, "json decoding emotes") {
				time.Sleep(10 * time.Minute)
				continue
			}

			emotes.Lock()

			emotes.sevenTV = map[string]string{}

			for _, e := range emotesRaw.Set.Emotes {
				emotes.sevenTV[e.Name] = e.ID
			}

			emotes.Unlock()

			time.Sleep(30 * time.Minute)
		}
	}()
}

func processEmotes(str string, base []*twitch.Emote) map[string]string {
	newEmotes := map[string]string{}

	emotes.RLock()

	for _, e := range base {
		if !emotes.blacklist[strings.ToLower(e.Name)] {
			newEmotes[strings.ToLower(e.Name)] = fmt.Sprintf(`https://static-cdn.jtvnw.net/emoticons/v2/%s/default/dark/1.0`, e.ID)
		}
	}

	spl := strings.Split(str, " ")

	for _, s := range spl {
		emoteID, ok := emotes.sevenTV[s]
		if !ok {
			continue
		}
		newEmotes[s] = fmt.Sprintf(`https://cdn.7tv.app/emote/%s/1x.webp`, emoteID)
	}

	emotes.RUnlock()

	return newEmotes
}
