package db

import (
	"encoding/json"
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
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	Pronouns    pronoun   `json:"pronouns"`
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
	ID      string    `json:"id"`
	Author  *Author   `json:"author"`
	Content []*MDNode `json:"content"`
	Time    time.Time `json:"time"`
	Bits    int       `json:"bits"`
	Action  bool      `json:"action"`
	Parent  *Message  `json:"reply"`
}

type emote struct {
	URL      string `json:"url"`
	Position [2]int
}
