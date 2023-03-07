package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/ShadiestGoat/colorutils"
	"github.com/ShadiestGoat/pronoundb"
	"github.com/ShadiestGoat/twitch-chat-backend/colors"
	"github.com/ShadiestGoat/twitch-chat-backend/log"
	"github.com/gempir/go-twitch-irc/v4"
)

// Add a message to the pool
// WARNING: NOT SAFE FOR CONCURRENT USE!
func addMessage(msg *Message) {
	db.Messages[msg.ID] = msg
	if db.MessagesByAuthor[msg.Author.ID] == nil {
		db.MessagesByAuthor[msg.Author.ID] = map[string]*Message{}
	}
	db.MessagesByAuthor[msg.Author.ID][msg.ID] = msg
	db.OldestMessages.Add(msg)
}

// Shifts all the messages in OldestMessages into a good order
// WARNING: NOT SAFE FOR CONCURRENT USE!
func fixOldMessages() {
	lastAvailable := 0

	for i, item := range db.OldestMessages {
		if item != nil {
			db.OldestMessages[lastAvailable] = item
			db.OldestMessages[i] = nil
			lastAvailable++
		}
	}
}

// Get a message by ID (from the message pool)
func FetchMessage(id string) *Message {
	db.Lock()
	defer db.Unlock()

	return db.Messages[id]
}

// Get all the available messages from the considered pool
func FetchMessages() []*Message {
	db.Lock()
	defer db.Unlock()
	msgs := []*Message{}

	for _, msg := range db.OldestMessages {
		if msg == nil {
			continue
		}

		msgs = append(msgs, msg)
	}

	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs
}

// Remove a message from the message pool
func DeleteMessage(id string) {
	db.Lock()
	defer db.Unlock()

	msg := db.Messages[id]

	delete(db.Messages, id)
	delete(db.MessagesByAuthor[msg.Author.ID], id)
	if len(db.MessagesByAuthor[msg.Author.ID]) == 0 {
		delete(db.MessagesByAuthor, msg.Author.ID)
	}

	deleteI := -1
	for i, msg := range db.OldestMessages {
		if msg.ID == id {
			deleteI = i
			break
		}
	}

	if deleteI != -1 {
		db.OldestMessages.Delete(deleteI)
	}
}

// Remove all the messages from the pool by author 'userID'
func DeleteMessagesByAuthor(userID string) {
	db.Lock()

	if len(db.MessagesByAuthor[userID]) == 0 {
		return
	}

	ids := map[string]bool{}

	for _, msg := range db.MessagesByAuthor[userID] {
		ids[msg.ID] = true
	}

	for id := range ids {
		msg := db.Messages[id]

		delete(db.Messages, id)
		delete(db.MessagesByAuthor[msg.Author.ID], id)
		if len(db.MessagesByAuthor[msg.Author.ID]) == 0 {
			delete(db.MessagesByAuthor, msg.Author.ID)
		}
	}

	for i, msg := range db.OldestMessages {
		if msg == nil || ids[msg.ID] {
			db.OldestMessages[i] = nil
		}
	}

	fixOldMessages()

	db.Unlock()
}

var pronounClient = pronoundb.NewClient()

// Parse a raw twitch message into a readable local message and add it to the pool
// If the message has already been added to the pool, no more processing will be done to it.
func ProcessMessage(raw *twitch.PrivateMessage) *Message {
	msg := FetchMessage(raw.ID)
	if msg != nil {
		return msg
	}

	content := raw.Message

	if raw.Reply != nil {
		content, _ = strings.CutPrefix(content, fmt.Sprintf("@%s ", raw.Reply.ParentDisplayName))
	}
	
	msg = &Message{
		ID:      raw.ID,
		Author:  ProcessAuthor(&raw.User),
		Content: content,
		Time:    raw.Time,
		Emotes:  ProcessEmotes(content, raw.Emotes),
		Bits:    raw.Bits,
		Action:  raw.Action,
	}

	if raw.Reply != nil {
		msg.Parent = FetchMessage(raw.Reply.ParentMsgID)
	}

	db.Lock()
	addMessage(msg)
	db.Unlock()

	return msg
}

// Parse a raw twitch user into a local author.
// If the db contains the user & they has not expired, it will return the already made author.
func ProcessAuthor(raw *twitch.User) *Author {
	db.Lock()
	defer db.Unlock()

	t := db.Authors[raw.ID]
	if t != nil && time.Since(t.TimeFetched) <= 15*time.Minute {
		return t
	}

	pr, err := pronounClient.Lookup(pronoundb.PLATFORM_TWITCH, raw.ID)
	if log.ErrorIfErr(err, "looking up pronouns") {
		pr = pronoundb.PR_UNSPECIFIED
	}

	a := Author{
		ID:          raw.ID,
		Name:        raw.DisplayName,
		Color:       raw.Color,
		Pronouns:    pronoun(pr),
		TimeFetched: time.Now(),
	}

	if a.Color == "" {
		if t != nil {
			a.Color = t.Color
		} else {
			r, g, b := colorutils.NewContrastColor(0x30, 0x36, 0x3d)
			a.Color = colorutils.Hexadecimal(r, g, b)
		}
	} else {
		a.Color = colors.FixColor(a.Color[1:])
	}

	if t != nil {
		*t = a
	} else {
		t = &a
	}

	db.Authors[raw.ID] = t

	return t
}
