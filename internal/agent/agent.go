package agent

import (
	"fmt"
	"time"

	_ "embed"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/llm"
	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
)

type Agent struct {
	Name string

	db  store.Store
	llm llm.LLM

	//Chat specific
	identity                string
	maxChatMessages         int
	maintainConversation    time.Duration
	maxConversationIdleTime time.Duration

	//Summary specific
	daemonTicker                           *time.Ticker
	summaryMinMessages                     int
	summaryMinConversationTime             time.Duration
	summaryMinMessagesToForceSummarization int
}

func NewAgent(name string, db store.Store, identity string, llm llm.LLM) *Agent {
	agent := &Agent{
		Name: name,
		db:   db,
		llm:  llm,

		identity:                identity,
		maxChatMessages:         20,
		maintainConversation:    10 * time.Minute,
		maxConversationIdleTime: 6 * time.Hour,

		daemonTicker: time.NewTicker(60 * time.Second),

		summaryMinMessages:                     5,
		summaryMinConversationTime:             5 * time.Minute,
		summaryMinMessagesToForceSummarization: 15,
	}

	go func() {
		for {
			<-agent.daemonTicker.C
			agent.RunDaemons()
		}
	}()

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
		fmt.Println("Chose conversation: ", msg.Conversation)
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

	if len(history.Messages) > agent.maxChatMessages {
		history.Messages = history.PastNMessages(agent.maxChatMessages)
	}

	// Load up summaries for user/agent conversations
	pastSummaries, err := agent.db.ListSummaries(store.Filter{
		Attributes: []*store.FilterAttribute{
			{
				Attribute: "agent",
				Value:     msg.Agent,
				Operation: store.EQ,
			},
			{
				Attribute: "user",
				Value:     msg.User,
				Operation: store.EQ,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Find associated facts from prior conversations
	// knowledge, err := agent.db.GetKnowlegeByAgentAndUser(msg.Agent, msg.User)
	// if err != nil {
	// 	return nil, err
	// }
	// Placeholder for now
	knowledge := []*memory.Knowledge{}

	// Have the LLM deal with the message as expected
	response, err := agent.llm.SendMessage(
		agent.identity,
		history,
		pastSummaries,
		knowledge,
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
	err = agent.db.SaveMessage(
		response.ToMessage(
			msg.User,
			agent.Name,
			msg.Conversation,
		),
	)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (agent *Agent) GenerateOrFindConversation(msg *chat.Message) (string, error) {
	conversation, timestamp, err := agent.db.GetLatestConversation(msg.Agent, msg.User)
	if err != nil {
		return "", err
	} else if time.Now().Add(-1 * agent.maintainConversation).Before(timestamp) {
		// This handles the conversation existing, and it being within the maintain
		// window of a conversation
		return conversation, nil
	} else if conversation == "" {
		// This handles the situation of there is no existing conversation at all
		conversation = uuid.New().String()
	} else if time.Now().Add(-1 * agent.maxConversationIdleTime).After(timestamp) {
		// We have a conversation that exists, but it's beyond the allowed time to
		// continue the conversation; therefore it's an automatic new one
		conversation = uuid.New().String()
	} else {
		// Finally, this sees via the LLM if it should be a conversation
		// continuance or a new conversation.
		retrievedConversation, err := agent.db.GetConversation(conversation)
		if err != nil {
			return "", err
		}
		summaries, err := agent.db.ListSummaries(store.Filter{
			Attributes: []*store.FilterAttribute{
				{
					Attribute: "conversation",
					Value:     conversation,
					Operation: store.EQ,
				},
			},
		})
		if err != nil {
			return "", err
		}
		var summary *memory.Summary
		if len(summaries) != 0 {
			summary = summaries[0]
		}

		if summary == nil {
			return uuid.New().String(), nil
		}

		shouldContinue, err := agent.llm.ConversationContinuance(
			msg,
			retrievedConversation,
			summary,
		)
		if err != nil {
			return "", nil
		} else if shouldContinue {
			return conversation, nil
		} else {
			return uuid.New().String(), nil
		}
	}
	return conversation, nil
}
