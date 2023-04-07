package llm

import (
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
)

type LLM interface {
	/*
		SendMessage will send a new message in a given conversation and
		generate a response per the identity of the agent.
	*/
	SendMessage(
		identity string,
		conversation *chat.Conversation,
		previousConversations []*memory.Summary,
		knowledge []*memory.Knowledge,
		message *chat.Message,
	) (*chat.Response, error)

	/*
		ConversationContinuance attempts to determine whether or not
		a given new message is a continuance of the previous conversation
		or a new conversation entirely.
	*/
	ConversationContinuance(
		message *chat.Message,
		conversation *chat.Conversation,
		summary *memory.Summary,
	) (bool, error)

	/*
		Summarize will, given a conversation and possibly a previous
		summary, attempt to create a set of keywords and single sentence
		summary of the conversation to aid the agent in remembering past
		conversations.
	*/
	Summarize(
		history *chat.Conversation,
		previousSummary *memory.Summary,
	) (*memory.Summary, error)

	/*
		Learn will take a conversation and summary and attempt to
		extract short knowledge sets from the conversation to improve
		the agent's memory.
	*/
	Learn(
		history *chat.Conversation,
		summary *memory.Summary,
	) ([]*memory.Knowledge, error)

	/*
		EstimateTokens will take a string and attempt to estimate
		the total token count. This is generally a rule of thumb
		operation
	*/
	EstimateTokens(text string) int
}
