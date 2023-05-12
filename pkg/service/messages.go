package service

import (
	"fmt"
	"time"

	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/chat"
)

type MessageService struct {
	db store.Store
}

func NewMessageService(db store.Store) *MessageService {
	return &MessageService{
		db: db,
	}
}

func (service *MessageService) GetMessage(id string) (*chat.Message, error) {
	return service.db.GetMessage(id)
}

type GetMessagesRequest struct {
	Agent  string
	User   string
	Time   time.Time
	Before bool
	Limit  int
}

func (request *GetMessagesRequest) Valid() error {
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

func (request *GetMessagesRequest) getFilters() ([]*store.FilterAttribute, error) {
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

func (service *MessageService) GetMessages(request *GetMessagesRequest) ([]*chat.Message, error) {
	filters, err := request.getFilters()
	if err != nil {
		return nil, err
	}

	return service.db.ListMessages(store.Filter{
		Attributes: filters,
		OrderBy: store.OrderBy{
			Attribute: "created_at",
			Ascending: false,
		},
		Limit: request.Limit,
	})
}

func (service *MessageService) DeleteMessage(id string) error {
	return service.db.DeleteMessage(id)
}

func (service *MessageService) GetConversation(conversationId string) (*chat.Conversation, error) {
	return service.db.GetConversation(conversationId)
}

type GetConversationsRequest struct {
	Agent  string
	User   string
	Time   time.Time
	Before bool
	Limit  int
}

func (request *GetConversationsRequest) Valid() error {
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

func (request *GetConversationsRequest) getFilters() ([]*store.FilterAttribute, error) {
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

func (service *MessageService) GetConversations(request *GetConversationsRequest) ([]*chat.Conversation, error) {
	filters, err := request.getFilters()
	if err != nil {
		return nil, err
	}

	return service.db.ListConversations(store.Filter{
		Attributes: filters,
		OrderBy: store.OrderBy{
			Attribute: "created_at",
			Ascending: false,
		},
		Limit: request.Limit,
	})
}

func (service *MessageService) DeleteConversation(id string) error {
	return service.db.DeleteConversation(id)
}
