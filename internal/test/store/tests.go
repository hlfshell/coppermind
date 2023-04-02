package store

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/chat"
	"github.com/hlfshell/coppermind/internal/memory"
	"github.com/hlfshell/coppermind/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetConversation(t *testing.T, store store.Store) {
	conversationId := uuid.New().String()
	agent := "Rose"
	user := "Keith"

	//Assert that it returns nothing in the nil case first
	nullConversation, err := store.GetConversation(conversationId)
	require.Nil(t, err)
	assert.Nil(t, nullConversation)

	oldest := &chat.Message{
		ID:           uuid.New().String(),
		Agent:        agent,
		User:         user,
		Conversation: conversationId,
		Content:      "Hub-bub",
		Tone:         "sarcastic",
		CreatedAt:    time.Now().Add(-1 * time.Hour),
	}
	older := &chat.Message{
		ID:           uuid.New().String(),
		Agent:        agent,
		User:         user,
		Conversation: conversationId,
		Content:      "Hub-bub",
		Tone:         "sarcastic",
		CreatedAt:    time.Now().Add(-30 * time.Minute),
	}
	newest := &chat.Message{
		ID:           uuid.New().String(),
		Agent:        agent,
		User:         user,
		Conversation: conversationId,
		Content:      "Hub-bub",
		Tone:         "sarcastic",
		CreatedAt:    time.Now().Add(-5 * time.Minute),
	}
	redHerring := &chat.Message{
		ID:           uuid.New().String(),
		Agent:        agent,
		User:         "Not User",
		Conversation: conversationId,
		Content:      "Hub-bub",
		Tone:         "sarcastic",
		CreatedAt:    time.Now(),
	}
	msgs := []*chat.Message{oldest, older, newest, redHerring}
	for _, msg := range msgs {
		err = store.SaveMessage(msg)
		require.Nil(t, err)
	}

	retrievedConversation, err := store.GetConversation(conversationId)
	require.Nil(t, err)
	assert.NotNil(t, retrievedConversation)

	expectedConversation := &chat.Conversation{
		ID:        conversationId,
		User:      user,
		Agent:     agent,
		CreatedAt: oldest.CreatedAt,
		Messages:  msgs,
	}
	assert.True(t, expectedConversation.Equal(retrievedConversation))
}

func GetLatestConversation(t *testing.T, store store.Store) {
	agent := "Rose"
	user := "Keith"

	//Assert that it returns nothing in the nil case first
	latestConversation, timestamp, err := store.GetLatestConversation(agent, user)
	require.Nil(t, err)
	assert.Equal(t, latestConversation, "")
	assert.Equal(t, time.Time{}, timestamp)

	oldest := chat.Message{
		ID:           uuid.New().String(),
		Agent:        agent,
		User:         user,
		CreatedAt:    time.Now().Add(-time.Hour),
		Content:      "Blah blah blah",
		Tone:         "sarcastic",
		Conversation: uuid.New().String(),
	}
	older := chat.Message{
		ID:           uuid.New().String(),
		Agent:        agent,
		User:         user,
		CreatedAt:    time.Now().Add(-30 * time.Minute),
		Content:      "Blah blah blah",
		Tone:         "sassy",
		Conversation: uuid.New().String(),
	}
	latest := chat.Message{
		ID:           uuid.New().String(),
		Agent:        agent,
		User:         user,
		CreatedAt:    time.Now(),
		Content:      "Blah blah blah",
		Tone:         "malicious",
		Conversation: uuid.New().String(),
	}

	for _, message := range []*chat.Message{&oldest, &older, &latest} {
		err := store.SaveMessage(message)
		require.Nil(t, err)
	}

	latestConversation, timestamp, err = store.GetLatestConversation(agent, user)
	require.Nil(t, err)
	assert.Equal(t, latest.Conversation, latestConversation)
	assert.WithinDuration(t, latest.CreatedAt, timestamp, time.Second)
}

func SaveMessage(t *testing.T, store store.Store) {
	message := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Huey",
		Agent:        "Luey",
		Content:      "Where's Dewy?",
		Tone:         "inquisitive",
		Conversation: uuid.New().String(),
	}

	conversation, err := store.GetConversation(message.Conversation)
	require.Nil(t, err)
	assert.Nil(t, conversation)

	err = store.SaveMessage(message)
	require.Nil(t, err)

	conversation, err = store.GetConversation(message.Conversation)
	require.Nil(t, err)
	assert.Equal(t, 1, len(conversation.Messages))
	assert.True(t, message.Equal(conversation.Messages[0]))
}

func SaveSummary(t *testing.T, store store.Store) {
	conversation := uuid.New().String()

	//Check the null case
	retrievedSummary, err := store.GetSummaryByConversation(conversation)
	require.Nil(t, err)
	assert.Nil(t, retrievedSummary)

	//Create the summary and save, ensure we can read it back
	summary := &memory.Summary{
		ID:                    uuid.New().String(),
		Agent:                 "R2D2",
		User:                  "Luke",
		Conversation:          conversation,
		Keywords:              []string{"astronavigation", "ship repairs"},
		Summary:               "boop beep beep boop",
		ConversationStartedAt: time.Now(),
		UpdatedAt:             time.Now(),
	}
	err = store.SaveSummary(summary)
	require.Nil(t, err)

	retrievedSummary, err = store.GetSummaryByConversation(conversation)
	require.Nil(t, err)
	assert.NotNil(t, retrievedSummary)
	assert.True(t, summary.Equal(retrievedSummary))

	// Now that we have an existing summary, we should be able to
	// update it. The SaveSummary function assumes "upsert"
	// functionality. Likewise, we wait 1.5 seconds to ensure that
	// the updated timestamp is different (we don't have better than
	// second granularity). The time should be updated at the point
	// of saving the summary as we depend upon it being updated.
	time.Sleep(1*time.Second + 500*time.Millisecond)
	summaryOriginalTime := summary.UpdatedAt

	newSummary := "Beep boop boop beep"
	summary.Summary = newSummary
	err = store.SaveSummary(summary)
	require.Nil(t, err)
	retrievedSummary, err = store.GetSummaryByConversation(conversation)
	require.Nil(t, err)
	assert.NotNil(t, retrievedSummary)
	assert.True(t, summary.Equal(retrievedSummary))
	assert.Equal(t, newSummary, retrievedSummary.Summary)
	assert.Less(t, summaryOriginalTime, retrievedSummary.UpdatedAt)
}

func GetConversationsToSummarize(t *testing.T, store store.Store) {
	// 1. First let's ensure the null case - there are no summaries to find.
	conversations, err := store.GetConversationsToSummarize(0, 0, 0)
	require.Nil(t, err)
	assert.Equal(t, 0, len(conversations))

	// 2. Now we create a set of messages with summaries that
	// should be excluded for that reason
	//Create a situation where all summaries have been
	//summarized and thus we still have a null return
	msg1 := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Abby",
		Agent:        "Rebecca",
		Content:      "Carrot?",
		Tone:         "inquisitive",
		Conversation: uuid.New().String(),
		CreatedAt:    time.Now().Add(-5 * time.Minute),
	}
	summary1 := &memory.Summary{
		ID:                    uuid.New().String(),
		Agent:                 msg1.Agent,
		User:                  msg1.User,
		Conversation:          msg1.Conversation,
		Keywords:              []string{"food", "puppy"},
		Summary:               "A wonderful puppy inquires about a snack",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: msg1.CreatedAt,
	}
	msg2 := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Abby",
		Agent:        "Rebecca",
		Content:      "Leucadia's?",
		Tone:         "inquisitive",
		Conversation: uuid.New().String(),
		CreatedAt:    time.Now().Add(-5 * time.Minute),
	}
	summary2 := &memory.Summary{
		ID:                    uuid.New().String(),
		Agent:                 msg1.Agent,
		User:                  msg1.User,
		Conversation:          msg2.Conversation,
		Keywords:              []string{"food", "puppy"},
		Summary:               "A cute puppy inquires about a proper meal",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: msg2.CreatedAt,
	}

	for _, msg := range []*chat.Message{msg1, msg2} {
		err = store.SaveMessage(msg)
		require.Nil(t, err)
	}
	for _, summary := range []*memory.Summary{summary1, summary2} {
		err = store.SaveSummary(summary)
		require.Nil(t, err)
	}

	conversations, err = store.GetConversationsToSummarize(0, 0, 0)
	require.Nil(t, err)
	assert.Equal(t, 0, len(conversations))

	// 3. No let's create a set of messages with no summaries so that they get
	// listed.
	msg3 := &chat.Message{
		ID:           uuid.New().String(),
		User:         "Abby",
		Agent:        "Rebecca",
		Content:      "How about a quick calzone?",
		Tone:         "worried",
		Conversation: uuid.New().String(),
		CreatedAt:    time.Now().Add(-5 * time.Minute),
	}
	err = store.SaveMessage(msg3)
	require.Nil(t, err)
	conversations, err = store.GetConversationsToSummarize(0, 0, 0)
	require.Nil(t, err)
	assert.Equal(t, 1, len(conversations))
	assert.Equal(t, msg3.Conversation, conversations[0])

	// 4. Now set the minimum age so the conversation is ignore as it's
	// not old enough to be detected
	conversations, err = store.GetConversationsToSummarize(0, 10*time.Minute, 5)
	require.Nil(t, err)
	assert.Equal(t, 0, len(conversations))

	// 5. Revert the age limit and raise the min messages limit to show that it's
	// not detected either
	conversations, err = store.GetConversationsToSummarize(3, 0, 5)
	require.Nil(t, err)
	assert.Equal(t, 0, len(conversations))

	// 6. Mark the message as excluded so it's ignored
	err = store.ExcludeConversationFromSummary(msg3.Conversation)
	require.Nil(t, err)
	conversations, err = store.GetConversationsToSummarize(0, 0, 0)
	require.Nil(t, err)
	assert.Equal(t, 0, len(conversations))

	// 7. Create a long conversation. Make sure the latest message is above
	// the minimum age cutoff (too new), but has 1 less than the max messages
	// allowed before summarizaton. It should not be picked up. Add another
	// message and then immediately recheck to see if get picked up
	messages := []*chat.Message{}
	conversation := uuid.New().String()
	for i := 0; i < 4; i++ {
		messages = append(messages, &chat.Message{
			ID:           uuid.New().String(),
			Conversation: conversation,
			Agent:        "Rebecca",
			User:         "Abby",
			Content:      "This is a long conversation",
			CreatedAt:    time.Now().Add(-5*time.Minute + time.Duration(i)*time.Minute),
			Tone:         "neutral",
		})
	}
	for _, msg := range messages {
		err = store.SaveMessage(msg)
		require.Nil(t, err)
	}

	conversations, err = store.GetConversationsToSummarize(3, 10*time.Minute, 5)
	require.Nil(t, err)
	assert.Equal(t, 0, len(conversations))

	err = store.SaveMessage(&chat.Message{
		ID:           uuid.New().String(),
		Conversation: conversation,
		Agent:        "Rebecca",
		User:         "Abby",
		Content:      "This is a long conversation",
		CreatedAt:    time.Now(),
		Tone:         "neutral",
	})
	require.Nil(t, err)

	conversations, err = store.GetConversationsToSummarize(3, 10*time.Minute, 5)
	require.Nil(t, err)
	assert.Equal(t, 1, len(conversations))
	assert.Equal(t, conversation, conversations[0])

	// 8. If the above conversation has a summary, however, it should not
	// be picked up by our function unless we have some max messages ABOVE
	// the last updated_at date of the summary. Thus not until we add
	// multiple new messages will this summray be picked up again
	longSummary := &memory.Summary{
		ID:           uuid.New().String(),
		Agent:        "Rebecca",
		User:         "Abby",
		Conversation: conversation,
		Keywords:     []string{"meaning of life", "puppy"},
		Summary:      "A long conversation about fetch",
		UpdatedAt:    time.Now(),
	}
	err = store.SaveSummary(longSummary)
	require.Nil(t, err)

	// Ensure that it's no longer listed post summary
	conversations, err = store.GetConversationsToSummarize(3, 10*time.Minute, 5)
	require.Nil(t, err)
	assert.Equal(t, 0, len(conversations))

	// Now add a few messages (under the 5 needed to trigger a new summary)
	// and ensure it's still not picked up. Note that each message should
	// have a CreatedAt time after our summary's updated_at
	for i := 0; i < 4; i++ {
		err = store.SaveMessage(&chat.Message{
			ID:           uuid.New().String(),
			Conversation: conversation,
			Agent:        "Rebecca",
			User:         "Abby",
			Content:      "This is a long conversation",
			CreatedAt:    time.Now().Add(time.Minute * time.Duration(i+1)),
			Tone:         "neutral",
		})
		require.Nil(t, err)
	}
	conversations, err = store.GetConversationsToSummarize(3, 10*time.Minute, 5)
	require.Nil(t, err)
	assert.Equal(t, 0, len(conversations))

	// Finally add the fifth message to this forcing the conversation to be
	// selected
	err = store.SaveMessage(&chat.Message{
		ID:           uuid.New().String(),
		Conversation: conversation,
		Agent:        "Rebecca",
		User:         "Abby",
		Content:      "This is a long conversation",
		CreatedAt:    time.Now().Add(time.Minute * 6),
		Tone:         "neutral",
	})
	require.Nil(t, err)
	conversations, err = store.GetConversationsToSummarize(3, 10*time.Minute, 5)
	require.Nil(t, err)
	require.Equal(t, 1, len(conversations))
	assert.Equal(t, conversation, conversations[0])
}

func GetSummaryByConversation(t *testing.T, store store.Store) {
	conversation := uuid.New().String()
	msg := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: conversation,
		Agent:        "Bob",
		User:         "Alice",
		Content:      "Super secret stuff",
		CreatedAt:    time.Now(),
		Tone:         "hushed",
	}
	err := store.SaveMessage(msg)
	require.Nil(t, err)

	//Ensure nothing is returned as no summary is created
	//yet
	summary, err := store.GetSummaryByConversation(conversation)
	require.Nil(t, err)
	assert.Nil(t, summary)

	//Now create the summary
	createdSummary := &memory.Summary{
		ID:                    uuid.New().String(),
		Agent:                 "Bob",
		User:                  "Alice",
		Keywords:              []string{"secret", "stuff"},
		Summary:               "Bob plots with Alice",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: msg.CreatedAt,
		Conversation:          conversation,
	}
	err = store.SaveSummary(createdSummary)
	require.Nil(t, err)

	summary, err = store.GetSummaryByConversation(conversation)
	require.Nil(t, err)
	assert.NotNil(t, summary)
	assert.True(t, createdSummary.Equal(summary))
}

func GetSummariesByAgentAndUser(t *testing.T, store store.Store) {
	agent := "Bill"
	user := "Ted"
	summaries, err := store.GetSummariesByAgentAndUser(agent, user)
	require.Nil(t, err)
	assert.Equal(t, 0, len(summaries))

	summary := &memory.Summary{
		ID:                    uuid.New().String(),
		Conversation:          uuid.New().String(),
		Agent:                 agent,
		User:                  user,
		Keywords:              []string{"woah", "dude"},
		Summary:               "Be excellent to eachother",
		UpdatedAt:             time.Now(),
		ConversationStartedAt: time.Now().Add(-time.Hour),
	}

	err = store.SaveSummary(summary)
	require.Nil(t, err)

	summaries, err = store.GetSummariesByAgentAndUser(agent, user)
	require.Nil(t, err)
	assert.Equal(t, 1, len(summaries))
	assert.True(t, summary.Equal(summaries[0]))
}

func ExcludeConversationFromSummary(t *testing.T, store store.Store) {
	// Create a conversation, ensure we can detect it, then create the
	// exclusion. We should then not see it.

	conversation := uuid.New().String()
	msg := &chat.Message{
		ID:           uuid.New().String(),
		Conversation: conversation,
		User:         "Abby",
		Agent:        "Abbigators",
		Content:      "You do bring up a good point about the socioeconomic implications of that",
		Tone:         "resigned",
		CreatedAt:    time.Now(),
	}
	err := store.SaveMessage(msg)
	require.Nil(t, err)

	conversations, err := store.GetConversationsToSummarize(0, 0, 0)
	require.Nil(t, err)
	require.Equal(t, 1, len(conversations))
	assert.Equal(t, conversation, conversations[0])

	err = store.ExcludeConversationFromSummary(conversation)
	require.Nil(t, err)

	conversations, err = store.GetConversationsToSummarize(0, 0, 0)
	require.Nil(t, err)
	assert.Equal(t, 0, len(conversations))
}
