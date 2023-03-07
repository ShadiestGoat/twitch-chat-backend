package db

import (
	"sync"
	"time"
)

type DB struct {
	Messages         map[string]*Message
	OldestMessages   Ring[*Message]
	Authors          map[string]*Author
	MessagesByAuthor map[string]map[string]*Message

	*sync.Mutex
}

var db = &DB{
	Messages:         map[string]*Message{},
	OldestMessages:   NewRing[*Message](200),
	Authors:          map[string]*Author{},
	MessagesByAuthor: map[string]map[string]*Message{},
	Mutex:            &sync.Mutex{},
}

func DBClearLoop(c chan bool) {
	for {
		time.Sleep(time.Minute * 6)
		db.cleanDB()
		c <- true
	}
}

func (db *DB) cleanDB() {
	db.Lock()
	defer db.Unlock()

	s := len(db.Messages)

	for i := s - 1; i >= 0; i-- {
		msg := db.OldestMessages[i]
		if msg == nil {
			continue
		}

		if time.Since(msg.Time) > 7 * time.Minute {
			db.OldestMessages[i] = nil
			delete(db.Messages, msg.ID)
			delete(db.MessagesByAuthor[msg.Author.ID], msg.ID)
			if len(db.MessagesByAuthor[msg.Author.ID]) == 0 {
				delete(db.MessagesByAuthor, msg.Author.ID)
			}
		}
	}

	fixOldMessages()

	authors := map[string]bool{}

	for _, msg := range db.Messages {
		authors[msg.Author.ID] = true
	}

	for _, a := range db.Authors {
		if a == nil || (!authors[a.ID] && time.Since(a.TimeFetched) > time.Minute*45) {
			delete(db.Authors, a.ID)
		}
	}
}
