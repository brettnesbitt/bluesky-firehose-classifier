// Types for handling firehose messages

package domain

import (
	"encoding/json"
	"fmt"
)

type Reply struct {
	Parent interface{} `json:"parent"`
	Root   interface{} `json:"root"`
}

func (r Reply) String() string {
	return fmt.Sprintf("Parent: %s \nRoot: %s", r.Parent, r.Root)
}

type Record struct {
	Type      string      `json:"$type"`
	CreatedAt string      `json:"createdAt"`
	Langs     []string    `json:"langs"`
	Text      string      `json:"text"`
	Embed     interface{} `json:"embed"`
	Reply     Reply       `json:"reply"`
}

func (r Record) String() string {
	return fmt.Sprintf("Type %s \nCreatedAt: %s \nLangs: %s \nText: %s \nEmbed: %s \nReply:\n%s", r.Type, r.CreatedAt, r.Langs, r.Text, r.Embed, r.Reply)
}

type Commit struct {
	Rev        string `json:"rev"`
	Operation  string `json:"operation"`
	Collection string `json:"collection"`
	Rkey       string `json:"rkey"`
	Record     Record
	CID        string `json:"cid"`
}

func (c Commit) String() string {
	return fmt.Sprintf("Rev: %s \nOperation: %s \nCollection: %s \nRkey: %s \nCID: %s \nRecord:\n%s", c.Rev, c.Operation, c.Collection, c.Rkey, c.CID, c.Record)
}

type Message struct {
	DID          string  `json:"did"`
	TimeUS       float32 `json:"time_us"`
	Kind         string  `json:"kind"`
	Commit       Commit  `json:"commit"`
	Categories   []string
	FinSentiment string
}

func (m Message) ToJSON() (string, error) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func (m Message) String() string {
	return fmt.Sprintf("DID: %s \nTimeUS: %d \nKind: %s \nCommit:\n%s \nCategories: %s\nFinSentiment: %s", m.DID, int(m.TimeUS), m.Kind, m.Commit, m.Categories, m.FinSentiment)
}

func (m *Message) UnmarshalJSON(b []byte) error {
	var wrapper json.RawMessage
	err := json.Unmarshal(b, &wrapper)
	if err == nil {
		type message Message
		h2 := (*message)(m)
		if err = json.Unmarshal(b, &h2); err != nil {
			return err
		}
	}
	return err
}

// Convenience method for checking if post is in English.
func (m Message) IsEnglish() bool {
	return len(m.Commit.Record.Langs) > 0 && m.Commit.Record.Langs[0] == "en"
}

// Vonvenience method for checking if the reply is empty, if so then we have a root message.
func (m Message) IsRoot() bool {
	return m.Commit.Record.Reply.Root == nil
}

// Convencience method for checking if the message is too short to be useful.
func (m Message) IsTooShort() bool {
	return len(m.Commit.Record.Text) < 20
}

// Convenience method for complete validity.
func (m Message) IsValid() bool {
	return m.IsEnglish() && m.IsRoot() && !m.IsTooShort()
}
