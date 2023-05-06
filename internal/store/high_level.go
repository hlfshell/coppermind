package store

import (
	"time"

	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
)

// HighLevelStore interface is the heart of our system's memory
// - it represents the totality of everything Coppermind needs
// to operate functionality. The interface below is broken down
// into categorical sections in order.
// High level store functions *may* utilize low level store
// functions *or* just choose to implement similar functionality
// in their own unique way to optimize for their given underlying
// datastore.
type HighLevelStore interface {
	//===============================
	// Management functions
	//===============================

	/*
		Migrate:
		This may not be necessary for your given database store,
		but any initiailization/version upgrades to the store
		need to be handled here. Migrate() will be called for
		the chosen store on initialization on startup of an
		agent
	*/
	Migrate() error

	//===============================
	// Messages
	//===============================

	// SaveMessage will upsert save a given message.
	SaveMessage(msg *chat.Message) error

	/*
		GetConversation will, given a conversation ID,
		return all messages in that conversation sorted in oldest
		to latest order of time
	*/
	GetConversation(conversation string) (*chat.Conversation, error)

	/*
		GetLatestConversation will, given the agent and user, return
		the latest conversation and timestamp of their last
		conversation. If none exist, the conversation ID will be
		"" and the time.Time will be a fresh uninit'ed one.
	*/
	GetLatestConversation(agent string, user string) (string, time.Time, error)

	//===============================
	// Summaries
	//===============================

	// SaveSummary will upsert a given summary into the store
	SaveSummary(summary *memory.Summary) error

	/*
		GetConversationsToSummarize will find any conversation past
		a certain size or age that does not yet have a summary,
		or return summaries that have summaries but have
		additional messages to include in its summary
		consideration.
		Summaries are to be generated for conversations with the
		following qualifications:

			1. A summary does not already exist for the conversation
				with an updated_at greater than the last message
			2. The conversation is not marked as "excluded" via the
				ExcludeFromSummary() function and however the store
				tracks this
			3. The agent specifies a minimum number of messages required
				for a summary to be created. This could be zero to
				essentially disable this.
			4. A minimum duration unless overrided by a minimum length.
				Going into further detail, a conversation will not have
				a summary generated for it until its last message is
				older than a given duration. The exception to this is
				the next specified qualifation.
			5. A max length allowed for a summary to be unsummarized
				before forcing a summary onto it. If messages trade
				fast enough, we will run out of room to pass message
				history and we can't wait for it to age for a summary;
				thus we need to create summaries regularly and update
				it as the conversation moves along. This max message
				length needs to be both on first coverage at a summary
				and subequent updates. Thus future summarization needs
				to consider if there has been the max length has
				occurred since the last update.
	*/
	GetConversationsToSummarize(minMessages int, minAge time.Duration, maxLength int) ([]string, error)

	/*
		GetSummaryByConversation will return the summary associated
		with a specific summmary. If none exists, summary pointer
		will be nil
	*/
	GetSummaryByConversation(conversation string) (*memory.Summary, error)

	/*
		GetsummariesByAgentAndUser will return all summaries associated
		with a given user in oldest to latest order of time
	*/
	GetSummariesByAgentAndUser(agent string, user string) ([]*memory.Summary, error)

	//===============================
	// Knowledge
	//===============================

	/*
		SaveKnowledge takes a given bit of knowledge and saves
		it.
	*/
	SaveKnowledge(knowledge *memory.Knowledge) error

	/*
		GetConversationsToExtractKnowledge grabs any updates
		to any conversation it can. It has less stringent rules
		than the summarization model since we have no need to
		hold off on waiting to re-extract since we ask the
		LLM to avoid duplication of knowledge.
	*/
	GetConversationsToExtractKnowledge() ([]string, error)

	/*
		SetconversationAsKnowledgeExtracted marks a given conversation
		as having its knowledge extracted. This should prevent the
		conversation from being scanned again unless new messages are
		added
	*/
	SetConversationAsKnowledgeExtracted(conversation string) error

	/*
		GetKnowledgeByAgentAndUser will return all knowledge generated
		from conversation between the user and agent. Expired knowledge
		should not be included.
	*/
	GetKnowlegeByAgentAndUser(agent string, user string) ([]*memory.Knowledge, error)

	/*
		GetKnowledgeGroupedByAgentAndUser will return all knowledge across
		all agents, but group by agent/user combination. Again, expired
		knowledge is not returned.
	*/
	GetKnowledgeGroupedByAgentAndUser(agent string, user string) (map[string]map[string][]*memory.Knowledge, error)

	/*
		ExpireKnowledge erases all knowledge that should have been expired
	*/
	ExpireKnowledge() error
}
