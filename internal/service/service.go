package service

import (
	"github.com/hlfshell/coppermind/internal/agent"
	"github.com/hlfshell/coppermind/internal/service/chat"
)

type Service struct {
	Chat   *chat.ChatService
	agents map[string]*agent.Agent
}

func NewService(agent *agent.Agent) *Service {
	return &Service{
		Chat:   chat.NewChatService(agent),
		agents: map[string]*agent.Agent{},
	}
}

func (service *Service) GetAgent(id string) (*agent.Agent, error) {
	if agent, ok := service.agents[id]; !ok {
		agent = nil
		service.agents[id] = agent
	} else {
		return agent, nil
	}
	return nil, nil
}
