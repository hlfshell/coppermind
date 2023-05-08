package sqlite

import (
	"testing"

	"github.com/hlfshell/coppermind/internal/store"
	storeTest "github.com/hlfshell/coppermind/internal/test/store"
	"github.com/stretchr/testify/require"
)

func createSqlLiteStore() (*SqliteStore, error) {
	store, err := NewSqliteStore(":memory:")
	// store, err := NewSqliteStore("/home/keith/projects/coppermind/test.db")
	if err != nil {
		return nil, err
	}
	if err = store.Migrate(); err != nil {
		return nil, err
	}
	return store, nil
}

func TestMigrate(t *testing.T) {

}

func TestLowLevelSqlite(t *testing.T) {
	tests := map[string]func(*testing.T, store.LowLevelStore){
		"SaveAndGetUser":                 storeTest.SaveAndCreatetUser,
		"GetUserAuth":                    storeTest.GetUserAuth,
		"GenerateUserPasswordResetToken": storeTest.GenerateUserPasswordResetToken,
		"ResetPassword":                  storeTest.ResetPassword,
		"DeleteUser":                     storeTest.DeleteUser,
		"SaveAndGetMessage":              storeTest.SaveAndGetMessage,
		"DeleteMessage":                  storeTest.DeleteMessage,
		"ListMessages":                   storeTest.ListMessages,
		"GetConversation":                storeTest.GetAndDeleteConversation,
		"ListConversation":               storeTest.ListConversations,
		"SaveAndGetAgent":                storeTest.SaveAndGetAgent,
		"DeleteAgent":                    storeTest.DeleteAgent,
		"ListAgents":                     storeTest.ListAgents,
		"SaveAndGetSummary":              storeTest.SaveAndGetSummary,
		"DeleteSummary":                  storeTest.DeleteSummary,
		"ListSummaries":                  storeTest.ListSummaries,
	}

	for name, test := range tests {
		t.Run("TestLowLevelSqlite"+name, func(t *testing.T) {
			sqlite, err := createSqlLiteStore()

			require.Nil(t, err)
			test(t, sqlite)
		})
	}
}

func TestSqlite(t *testing.T) {
	tests := map[string]func(*testing.T, store.Store){
		"GetLatestConversation":          storeTest.GetLatestConversation,
		"GetConversationsToSummarize":    storeTest.GetConversationsToSummarize,
		"ExcludeConversationFromSummary": storeTest.ExcludeConversationFromSummary,
		// "ExpireKnowledge":                     storeTest.ExpireKnowledge,
		// "SetConversationAsKnowledgeExtracted": storeTest.SetConversationAsKnowledgeExtracted,
	}

	for name, test := range tests {
		t.Run("TestSqlite"+name, func(t *testing.T) {
			sqlite, err := createSqlLiteStore()

			require.Nil(t, err)
			test(t, sqlite)
		})
	}
}
