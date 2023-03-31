package llm

import (
	"github.com/hlfshell/coppermind/internal/chat"
	"github.com/hlfshell/coppermind/internal/memory"
)

type LLM interface {
	SendMessage(
		instructions []*chat.Prompt,
		identity []*chat.Prompt,
		conversation *chat.Conversation,
		previousConversations []*memory.Summary,
		summary *memory.Summary,
		message *chat.Message,
	) (*chat.Response, error)
	Summarize(
		instructions []*chat.Prompt,
		history *chat.Conversation,
		previousSummary *memory.Summary,
	) (*memory.Summary, error)
	Learn(
		instructions []*chat.Prompt,
		history *chat.Conversation,
	) ([]*memory.Knowledge, error)
}
