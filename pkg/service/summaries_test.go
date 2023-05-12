package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/llm/mock"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarize(t *testing.T) {
	llm := mock.NewMockLLM()

	service, store, err := createMockService(llm)
	require.Nil(t, err)
	require.NotNil(t, service)
	require.NotNil(t, store)

	// ==== Nonexistent conversation targeted ====
	summary, err := service.Summarize(uuid.New().String())
	require.NotNil(t, err)
	require.Nil(t, summary)

	// Create a conversation by creating multiple messages
	// that will ultimately be our target for summarizaton
	msg1 := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: uuid.New().String(),
		Agent:        testAgent.ID,
		User:         testUser.ID,
		From:         testUser.ID,
		Content:      "In a hole in the ground there lived a hobbit.",
		CreatedAt:    time.Now().Add(-10 * time.Minute),
	}
	msg2 := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: msg1.Conversation,
		Agent:        testAgent.ID,
		User:         testUser.ID,
		From:         testAgent.ID,
		Content:      "Not a nasty, dirty, wet hole, filled with the ends of worms and an oozy smell...",
		CreatedAt:    time.Now().Add(-5 * time.Minute),
	}
	msg3 := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: msg1.Conversation,
		Agent:        testAgent.ID,
		User:         testUser.ID,
		From:         testUser.ID,
		Content:      "Nor yet a dry, bare, sandy hole with nothing in it to sit down on or to eat...",
		CreatedAt:    time.Now().Add(-1 * time.Minute),
	}

	for _, msg := range []*chat.Message{msg1, msg2, msg3} {
		err = store.SaveMessage(msg)
		require.Nil(t, err)
	}

	// First we will try with no existing summary.
	returnedSummary := &memory.Summary{
		ID:                    uuid.New().String(),
		Agent:                 testAgent.ID,
		User:                  testUser.ID,
		Conversation:          msg1.Conversation,
		Keywords:              []string{"hobbit", "second breakfast"},
		UpdatedAt:             time.Now(),
		ConversationStartedAt: msg1.CreatedAt,
	}
	llm.AddSummarizeResponse(returnedSummary, nil)

	summary, err = service.Summarize(msg1.Conversation)
	require.Nil(t, err)
	require.NotNil(t, summary)
	assert.True(t, returnedSummary.Equal(summary))
	conversation, existingSummary := llm.GetSummarizeInputs()
	require.NotNil(t, conversation)
	assert.Nil(t, existingSummary)

	// Compare the summary to what exists for our messages
	expectedConversation, err := store.GetConversation(msg1.Conversation)
	require.Nil(t, err)
	assert.True(t, expectedConversation.Equal(conversation))

	// Now we will try again, but return nil as our summary
	// This should result in an exclusion being noted
	// for our conversation

	// First let's ensure that we would have seen the conversation
	// via our GetConversationsToSummarize call
	// First we have to remove the existing summary
	err = store.DeleteSummary(summary.ID)
	require.Nil(t, err)

	conversations, err := store.GetConversationsToSummarize(
		0,
		0,
		0,
	)
	require.Nil(t, err)
	require.NotNil(t, conversations)
	require.Len(t, conversations, 1)
	assert.Equal(t, msg1.Conversation, conversations[0])

	llm.ClearMemory()
	err = store.DeleteSummary(summary.ID)
	require.Nil(t, err)
	llm.AddSummarizeResponse(nil, nil)
	summary, err = service.Summarize(msg1.Conversation)
	require.Nil(t, err)
	assert.Nil(t, summary)

	// Now we should no longer see it as an option to recall
	conversations, err = store.GetConversationsToSummarize(
		0,
		0,
		0,
	)
	require.Nil(t, err)
	require.NotNil(t, conversations)
	assert.Len(t, conversations, 0)

	// Now we will run it through with a previously existing
	// summary and ensure it's included and sent to the
	// LLM service
	err = store.SaveSummary(returnedSummary)
	require.Nil(t, err)
	err = store.DeleteSummaryExclusion(msg1.Conversation)
	require.Nil(t, err)

	llm.ClearMemory()
	llm.AddSummarizeResponse(returnedSummary, nil)

	summary, err = service.Summarize(msg1.Conversation)
	require.Nil(t, err)
	require.NotNil(t, summary)
	assert.True(t, returnedSummary.Equal(summary))
	conversation, existingSummary = llm.GetSummarizeInputs()
	require.NotNil(t, conversation)
	require.NotNil(t, existingSummary)
	assert.Equal(t, msg1.Conversation, conversation.ID)
	assert.True(t, returnedSummary.Equal(existingSummary))
}

func TestGetSummaries(t *testing.T) {
	llm := mock.NewMockLLM()

	service, store, err := createMockService(llm)
	require.Nil(t, err)
	require.NotNil(t, service)
	require.NotNil(t, store)

	// --- Test Invalid Requests ---
	// No agent
	request := &GetSummariesRequest{
		Agent:  "",
		User:   testUser.ID,
		Time:   time.Now(),
		Before: false,
	}
	summaries, err := service.Summary.GetSummaries(request)
	assert.NotNil(t, err)
	assert.Nil(t, summaries)

	// No user
	request = &GetSummariesRequest{
		Agent:  testAgent.ID,
		User:   "",
		Time:   time.Now(),
		Before: false,
	}
	summaries, err = service.Summary.GetSummaries(request)
	assert.NotNil(t, err)
	assert.Nil(t, summaries)

	// Zero time
	request = &GetSummariesRequest{
		Agent:  testAgent.ID,
		User:   testUser.ID,
		Time:   time.Time{},
		Before: false,
	}
	summaries, err = service.Summary.GetSummaries(request)
	assert.NotNil(t, err)
	assert.Nil(t, summaries)

	// --- Test Valid Requests ---

	// Create three summaries for us to search through - two of
	// the same user/agent combo, the third of a different user
	summary1 := &memory.Summary{
		ID:                    uuid.New().String(),
		Agent:                 testAgent.ID,
		User:                  testUser.ID,
		Conversation:          uuid.New().String(),
		Keywords:              []string{"amazing", "show"},
		Summary:               "Life is like a hurricane, here in Duckburg",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: time.Now().Add(-time.Hour),
	}
	summary2 := &memory.Summary{
		ID:                    uuid.New().String(),
		Agent:                 testAgent.ID,
		User:                  testUser.ID,
		Conversation:          uuid.New().String(),
		Keywords:              []string{"amazing", "show"},
		Summary:               "Race cars, lasers, aeroplanes, it's a duck blur",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: time.Now().Add(-30 * time.Minute),
	}
	summary3 := &memory.Summary{
		ID:                    uuid.New().String(),
		Agent:                 testAgent.ID,
		User:                  uuid.New().String(),
		Conversation:          uuid.New().String(),
		Keywords:              []string{"amazing", "show"},
		Summary:               "Might solve a mystery, or rewrite history",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: time.Now().Add(-time.Hour),
	}

	for _, summary := range []*memory.Summary{summary1, summary2, summary3} {
		err = store.SaveSummary(summary)
		require.Nil(t, err)
	}

	// Test getting all summaries for testUser with a time set to
	// before the first summary's conversation start
	request = &GetSummariesRequest{
		Agent:  testAgent.ID,
		User:   testUser.ID,
		Time:   summary1.ConversationStartedAt.Add(-time.Minute),
		Before: false,
	}
	summaries, err = service.Summary.GetSummaries(request)
	require.Nil(t, err)
	require.NotNil(t, summaries)
	require.Len(t, summaries, 2)
	assert.True(t, summary1.Equal(summaries[0]))
	assert.True(t, summary2.Equal(summaries[1]))

	// Confirm we can get back the other user's summaries
	request = &GetSummariesRequest{
		Agent:  testAgent.ID,
		User:   summary3.User,
		Time:   summary3.ConversationStartedAt.Add(-time.Minute),
		Before: false,
	}
	summaries, err = service.Summary.GetSummaries(request)
	require.Nil(t, err)
	require.NotNil(t, summaries)

	// We expect these to be in ConversationCreatedAt order
	require.Len(t, summaries, 1)
	assert.True(t, summary3.Equal(summaries[0]))

	// Ensure that if we set the time between our first two summaries
	// we appropriately filter based on ConversationCreatedAt
	request = &GetSummariesRequest{
		Agent:  testAgent.ID,
		User:   testUser.ID,
		Time:   summary1.ConversationStartedAt.Add(15 * time.Minute),
		Before: false,
	}

	summaries, err = service.Summary.GetSummaries(request)
	require.Nil(t, err)
	require.NotNil(t, summaries)
	require.Len(t, summaries, 1)
	assert.True(t, summary2.Equal(summaries[0]))

	// If we swap the Before setting, we should be looking
	// at older summaries than our specified time, so summary1
	// will be returned
	request.Before = true
	summaries, err = service.Summary.GetSummaries(request)
	require.Nil(t, err)
	require.NotNil(t, summaries)
	require.Len(t, summaries, 1)

	assert.True(t, summary1.Equal(summaries[0]))
}
