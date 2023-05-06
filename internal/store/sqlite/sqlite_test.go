package sqlite

import (
	"testing"

	"github.com/hlfshell/coppermind/internal/store"
	storeTest "github.com/hlfshell/coppermind/internal/test/store"
	"github.com/stretchr/testify/require"
)

func createSqlLiteStore() (*SqliteStore, error) {
	store, err := NewSqliteStore(":memory:")
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

}

func TestSqlite(t *testing.T) {
	tests := map[string]func(*testing.T, store.Store){
		"GetLatestConversation":               storeTest.GetLatestConversation,
		"GetConversations":                    storeTest.GetConversation,
		"SaveMessage":                         storeTest.SaveMessage,
		"GetConversationsToSummarize":         storeTest.GetConversationsToSummarize,
		"GetSummaryByConversation":            storeTest.GetSummaryByConversation,
		"SaveSummary":                         storeTest.SaveSummary,
		"GetSummariesByAgentAndUser":          storeTest.GetSummariesByAgentAndUser,
		"ExcludeConversationFromSummary":      storeTest.ExcludeConversationFromSummary,
		"ExpireKnowledge":                     storeTest.ExpireKnowledge,
		"SetConversationAsKnowledgeExtracted": storeTest.SetConversationAsKnowledgeExtracted,
	}

	for name, test := range tests {
		t.Run("TestSqlite"+name, func(t *testing.T) {
			sqlite, err := createSqlLiteStore()

			require.Nil(t, err)
			test(t, sqlite)
		})
	}
}
