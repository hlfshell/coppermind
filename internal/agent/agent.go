package agent

import (
	"fmt"
	"os"
	"time"

	_ "embed"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/chat"
	"github.com/hlfshell/coppermind/internal/llm"
	"github.com/hlfshell/coppermind/internal/prompts"
	"github.com/hlfshell/coppermind/internal/store"
)

type Agent struct {
	Name string

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

func NewAgent(name string, db store.Store, llm llm.LLM) *Agent {
	instructions := []*chat.Prompt{&chat.Prompt{Type: chat.SetupPrompt, Content: prompts.Instructions}}
	identity := []*chat.Prompt{&chat.Prompt{Type: chat.SetupPrompt, Content: prompts.Identity}}

	summaryInstructions := []*chat.Prompt{&chat.Prompt{Type: chat.SetupPrompt, Content: prompts.Summary}}

	knowledgeInstructions := []*chat.Prompt{&chat.Prompt{Type: chat.SetupPrompt, Content: prompts.Knowledge}}

	agent := &Agent{
		Name:                name,
		db:                  db,
		llm:                 llm,
		chatInstructions:    instructions,
		identity:            identity,
		summaryInstructions: summaryInstructions,
		summaryTicker:       time.NewTicker(60 * time.Second),

		knowledgeInstructions: knowledgeInstructions,
	}

	err := agent.SummaryDaemon()
	fmt.Println("done")
	fmt.Println(err)
	os.Exit(3)

	// err := agent.KnowledgeDaemon()
	// fmt.Println("done")
	// fmt.Println(err)
	// os.Exit(3)

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
	// If no conversation is set, lookup to see if we have an old conversation
	// that we can load up and join (based on how long since it's been) the
	// last message in that conversation
	if msg.Conversation == "" {
		conversation, err := agent.GenerateOrFindConversation(msg)
		if err != nil {
			return nil, err
		}
		msg.Conversation = conversation
	}

	// Load up the history if any exists
	history, err := agent.db.GetConversation(msg.Conversation)
	if err != nil {
		return nil, err
	} else if history == nil {
		history = &chat.Conversation{
			ID:        uuid.New().String(),
			Agent:     msg.Agent,
			User:      msg.User,
			CreatedAt: msg.CreatedAt,
			Messages:  []*chat.Message{},
		}
	}

	// Load up summaries for user/agent conversations
	summary, err := agent.db.GetSummaryByConversation(msg.Conversation)
	if err != nil {
		return nil, err
	}
	pastSummaries, err := agent.db.GetSummariesByAgentAndUser(msg.Agent, msg.User)
	if err != nil {
		return nil, err
	}

	// Have the LLM deal with the message as expected
	response, err := agent.llm.SendMessage(
		agent.chatInstructions,
		agent.identity,
		history,
		pastSummaries,
		summary,
		msg,
	)
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

func (agent *Agent) GenerateOrFindConversation(msg *chat.Message) (string, error) {
	conversation, timestamp, err := agent.db.GetLatestConversation(msg.Agent, msg.User)
	if err != nil {
		return "", err
	}
	if time.Now().After(timestamp.Add(5*time.Minute)) || conversation == "" {
		conversation = uuid.New().String()
	}
	return conversation, nil
}
