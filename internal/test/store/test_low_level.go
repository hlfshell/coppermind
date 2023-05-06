package store

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/agents"
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
	message := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Huey",
		Agent:        "Luey",
		From:         "Huey",
		Content:      "Where's Dewy?",
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
	message := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Yoshi",
		Agent:        "Mario",
		From:         "Yoshi",
		Content:      "Need a ride?",
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

	err = store.DeleteMessage(message.ID)
	require.Nil(t, err)

	msg, err = store.GetMessage(message.ID)
	require.Nil(t, err)
	assert.Nil(t, msg)
}

func ListMessages(t *testing.T, s store.LowLevelStore) {
	// Create a number of messages for us to query back
	// and test the list feature with
	msg1 := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Peach",
		Agent:        "Bowswer",
		From:         "Peach",
		Content:      "I need some space...",
		Conversation: uuid.New().String(),
		CreatedAt:    time.Now().Add(-5 * time.Minute),
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
	msg3 := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Peach",
		Agent:        "Mario",
		From:         "Peach",
		Content:      "I just headed over to another castle...",
		Conversation: uuid.New().String(),
		CreatedAt:    time.Now(),
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
		Columns: []*store.FilterColumn{
			{
				Column:    "user",
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

	//TODO - plenty of other tests
}

// ===============================
// Conversations
// ===============================

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

func ListSummaries(t *testing.T, store store.LowLevelStore) {
	//TODO
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
	assert.NotNil(t, readUser)
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
	assert.Equal(t, "invaid password", err.Error())
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

}

func ChangePassword(t *testing.T, store store.LowLevelStore) {
	// Generate user and ensure it's created with the expected
	// password
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

	// Now we will change the password and ensure it's updated
	newPassword := "even more super secret"
	err = store.ChangePassword(user.ID, auth.Password, newPassword)
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

	err = store.SaveUser(user, password)
	require.Nil(t, err)

	readUser, err = store.GetUser(user.ID)
	require.Nil(t, err)
	assert.NotNil(t, readUser)
	assert.True(t, user.Equal(readUser))

	err = store.DeleteUser(user.ID)
	require.Nil(t, err)

	readUser, err = store.GetUser(user.ID)
	require.Nil(t, err)
	assert.Nil(t, readUser)
}
