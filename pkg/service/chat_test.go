package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/llm/mock"
	"github.com/hlfshell/coppermind/pkg/artifacts"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendMessage(t *testing.T) {
	// ==== Happy paths ====

	// -- No conversation history -- //

	llm := mock.NewMockLLM()

	service, store, err := createMockService(llm)
	require.Nil(t, err)
	require.NotNil(t, service)

	returnMsg := &chat.Message{
		ID:        uuid.New().String(),
		Agent:     testAgent.ID,
		User:      testUser.ID,
		From:      testAgent.ID,
		Content:   "Hello, world!",
		Artifacts: []*artifacts.ArtifactData{},
		CreatedAt: time.Now(),
	}
	conversationId := returnMsg.Conversation

	llm.AddSendMessageResponse(returnMsg, nil)

	msg, err := service.SendMessage(returnMsg)
	require.Nil(t, err)
	require.NotNil(t, msg)

	// We expect the conversation to get set by the service
	// in this situation, so we expect it to not be blank
	assert.NotEqual(t, conversationId, msg.Conversation)
	assert.True(t, returnMsg.Equal(msg))

	// -- No conversation history, conversation specified -- //

	llm.ClearMemory()
	returnMsg.Conversation = uuid.New().String()
	conversationId = returnMsg.Conversation
	llm.AddSendMessageResponse(returnMsg, nil)
	msg, err = service.SendMessage(returnMsg)
	require.Nil(t, err)
	require.NotNil(t, msg)

	assert.Equal(t, conversationId, msg.Conversation)
	assert.True(t, returnMsg.Equal(msg))

	// -- Conversation history, new conversation -- //

	llm.ClearMemory()
	llm.AddConversationContinuanceResponse(false, nil)
	// Reset the conversation id
	returnMsg.Conversation = ""
	conversationId = returnMsg.Conversation
	llm.AddSendMessageResponse(returnMsg, nil)

	msg, err = service.SendMessage(returnMsg)
	require.Nil(t, err)
	require.NotNil(t, msg)

	// We expect the conversation to get set by the service
	assert.NotEqual(t, conversationId, msg.Conversation)
	assert.True(t, returnMsg.Equal(msg))

	// Now we test to ensure if a conversation exists
	// already, that we utilize that conversation ID

	// -- Conversation history, continue conversation -- //

	llm.ClearMemory()
	llm.AddConversationContinuanceResponse(true, nil)
	// Save the old message as a prior conversation
	store.SaveMessage(returnMsg)
	// Reset the conversation id
	oldConversationId := returnMsg.Conversation
	returnMsg.Conversation = ""
	conversationId = returnMsg.Conversation
	llm.AddSendMessageResponse(returnMsg, nil)

	msg, err = service.SendMessage(returnMsg)
	require.Nil(t, err)
	require.NotNil(t, msg)

	assert.NotEqual(t, conversationId, msg.Conversation)
	assert.Equal(t, oldConversationId, msg.Conversation)

	// -- Conversation history, includes summaries -- //

	llm.ClearMemory()

	// Create a summary for the existing message conversation
	summary := &memory.Summary{
		ID:                    uuid.New().String(),
		Conversation:          returnMsg.Conversation,
		Agent:                 returnMsg.Agent,
		User:                  returnMsg.User,
		Keywords:              []string{"hello", "world"},
		Summary:               "A fake summary for a fake conversation",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: time.Now(),
	}
	err = store.SaveSummary(summary)
	require.Nil(t, err)

	llm.AddConversationContinuanceResponse(true, nil)
	llm.AddSendMessageResponse(returnMsg, nil)

	msg, err = service.SendMessage(returnMsg)
	require.Nil(t, err)
	require.NotNil(t, msg)

	assert.NotEqual(t, conversationId, msg.Conversation)
	assert.True(t, returnMsg.Equal(msg))

	// Let's look at the incoming summaries and confirm that the
	// summary was included and passed.
	_, _, pastSummaries, _, _ := llm.GetSendMessageInputs()
	require.Equal(t, 1, len(pastSummaries))
	assert.True(t, summary.Equal(pastSummaries[0]))

	// ==== Error paths ====

	// -- Agent does not exist -- //

	llm.ClearMemory()
	returnMsg.Agent = uuid.New().String()
	msg, err = service.SendMessage(returnMsg)
	require.NotNil(t, err)
	assert.Nil(t, msg)
}
