package service

import (
	"github.com/hlfshell/coppermind/internal/agent"
	"github.com/hlfshell/coppermind/internal/service/chat"
)

type Service struct {
	Chat *chat.ChatService
}

func NewService(agent *agent.Agent) *Service {
	return &Service{
		Chat: chat.NewChatService(agent),
	}
}
