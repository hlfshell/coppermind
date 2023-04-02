package chat

import (
	"fmt"

	"github.com/hlfshell/coppermind/internal/agent"
	"github.com/hlfshell/coppermind/pkg/chat"
)

type ChatService struct {
	agent *agent.Agent
}

func NewChatService(agent *agent.Agent) *ChatService {
	return &ChatService{
		agent: agent,
	}
}

func (service *ChatService) SendMessage(message *chat.Message) (*chat.Response, error) {
	fmt.Println("Received", message)
	resp, err := service.agent.SendMessage(message)
	fmt.Println("responding", err, resp)
	return resp, err
}
