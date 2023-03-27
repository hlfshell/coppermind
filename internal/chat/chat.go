package chat

import (
	"encoding/json"
	"strings"
	"time"
)

type Message struct {
	ID           string    `json:"id,omitempty"`
	Conversation string    `json:"conversation,omitempty"`
	User         string    `json:"user,omitempty"`
	Content      string    `json:"content,omitempty"`
	Tone         string    `json:"tone,omitempty"`
	Time         time.Time `json:"time,omitempty"`
}

func (msg *Message) String() string {
	var str strings.Builder
	str.WriteString(msg.User)
	str.WriteString(" | ")
	str.WriteString(msg.Tone)
	str.WriteString(" | ")
	str.WriteString(msg.Content)
	str.WriteString(" | ")
	str.WriteString(msg.Time.String())

	return str.String()
}

func (msg *Message) JSON() (string, error) {
	b, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type Response struct {
	Name    string `json:"name,omitempty"`
	Tone    string `json:"tone,omitempty"`
	Content string `json:"content,omitempty"`
}

func (response *Response) ToMessage(conversation string) *Message {
	return &Message{
		User:         response.Name,
		Tone:         response.Tone,
		Content:      response.Content,
		Time:         time.Now(),
		Conversation: conversation,
	}
}

type Prompt struct {
	Content string
	Type    PromptType
}

type PromptType string

const IdentityPrompt PromptType = "identity"
const SetupPrompt PromptType = "setup"
