package service

import (
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

func (service *MessageService) GetMessages(filter store.Filter) ([]*chat.Message, error) {
	return service.db.ListMessages(filter)
}

func (service *MessageService) DeleteMessage(id string) error {
	return service.db.DeleteMessage(id)
}

func (service *MessageService) GetConversation(conversationId string) (*chat.Conversation, error) {
	return service.db.GetConversation(conversationId)
}

func (service *MessageService) GetConversations(filter store.Filter) ([]*chat.Conversation, error) {
	return service.db.ListConversations(filter)
}

func (service *MessageService) DeleteConversation(id string) error {
	return service.db.DeleteConversation(id)
}
