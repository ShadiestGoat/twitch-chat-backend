package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ShadiestGoat/pronoundb"
)

type pronoun pronoundb.Pronoun

func (p pronoun) MarshalJSON() ([]byte, error) {
	pr := pronoundb.Pronoun(p)
	abb := ""

	if pr != pronoundb.PR_UNSPECIFIED {
		abb = pr.Abbreviation()
	}

	return json.Marshal(abb)
}

type Author struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Pronouns    pronoun `json:"pronouns"`
	TimeFetched time.Time `json:"-"`
}

func (a Author) String() string {
	r := a.Name
	if pronoundb.Pronoun(a.Pronouns) != pronoundb.PR_UNSPECIFIED {
		r += " " + pronoundb.Pronoun(a.Pronouns).Abbreviation()
	}
	return r
}

type Message struct {
	ID      string          `json:"id"`
	Author  *Author         `json:"author"`
	Content string          `json:"content"`
	Time    time.Time       `json:"time"`
	Emotes  []*Emote		`json:"emotes"`
	Bits    int             `json:"bits"`
	Action  bool            `json:"action"`
	Parent  *Message        `json:"reply"`
}

func (msg Message) String() string {
	parent := ""
	if msg.Parent != nil {
		parent = msg.Parent.String()
		parent = "╭─" + parent + "\n"
	}
	return fmt.Sprintf("%s[%v] %s", parent, msg.Author, msg.Content)
}

type Emote struct {
	URL string `json:"url"`
	// [][start, end]
	Positions [][2]int `json:"positions"`
}

// type Message struct {
// 	ID string
// 	Author *Author
// 	Reply *Message
// 	Emotes map[string]string
// 	Time time.Time
// }
