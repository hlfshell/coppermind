package mysql

import (
	"strings"
	"testing"
	"time"

	"github.com/hlfshell/docker-harness/databases/mysql"
	"github.com/stretchr/testify/require"
)

func createMysqlStoreContainer(t *testing.T) (*MySQLStore, *mysql.Mysql, error) {
	name := t.Name()
	// Check for a "/" in the name - if it's there, grab only the
	// sub test at the end
	namesplit := strings.Split(name, "/")
	if len(namesplit) > 1 {
		name = namesplit[len(namesplit)-1]
	}

	container, err := mysql.NewMysql(
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

	db, err := container.ConnectWithTimeout(20 * time.Second)
	if err != nil {
		return nil, nil, err
	}

	return &MySQLStore{
		db: db,
	}, container, nil
}

func TestMysql(t *testing.T) {
	store, container, err := createMysqlStoreContainer(t)
	require.Nil(t, err)
	defer container.Cleanup()

	err = store.Migrate()
	require.Nil(t, err)
}

// func TestLowLevelSqlite(t *testing.T) {
// 	tests := map[string]func(t *testing.T, store store.LowLevelStore){
// 		"SaveAndGetUser":                 storeTest.SaveAndCreatetUser,
// 		"GetUserAuth":                    storeTest.GetUserAuth,
// 		"GenerateUserPasswordResetToken": storeTest.GenerateUserPasswordResetToken,
// 		"ResetPassword":                  storeTest.ResetPassword,
// 		"DeleteUser":                     storeTest.DeleteUser,
// 		"SaveAndGetMessage":              storeTest.SaveAndGetMessage,
// 		"DeleteMessage":                  storeTest.DeleteMessage,
// 		"ListMessages":                   storeTest.ListMessages,
// 		"GetConversation":                storeTest.GetAndDeleteConversation,
// 		"ListConversation":               storeTest.ListConversations,
// 		"SaveAndGetAgent":                storeTest.SaveAndGetAgent,
// 		"DeleteAgent":                    storeTest.DeleteAgent,
// 		"ListAgents":                     storeTest.ListAgents,
// 		"SaveAndGetSummary":              storeTest.SaveAndGetSummary,
// 		"DeleteSummary":                  storeTest.DeleteSummary,
// 		"ListSummaries":                  storeTest.ListSummaries,
// 	}

// 	for name, _ := range tests {
// 		t.Run("TestLowLevelPostgres"+name, func(t *testing.T) {
// 			t.Parallel()
// 			store, container, err := createMysqlStoreContainer(t)
// 			require.Nil(t, err)
// 			defer container.Cleanup()

// 			err = store.Migrate()
// 			require.Nil(t, err)

// 			tests[name](t, store)
// 		})
// 	}
// }
