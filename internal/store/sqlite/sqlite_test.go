package sqlite

import (
	"testing"

	storeTest "github.com/hlfshell/coppermind/internal/test/store"
	"github.com/stretchr/testify/assert"
)

func CreateSqlLiteStore() (*SqliteStore, error) {
	store, err := NewSqliteStore(":memory:")
	// store, err := NewSqliteStore("./testing.db")
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

// func TestGetConversation(t *testing.T) {
// 	sqlite, err := CreateSqlLiteStore()
// 	assert.Nil(t, err)

//		t.Run("TestSqliteGetConversation", func(t *testing.T) {
//			storeTest.GetConversation(t, sqlite)
//		})
//	}
func TestGetConversations(t *testing.T) {
	sqlite, err := CreateSqlLiteStore()
	assert.Nil(t, err)

	t.Run("TestSqliteGetConversation", func(t *testing.T) {
		storeTest.GetConversation(t, sqlite)
	})
}

// func TestSaveConversation(t *testing.T) {
// 	sqlite, err := CreateSqlLiteStore()
// 	assert.Nil(t, err)

// 	t.Run("TestSqliteSaveConversation", func(t *testing.T) {
// 		storeTest.SaveConversation(t, sqlite)
// 	})
// }

func TestGetLatestConversation(t *testing.T) {
	sqlite, err := CreateSqlLiteStore()
	assert.Nil(t, err)

	t.Run("TestSqliteGetLatestConversation", func(t *testing.T) {
		storeTest.GetConversation(t, sqlite)
	})
}

func TestSaveMessage(t *testing.T) {
	sqlite, err := CreateSqlLiteStore()
	assert.Nil(t, err)

	t.Run("TestSqliteSaveMessage", func(t *testing.T) {
		storeTest.SaveMessage(t, sqlite)
	})
}

func TestGetConversationsToSummarize(t *testing.T) {
	sqlite, err := CreateSqlLiteStore()
	assert.Nil(t, err)

	t.Run("TestSqliteGetConversationsToSummarize", func(t *testing.T) {
		storeTest.GetConversationsToSummarize(t, sqlite)
	})
}

func TestGetSummaryByConversation(t *testing.T) {
	sqlite, err := CreateSqlLiteStore()
	assert.Nil(t, err)

	t.Run("TestSqliteGetSummaryByConversation", func(t *testing.T) {
		storeTest.GetSummaryByConversation(t, sqlite)
	})
}

func TestSaveSummary(t *testing.T) {
	sqlite, err := CreateSqlLiteStore()
	assert.Nil(t, err)

	t.Run("TestSqliteSaveSummary", func(t *testing.T) {
		storeTest.SaveSummary(t, sqlite)
	})
}
