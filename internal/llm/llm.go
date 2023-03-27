package llm

import (
	"github.com/hlfshell/coppermind/internal/chat"
	"github.com/hlfshell/coppermind/internal/memory"
)

type LLM interface {
	SendMessage(instructions []*chat.Prompt, identity []*chat.Prompt, history []*chat.Message, message *chat.Message) (*chat.Response, error)
	Summarize(instructions []*chat.Prompt, history []*chat.Message, previousSummary *memory.Summary) (*memory.Summary, error)
	Learn(instructions []*chat.Prompt, history []*chat.Message) ([]*memory.Knowledge, error)
}
