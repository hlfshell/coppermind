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

func (llm *mockLLM) SendMessage(identity string, conversation *chat.Conversation, previousConversations []*memory.Summary, knowledge []*memory.Knowledge, message *chat.Message) (*chat.Response, error) {
	return llm.sendMessageResponse, llm.sendMessageError
}

func (llm *mockLLM) ConversationContinuance(
	message *chat.Message,
	conversation *chat.Conversation,
	summary *memory.Summary,
) (bool, error) {
	return llm.conversationContinuanceResponse, llm.conversationContinuanceError
}

func (llm *mockLLM) Summarize(
	history *chat.Conversation,
	previousSummary *memory.Summary,
) (*memory.Summary, error) {
	return nil, nil
}

func (llm *mockLLM) Learn(
	history *chat.Conversation,
	summary *memory.Summary,
) ([]*memory.Knowledge, error) {
	return nil, nil
}

func (llm *mockLLM) EstimateTokens(input string) int {
	return int(len(input) / 4)
}
