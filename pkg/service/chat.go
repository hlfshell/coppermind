package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
)

func (service *Service) SendMessage(msg *chat.Message) (*chat.Message, error) {
	// Get the agent for the message
	agent, err := service.db.GetAgent(msg.Agent)
	if err != nil {
		return nil, err
	} else if agent == nil {
		return nil, fmt.Errorf("agent %s not found", msg.Agent)
	}

	// If no conversation is set, lookup to see if we have an old conversation
	// that we can load up and join (based on how long since it's been) the
	// last message in that conversation
	if msg.Conversation == "" {
		conversationId, err := service.generateOrFindConversation(msg)
		if err != nil {
			return nil, err
		}
		msg.Conversation = conversationId
		fmt.Println("Chose conversation: ", msg.Conversation)
	}

	// Get the current conversation history if any exists
	conversation, err := service.db.GetConversation(msg.Conversation)
	if err != nil {
		return nil, err
	} else if conversation == nil {
		conversation = &chat.Conversation{
			ID:        msg.Conversation,
			Agent:     msg.Agent,
			User:      msg.User,
			CreatedAt: msg.CreatedAt,
			Messages:  []*chat.Message{msg},
		}
	}

	// Find the summaries of prior conversations if any exist
	pastSummaries, err := service.previousSummaries(msg.Agent, msg.User)
	if err != nil {
		return nil, err
	}

	// Placeholder for anything to do with injected knowledge
	// here
	knowledge := []*memory.Knowledge{}

	// Now we have the LLM deal with the message
	response, err := service.llm.SendMessage(
		agent,
		conversation,
		pastSummaries,
		knowledge,
		msg,
	)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *Service) generateOrFindConversation(msg *chat.Message) (string, error) {
	conversation, timestamp, err := service.db.GetLatestConversation(msg.Agent, msg.User)
	if err != nil {
		return "", err
	} else if time.Now().Add(time.Duration(-1*service.config.Chat.ConversationMaintainanceDurationSeconds) * time.Second).Before(timestamp) {
		// This handles the conversation existing, and it being within the maintain
		// window of a conversation
		return conversation, nil
	} else if conversation == "" {
		// This handles the situation of there is no existing conversation at all
		conversation = uuid.New().String()
	} else if time.Now().Add(time.Duration(-1*service.config.Chat.MaxConversationIdleTimeSeconds) * time.Second).After(timestamp) {
		// We have a conversation that exists, but it's beyond the allowed time to
		// continue the conversation; therefore it's an automatic new one
		conversation = uuid.New().String()
	} else {
		// Finally, this sees via the LLM if it should be a conversation
		// continuance or a new conversation. This occurs if were beyond the
		// auto-continue time of ConversationMaintainanceDuration, but within
		// the MaxConversationIdleTime.
		retrievedConversation, err := service.db.GetConversation(conversation)
		if err != nil {
			return "", err
		}
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
			return "", err
		}
		var summary *memory.Summary
		if len(summaries) != 0 {
			summary = summaries[0]
		}

		if summary == nil {
			return uuid.New().String(), nil
		}

		shouldContinue, err := service.llm.ConversationContinuance(
			msg,
			retrievedConversation,
			summary,
		)
		if err != nil {
			return "", nil
		} else if shouldContinue {
			return conversation, nil
		} else {
			return uuid.New().String(), nil
		}
	}
	return conversation, nil
}

func (service Service) previousSummaries(agent string, user string) ([]*memory.Summary, error) {
	return service.db.ListSummaries(store.Filter{
		Attributes: []*store.FilterAttribute{
			{
				Attribute: "agent",
				Value:     agent,
				Operation: store.EQ,
			},
			{
				Attribute: "user",
				Value:     user,
				Operation: store.EQ,
			},
		},
		Limit: service.config.Chat.MaxSummariesToInclude,
		OrderBy: store.OrderBy{
			Attribute: "conversation_started_at",
			Ascending: false,
		},
	})
}

type GetRecentConversationsRequest struct {
	Agent      string
	User       string
	Time       time.Time
	Before     bool
	Limit      int
	MostRecent bool
}

func (request *GetRecentConversationsRequest) Valid() error {
	if request.Agent == "" {
		return fmt.Errorf("agent cannot be empty")
	}
	if request.User == "" {
		return fmt.Errorf("user cannot be empty")
	}
	if request.Time.IsZero() {
		return fmt.Errorf("time must be set")
	}
}

func (request *GetRecentConversationsRequest) GetFilters() ([]*store.FilterAttribute, error) {
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

	var operation string
	if request.Before {
		operation = store.LTE
	} else {
		operation = store.GTE
	}

	attributes = append(
		attributes,
		&store.FilterAttribute{
			Attribute: "created_at",
			Value:     request.Time,
			Operation: operation,
		},
	)

	return attributes, nil
}

func (service *Service) GetRecentConversations(request *GetRecentConversationsRequest) ([]*chat.Conversation, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	attributes, err := request.GetFilters()
	if err != nil {
		return nil, err
	}

	return service.db.ListConversations(store.Filter{
		Attributes: attributes,
		OrderBy: store.OrderBy{
			Attribute: "created_at",
			Ascending: true,
		},
	})
}
