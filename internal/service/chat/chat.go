package chat

import (
	"github.com/hlfshell/coppermind/internal/agent"
	"github.com/hlfshell/coppermind/internal/chat"
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
	return service.agent.SendMessage(message)
}
