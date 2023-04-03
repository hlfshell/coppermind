package agent

import (
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
)

// mockLLM is a mock LLM struct for testing purposes
type mockLLM struct {
	sendMessageResponse *chat.Response
	sendMessageError    error

	conversationContinuanceResponse bool
	conversationContinuanceError    error
}

func (llm *mockLLM) SendMessage(instructions []*chat.Prompt, identity []*chat.Prompt, conversation *chat.Conversation, previousConversations []*memory.Summary, message *chat.Message) (*chat.Response, error) {
	return llm.sendMessageResponse, llm.sendMessageError
}

func (llm *mockLLM) ConversationContinuance(
	instructions []*chat.Prompt,
	conversation *chat.Conversation,
	summary *memory.Summary,
) (bool, error) {
	return llm.conversationContinuanceResponse, llm.conversationContinuanceError
}

func (llm *mockLLM) Summarize(
	instructions []*chat.Prompt,
	history *chat.Conversation,
	previousSummary *memory.Summary,
) (*memory.Summary, error) {
	return nil, nil
}

func (llm *mockLLM) Learn(
	instructions []*chat.Prompt,
	history *chat.Conversation,
) ([]*memory.Knowledge, error) {
	return nil, nil
}
