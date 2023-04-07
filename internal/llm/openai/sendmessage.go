package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/hlfshell/coppermind/internal/prompts"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/sashabaranov/go-openai"
	"github.com/wissance/stringFormatter"
)

func (ai *OpenAI) SendMessage(
	identity string,
	conversation *chat.Conversation,
	previousConversations []*memory.Summary,
	knowledge []*memory.Knowledge,
	message *chat.Message,
) (*chat.Response, error) {
	data, err := ai.prepareChatMessage(
		ai.chatPrompt,
		identity,
		conversation.Messages,
		previousConversations,
		knowledge,
		message,
	)
	if err != nil {
		return nil, err
	}

	resp, err := ai.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{data},
		},
	)

	if err != nil {
		return nil, err
	} else if len(resp.Choices) < 1 {
		return nil, OpenAIResponseError{msg: "No proper response returned"}
	}

	fmt.Println("Token usage", resp.Usage)
	fmt.Println(resp.Choices[0].Message.Content)

	return &chat.Response{
		Name:    "Rose",
		Tone:    "neutral",
		Content: resp.Choices[0].Message.Content,
	}, nil
}

func (ai *OpenAI) prepareChatMessage(
	prompt string,
	identity string,
	history []*chat.Message,
	previousConversations []*memory.Summary,
	knowledge []*memory.Knowledge,
	message *chat.Message,
) (openai.ChatCompletionMessage, error) {
	var tokenCount int

	tokenCount += ai.EstimateTokens(prompt)

	tokenCount += ai.EstimateTokens(identity)

	var previousSummary *memory.Summary
	summariesString := ""
	if len(previousConversations) > 0 {
		for index, summary := range previousConversations {
			if index > 0 {
				summariesString += ","
			}
			// Do not include the summary for this conversation
			// if it's included
			if summary.Conversation != message.Conversation {
				summariesString += summary.String()
			} else {
				previousSummary = summary
			}
		}
		tokenCount += ai.EstimateTokens(summariesString)
	}

	previousSummaryString := ""
	if previousSummary != nil {
		previousSummaryString = prompts.PreviousSummary
		previousSummaryString += previousSummary.String()
		tokenCount += ai.EstimateTokens(previousSummaryString)
	}

	var knowledgeString string
	if len(knowledge) > 0 {
		knowledgeString = "The following facts have been extracted from prior converastions and should be considered when forming your responses."

		for index, fact := range knowledge {
			if index > 0 {
				knowledgeString += "\n"
			}
			knowledgeString += fact.String()
		}

		tokenCount += ai.EstimateTokens(knowledgeString)
	}

	var messagesHistory string
	if len(history) > 0 {
		messagesHistory = `The following is the message log, where it shares when the message occured, who is talking, and the message itself, delimited by the "|" character.`
		tokenCount += ai.EstimateTokens(messagesHistory)
	}
	for _, msg := range history {
		content := msg.DatedString()
		messagesHistory += content + "\n"
		tokenCount += ai.EstimateTokens(content)
	}

	content := message.DatedString() + "\n"
	messagesHistory += content
	tokenCount += ai.EstimateTokens(content)

	prompt = stringFormatter.FormatComplex(
		prompt,
		map[string]interface{}{
			"name":             "Rose",
			"identity":         identity,
			"summaries":        summariesString,
			"knowledge":        knowledgeString,
			"previous_summary": previousSummaryString,
			"message_history":  messagesHistory,
		},
	)

	fmt.Println("output", prompt)
	fmt.Println("estimated tokens", tokenCount)

	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: prompt,
	}, nil
}

func (ai *OpenAI) ConversationContinuance(
	msg *chat.Message,
	conversation *chat.Conversation,
	summary *memory.Summary,
) (bool, error) {
	// data, err := ai.prepareConversationContinuanceMessage(instructions, conversation, summary)
	data, err := ai.prepareConversationContinuanceMessage(
		prompts.ConversationContinuance,
		msg,
		conversation,
		summary,
	)
	if err != nil {
		return false, err
	}
	fmt.Println("Continuance prompt")
	fmt.Println(data)

	resp, err := ai.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{{
				Role:    openai.ChatMessageRoleSystem,
				Content: data,
			}},
		},
	)
	if err != nil {
		return false, err
	}
	fmt.Println("continuance test", resp)
	return ai.parseConversationContinuanceResponse(resp.Choices[0].Message.Content), nil
}

func (ai *OpenAI) prepareConversationContinuanceMessage(
	instructions string,
	msg *chat.Message,
	conversation *chat.Conversation,
	summary *memory.Summary,
) (string, error) {
	tokenCount := 0

	var output string

	output += instructions
	tokenCount += ai.EstimateTokens(output)

	var summaryText string
	if summary != nil {
		summaryText := "Summary:\n"
		summaryText += summary.Summary + "\n"
		tokenCount += ai.EstimateTokens(summaryText)
	}

	// We need to iterate through conversation.Messages
	// in reverse, not allowing the tokenCount to
	// exceed ai.maxTokens
	messageContent := ""
	for i := len(conversation.Messages) - 1; i >= 0; i-- {
		targetMessage := conversation.Messages[i]
		content := targetMessage.SimpleString()
		content = content + "\n"
		if tokenCount+ai.EstimateTokens(content) > ai.maxInput {
			break
		}
		// Because of the reverse order, we wish to prepend the
		// new message to our message history content
		tokenCount += ai.EstimateTokens(content)
		messageContent = fmt.Sprintf("%s%s", content, messageContent)
	}

	output = stringFormatter.FormatComplex(
		output,
		map[string]interface{}{
			"summary":         summaryText,
			"message_history": messageContent,
			"new_message":     msg.SimpleString(),
		},
	)

	return output, nil
}

func (ai *OpenAI) parseConversationContinuanceResponse(raw string) bool {
	return strings.ToLower(raw) == "true"
}
