package service

import (
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
	}

	//Determine if a summary already exists for this conversation
	var existingSummary *memory.Summary
	summaries, err := service.db.ListSummaries(store.Filter{
		Attributes: []*store.FilterAttribute{
			{
				Attribute: "conversation",
				Value:     conversation,
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
