package store

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/agents"
	"github.com/hlfshell/coppermind/pkg/artifacts"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/hlfshell/coppermind/pkg/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===============================
// Messages
// ===============================

func SaveAndGetMessage(t *testing.T, store store.LowLevelStore) {
	id := uuid.New().String()
	message := &chat.Message{
		ID:      id,
		User:    "Huey",
		Agent:   "Luey",
		From:    "Huey",
		Content: "Where's Dewy?",
		Artifacts: []*artifacts.ArtifactData{
			{
				ID:        uuid.New().String(),
				Message:   id,
				Type:      "image",
				CreatedAt: time.Now(),
				Data:      json.RawMessage(`{"url": "https://www.picturesofdogs.com/img1"}`),
			},
			{
				ID:        uuid.New().String(),
				Message:   id,
				Type:      "voice",
				CreatedAt: time.Now(),
				Data:      []byte{0, 1, 2},
			},
		},
		Conversation: uuid.New().String(),
	}

	msg, err := store.GetMessage(message.ID)
	require.Nil(t, err)
	assert.Nil(t, msg)

	err = store.SaveMessage(message)
	require.Nil(t, err)

	msg, err = store.GetMessage(message.ID)
	require.Nil(t, err)
	assert.NotNil(t, msg)
	assert.True(t, message.Equal(msg))
}

func DeleteMessage(t *testing.T, store store.LowLevelStore) {
	id := uuid.New().String()
	message := &chat.Message{
		ID:           id,
		User:         "Yoshi",
		Agent:        "Mario",
		From:         "Yoshi",
		Content:      "Need a ride?",
		Conversation: uuid.New().String(),
		Artifacts: []*artifacts.ArtifactData{
			{
				ID:        uuid.New().String(),
				Message:   id,
				Type:      "image",
				CreatedAt: time.Now(),
				Data:      json.RawMessage(`{"url": "https://www.picturesofdogs.com/img1"}`),
			},
			{
				ID:        uuid.New().String(),
				Message:   id,
				Type:      "voice",
				CreatedAt: time.Now(),
				Data:      []byte{0, 1, 2},
			},
		},
	}

	msg, err := store.GetMessage(message.ID)
	require.Nil(t, err)
	assert.Nil(t, msg)

	err = store.SaveMessage(message)
	require.Nil(t, err)

	msg, err = store.GetMessage(message.ID)
	require.Nil(t, err)
	assert.NotNil(t, msg)
	assert.True(t, message.Equal(msg))

	err = store.DeleteMessage(message.ID)
	require.Nil(t, err)

	msg, err = store.GetMessage(message.ID)
	require.Nil(t, err)
	assert.Nil(t, msg)
}

func ListMessages(t *testing.T, s store.LowLevelStore) {
	// Create a number of messages for us to query back
	// and test the list feature with
	msg1Id := uuid.New().String()
	msg1 := &chat.Message{
		ID:           msg1Id,
		User:         "Peach",
		Agent:        "Bowser",
		From:         "Peach",
		Content:      "I need some space...",
		Conversation: uuid.New().String(),
		CreatedAt:    time.Now().Add(-5 * time.Minute),
		Artifacts: []*artifacts.ArtifactData{
			{
				ID:        uuid.New().String(),
				Message:   msg1Id,
				Type:      "image",
				CreatedAt: time.Now(),
				Data:      json.RawMessage(`{"url": "https://www.picturesofdogs.com/img1"}`),
			},
		},
	}
	msg2 := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Yoshi",
		Agent:        "Mario",
		From:         "Yoshi",
		Content:      "Need a ride?",
		Conversation: uuid.New().String(),
		CreatedAt:    time.Now().Add(-24 * time.Hour),
	}
	msg3Id := uuid.New().String()
	msg3 := &chat.Message{
		ID:           msg3Id,
		User:         "Peach",
		Agent:        "Mario",
		From:         "Peach",
		Content:      "I just headed over to another castle...",
		Conversation: uuid.New().String(),
		CreatedAt:    time.Now(),
		Artifacts: []*artifacts.ArtifactData{
			{
				ID:        uuid.New().String(),
				Message:   msg3Id,
				Type:      "voice",
				CreatedAt: time.Now(),
				Data:      []byte{0, 1, 2},
			},
		},
	}

	err := s.SaveMessage(msg1)
	require.Nil(t, err)
	err = s.SaveMessage(msg2)
	require.Nil(t, err)
	err = s.SaveMessage(msg3)
	require.Nil(t, err)

	// First we will ensure we can get all back with a blank
	// filter
	messages, err := s.ListMessages(store.Filter{})
	require.Nil(t, err)
	assert.Equal(t, 3, len(messages))

	// Now we will test the user filter to get back all messages
	// with a singular user
	messages, err = s.ListMessages(store.Filter{
		Attributes: []*store.FilterAttribute{
			{
				Attribute: "user",
				Operation: store.EQ,
				Value:     "Peach",
			},
		},
	})
	require.Nil(t, err)
	assert.Equal(t, 2, len(messages))
	// Assert that we get the right messages back and
	// are in the expected older-first order
	assert.True(t, msg1.Equal(messages[0]))
	assert.True(t, msg3.Equal(messages[1]))

	// Test the limit feature
	messages, err = s.ListMessages(store.Filter{
		Attributes: []*store.FilterAttribute{
			{
				Attribute: "user",
				Operation: store.EQ,
				Value:     "Peach",
			},
		},
		Limit: 1,
	})
	require.Nil(t, err)
	assert.Equal(t, 1, len(messages))
	assert.True(t, msg1.Equal(messages[0]))
}

// ===============================
// Conversations
// ===============================

func GetAndDeleteConversation(t *testing.T, store store.LowLevelStore) {
	msg := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Abby",
		Agent:        "Carrot",
		From:         "Abby",
		Content:      "You look tasty, carrot",
		Conversation: uuid.New().String(),
		CreatedAt:    time.Now().Add(-5 * time.Minute),
	}
	msg2 := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Abby",
		Agent:        "Carrot",
		From:         "Carrot",
		Content:      "Please don't eat me",
		Conversation: msg.Conversation,
		CreatedAt:    time.Now(),
	}
	msg3 := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Abby",
		Agent:        "Carrot",
		Content:      "...and that's why I believe that capitalism is ultimately going to lead to...",
		Conversation: uuid.New().String(),
		CreatedAt:    time.Now(),
	}

	// Ensure the conversation doesn't return yet
	conversation, err := store.GetConversation(msg.Conversation)
	require.Nil(t, err)
	assert.Nil(t, conversation)

	// Create our messages
	err = store.SaveMessage(msg)
	require.Nil(t, err)
	err = store.SaveMessage(msg2)
	require.Nil(t, err)
	err = store.SaveMessage(msg3)
	require.Nil(t, err)

	// We can recover the conversation of 2 messages
	conversation, err = store.GetConversation(msg.Conversation)
	require.Nil(t, err)
	assert.NotNil(t, conversation)
	assert.Equal(t, msg.Conversation, conversation.ID)
	assert.Equal(t, msg.Agent, conversation.Agent)
	assert.Equal(t, msg.User, conversation.User)
	assert.Equal(t, 2, len(conversation.Messages))
	assert.WithinDuration(t, msg.CreatedAt, conversation.CreatedAt, time.Second)
	assert.True(t, msg.Equal(conversation.Messages[0]))
	assert.True(t, msg2.Equal(conversation.Messages[1]))

	// Delete and ensure the conversation and messages are gone, but
	// message 3 is still there
	err = store.DeleteConversation(msg.Conversation)
	require.Nil(t, err)

	conversation, err = store.GetConversation(msg.Conversation)
	require.Nil(t, err)
	assert.Nil(t, conversation)

	msg, err = store.GetMessage(msg.ID)
	require.Nil(t, err)
	assert.Nil(t, msg)

	msg2, err = store.GetMessage(msg2.ID)
	require.Nil(t, err)
	assert.Nil(t, msg2)

	msg3, err = store.GetMessage(msg3.ID)
	require.Nil(t, err)
	assert.NotNil(t, msg3)
}

func ListConversations(t *testing.T, db store.LowLevelStore) {
	// Confirm that a blank filter returns no conversations as none
	// exists
	conversations, err := db.ListConversations(store.Filter{})
	require.Nil(t, err)
	assert.Equal(t, 0, len(conversations))

	// Create several conversations, of at least 3 messages each for
	// now
	messages := map[string][]*chat.Message{}
	numConvos := 3
	numMsgs := 3
	for i := 0; i < numConvos; i++ {
		convo := uuid.New().String()
		user := uuid.New().String()
		agent := uuid.New().String()

		for j := 0; j < numMsgs; j++ {
			msg := &chat.Message{
				ID:           uuid.New().String(),
				User:         user,
				Agent:        agent,
				From:         user,
				Content:      uuid.New().String(),
				CreatedAt:    time.Now().Add(-1*time.Duration(j) + (-5 * time.Minute)),
				Conversation: convo,
			}
			err = db.SaveMessage(msg)
			require.Nil(t, err)
			messages[convo] = append(messages[convo], msg)
		}
	}

	// Now we try for all conversations again and hope to get all three
	conversations, err = db.ListConversations(store.Filter{})
	require.Nil(t, err)
	require.Equal(t, numConvos, len(conversations))
	for _, convo := range conversations {
		assert.Equal(t, numMsgs, len(convo.Messages))
		for i := 0; i < numMsgs; i++ {
			assert.True(t, messages[convo.ID][i].Equal(convo.Messages[i]))
		}
		assert.WithinDuration(t, messages[convo.ID][0].CreatedAt, convo.CreatedAt, time.Second)
	}

	// Test limiting by user/agent
	conversations, err = db.ListConversations(store.Filter{
		Attributes: []*store.FilterAttribute{
			{
				Attribute: "user",
				Operation: store.EQ,
				Value:     messages[conversations[0].ID][0].User,
			},
			{
				Attribute: "agent",
				Operation: store.EQ,
				Value:     messages[conversations[0].ID][0].Agent,
			},
		},
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(conversations))
	assert.Equal(t, numMsgs, len(conversations[0].Messages))
	for i := 0; i < numMsgs; i++ {
		assert.True(t, messages[conversations[0].ID][i].Equal(conversations[0].Messages[i]))
	}
}

// ===============================
// Summaries
// ===============================

func SaveAndGetSummary(t *testing.T, store store.LowLevelStore) {
	summary := &memory.Summary{
		ID:                    uuid.New().String(),
		Agent:                 "Michelangelo",
		Conversation:          uuid.New().String(),
		Keywords:              []string{"pizza", "anchovies", "sewer surfing"},
		Summary:               "Cowabunga!",
		User:                  "Donatello",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: time.Now().Add(-5 * time.Minute),
	}

	readSummary, err := store.GetSummary(summary.ID)
	require.Nil(t, err)
	assert.Nil(t, readSummary)

	err = store.SaveSummary(summary)
	require.Nil(t, err)

	readSummary, err = store.GetSummary(summary.ID)
	require.Nil(t, err)
	assert.NotNil(t, readSummary)
	assert.True(t, summary.Equal(readSummary))
}

func DeleteSummary(t *testing.T, store store.LowLevelStore) {
	summary := &memory.Summary{
		ID:                    uuid.New().String(),
		Agent:                 "Michelangelo",
		Conversation:          uuid.New().String(),
		Keywords:              []string{"pizza", "anchovies", "sewer surfing"},
		Summary:               "Master splinter's training is too harsh",
		User:                  "Leonardo",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: time.Now().Add(-5 * time.Minute),
	}
	err := store.SaveSummary(summary)
	require.Nil(t, err)

	readSummary, err := store.GetSummary(summary.ID)
	require.Nil(t, err)
	assert.NotNil(t, readSummary)
	assert.True(t, summary.Equal(readSummary))

	err = store.DeleteSummary(summary.ID)
	require.Nil(t, err)

	readSummary, err = store.GetSummary(summary.ID)
	require.Nil(t, err)
	assert.Nil(t, readSummary)
}

func ListSummaries(t *testing.T, db store.LowLevelStore) {
	// Ensure that we have no summaries to start
	summaries, err := db.ListSummaries(store.Filter{})
	require.Nil(t, err)
	assert.Equal(t, 0, len(summaries))

	// Create summaries for testing
	summary1 := &memory.Summary{
		ID:                    uuid.New().String(),
		User:                  "User 1",
		Agent:                 "Agent 1",
		Conversation:          uuid.New().String(),
		Keywords:              []string{"keyword 1 1", "keyword 1 2", "keyword 1 3"},
		Summary:               "Summary content 1",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: time.Now().Add(-3 * time.Minute),
	}
	summary2 := &memory.Summary{
		ID:                    uuid.New().String(),
		User:                  "User 2",
		Agent:                 "Agent 2",
		Conversation:          uuid.New().String(),
		Keywords:              []string{"keyword 2 1", "keyword 2 2", "keyword 2 3"},
		Summary:               "Summary content 2",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: time.Now().Add(-2 * time.Minute),
	}
	summary3 := &memory.Summary{
		ID:                    uuid.New().String(),
		User:                  "User 3",
		Agent:                 "Agent 3",
		Conversation:          uuid.New().String(),
		Keywords:              []string{"keyword 3 1", "keyword 3 2", "keyword 3 3"},
		Summary:               "Summary content 3",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: time.Now().Add(-1 * time.Minute),
	}
	err = db.SaveSummary(summary1)
	require.Nil(t, err)
	err = db.SaveSummary(summary2)
	require.Nil(t, err)
	err = db.SaveSummary(summary3)
	require.Nil(t, err)

	// Test that we can get all summaries back
	summaries, err = db.ListSummaries(store.Filter{})
	require.Nil(t, err)
	assert.Equal(t, 3, len(summaries))

	// They should be returned oldest first
	assert.True(t, summary1.Equal(summaries[0]))
	assert.True(t, summary2.Equal(summaries[1]))
	assert.True(t, summary3.Equal(summaries[2]))

	// Ensure the limit option works
	summaries, err = db.ListSummaries(store.Filter{
		Limit: 1,
	})
	require.Nil(t, err)
	assert.Equal(t, 1, len(summaries))
	assert.True(t, summary1.Equal(summaries[0]))

	// Test that we can get summaries back by user
	summaries, err = db.ListSummaries(store.Filter{
		Attributes: []*store.FilterAttribute{
			{
				Attribute: "user",
				Operation: store.EQ,
				Value:     summary1.User,
			},
		},
	})
	require.Nil(t, err)
	assert.Equal(t, 1, len(summaries))
	assert.True(t, summary1.Equal(summaries[0]))
}

// ===============================
// Agents
// ===============================

func SaveAndGetAgent(t *testing.T, store store.LowLevelStore) {
	agent := &agents.Agent{
		ID:       uuid.New().String(),
		Name:     "Hal",
		Identity: "Super helpful, nothing but",
	}

	readAgent, err := store.GetAgent(agent.ID)
	require.Nil(t, err)
	assert.Nil(t, readAgent)

	err = store.SaveAgent(agent)
	require.Nil(t, err)

	readAgent, err = store.GetAgent(agent.ID)
	require.Nil(t, err)
	assert.NotNil(t, readAgent)
	assert.Equal(t, agent, readAgent)
}

func DeleteAgent(t *testing.T, store store.LowLevelStore) {
	agent := &agents.Agent{
		ID:       uuid.New().String(),
		Name:     "Rose",
		Identity: "Sassy and cynical",
	}

	readAgent, err := store.GetAgent(agent.ID)
	require.Nil(t, err)
	assert.Nil(t, readAgent)

	err = store.SaveAgent(agent)
	require.Nil(t, err)

	readAgent, err = store.GetAgent(agent.ID)
	require.Nil(t, err)
	assert.NotNil(t, readAgent)
	assert.Equal(t, agent, readAgent)

	err = store.DeleteAgent(agent.ID)
	require.Nil(t, err)

	readAgent, err = store.GetAgent(agent.ID)
	require.Nil(t, err)
	assert.Nil(t, readAgent)
}

func ListAgents(t *testing.T, db store.LowLevelStore) {
	agnts, err := db.ListAgents()
	require.Nil(t, err)
	assert.Equal(t, 0, len(agnts))

	numAgents := 3
	agentsMap := map[string]*agents.Agent{}
	for i := 0; i < numAgents; i++ {
		agent := &agents.Agent{
			ID:       uuid.New().String(),
			Name:     uuid.New().String(),
			Identity: uuid.New().String(),
		}
		err := db.SaveAgent(agent)
		require.Nil(t, err)
		agentsMap[agent.ID] = agent
	}

	agnts, err = db.ListAgents()
	require.Nil(t, err)
	assert.Equal(t, numAgents, len(agnts))
	for _, agent := range agnts {
		assert.Equal(t, agentsMap[agent.ID], agent)
	}
}

// ===============================
// Users
// ===============================

func SaveAndCreatetUser(t *testing.T, store store.LowLevelStore) {
	user := &users.User{
		ID:        uuid.New().String(),
		Name:      "Keith",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	password := "supersecret"

	readUser, err := store.GetUser(user.ID)
	require.Nil(t, err)
	assert.Nil(t, readUser)

	err = store.CreateUser(user, password)
	require.Nil(t, err)

	readUser, err = store.GetUser(user.ID)
	require.Nil(t, err)
	require.NotNil(t, readUser)
	assert.True(t, user.Equal(readUser))

	// Test user creation w/ a password that's too small
	user = &users.User{
		ID:        uuid.New().String(),
		Name:      "Karen",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	password = "short"

	err = store.CreateUser(user, password)
	require.NotNil(t, err)
	assert.Equal(t, "invalid password", err.Error())
}

func GetUserAuth(t *testing.T, store store.LowLevelStore) {
	user := &users.User{
		ID:        uuid.New().String(),
		Name:      "Pepper",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	password := "supersecret"

	readUser, err := store.GetUser(user.ID)
	require.Nil(t, err)
	assert.Nil(t, readUser)

	err = store.CreateUser(user, password)
	require.Nil(t, err)

	readUser, err = store.GetUser(user.ID)
	require.Nil(t, err)

	auth, err := store.GetUserAuth(user.ID)
	require.Nil(t, err)
	assert.NotNil(t, auth)

	assert.NotEqual(t, password, auth.Password)
	assert.True(t, auth.CheckPassword(password))
}

func GenerateUserPasswordResetToken(t *testing.T, store store.LowLevelStore) {
	// Generate user
	user := &users.User{
		ID:        uuid.New().String(),
		Name:      "Pepper",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	password := "supersecret"

	err := store.CreateUser(user, password)
	require.Nil(t, err)

	readUser, err := store.GetUser(user.ID)
	require.Nil(t, err)
	require.NotNil(t, readUser)
	assert.True(t, user.Equal(readUser))

	// Generate token
	token, err := store.GenerateUserPasswordResetToken(user.ID)
	require.Nil(t, err)
	assert.NotEmpty(t, token)

	// Read back the UserAuth and ensure the reset token matches
	// and our attempt useage is reset
	auth, err := store.GetUserAuth(user.ID)
	require.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, token, auth.ResetToken)
	assert.Equal(t, 0, auth.ResetTokenAttempts)
	assert.WithinDuration(t, time.Now(), auth.ResetTokenGeneratedAt, 2*time.Second)
}

func ResetPassword(t *testing.T, store store.LowLevelStore) {
	// Generate user and ensure it's created with the expected
	// password
	user := &users.User{
		ID:        uuid.New().String(),
		Name:      "Pepper",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	password := "supersecret"

	err := store.CreateUser(user, password)
	require.Nil(t, err)

	auth, err := store.GetUserAuth(user.ID)
	require.Nil(t, err)
	require.NotNil(t, auth)

	assert.NotEqual(t, password, auth.Password)
	assert.True(t, auth.CheckPassword(password))

	// Generate our reset token
	token, err := store.GenerateUserPasswordResetToken(user.ID)
	require.Nil(t, err)
	assert.NotEmpty(t, token)

	// Attempt a reset with a bad token. We should see no
	// change, get an error, and see an increment in attempts
	err = store.ResetPassword(user.ID, "bad token", "new password")
	require.NotNil(t, err)

	auth, err = store.GetUserAuth(user.ID)
	require.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, 1, auth.ResetTokenAttempts)

	// Now we will change the password and ensure it's updated
	newPassword := "even more super secret"
	err = store.ResetPassword(user.ID, auth.ResetToken, newPassword)
	require.Nil(t, err)

	auth, err = store.GetUserAuth(user.ID)
	require.Nil(t, err)
	assert.NotNil(t, auth)

	assert.False(t, auth.CheckPassword(password))
	assert.True(t, auth.CheckPassword(newPassword))
}

func DeleteUser(t *testing.T, store store.LowLevelStore) {
	user := &users.User{
		ID:        uuid.New().String(),
		Name:      "Rebecca",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	password := "supersecret"

	readUser, err := store.GetUser(user.ID)
	require.Nil(t, err)
	assert.Nil(t, readUser)

	err = store.CreateUser(user, password)
	require.Nil(t, err)

	readUser, err = store.GetUser(user.ID)
	require.Nil(t, err)
	require.NotNil(t, readUser)
	assert.True(t, user.Equal(readUser))

	err = store.DeleteUser(user.ID)
	require.Nil(t, err)

	readUser, err = store.GetUser(user.ID)
	require.Nil(t, err)
	assert.Nil(t, readUser)
}

// ===============================
// Knowledge
// ===============================

func SaveAndGetKnowledge(t *testing.T, store store.LowLevelStore) {
	knowledge := &memory.Knowledge{
		ID:      uuid.New().String(),
		User:    uuid.New().String(),
		Agent:   uuid.New().String(),
		Content: "Has anyone really been far even as decided to use even go want to do look more like?",
		Metadata: map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
		CreatedAt: time.Now(),
		Vector:    []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7},
	}

	// Ensure the knowledge is not yet there
	readKnowledge, err := store.GetKnowledge(knowledge.ID)
	require.Nil(t, err)
	assert.Nil(t, readKnowledge)

	// Write the knowledge...
	err = store.SaveKnowledge(knowledge)
	require.Nil(t, err)

	// And recall it
	readKnowledge, err = store.GetKnowledge(knowledge.ID)
	require.Nil(t, err)
	require.NotNil(t, readKnowledge)
	assert.True(t, knowledge.Equal(readKnowledge))

	// Ensure that updating it and writing it with the new information
	// works
	knowledge.Content = "A good wisdom: If your ball is too big for your mouth, it's not yours."
	knowledge.Metadata["foo"] = "baz"

	err = store.SaveKnowledge(knowledge)
	require.Nil(t, err)

	readKnowledge, err = store.GetKnowledge(knowledge.ID)
	require.Nil(t, err)
	require.NotNil(t, readKnowledge)
}

func DeleteKnowledge(t *testing.T, store store.LowLevelStore) {
	knowledge := &memory.Knowledge{
		ID:      uuid.New().String(),
		User:    uuid.New().String(),
		Agent:   uuid.New().String(),
		Content: "The term for a group of people who do not immediately kick nazis out of their presence is 'nazis'.",
		Metadata: map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
		CreatedAt: time.Now(),
		Vector:    []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7},
	}

	// Ensure the knowledge is not yet there
	readKnowledge, err := store.GetKnowledge(knowledge.ID)
	require.Nil(t, err)
	assert.Nil(t, readKnowledge)

	// Write the knowledge...
	err = store.SaveKnowledge(knowledge)
	require.Nil(t, err)

	// And recall it
	readKnowledge, err = store.GetKnowledge(knowledge.ID)
	require.Nil(t, err)
	require.NotNil(t, readKnowledge)
	require.True(t, knowledge.Equal(readKnowledge))

	// Delete it
	err = store.DeleteKnowledge(knowledge.ID)
	require.Nil(t, err)

	// And ensure it's gone
	readKnowledge, err = store.GetKnowledge(knowledge.ID)
	require.Nil(t, err)
	assert.Nil(t, readKnowledge)
}

func ListKnowledge(t *testing.T, s store.LowLevelStore) {
	// Create a set of knowledge for two different agents, two
	// different agents, and varying creation times for search.
	user1 := uuid.New().String()
	user2 := uuid.New().String()
	agent1 := uuid.New().String()
	agent2 := uuid.New().String()

	knowledge1 := &memory.Knowledge{
		ID:      uuid.New().String(),
		User:    user1,
		Agent:   agent1,
		Source:  "conversation",
		Content: "The best proof of extraterrestrial intelligence is that they haven't contacted us",
		Metadata: map[string]string{
			"Spaceman":      "Spiff",
			"Cardboard Box": "Transmorgrifier",
		},
		CreatedAt: time.Now().Add(-time.Hour * 24 * 7),
		Vector:    []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
	}
	knowledge2 := &memory.Knowledge{
		ID:      uuid.New().String(),
		User:    user1,
		Agent:   agent2,
		Source:  "document",
		Content: "The question is not whether you will love, hurt, dream, and die. It is what you will love, why you will hurt, when you will dream, and how you will die. This is your choice. You cannot pick the destination, only the path.",
		Metadata: map[string]string{
			"Shallan": "Is a bit of a nerd",
			"Adolin":  "Is a himbo",
		},
		CreatedAt: time.Now().Add(-time.Hour * 24 * 3),
		Vector:    []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
	}
	knowledge3 := &memory.Knowledge{
		ID:      uuid.New().String(),
		User:    user2,
		Agent:   agent1,
		Source:  "document",
		Content: "Moments of beauty sustain us through hours of ugliness",
		Metadata: map[string]string{
			"Kip":        "fat",
			"Blackguard": "cool",
		},
		CreatedAt: time.Now().Add(-time.Hour * 24 * 2),
		Vector:    []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
	}
	knowledge4 := &memory.Knowledge{
		ID:      uuid.New().String(),
		User:    user2,
		Agent:   agent2,
		Source:  "conversation",
		Content: "Stick them with the pointy end",
		Metadata: map[string]string{
			"Book": "better",
			"Arya": "bestest",
		},
		CreatedAt: time.Now().Add(-time.Hour * 24 * 1),
		Vector:    []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
	}
	knowledge5 := &memory.Knowledge{
		ID:      uuid.New().String(),
		User:    user1,
		Agent:   agent1,
		Source:  "user-generated",
		Content: "The man who wants you to trust him is the one you must fear the most",
		Metadata: map[string]string{
			"Ash": "fell from the sky",
		},
		CreatedAt: time.Now(),
		Vector:    []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
	}

	// Create and save all the knowledge pieces
	// They should be in order of oldest to newest
	knowledges := []*memory.Knowledge{knowledge1, knowledge2, knowledge3, knowledge4, knowledge5}
	for _, knowledge := range knowledges {
		err := s.SaveKnowledge(knowledge)
		require.Nil(t, err)
	}

	// First ensure that a blank filter and a nil filter return all knowledge
	foundKnowledge, err := s.ListKnowledge(&store.KnowledgeFilter{})
	require.Nil(t, err)
	require.NotNil(t, foundKnowledge)
	require.Len(t, foundKnowledge, len(knowledges))
	assert.True(t, knowledge1.Equal(foundKnowledge[0]))
	assert.True(t, knowledge2.Equal(foundKnowledge[1]))
	assert.True(t, knowledge3.Equal(foundKnowledge[2]))
	assert.True(t, knowledge4.Equal(foundKnowledge[3]))
	assert.True(t, knowledge5.Equal(foundKnowledge[4]))

	foundKnowledge, err = s.ListKnowledge(nil)
	require.Nil(t, err)
	require.NotNil(t, foundKnowledge)
	require.Len(t, foundKnowledge, len(knowledges))
	assert.True(t, knowledge1.Equal(foundKnowledge[0]))
	assert.True(t, knowledge2.Equal(foundKnowledge[1]))
	assert.True(t, knowledge3.Equal(foundKnowledge[2]))
	assert.True(t, knowledge4.Equal(foundKnowledge[3]))
	assert.True(t, knowledge5.Equal(foundKnowledge[4]))

	// Test the ascending flag - we should get it in reverse order
	foundKnowledge, err = s.ListKnowledge(&store.KnowledgeFilter{OldestFirst: true})
	require.Nil(t, err)
	require.NotNil(t, foundKnowledge)
	require.Len(t, foundKnowledge, len(knowledges))
	assert.True(t, knowledge5.Equal(foundKnowledge[0]))
	assert.True(t, knowledge4.Equal(foundKnowledge[1]))
	assert.True(t, knowledge3.Equal(foundKnowledge[2]))
	assert.True(t, knowledge2.Equal(foundKnowledge[3]))
	assert.True(t, knowledge1.Equal(foundKnowledge[4]))

	// Then we filter by the user and the agent
	foundKnowledge, err = s.ListKnowledge(&store.KnowledgeFilter{
		User:  &store.FilterString{Operation: store.EQ, Value: user1},
		Agent: &store.FilterString{Operation: store.EQ, Value: agent1},
	})
	require.Nil(t, err)
	require.NotNil(t, foundKnowledge)
	require.Len(t, foundKnowledge, 2)
	assert.True(t, knowledge1.Equal(foundKnowledge[0]))
	assert.True(t, knowledge5.Equal(foundKnowledge[1]))

	// Try to further narrow it by using an IN filter w/ source
	foundKnowledge, err = s.ListKnowledge(&store.KnowledgeFilter{
		User:   &store.FilterString{Operation: store.EQ, Value: user1},
		Source: &store.FilterString{Operation: store.IN, Value: "conversation, user-generated"},
	})
	require.Nil(t, err)
	require.NotNil(t, foundKnowledge)
	require.Len(t, foundKnowledge, 2)
	assert.True(t, knowledge1.Equal(foundKnowledge[0]))
	assert.True(t, knowledge5.Equal(foundKnowledge[1]))

	// Test the limit function
	foundKnowledge, err = s.ListKnowledge(&store.KnowledgeFilter{
		User:  &store.FilterString{Operation: store.EQ, Value: user2},
		Limit: 2,
	})
	require.Nil(t, err)
	require.NotNil(t, foundKnowledge)
	require.Len(t, foundKnowledge, 2)
	assert.True(t, knowledge3.Equal(foundKnowledge[0]))
	assert.True(t, knowledge4.Equal(foundKnowledge[1]))

	// Another ascending test
	foundKnowledge, err = s.ListKnowledge(&store.KnowledgeFilter{
		User:        &store.FilterString{Operation: store.EQ, Value: user2},
		Limit:       2,
		OldestFirst: true,
	})
	require.Nil(t, err)
	require.NotNil(t, foundKnowledge)
	require.Len(t, foundKnowledge, 2)
	assert.True(t, knowledge4.Equal(foundKnowledge[0]))
	assert.True(t, knowledge3.Equal(foundKnowledge[1]))

	foundKnowledge, err = s.ListKnowledge(&store.KnowledgeFilter{
		User:  &store.FilterString{Operation: store.EQ, Value: user2},
		Limit: 1,
	})
	require.Nil(t, err)
	require.NotNil(t, foundKnowledge)
	require.Len(t, foundKnowledge, 1)
	assert.True(t, knowledge3.Equal(foundKnowledge[0]))

	// Test the time/created at filters
	foundKnowledge, err = s.ListKnowledge(&store.KnowledgeFilter{
		CreatedAt: &store.FilterTime{Operation: store.GT, Value: time.Now().Add((-time.Hour * 24 * 2) - time.Minute)},
	})
	require.Nil(t, err)
	require.NotNil(t, foundKnowledge)
	require.Len(t, foundKnowledge, 3)
	assert.True(t, knowledge3.Equal(foundKnowledge[0]))
	assert.True(t, knowledge4.Equal(foundKnowledge[1]))
	assert.True(t, knowledge5.Equal(foundKnowledge[2]))

	foundKnowledge, err = s.ListKnowledge(&store.KnowledgeFilter{
		CreatedAt: &store.FilterTime{Operation: store.LT, Value: time.Now().Add((-time.Hour * 24 * 2) + time.Minute)},
	})
	require.Nil(t, err)
	require.NotNil(t, foundKnowledge)
	require.Len(t, foundKnowledge, 3)
	assert.True(t, knowledge1.Equal(foundKnowledge[0]))
	assert.True(t, knowledge2.Equal(foundKnowledge[1]))
	assert.True(t, knowledge3.Equal(foundKnowledge[2]))
}
