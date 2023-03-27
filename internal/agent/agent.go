package agent

import (
	"time"

	_ "embed"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/chat"
	"github.com/hlfshell/coppermind/internal/llm"
	"github.com/hlfshell/coppermind/internal/prompts"
	"github.com/hlfshell/coppermind/internal/store"
)

type Agent struct {
	db  store.Store
	llm llm.LLM

	//Chat specific
	chatInstructions []*chat.Prompt
	identity         []*chat.Prompt

	//Summary specific
	summaryInstructions []*chat.Prompt
	summaryTicker       *time.Ticker

	//Knowledge specific
	knowledgeInstructions []*chat.Prompt
}

func NewAgent(db store.Store, llm llm.LLM) *Agent {
	instructions := []*chat.Prompt{&chat.Prompt{Type: chat.SetupPrompt, Content: prompts.Instructions}}
	identity := []*chat.Prompt{&chat.Prompt{Type: chat.SetupPrompt, Content: prompts.Identity}}

	summaryInstructions := []*chat.Prompt{&chat.Prompt{Type: chat.SetupPrompt, Content: prompts.Summary}}

	knowledgeInstructions := []*chat.Prompt{&chat.Prompt{Type: chat.SetupPrompt, Content: prompts.Knowledge}}

	agent := &Agent{
		db:                  db,
		llm:                 llm,
		chatInstructions:    instructions,
		identity:            identity,
		summaryInstructions: summaryInstructions,
		summaryTicker:       time.NewTicker(60 * time.Second),

		knowledgeInstructions: knowledgeInstructions,
	}

	// go func() {
	// 	for {
	// 		<-agent.summaryTicker.C
	// 		fmt.Println("Summary Daemon triggered")
	// 		err := agent.SummaryDaemon()
	// 		if err != nil {
	// 			fmt.Println("Summary error")
	// 			fmt.Println(err)
	// 			os.Exit(3)
	// 		}

	// 	}
	// }()

	return agent
}

func (agent *Agent) SendMessage(msg *chat.Message) (*chat.Response, error) {
	conversation, err := agent.GenerateOrFindConversation(msg.User)
	if err != nil {
		return nil, err
	}
	msg.Conversation = conversation

	history, err := agent.loadConversationHistory(msg.Conversation)
	if err != nil {
		return nil, err
	}
	response, err := agent.llm.SendMessage(agent.chatInstructions, agent.identity, history, msg)
	if err != nil {
		return nil, err
	}

	// Save both the incoming message and response to the history
	err = agent.db.SaveMessage(msg)
	if err != nil {
		return nil, err
	}
	err = agent.db.SaveMessage(response.ToMessage(msg.Conversation))
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (agent *Agent) GenerateOrFindConversation(user string) (string, error) {
	conversation, timestamp, err := agent.db.GetLatestConversation(user)
	if err != nil {
		return "", nil
	}
	if time.Now().After(timestamp.Add(5*time.Minute)) || conversation == "" {
		conversation = uuid.New().String()
	}
	return conversation, nil
}

func (agent *Agent) loadConversationHistory(conversation string) ([]*chat.Message, error) {
	return agent.db.LoadConversationHistory(conversation)
}
