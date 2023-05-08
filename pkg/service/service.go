package service

import (
	"time"

	"github.com/hlfshell/coppermind/internal/llm"
	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/config"
)

type Service struct {
	db     store.Store
	llm    llm.LLM
	config config.Config

	// Services
	Messages *MessageService
	Summary  *SummaryService
	Agents   *AgentService
	Users    *UserService

	// Daemon services
	summarizationTicker *time.Ticker
	knowledgeTicker     *time.Ticker
}

func NewService(db store.Store, llm llm.LLM, config *config.Config) *Service {
	service := &Service{
		db:     db,
		llm:    llm,
		config: *config,

		Messages: NewMessageService(db),
		Summary:  NewSummaryService(db),
		Agents:   NewAgentService(db),
		Users:    NewUserService(db),

		summarizationTicker: time.NewTicker(config.Summary.SummaryDaemonInterval * time.Second),
		knowledgeTicker:     time.NewTicker(60 * time.Second),
	}

	return service
}

func NewServiceFromConfig(config *config.Config) (*Service, error) {
	db, err := NewStoreFromConfig(config)
	if err != nil {
		return nil, err
	}

	llm, err := NewLLMFromConfig(config)
	if err != nil {
		return nil, err
	}

	return NewService(db, llm, config), nil
}

func (service *Service) InitDaemons() {
	if service.config.Summary.SummaryDaemonInterval > 0 {
		go func() {
			for {
				<-service.summarizationTicker.C
				service.SummaryDaemon()
			}
		}()
	}
}
