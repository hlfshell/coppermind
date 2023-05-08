package service

import (
	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/agents"
)

type AgentService struct {
	db store.Store
}

func NewAgentService(db store.Store) *AgentService {
	return &AgentService{
		db: db,
	}
}

func (service *AgentService) CreateAgent(agent *agents.Agent) error {
	return service.db.SaveAgent(agent)
}

func (service *AgentService) GetAgent(id string) (*agents.Agent, error) {
	return service.db.GetAgent(id)
}

func (service *AgentService) GetAgents() ([]*agents.Agent, error) {
	return service.db.ListAgents()
}
