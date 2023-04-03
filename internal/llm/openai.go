package llm

import (
	"context"
	"encoding/json"
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

type OpenAI struct {
	apiKey string
	client *openai.Client
}

type OpenAIResponseError struct {
	msg string
}

func (err OpenAIResponseError) Error() string {
	return err.msg
}

func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{
		apiKey: apiKey,
		client: openai.NewClient(apiKey),
	}
}

func (ai *OpenAI) SendMessage(instructions []*chat.Prompt, identity []*chat.Prompt, conversation *chat.Conversation, previousConversations []*memory.Summary, message *chat.Message) (*chat.Response, error) {
	data, err := ai.prepareChatMessage(
		instructions,
		identity,
		conversation.Messages,
		previousConversations,
		message,
	)
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

	return ai.parseChatResponse(resp.Choices[0].Message.Content)
}

func (ai *OpenAI) prepareChatMessage(
	instructions []*chat.Prompt,
	identity []*chat.Prompt,
	history []*chat.Message,
	previousConversations []*memory.Summary,
	message *chat.Message,
) ([]openai.ChatCompletionMessage, error) {
	msgs := []openai.ChatCompletionMessage{}

	for _, instruction := range instructions {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: instruction.Content,
		})
	}

	for _, fact := range identity {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: fact.Content,
		})
	}

	var previousSummary *memory.Summary
	if len(previousConversations) > 0 {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompts.SummaryIncluded,
		})
		summariesString := ""
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
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: summariesString,
		})
	}

	if previousSummary != nil {
		msgs = append(
			msgs,
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompts.PreviousSummary,
			},
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: previousSummary.String(),
			},
		)
	}

	for _, msg := range history {
		content, err := msg.JSON()
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: content,
		})
	}

	lastMsg, err := message.JSON()
	if err != nil {
		return nil, err
	}

	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: lastMsg,
	})

	return msgs, nil
}

func (ai *OpenAI) parseChatResponse(raw string) (*chat.Response, error) {
	fmt.Println(raw)
	msg := &chat.Response{}
	err := json.Unmarshal([]byte(raw), msg)

	return msg, err
}

func (ai *OpenAI) Summarize(
	instructions []*chat.Prompt,
	conversation *chat.Conversation,
	previousSummary *memory.Summary,
) (*memory.Summary, error) {
	data, err := ai.prepareSummaryMessage(instructions, conversation, previousSummary)
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

	return ai.parseSummaryResponse(conversation, resp.Choices[0].Message.Content)
}

func (ai *OpenAI) prepareSummaryMessage(
	instructions []*chat.Prompt,
	conversation *chat.Conversation,
	previousSummary *memory.Summary,
) ([]openai.ChatCompletionMessage, error) {
	msgs := []openai.ChatCompletionMessage{}

	for _, instruction := range instructions {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: instruction.Content,
		})
	}

	//Handle the case of an existing summary already exists for the summary
	if previousSummary != nil {
		previousSummaryText := stringFormatter.Format(prompts.ExistingSummary, map[string]string{
			"summary": previousSummary.String(),
		})
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: previousSummaryText,
		})
	}

	for _, msg := range conversation.Messages {
		content, err := msg.JSON()
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: content,
		})
	}

	return msgs, nil
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

func (ai *OpenAI) Learn(instructions []*chat.Prompt, conversation *chat.Conversation) ([]*memory.Knowledge, error) {
	data, err := ai.prepareLearnMessage(instructions, conversation)
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
	fmt.Println("back from opeani")
	fmt.Println(err)
	fmt.Println(resp)
	fmt.Println("tokens")
	fmt.Println(resp.Usage)
	ai.parseLearnResponse(conversation, resp.Choices[0].Message.Content)

	return nil, nil
}

func (ai *OpenAI) prepareLearnMessage(instructions []*chat.Prompt, conversation *chat.Conversation) ([]openai.ChatCompletionMessage, error) {
	msgs := []openai.ChatCompletionMessage{}

	for _, instruction := range instructions {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: instruction.Content,
		})
	}

	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "Converstaion History:\n",
	})

	lastMsgJson := conversation.Messages[len(conversation.Messages)-1].SimpleString()

	for _, msg := range conversation.Messages {
		content := msg.SimpleString()
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: content + "\n",
		})
	}
	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "Input:\n",
	})
	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: lastMsgJson + "\n",
	})
	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "Output:\n",
	})

	return msgs, nil
}

func (ai *OpenAI) parseLearnResponse(conversation *chat.Conversation, raw string) (*memory.Summary, error) {
	split := strings.Split(raw, "|")
	fmt.Println("returns")
	fmt.Println(split)
	//Validate and protect this
	split[0] = strings.TrimSpace(split[0])
	split[1] = strings.TrimSpace(split[1])

	if split[0] == "none" {
		fmt.Println("NONE REPORTED")
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

func (ai *OpenAI) ConversationContinuance(
	instructions []*chat.Prompt,
	conversation *chat.Conversation,
	summary *memory.Summary,
) (bool, error) {
	data, err := ai.prepareConversationContinuanceMessage(instructions, conversation, summary)
	if err != nil {
		return false, err
	}

	resp, err := ai.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: data,
		},
	)
	if err != nil {
		return false, err
	}
	fmt.Println("continuance test", resp)
	return ai.parseConversationContinuanceResponse(resp.Choices[0].Message.Content), nil
}

func (ai *OpenAI) prepareConversationContinuanceMessage(
	instructions []*chat.Prompt,
	conversation *chat.Conversation,
	summary *memory.Summary,
) ([]openai.ChatCompletionMessage, error) {
	msgs := []openai.ChatCompletionMessage{}

	for _, instruction := range instructions {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: instruction.Content,
		})
	}

	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "Summary:\n",
	})

	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: summary.String(),
	})

	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "Converstaion History:\n",
	})

	for _, msg := range conversation.Messages {
		content, err := msg.JSON()
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: content + "\n",
		})
	}

	return msgs, nil
}

func (ai *OpenAI) parseConversationContinuanceResponse(raw string) bool {
	return strings.ToLower(raw) == "true"
}
