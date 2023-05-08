package store

import (
	"github.com/hlfshell/coppermind/pkg/agents"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/hlfshell/coppermind/pkg/users"

	users_internal "github.com/hlfshell/coppermind/internal/users"
)

// LowLevelStore is the simple CRUDL (create, read, update
// delete, list) interface for coppermind. This interface
// is for generic object access as opposed to the
// HighLevelStore's more functionality specific functions.
// Each object in the scope of coppermind should have associated
// CRDL functions w/ expected behaviour.
type LowLevelStore interface {
	//===============================
	// Users
	//===============================

	/*
		SaveUser will create a given user object. Note that
		the password is passed unlike other attributes in
		that it's never returned after writing save on
		specific request.

		Unlike other objects, SaveUser is a one time write,
		and will error if it doesn't exists.
	*/
	CreateUser(user *users.User, password string) error

	/*
		GetUser will return a user given its ID
	*/
	GetUser(id string) (*users.User, error)

	/*
		GetUserAuth will return a UserAuth given its ID.
		This is different than the user object as it's
		just the authentication related information for
		the user.
	*/
	GetUserAuth(id string) (*users_internal.UserAuth, error)

	/*
		SaveUserAuth will save a user's authentication
		information. It may be its own table/object based
		on the store, or a part of the user. Generally
		this is called separately from CreateUser, and
		assumes that the user is already created.
	*/
	SaveUserAuth(auth *users_internal.UserAuth) error

	/*
		GenerateUserPasswordResetToken will generate a new
		token for resetting a password for the given user,
		as well as reset the attempts and reset time.
	*/
	GenerateUserPasswordResetToken(id string) (string, error)

	/*
		ResetPassword updates the user's password only.
		Any rules around password changes are handled elsewhere.
		Note that the last ResetTokenAttempts should be
		incremented on each use of the token. If the token is
		successfully utilized, then the token and usage tracking
		should be cleared.
	*/
	ResetPassword(id string, token string, password string) error

	/*
		DeleteUser will delete a user given its ID.
		Note that this does *not* remove any of the
		user's histories or other data.
	*/
	DeleteUser(id string) error

	//===============================
	// Agents
	//===============================

	/*
		SaveAgent will upsert save a given agent.
	*/
	SaveAgent(agent *agents.Agent) error

	/*
		GetAgent will return an agent given its ID
	*/
	GetAgent(id string) (*agents.Agent, error)

	/*
		DeleteAgent will delete an agent given its ID.
		Note that this does *not* remove any of the
		agent's histories or other data.
	*/
	DeleteAgent(id string) error

	/*
		ListAgents will return all agents in the store
	*/
	ListAgents() ([]*agents.Agent, error)

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
