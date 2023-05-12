package service

import (
	"fmt"
	"time"

	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/memory"
)

func (service *Service) SummaryDaemon() error {
	conversations, err := service.db.GetConversationsToSummarize(
		service.config.Summary.MinMessagesToSummarize,
		time.Duration(service.config.Summary.MinConversationTimeToWaitSeconds)*time.Second,
		service.config.Summary.MinMessagesToForceSummarization,
	)
	if err != nil {
		return err
	}

	for _, conversation := range conversations {
		_, err := service.Summarize(conversation)
		if err != nil {
			return err
		}
	}

	return nil
}

func (service *Service) Summarize(conversationId string) (*memory.Summary, error) {
	// First we get all of the messages in that conversation
	// that we'll be trying to summarize
	conversation, err := service.db.GetConversation(conversationId)
	if err != nil {
		return nil, err
	} else if conversation == nil {
		return nil, fmt.Errorf("conversation %s not found", conversationId)
	}

	//Determine if a summary already exists for this conversation
	var existingSummary *memory.Summary
	summaries, err := service.db.ListSummaries(store.Filter{
		Attributes: []*store.FilterAttribute{
			{
				Attribute: "conversation",
				Value:     conversation.ID,
				Operation: store.EQ,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(summaries) != 0 {
		existingSummary = summaries[0]
	}

	// Ask the llm to generate the summaries
	summary, err := service.llm.Summarize(conversation, existingSummary)
	if err != nil {
		return nil, err
	} else if summary == nil {
		service.db.ExcludeConversationFromSummary(conversationId)
		return nil, nil
	}

	err = service.db.SaveSummary(summary)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

type SummaryService struct {
	db store.Store
}

func NewSummaryService(db store.Store) *SummaryService {
	return &SummaryService{
		db: db,
	}
}

func (service *SummaryService) GetSummary(id string) (*memory.Summary, error) {
	return service.db.GetSummary(id)
}

type GetSummariesRequest struct {
	Agent  string
	User   string
	Time   time.Time
	Before bool
	Limit  int
}

func (request *GetSummariesRequest) Valid() error {
	if request.Agent == "" {
		return fmt.Errorf("agent cannot be empty")
	}
	if request.User == "" {
		return fmt.Errorf("user cannot be empty")
	}
	if request.Time.IsZero() {
		return fmt.Errorf("time must be set")
	}
	return nil
}

func (request *GetSummariesRequest) GetFilters() ([]*store.FilterAttribute, error) {
	err := request.Valid()
	if err != nil {
		return nil, err
	}

	attributes := []*store.FilterAttribute{
		{
			Attribute: "agent",
			Value:     request.Agent,
			Operation: store.EQ,
		},
		{
			Attribute: "user",
			Value:     request.User,
			Operation: store.EQ,
		},
	}

	if request.Before {
		attributes = append(attributes, &store.FilterAttribute{
			Attribute: "conversation_started_at",
			Value:     request.Time,
			Operation: store.LT,
		})
	} else {
		attributes = append(attributes, &store.FilterAttribute{
			Attribute: "conversation_started_at",
			Value:     request.Time,
			Operation: store.GT,
		})
	}

	return attributes, nil
}

func (service *SummaryService) GetSummaries(request *GetSummariesRequest) ([]*memory.Summary, error) {
	filters, err := request.GetFilters()
	if err != nil {
		return nil, err
	}

	return service.db.ListSummaries(store.Filter{
		Attributes: filters,
		Limit:      request.Limit,
		OrderBy: store.OrderBy{
			Attribute: "conversation_started_at",
			Ascending: request.Before,
		},
	})
}

func (service *SummaryService) DeleteSummary(id string) error {
	return service.db.DeleteSummary(id)
}
