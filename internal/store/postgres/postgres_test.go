package postgres

import (
	"strings"
	"testing"
	"time"

	"github.com/hlfshell/coppermind/internal/store"
	storeTest "github.com/hlfshell/coppermind/internal/test/store"
	"github.com/stretchr/testify/require"

	"github.com/hlfshell/docker-harness/databases/postgres"
)

func createPostgresStoreContainer(t *testing.T) (*PostgresStore, *postgres.Postgres, error) {
	name := t.Name()
	// Check for a "/" in the name - if it's there, grab only the
	// sub test at the end
	nameSplit := strings.Split(name, "/")
	if len(nameSplit) > 1 {
		name = nameSplit[len(nameSplit)-1]
	}

	container, err := postgres.NewPostgres(
		name,
		"",
		"username",
		"password",
		"coppermind",
	)
	if err != nil {
		return nil, nil, err
	}

	err = container.Create()
	if err != nil {
		return nil, nil, err
	}

	db, err := container.ConnectWithTimeout(10 * time.Second)
	if err != nil {
		return nil, nil, err
	}

	return &PostgresStore{
		db: db,
	}, container, nil
}

func TestPostgres(t *testing.T) {
	store, container, err := createPostgresStoreContainer(t)
	require.Nil(t, err)
	defer container.Cleanup()

	err = store.Migrate()
	require.Nil(t, err)
}

func TestLowLevelSqlite(t *testing.T) {
	tests := map[string]func(t *testing.T, store store.LowLevelStore){
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

	for name, _ := range tests {
		t.Run("TestLowLevelPostgres"+name, func(t *testing.T) {
			t.Parallel()
			store, container, err := createPostgresStoreContainer(t)
			require.Nil(t, err)
			defer container.Cleanup()

			err = store.Migrate()
			require.Nil(t, err)

			tests[name](t, store)
		})
	}
}

func TestHighLevelPostgres(t *testing.T) {
	tests := map[string]func(t *testing.T, store store.Store){
		"GetLatestConversation":          storeTest.GetLatestConversation,
		"GetConversationsToSummarize":    storeTest.GetConversationsToSummarize,
		"ExcludeConversationFromSummary": storeTest.ExcludeConversationFromSummary,
		// "ExpireKnowledge":                     storeTest.ExpireKnowledge,
		// "SetConversationAsKnowledgeExtracted": storeTest.SetConversationAsKnowledgeExtracted,
	}

	for name, _ := range tests {
		t.Run("TestSqlite"+name, func(t *testing.T) {
			t.Parallel()
			store, container, err := createPostgresStoreContainer(t)
			require.Nil(t, err)
			defer container.Cleanup()

			err = store.Migrate()
			require.Nil(t, err)

			tests[name](t, store)
		})
	}
}
