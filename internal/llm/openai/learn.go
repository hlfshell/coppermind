package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/sashabaranov/go-openai"
)

func (ai *OpenAI) Learn(
	history *chat.Conversation,
	summary *memory.Summary,
) ([]*memory.Knowledge, error) {
	data, err := ai.prepareLearnMessage(
		ai.knowledgePrompt,
		history,
		summary,
	)
	fmt.Println("prepped")
	// fmt.Println(data)
	if err != nil {
		return nil, err
	}

	resp, err := ai.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: data,
		},
	)
	if err != nil {
		return nil, err
	}
	fmt.Println("Knowledge extacted", resp.Usage)
	fmt.Println("resp", resp.Choices[0].Message.Content)
	facts, err := ai.parseLearnResponse(history, resp.Choices[0].Message.Content)
	if err != nil {
		return nil, err
	}
	fmt.Println("facts", facts)
	return facts, nil
}

func (ai *OpenAI) prepareLearnMessage(
	instructions string,
	conversation *chat.Conversation,
	summary *memory.Summary,
) ([]openai.ChatCompletionMessage, error) {
	content := strings.Builder{}

	content.WriteString(instructions)

	if summary != nil {
		content.WriteString("Summary: ")
		content.WriteString(summary.Summary)
		content.WriteString("\n")
	}

	content.WriteString("Conversation History:\n")

	for _, msg := range conversation.Messages {
		content.WriteString(msg.SimpleString() + "\n")
	}

	content.WriteString("Output:\n")

	return []openai.ChatCompletionMessage{{
		Role:    openai.ChatMessageRoleSystem,
		Content: content.String(),
	}}, nil
}

func (ai *OpenAI) parseLearnResponse(
	conversation *chat.Conversation,
	raw string,
) ([]*memory.Knowledge, error) {
	var responses []*memory.LearnResponse

	err := json.Unmarshal([]byte(raw), &responses)
	if err != nil {
		return nil, err
	}
	derivedFacts := []*memory.Knowledge{}

	for _, response := range responses {
		fact, err := memory.ToKnowledge(
			response,
			conversation.Agent,
			conversation.User,
			conversation.ID,
		)
		if err != nil {
			return nil, err
		}
		derivedFacts = append(derivedFacts, fact)
	}

	return derivedFacts, err
}
