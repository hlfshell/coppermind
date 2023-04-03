package llm

import (
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
)

type LLM interface {
	SendMessage(
		instructions []*chat.Prompt,
		identity []*chat.Prompt,
		conversation *chat.Conversation,
		previousConversations []*memory.Summary,
		message *chat.Message,
	) (*chat.Response, error)
	ConversationContinuance(
		instructions []*chat.Prompt,
		conversation *chat.Conversation,
		summary *memory.Summary,
	) (bool, error)
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
