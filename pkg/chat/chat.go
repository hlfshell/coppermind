package chat

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/hlfshell/coppermind/pkg/artifacts"
)

type Conversation struct {
	ID        string     `json:"id,omitempty" db:"id"`
	CreatedAt time.Time  `json:"created_at,omitempty" db:"created_at"`
	Agent     string     `json:"agent,omitempty" db:"agent"`
	User      string     `json:"user,omitempty" db:"user"`
	Messages  []*Message `json:"messages,omitempty"`
}

func (conversation *Conversation) Equal(other *Conversation) bool {
	//Check to see if the time is within a small difference.
	//This is because storing time and reading it back is
	//unlikely to match due to Go's nanosecond precision.
	//We will say anything within a second is equivalent for
	//simplictiy
	timeDifference := conversation.CreatedAt.Sub(other.CreatedAt)
	if timeDifference < 0 {
		timeDifference = -timeDifference
	}

	return conversation.ID == other.ID &&
		conversation.Agent == other.Agent &&
		conversation.User == other.User &&
		timeDifference < time.Second
}

func (conversation *Conversation) PastNMessages(n int) []*Message {
	if n > len(conversation.Messages) {
		n = len(conversation.Messages)
	}
	return conversation.Messages[len(conversation.Messages)-n:]
}

type Message struct {
	ID           string     `json:"id,omitempty"`
	Conversation string     `json:"conversation,omitempty" db:"conversation"`
	User         string     `json:"user,omitempty" db:"user"`
	Agent        string     `json:"agent,omitempty" db:"agent"`
	From         string     `json:"from,omitempty" db:"from"`
	Content      string     `json:"content,omitempty" db:"content"`
	Artifacts    []Artifact `json:"artifacts,omitempty"`
	CreatedAt    time.Time  `json:"created_at,omitempty" db:"created_at"`
}

func (msg *Message) Equal(other *Message) bool {
	timeDifference := msg.CreatedAt.Sub(other.CreatedAt)
	if timeDifference < 0 {
		timeDifference = -timeDifference
	}

	return msg.ID == other.ID &&
		msg.Agent == other.Agent &&
		msg.User == other.User &&
		msg.Content == other.Content &&
		msg.From == other.From &&
		msg.Conversation == other.Conversation &&
		timeDifference < time.Second
}

func (msg *Message) DatedString() string {
	var str strings.Builder
	str.WriteString(msg.CreatedAt.Format("Jan, 2 06 15:04"))
	str.WriteString(" | ")
	str.WriteString(msg.From)
	str.WriteString(" | ")
	str.WriteString(msg.Content)

	return str.String()
}

func (msg *Message) SimpleString() string {
	var str strings.Builder
	str.WriteString(msg.From)
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

/*
The Response struct is used to handle all the generated
output from the LLM. For now, that's just the Content,
and can ultimately be converted to a resulting Message.

In the future, this will include types of Artifacts,
such as generated images, search results, PDFs, sound
files for generated voices, etc. The resulting service
that receives the response has to then determine if it
can and how it should utilize responses with artifacts.
*/
type Response struct {
	Content   string               `json:"content,omitempty"`
	Artifacts []artifacts.Artifact `json:"artifacts,omitempty"`
}

/*
ToMessage takes a response and converts it to a message
given other required information.
*/
func (response *Response) ToMessage(user string, agent string, conversation string) *Message {
	return &Message{
		ID:           uuid.New().String(),
		Agent:        agent,
		User:         user,
		From:         agent,
		Content:      response.Content,
		CreatedAt:    time.Now(),
		Conversation: conversation,
	}
}
