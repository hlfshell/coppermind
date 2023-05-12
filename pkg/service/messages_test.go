package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/llm/mock"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMessages(t *testing.T) {
	llm := mock.NewMockLLM()

	service, store, err := createMockService(llm)
	require.Nil(t, err)
	require.NotNil(t, service)
	require.NotNil(t, store)

	// --- Test Invalid Requests ---
	// No agent
	request := &GetMessagesRequest{
		Agent:  "",
		User:   testUser.ID,
		Time:   time.Now(),
		Before: false,
	}
	messages, err := service.Messages.GetMessages(request)
	assert.NotNil(t, err)
	assert.Nil(t, messages)

	// No user
	request = &GetMessagesRequest{
		Agent:  testAgent.ID,
		User:   "",
		Time:   time.Now(),
		Before: false,
	}
	messages, err = service.Messages.GetMessages(request)
	assert.NotNil(t, err)
	assert.Nil(t, messages)

	// Zero time
	request = &GetMessagesRequest{
		Agent:  testAgent.ID,
		User:   testUser.ID,
		Time:   time.Time{},
		Before: false,
	}
	messages, err = service.Messages.GetMessages(request)
	assert.NotNil(t, err)
	assert.Nil(t, messages)

	// --- Test Valid Requests ---

	// First we create three messages amongst the same user
	// and agent, and another two amongst another user
	secondUser := uuid.New().String()

	msg1 := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: uuid.New().String(),
		User:         testUser.ID,
		Agent:        testAgent.ID,
		From:         testUser.Name,
		Content:      "This is the song that never ends",
		CreatedAt:    time.Now().Add(-5 * time.Minute),
	}
	msg2 := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: msg1.Conversation,
		User:         testUser.ID,
		Agent:        testAgent.ID,
		From:         testAgent.Name,
		Content:      "Yes it goes on and on my friends",
		CreatedAt:    time.Now().Add(-4 * time.Minute),
	}
	msg3 := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: uuid.New().String(),
		User:         testUser.ID,
		Agent:        testAgent.ID,
		From:         testUser.Name,
		Content:      "Some people started singing it not knowing what it was",
		CreatedAt:    time.Now().Add(-3 * time.Minute),
	}
	// Two for another user
	msg4 := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: uuid.New().String(),
		User:         secondUser,
		Agent:        testAgent.ID,
		From:         secondUser,
		Content:      "And they'll continue singing it forever just because",
		CreatedAt:    time.Now().Add(-5 * time.Minute),
	}
	msg5 := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: msg4.Conversation,
		User:         secondUser,
		Agent:        testAgent.ID,
		From:         testAgent.Name,
		Content:      "This is the song that never ends!",
		CreatedAt:    time.Now().Add(-4 * time.Minute),
	}

	// Create all the messages
	for _, msg := range []*chat.Message{msg1, msg2, msg3, msg4, msg5} {
		err = store.SaveMessage(msg)
		require.Nil(t, err)
	}

	// For our first attempt we will request all messages with the first
	// user and agent that were created after the first message
	request = &GetMessagesRequest{
		Agent: testAgent.ID,
		User:  testUser.ID,
		Time:  msg1.CreatedAt.Add(-1 * time.Second),
	}

	// We expect to get back all three messages
	messages, err = service.Messages.GetMessages(request)
	require.Nil(t, err)
	require.Len(t, messages, 3)

	// We expect the messages to be in the order they were created
	assert.True(t, msg1.Equal(messages[0]))
	assert.True(t, msg2.Equal(messages[1]))
	assert.True(t, msg3.Equal(messages[2]))

	// Recall the other user/agent combo
	request = &GetMessagesRequest{
		Agent:  testAgent.ID,
		User:   secondUser,
		Time:   msg4.CreatedAt.Add(-1 * time.Second),
		Before: false,
	}

	// We expect to get back both messages
	messages, err = service.Messages.GetMessages(request)
	require.Nil(t, err)
	require.Len(t, messages, 2)

	// We expect the messages to be in the order they were created
	assert.True(t, msg4.Equal(messages[0]))
	assert.True(t, msg5.Equal(messages[1]))

	// Let's search for the first user/agent combo, but utilize
	// a time between the first and last two messages
	request = &GetMessagesRequest{
		Agent:  testAgent.ID,
		User:   testUser.ID,
		Time:   msg1.CreatedAt.Add(30 * time.Second),
		Before: false,
	}

	// We expect to get back the last two messages
	messages, err = service.Messages.GetMessages(request)
	require.Nil(t, err)

	assert.True(t, msg2.Equal(messages[0]))
	assert.True(t, msg3.Equal(messages[1]))

	// If we swap the Before to True, we should get back the first
	// message
	request.Before = true
	messages, err = service.Messages.GetMessages(request)
	require.Nil(t, err)
	require.Len(t, messages, 1)

	assert.True(t, msg1.Equal(messages[0]))
}
