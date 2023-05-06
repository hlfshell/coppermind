package store

import (
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
)

// LowLevelStore is the simple CRUDL (create, read, update
// delete, list) interface for coppermind. This interface
// is for generic object access as opposed to the
// HighLevelStore's more functionality specific functions.
// Each object in the scope of coppermind should have associated
// CRDL functions w/ expected behaviour.
type LowLevelStore interface {
	//===============================
	// Messages
	//===============================

	/*
		SaveMessage will upsert save a given message.
	*/
	SaveMessage(msg *chat.Message) error

	/*
		GetMessage will return a message given its ID
	*/
	GetMessage(id string) (*chat.Message, error)

	/*
		DeleteMessage will delete a message given its ID
	*/
	DeleteMessage(id string) error

	/*
		ListMessages will return all messages in the store
	*/
	ListMessages(query Filter) ([]*chat.Message, error)

	//===============================
	// Conversations
	//===============================
	/*
		Conversatons are not necessarily their own object; that
		depends upon the implementor of the store. It may be
		ideal to merely treat conversations as a UUID of the
		message and act accordingly. Still, we create these
		functions as some stores may find benefit in treating
		them as separate organizational entities, and the
		helper functions are useful on their own.

		Note that we don't have a Create for conversations as
		a null conversation is currently not useful to us.
	*/

	/*
		GetConversation will, given a conversation ID, return
		all messages in that conversation sorted in oldest to
		latest creation time
	*/
	GetConversation(conversation string) (*chat.Conversation, error)

	/*
		DeleteConversation will delete a conversation given its
		ID
	*/
	DeleteConversation(id string) error

	/*
		ListConversations will return all conversations that
		match a given filter's criteria
	*/
	ListConversations(query Filter) ([]*chat.Conversation, error)

	//===============================
	// Summaries
	//===============================

	/*
		SaveSummary will upsert a given summary into the store
	*/
	SaveSummary(summary *memory.Summary) error

	/*
		GetSummary will return a summary given its ID
	*/
	GetSummary(id string) (*memory.Summary, error)

	/*
		DeleteSummary will delete a summary given its ID
	*/
	DeleteSummary(id string) error

	/*
		ListSummaries will return all summaries in the store
	*/
	ListSummaries(query Filter) ([]*memory.Summary, error)

	//===============================
	// SummaryExclusions
	//===============================

	/*
		ExcludeConversationFromSummary marks a given conversation as
		one to ignore if a conversation
	*/
	ExcludeConversationFromSummary(conversation string) error

	/*
		DeleteSummaryExclusion removes exclusion from summarization
	*/
	DeleteSummaryExclusion(conversation string) error

	//===============================
	// Knowledge
	//===============================
}
