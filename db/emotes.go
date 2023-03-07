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
	ID string `json:"id"`
	Name string `json:"name"`
}

var emotes = &emoteDB{
	RWMutex:     &sync.RWMutex{},
	blacklist: map[string]bool{},
	sevenTV:   map[string]string{},
}

func init() {
	for _, e := range config.EMOTE_BLACKLIST {
		emotes.blacklist[e] = true
	}

	go func () {
		for {
			resp, err := http.Get(`https://7tv.io/v3/users/twitch/` + config.TWITCH_CHANNEL_ID)
			if log.ErrorIfErr(err, "fetching emotes") {
				fmt.Println("emote err", err)
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

func ProcessEmotes(str string, base []*twitch.Emote) []*Emote {
	newEmotes := []*Emote{}

	emotes.RLock()

	for _, e := range base {
		if !emotes.blacklist[strings.ToLower(e.Name)] {
			pos := [][2]int{}
			for _, p := range e.Positions {
				pos = append(pos, [2]int{p.Start, p.End})
			}

			newEmotes = append(newEmotes, &Emote{
				URL:       fmt.Sprintf(`https://static-cdn.jtvnw.net/emoticons/v2/%s/default/dark/1.0`, e.ID),
				Positions: pos,
			})
		}
	}
	
	spl := strings.Split(str, " ")

	emoteMap := map[string]*Emote{}

	l := 0

	for _, s := range spl {
		l += len(s) + 1
		emoteID, ok := emotes.sevenTV[s]
		if !ok {
			continue
		}

		if emoteMap[emoteID] == nil {
			emoteMap[emoteID] = &Emote{
				URL:       fmt.Sprintf(`https://cdn.7tv.app/emote/%s/1x.webp`, emoteID),
				Positions: [][2]int{},
			}
		}

		emoteMap[emoteID].Positions = append(emoteMap[emoteID].Positions, [2]int{
			l - len(s) - 1, l - 1,
		})
	}

	emotes.RUnlock()
	
	for _, e := range emoteMap {
		newEmotes = append(newEmotes, e)
	}

	return newEmotes
}
