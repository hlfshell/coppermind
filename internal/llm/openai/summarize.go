package openai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/prompts"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/sashabaranov/go-openai"
	"github.com/wissance/stringFormatter"
)

func (ai *OpenAI) Summarize(
	conversation *chat.Conversation,
	previousSummary *memory.Summary,
) (*memory.Summary, error) {
	data, lastMessage, err := ai.prepareSummaryMessage(ai.summaryPrompt, conversation, previousSummary)
	fmt.Println("prepped")
	fmt.Println(data)
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
	} else if len(resp.Choices) < 1 {
		return nil, OpenAIResponseError{msg: "No proper response returned"}
	}

	fmt.Println(resp.Usage)

	summary, err := ai.parseSummaryResponse(conversation, resp.Choices[0].Message.Content)
	if err != nil {
		return nil, err
	} else if lastMessage.ID != conversation.Messages[len(conversation.Messages)-1].ID {
		// If our lastMessage is NOT the last message in the conversation,
		// then we were cut short due to the token limit. We need to update
		// the summary to reflect that it's essentially not tracking any
		// messages past that point in time.
		summary.UpdatedAt = lastMessage.CreatedAt
	}

	return summary, nil
}

func (ai *OpenAI) prepareSummaryMessage(
	instructions string,
	conversation *chat.Conversation,
	previousSummary *memory.Summary,
) ([]openai.ChatCompletionMessage, *chat.Message, error) {
	var tokenCount int
	var output string

	output += instructions + "\n"
	tokenCount += ai.EstimateTokens(output)

	//Handle the case of an existing summary already exists for the summary
	if previousSummary != nil {
		previousSummaryText := stringFormatter.Format(prompts.ExistingSummary, map[string]string{
			"summary": previousSummary.String(),
		})
		tokenCount += ai.EstimateTokens(previousSummaryText)
		output += previousSummaryText + "\n"
	}

	contents := []string{}
	var start int
	var targetMessage *chat.Message

	for {
		if start >= len(conversation.Messages) {
			break
		}

		targetMessage = conversation.Messages[start]
		content := targetMessage.SimpleString()

		tokens := ai.EstimateTokens(content)

		if tokenCount+tokens > ai.maxInput {
			break
		} else {
			contents = append(contents, content)
			tokenCount += tokens
		}
		start++
	}

	for _, content := range contents {
		output += content + "\n"
	}

	msgs := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: output,
		},
	}

	return msgs, targetMessage, nil
}

func (ai *OpenAI) parseSummaryResponse(conversation *chat.Conversation, raw string) (*memory.Summary, error) {
	split := strings.Split(raw, "|")
	//Validate and protect this
	split[0] = strings.TrimSpace(split[0])
	split[1] = strings.TrimSpace(split[1])

	if split[0] == "none" {
		return nil, nil
	}

	summary := &memory.Summary{
		ID:           uuid.New().String(),
		Agent:        conversation.Agent,
		Conversation: conversation.ID,
		Summary:      split[1],
		User:         conversation.User,
		UpdatedAt:    time.Now(),
	}

	summary.StringToKeywords(split[0])
	return summary, nil
}
