package chat

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID        string    `json:"id,omitempty" db:"id"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	Agent     string    `json:"agent,omitempty" db:"agent"`
	User      string    `json:"user,omitempty" db:"user"`
}

type Message struct {
	ID           string    `json:"id,omitempty"`
	Conversation string    `json:"conversation,omitempty" db:"conversation"`
	User         string    `json:"user,omitempty" db:"user"`
	Content      string    `json:"content,omitempty" db:"message"`
	Tone         string    `json:"tone,omitempty" db:"tone"`
	CreatedAt    time.Time `json:"created_at,omitempty" db:"created_at"`
}

func (msg *Message) String() string {
	var str strings.Builder
	str.WriteString(msg.User)
	str.WriteString(" | ")
	str.WriteString(msg.Tone)
	str.WriteString(" | ")
	str.WriteString(msg.Content)
	str.WriteString(" | ")
	str.WriteString(msg.CreatedAt.String())

	return str.String()
}

func (msg *Message) SimpleString() string {
	var str strings.Builder
	str.WriteString(msg.User)
	str.WriteString(" | ")
	str.WriteString(msg.Content)
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
		ID:           uuid.New().String(),
		User:         response.Name,
		Tone:         response.Tone,
		Content:      response.Content,
		CreatedAt:    time.Now(),
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
