package service

import (
	"github.com/hlfshell/coppermind/internal/agent"
	"github.com/hlfshell/coppermind/internal/service/chat"
)

type Service struct {
	chat *chat.ChatService
}

func NewService(agent *agent.Agent) *Service {
	return &Service{
		chat: chat.NewChatService(agent),
	}
}
