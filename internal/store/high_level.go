package store

import (
	"time"
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

	//===============================
	// Knowledge
	//===============================

	// /*
	// 	GetConversationsToExtractKnowledge grabs any updates
	// 	to any conversation it can. It has less stringent rules
	// 	than the summarization model since we have no need to
	// 	hold off on waiting to re-extract since we ask the
	// 	LLM to avoid duplication of knowledge.
	// */
	// GetConversationsToExtractKnowledge() ([]string, error)

	// /*
	// 	SetConversationAsKnowledgeExtracted marks a given conversation
	// 	as having its knowledge extracted. This should prevent the
	// 	conversation from being scanned again unless new messages are
	// 	added
	// */
	// SetConversationAsKnowledgeExtracted(conversation string) error

	// /*
	// 	MarkKnowledgeAsUtilized will mark the given knowledge IDs as
	// 	pulled and utilized for a chat prompt.
	// */
	// MarkKnowledgeAsUtilized(ids []string) error
}
