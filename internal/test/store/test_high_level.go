package store

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		From:         agent,
		CreatedAt:    time.Now().Add(-time.Hour),
		Content:      "Blah blah blah",
		Conversation: uuid.New().String(),
	}
	older := chat.Message{
		ID:           uuid.New().String(),
		Agent:        agent,
		User:         user,
		From:         user,
		CreatedAt:    time.Now().Add(-30 * time.Minute),
		Content:      "Blah blah blah",
		Conversation: uuid.New().String(),
	}
	latest := chat.Message{
		ID:           uuid.New().String(),
		Agent:        agent,
		User:         user,
		From:         agent,
		CreatedAt:    time.Now(),
		Content:      "Blah blah blah",
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
		From:         "Abby",
		Content:      "Carrot?",
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
		From:         "Abby",
		Content:      "Leucadia's?",
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
		From:         "Abby",
		Content:      "How about a quick calzone?",
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
			From:         "Rebecca",
			Content:      "This is a long conversation",
			CreatedAt:    time.Now().Add(-5*time.Minute + time.Duration(i)*time.Minute),
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
		From:         "Rebecca",
		Content:      "This is a long conversation",
		CreatedAt:    time.Now(),
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
			From:         "Abby",
			Content:      "This is a long conversation",
			CreatedAt:    time.Now().Add(time.Minute * time.Duration(i+1)),
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
		From:         "Rebecca",
		Content:      "This is a long conversation",
		CreatedAt:    time.Now().Add(time.Minute * 6),
	})
	require.Nil(t, err)
	conversations, err = store.GetConversationsToSummarize(3, 10*time.Minute, 5)
	require.Nil(t, err)
	require.Equal(t, 1, len(conversations))
	assert.Equal(t, conversation, conversations[0])
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
		From:         "Abby",
		Content:      "You do bring up a good point about the socioeconomic implications of that",
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

	// Now ensure that we can remove the exclusion and see it
	err = store.DeleteSummaryExclusion(conversation)
	require.Nil(t, err)

	conversations, err = store.GetConversationsToSummarize(0, 0, 0)
	require.Nil(t, err)
	require.Equal(t, 1, len(conversations))
	assert.Equal(t, conversation, conversations[0])
}

// func ExpireKnowledge(t *testing.T, store store.Store) {
// 	// Create three knowledge entries. Ensure they can be read back.
// 	// Then ensure that we can remove them by age, leaving the non-
// 	// expired ones.
// 	knowledge1 := &memory.Knowledge{
// 		ID:        uuid.New().String(),
// 		Agent:     "Rose",
// 		User:      "Abby",
// 		Subject:   "Abby",
// 		Predicate: "is",
// 		Object:    "hungry",
// 		CreatedAt: time.Now().Add(-1 * time.Hour),
// 		ExpiresAt: time.Now().Add(time.Hour),
// 	}
// 	knowledge2 := &memory.Knowledge{
// 		ID:        uuid.New().String(),
// 		Agent:     "Rose",
// 		User:      "Keith",
// 		Subject:   "Keith",
// 		Predicate: "programmed",
// 		Object:    "Rose",
// 		CreatedAt: time.Now().Add(-2 * time.Hour),
// 		ExpiresAt: time.Now().Add(-1 * time.Hour),
// 	}

// 	for _, fact := range []*memory.Knowledge{knowledge1, knowledge2} {
// 		err := store.SaveKnowledge(fact)
// 		require.Nil(t, err)
// 	}

// 	// Ensure they're there.
// 	facts, err := store.GetKnowlegeByAgentAndUser(knowledge1.Agent, knowledge1.User)
// 	require.Nil(t, err)
// 	require.Equal(t, 1, len(facts))
// 	assert.True(t, knowledge1.Equal(facts[0]))

// 	facts, err = store.GetKnowlegeByAgentAndUser(knowledge2.Agent, knowledge2.User)
// 	require.Nil(t, err)
// 	require.Equal(t, 1, len(facts))
// 	assert.True(t, knowledge2.Equal(facts[0]))

// 	// Now expire the older fact (knowledge2)
// 	err = store.ExpireKnowledge()
// 	require.Nil(t, err)

// 	facts, err = store.GetKnowlegeByAgentAndUser(knowledge1.Agent, knowledge1.User)
// 	require.Nil(t, err)
// 	require.Equal(t, 1, len(facts))
// 	assert.True(t, knowledge1.Equal(facts[0]))

// 	facts, err = store.GetKnowlegeByAgentAndUser(knowledge2.Agent, knowledge2.User)
// 	require.Nil(t, err)
// 	require.Equal(t, 0, len(facts))
// }

// func SetConversationAsKnowledgeExtracted(t *testing.T, store store.Store) {
// 	conversations, err := store.GetConversationsToExtractKnowledge()
// 	require.Nil(t, err)
// 	assert.Equal(t, 0, len(conversations))

// 	conversation := uuid.New().String()
// 	message := &chat.Message{
// 		ID:           uuid.New().String(),
// 		Conversation: conversation,
// 		Agent:        "Rose",
// 		User:         "Keith",
// 		Content:      "Beep boop I'm a robot!",
// 		CreatedAt:    time.Now().Add(-5 * time.Minute),
// 	}
// 	err = store.SaveMessage(message)
// 	require.Nil(t, err)

// 	// Ensure we have the conversation present and selected
// 	conversations, err = store.GetConversationsToExtractKnowledge()
// 	require.Nil(t, err)
// 	require.Equal(t, 1, len(conversations))
// 	assert.Equal(t, conversation, conversations[0])

// 	// Now mark it as extracted
// 	err = store.SetConversationAsKnowledgeExtracted(conversation)
// 	require.Nil(t, err)

// 	conversations, err = store.GetConversationsToExtractKnowledge()
// 	require.Nil(t, err)
// 	assert.Equal(t, 0, len(conversations))
// }

// func GetConversationsToExtractKnowledge(t *testing.T, store store.Store) {
// 	conversations, err := store.GetConversationsToExtractKnowledge()
// 	require.Nil(t, err)
// 	assert.Equal(t, 0, len(conversations))

// 	// Create a few new conversation
// 	msg1 := &chat.Message{
// 		ID:           uuid.New().String(),
// 		Conversation: uuid.New().String(),
// 		Agent:        "Rose",
// 		User:         "Keith",
// 		Content:      "Beep boop I'm a robot!",
// 		CreatedAt:    time.Now().Add(-5 * time.Minute),
// 	}
// 	err = store.SaveMessage(msg1)
// 	require.Nil(t, err)
// 	msg2 := &chat.Message{
// 		ID:           uuid.New().String(),
// 		Conversation: uuid.New().String(),
// 		Agent:        "Rose",
// 		User:         "Keith",
// 		Content:      "Beep boop I'm a robot!",
// 		CreatedAt:    time.Now().Add(-5 * time.Minute),
// 	}
// 	err = store.SaveMessage(msg2)
// 	require.Nil(t, err)

// 	conversations, err = store.GetConversationsToExtractKnowledge()
// 	require.Nil(t, err)
// 	require.Equal(t, 2, len(conversations))
// 	assert.Contains(t, conversations, msg1.Conversation)
// 	assert.Contains(t, conversations, msg2.Conversation)

// 	// Mark them both as extracted and show that they don't
// 	// get picked up anymore
// 	err = store.SetConversationAsKnowledgeExtracted(msg1.Conversation)
// 	require.Nil(t, err)
// 	err = store.SetConversationAsKnowledgeExtracted(msg2.Conversation)
// 	require.Nil(t, err)

// 	conversations, err = store.GetConversationsToExtractKnowledge()
// 	require.Nil(t, err)
// 	assert.Equal(t, 0, len(conversations))

// 	// Create an additional message in the future so that it is
// 	// picked up despite the extraction
// 	msg3 := &chat.Message{
// 		ID:           uuid.New().String(),
// 		Conversation: msg1.Conversation,
// 		Agent:        "Rose",
// 		User:         "Keith",
// 		Content:      "Beep boop I'm a robot!",
// 		CreatedAt:    time.Now().Add(5 * time.Minute),
// 	}
// 	err = store.SaveMessage(msg3)
// 	require.Nil(t, err)

// 	conversations, err = store.GetConversationsToExtractKnowledge()
// 	require.Nil(t, err)
// 	require.Equal(t, 1, len(conversations))
// 	assert.Contains(t, conversations, msg1.Conversation)
// }
