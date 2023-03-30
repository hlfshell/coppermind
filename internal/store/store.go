package store

import (
	"time"

	"github.com/hlfshell/coppermind/internal/chat"
)

// Store interface is the heart of our system's memory - it
// represents the totality of everything Coppermind needs to
// operate functionality. The interface below is broken down
// into categorical sections in order
type Store interface {
	//===============================
	//Management functions
	//===============================
	//This may not be necessary for your given database store,
	//but any initiailization/version upgrades to the store
	//need to be handled here. Migrate() will be called for
	//the chosen store on initialization on startup of an
	//agent
	Migrate() error

	//===============================
	//Conversation Functions
	//===============================
	//GetConversation will return the metadata for a given
	//conversation
	// GetConversation(conversation string) (*chat.Conversation, error)
	//SaveConversation will create a new conversation
	// SaveConversation(conversation *chat.Conversation) error
	// //GetLatestConversation will, given an agent and user, return
	// //the latest configuration and the time of its last message
	// GetLatestConversation(agent string, user string) (string, time.Time, error)

	//===============================
	//Messages
	//===============================
	//SaveMessage will upsert save a given message.
	SaveMessage(msg *chat.Message) error
	//LoadConversationMessages will, given a conversation ID,
	//return all messages in that conversation sorted in oldest
	//to latest order of time
	// LoadConversationMessages(conversation string) ([]*chat.Message, error)
	GetConversation(conversation string) (*chat.Conversation, error)
	GetLatestConversation(agent string, user string) (string, time.Time, error)

	//===============================
	//Summaries
	//===============================
	//GetConversationsToUpdate will find any conversation past
	//a certain size or age that does not yet have a summary,
	//or return summaries that have summaries but have additional
	//messages to include in its summary consideration
	// GetConversationsToUpdate() ([]string, error)
	//GetSummbaryByconversation will return the summary associated
	//with a specific summmary. If none exists, summary pointer
	//will be nil
	// GetSummaryByConversation(conversation string) (*memory.Summary, error)
	//GetsummariesByUser will return all summaries associated
	//with a given user in oldest to latest order of time
	// GetSummariesByUser(user string) ([]*memory.Summary, error)
	//SaveSummary will upsert a given summary into the store
	// SaveSummary(summary *memory.Summary) error
}
