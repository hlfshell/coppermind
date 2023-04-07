package agent

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/prompts"
	"github.com/hlfshell/coppermind/internal/store/sqlite"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSqlLiteStore() (*sqlite.SqliteStore, error) {
	store, err := sqlite.NewSqliteStore(":memory:")
	if err != nil {
		return nil, err
	}
	if err = store.Migrate(); err != nil {
		return nil, err
	}
	return store, nil
}

func newTestingAgent(t *testing.T) *Agent {
	store, err := createSqlLiteStore()
	require.Nil(t, err)

	return &Agent{
		db:  store,
		llm: &mockLLM{},

		identity:                prompts.Rose,
		maxChatMessages:         20,
		maintainConversation:    10 * time.Minute,
		maxConversationIdleTime: 6 * time.Hour,

		daemonTicker:                           time.NewTicker(60 * time.Second),
		summaryMinMessages:                     5,
		summaryMinConversationTime:             5 * time.Minute,
		summaryMinMessagesToForceSummarization: 15,
	}
}

func TestGenerateOrFindConversation(t *testing.T) {
	agent := newTestingAgent(t)

	newMessage := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: "",
		User:         "Keith",
		Agent:        "Hal",
		Content:      "hub bub",
		CreatedAt:    time.Now(),
	}

	// Without any conversatons in the store, we expect a random
	// new conversation uuid
	newConversation, err := agent.GenerateOrFindConversation(newMessage)
	require.Nil(t, err)
	assert.NotEqual(t, "", newConversation)

	retrievedConversation, err := agent.db.GetConversation(newConversation)
	require.Nil(t, err)
	assert.Nil(t, retrievedConversation)

	// We have an existing conversation with a different user; this shouldn't
	// effect the random non associated conversation ID
	unaffiliatedMessage := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: uuid.New().String(),
		User:         "Abby",
		Agent:        "Hal",
		Content:      "hub bub",
		CreatedAt:    time.Now(),
	}
	err = agent.db.SaveMessage(unaffiliatedMessage)
	require.Nil(t, err)

	newConversation, err = agent.GenerateOrFindConversation(newMessage)
	require.Nil(t, err)
	assert.NotEqual(t, "", newConversation)
	assert.NotEqual(t, unaffiliatedMessage.Conversation, newConversation)

	retrievedConversation, err = agent.db.GetConversation(newConversation)
	require.Nil(t, err)
	assert.Nil(t, retrievedConversation)

	// Now we have a conversation with the same user within the
	// time limit of automatic continuance of the conversation
	// as per agent settings of agent.maintainConversation
	affiliatedMessage := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: uuid.New().String(),
		User:         "Keith",
		Agent:        "Hal",
		Content:      "hub bub",
		CreatedAt:    time.Now().Add(time.Duration(0.9 * -1 * float64(agent.maintainConversation))),
	}
	err = agent.db.SaveMessage(affiliatedMessage)
	require.Nil(t, err)

	newConversation, err = agent.GenerateOrFindConversation(newMessage)
	require.Nil(t, err)
	assert.NotEqual(t, "", newConversation)
	assert.Equal(t, affiliatedMessage.Conversation, newConversation)

	// Let's create a new conversation with a new user (swapping our target
	// message to this user) and set the maintain time outside of it. We shall
	// have no summaries available so as to prevent the continuation from
	// being called
	newMessage.User = "Rebecca"
	affiliatedMessage = &chat.Message{
		ID:           uuid.New().String(),
		Conversation: uuid.New().String(),
		User:         "Rebecca",
		Agent:        "Hal",
		Content:      "hub bub",
		CreatedAt:    time.Now().Add(-1*agent.maintainConversation - time.Minute),
	}
	err = agent.db.SaveMessage(affiliatedMessage)
	require.Nil(t, err)

	newConversation, err = agent.GenerateOrFindConversation(newMessage)
	require.Nil(t, err)
	assert.NotEqual(t, "", newConversation)
	assert.NotEqual(t, affiliatedMessage.Conversation, newConversation)

	// Create a summary for the conversation to have the agent ask the LLM
	// for whether or not it is appropriate to continue the conversation -
	// we run the test for both possible outcomes
	agent.llm = &mockLLM{
		conversationContinuanceResponse: true,
	}
	summary := &memory.Summary{
		ID:                    uuid.New().String(),
		Conversation:          affiliatedMessage.Conversation,
		User:                  affiliatedMessage.User,
		Agent:                 affiliatedMessage.Agent,
		Keywords:              []string{"hub", "bub"},
		Summary:               "a fake conversation",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: affiliatedMessage.CreatedAt,
	}
	agent.db.SaveSummary(summary)
	require.Nil(t, err)

	newConversation, err = agent.GenerateOrFindConversation(newMessage)
	require.Nil(t, err)
	assert.NotEqual(t, "", newConversation)
	assert.Equal(t, affiliatedMessage.Conversation, newConversation)

	// We'll create a new message and new summary for a new user, and test
	// the above case but with the LLM returning false
	newMessage.User = "Justin"
	agent.llm = &mockLLM{
		conversationContinuanceResponse: false,
	}
	affiliatedMessage = &chat.Message{
		ID:           uuid.New().String(),
		Conversation: uuid.New().String(),
		User:         "Justin",
		Agent:        "Hal",
		Content:      "hub bub",
		CreatedAt:    time.Now().Add(-1*agent.maintainConversation - time.Minute),
	}
	err = agent.db.SaveMessage(affiliatedMessage)
	require.Nil(t, err)
	summary = &memory.Summary{
		ID:                    uuid.New().String(),
		Conversation:          affiliatedMessage.Conversation,
		User:                  affiliatedMessage.User,
		Agent:                 affiliatedMessage.Agent,
		Keywords:              []string{"hub", "bub"},
		Summary:               "a fake conversation",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: affiliatedMessage.CreatedAt,
	}
	agent.db.SaveSummary(summary)
	require.Nil(t, err)

	newConversation, err = agent.GenerateOrFindConversation(newMessage)
	require.Nil(t, err)
	assert.NotEqual(t, "", newConversation)
	assert.NotEqual(t, affiliatedMessage.Conversation, newConversation)

	// Finally, test to ensure that a conversation sufficiently in the past
	// (past the agent.maxConversationIdleTime) is not continued
	newMessage.User = "Karen"
	affiliatedMessage = &chat.Message{
		ID:           uuid.New().String(),
		Conversation: uuid.New().String(),
		User:         "Karen",
		Agent:        "Hal",
		Content:      "hub bub",
		CreatedAt:    time.Now().Add(-1*agent.maxConversationIdleTime - time.Minute),
	}
	err = agent.db.SaveMessage(affiliatedMessage)
	require.Nil(t, err)

	newConversation, err = agent.GenerateOrFindConversation(newMessage)
	require.Nil(t, err)
	assert.NotEqual(t, "", newConversation)
	assert.NotEqual(t, affiliatedMessage.Conversation, newConversation)

	// Adding a summary does not change this
	summary = &memory.Summary{
		ID:                    uuid.New().String(),
		Conversation:          affiliatedMessage.Conversation,
		User:                  affiliatedMessage.User,
		Agent:                 affiliatedMessage.Agent,
		Keywords:              []string{"hub", "bub"},
		Summary:               "a fake conversation",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: affiliatedMessage.CreatedAt,
	}
	agent.db.SaveSummary(summary)
	require.Nil(t, err)

	newConversation, err = agent.GenerateOrFindConversation(newMessage)
	require.Nil(t, err)
	assert.NotEqual(t, "", newConversation)
	assert.NotEqual(t, affiliatedMessage.Conversation, newConversation)
}
