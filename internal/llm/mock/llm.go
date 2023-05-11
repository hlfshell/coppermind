package mock

import (
	"fmt"

	"github.com/hlfshell/coppermind/pkg/agents"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
)

type MockLLM struct {
	sendMessageResponses []*chat.Message
	sendMessageErrors    []error
	sendMessageInputs    []interface{}

	conversationContinuanceResponses []bool
	conversationContinuanceErrors    []error
	conversationContinuanceInputs    []interface{}

	summarizeResponses []*memory.Summary
	summarizeErrors    []error
	summarizeInputs    []interface{}

	learnResponses [][]*memory.Knowledge
	learnErrors    []error
	learnInputs    []interface{}

	charsPerToken int
}

func NewMockLLM() *MockLLM {
	return &MockLLM{
		sendMessageResponses: []*chat.Message{},
		sendMessageErrors:    []error{},
		sendMessageInputs:    []interface{}{},

		conversationContinuanceResponses: []bool{},
		conversationContinuanceErrors:    []error{},
		conversationContinuanceInputs:    []interface{}{},

		summarizeResponses: []*memory.Summary{},
		summarizeErrors:    []error{},
		summarizeInputs:    []interface{}{},

		learnResponses: [][]*memory.Knowledge{},
		learnErrors:    []error{},
		learnInputs:    []interface{}{},

		charsPerToken: 4,
	}
}

func (llm *MockLLM) ClearMemory() {
	llm.sendMessageResponses = []*chat.Message{}
	llm.sendMessageErrors = []error{}
	llm.sendMessageInputs = []interface{}{}

	llm.conversationContinuanceResponses = []bool{}
	llm.conversationContinuanceErrors = []error{}
	llm.conversationContinuanceInputs = []interface{}{}

	llm.summarizeResponses = []*memory.Summary{}
	llm.summarizeErrors = []error{}
	llm.summarizeInputs = []interface{}{}

	llm.learnResponses = [][]*memory.Knowledge{}
	llm.learnErrors = []error{}
	llm.learnInputs = []interface{}{}
}

func (llm *MockLLM) SetCharsPerToken(charsPerToken int) {
	llm.charsPerToken = charsPerToken
}

func (llm *MockLLM) AddSendMessageResponse(msg *chat.Message, err error) {
	llm.sendMessageResponses = append(llm.sendMessageResponses, msg)
	llm.sendMessageErrors = append(llm.sendMessageErrors, err)
}

func (llm *MockLLM) GetSendMessageInputs() (*agents.Agent, *chat.Conversation, []*memory.Summary, []*memory.Knowledge, *chat.Message) {
	if len(llm.sendMessageInputs) == 0 {
		return nil, nil, nil, nil, nil
	}

	// Pop the correct amount if tems from the sendMessageInputs and return
	// them typecasted to the correct type
	agent := llm.sendMessageInputs[0].(*agents.Agent)
	conversation := llm.sendMessageInputs[1].(*chat.Conversation)
	previousConversations := llm.sendMessageInputs[2].([]*memory.Summary)
	knowledge := llm.sendMessageInputs[3].([]*memory.Knowledge)
	message := llm.sendMessageInputs[4].(*chat.Message)

	llm.sendMessageInputs = llm.sendMessageInputs[5:]

	return agent, conversation, previousConversations, knowledge, message
}

func (llm *MockLLM) AddConversationContinuanceResponse(continueConversation bool, err error) {
	llm.conversationContinuanceResponses = append(llm.conversationContinuanceResponses, continueConversation)
	llm.conversationContinuanceErrors = append(llm.conversationContinuanceErrors, err)
}

func (llm *MockLLM) GetConversationContinuanceInputs() (*chat.Message, *chat.Conversation, *memory.Summary) {
	if len(llm.conversationContinuanceInputs) == 0 {
		return nil, nil, nil
	}

	// Pop the correct amount if tems from the sendMessageInputs and return
	// them typecasted to the correct type
	message := llm.conversationContinuanceInputs[0].(*chat.Message)
	conversation := llm.conversationContinuanceInputs[1].(*chat.Conversation)
	summary := llm.conversationContinuanceInputs[2].(*memory.Summary)

	llm.conversationContinuanceInputs = llm.conversationContinuanceInputs[3:]

	return message, conversation, summary
}

func (llm *MockLLM) AddSummarizeResponse(summary *memory.Summary, err error) {
	llm.summarizeResponses = append(llm.summarizeResponses, summary)
	llm.summarizeErrors = append(llm.summarizeErrors, err)
}

func (llm *MockLLM) GetSummarizeInputs() (*chat.Conversation, *memory.Summary) {
	if len(llm.summarizeInputs) == 0 {
		return nil, nil
	}

	// Pop the correct amount if tems from the sendMessageInputs and return
	// them typecasted to the correct type
	history := llm.summarizeInputs[0].(*chat.Conversation)
	summary := llm.summarizeInputs[1].(*memory.Summary)

	llm.summarizeInputs = llm.summarizeInputs[2:]

	return history, summary
}

func (llm *MockLLM) AddLearnResponse(knowledge []*memory.Knowledge, err error) {
	llm.learnResponses = append(llm.learnResponses, knowledge)
	llm.learnErrors = append(llm.learnErrors, err)
}

func (llm *MockLLM) GetLearnInputs() (*chat.Conversation, *memory.Summary) {
	if len(llm.learnInputs) == 0 {
		return nil, nil
	}

	// Pop the correct amount if tems from the sendMessageInputs and return
	// them typecasted to the correct type
	history := llm.learnInputs[0].(*chat.Conversation)
	summary := llm.learnInputs[1].(*memory.Summary)

	llm.learnInputs = llm.learnInputs[2:]

	return history, summary
}

func (llm *MockLLM) SendMessage(
	agent *agents.Agent,
	conversation *chat.Conversation,
	previousConversations []*memory.Summary,
	knowledge []*memory.Knowledge,
	message *chat.Message,
) (*chat.Message, error) {
	if len(llm.sendMessageResponses) == 0 {
		return nil, fmt.Errorf("no mocked responses included")
	}

	llm.sendMessageInputs = append(llm.sendMessageInputs, agent, conversation, previousConversations, knowledge, message)

	response := llm.sendMessageResponses[0]
	llm.sendMessageResponses = llm.sendMessageResponses[1:]

	err := llm.sendMessageErrors[0]
	llm.sendMessageErrors = llm.sendMessageErrors[1:]

	return response, err
}

func (llm *MockLLM) ConversationContinuance(
	message *chat.Message,
	conversation *chat.Conversation,
	summary *memory.Summary,
) (bool, error) {
	if len(llm.conversationContinuanceResponses) == 0 {
		return false, fmt.Errorf("no mocked responses included")
	}

	llm.conversationContinuanceInputs = append(llm.conversationContinuanceInputs, message, conversation, summary)

	response := llm.conversationContinuanceResponses[0]
	llm.conversationContinuanceResponses = llm.conversationContinuanceResponses[1:]

	err := llm.conversationContinuanceErrors[0]
	llm.conversationContinuanceErrors = llm.conversationContinuanceErrors[1:]

	return response, err
}

func (llm *MockLLM) Summarize(
	history *chat.Conversation,
	previousSummary *memory.Summary,
) (*memory.Summary, error) {
	if len(llm.summarizeResponses) == 0 {
		return nil, fmt.Errorf("no mocked responses included")
	}

	llm.summarizeInputs = append(llm.summarizeInputs, history, previousSummary)

	response := llm.summarizeResponses[0]
	llm.summarizeResponses = llm.summarizeResponses[1:]

	err := llm.summarizeErrors[0]
	llm.summarizeErrors = llm.summarizeErrors[1:]

	return response, err
}

func (llm *MockLLM) Learn(
	history *chat.Conversation,
	summary *memory.Summary,
) ([]*memory.Knowledge, error) {
	if len(llm.learnResponses) == 0 {
		return nil, fmt.Errorf("no mocked responses included")
	}

	llm.learnInputs = append(llm.learnInputs, history, summary)

	response := llm.learnResponses[0]
	llm.learnResponses = llm.learnResponses[1:]

	err := llm.learnErrors[0]
	llm.learnErrors = llm.learnErrors[1:]

	return response, err
}

func (llm *MockLLM) EstimateTokens(input string) int {
	return len(input) / llm.charsPerToken
}
