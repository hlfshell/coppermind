package store

import (
	"time"

	"github.com/hlfshell/coppermind/internal/chat"
	"github.com/hlfshell/coppermind/internal/memory"
)

type Store interface {
	SaveMessage(msg *chat.Message) error
	GetLatestConversation(user string) (string, time.Time, error)
	LoadConversationHistory(conversation string) ([]*chat.Message, error)
	GetSummaryByConversation(conversation string) (*memory.Summary, error)
	GetSummariesByUser(user string) ([]*memory.Summary, error)
	SaveSummary(summary *memory.Summary) error
	Migrate() error
}
