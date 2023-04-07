package openai

import (
	"github.com/hlfshell/coppermind/internal/prompts"
	"github.com/sashabaranov/go-openai"
)

type OpenAI struct {
	apiKey   string
	client   *openai.Client
	tokenMax int
	maxInput int

	// Prompts
	chatPrompt                      string
	conversationalContinuancePrompt string
	summaryPrompt                   string
	knowledgePrompt                 string
}

func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{
		apiKey:   apiKey,
		client:   openai.NewClient(apiKey),
		tokenMax: 4025,
		maxInput: 2800,

		chatPrompt:                      prompts.Instructions,
		conversationalContinuancePrompt: prompts.ConversationContinuance,
		summaryPrompt:                   prompts.Summary,
		knowledgePrompt:                 prompts.Knowledge,
	}
}

func (ai *OpenAI) EstimateTokens(text string) int {
	return int(len(text) / 4)
}

type OpenAIResponseError struct {
	msg string
}

func (err OpenAIResponseError) Error() string {
	return err.msg
}
