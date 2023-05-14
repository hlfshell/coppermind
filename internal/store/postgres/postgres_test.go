package postgres

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/hlfshell/coppermind/internal/store"
	storeTest "github.com/hlfshell/coppermind/internal/test/store"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

type postgresContainer struct {
	client       *docker.Client
	id           string
	image        string
	name         string
	port         string
	username     string
	password     string
	databaseName string
}

func createPostgresContainer(t *testing.T) (*postgresContainer, error) {
	// Derive the container name. The test might have an illegal
	// "/" in it if its a subtest; if so, grab the last piece of
	// a split
	name := t.Name()
	nameSplit := strings.Split(name, "/")
	if len(nameSplit) > 1 {
		name = nameSplit[len(nameSplit)-1]
	}

	pgContainer := &postgresContainer{
		client: nil,
		id:     "",
		image:  "postgres:latest",
		name:   name,
		port:   "",
	}

	client, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}

	pgContainer.client = client
	pgContainer.username = "postgres"
	pgContainer.password = "postgres"
	pgContainer.databaseName = "coppermind"

	out, err := client.ImagePull(context.Background(), pgContainer.image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}
	defer out.Close()

	io.Copy(os.Stdout, out)

	containerConfig := &container.Config{
		Image: "postgres:latest",
		Env: []string{
			fmt.Sprintf("POSTGRES_USER=%s", pgContainer.username),
			fmt.Sprintf("POSTGRES_PASSWORD=%s", pgContainer.password),
			fmt.Sprintf("POSTGRES_DB=%s", pgContainer.databaseName),
		},
		ExposedPorts: nat.PortSet{
			"5432/tcp": struct{}{},
		},
	}

	// Find a free port
	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}
	pgContainer.port = fmt.Sprint(port)

	hostConfig := &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			"5432/tcp": {
				{
					HostIP:   "0.0.0.0",
					HostPort: pgContainer.port,
				},
			},
		},
	}

	resp, err := client.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		pgContainer.name,
	)
	if err != nil {
		return nil, err
	}

	pgContainer.id = resp.ID

	client.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{})

	return pgContainer, nil
}

func (pg *postgresContainer) close() error {
	// Find our current container to find its mounts/volumes
	list, err := pg.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return err
	}

	// Find our container and identify attached volumes for cleanup
	var targetContainer *types.Container
	for _, c := range list {
		if c.ID == pg.id {
			targetContainer = &c
			break
		}
	}
	if targetContainer == nil {
		return fmt.Errorf("could not find container to destroy its resources")
	}

	// Stop the container - since it's ephermeral we can force a quick
	// kill as we don't care about data loss
	timeout := 0
	err = pg.client.ContainerStop(context.Background(), pg.id, container.StopOptions{
		Timeout: &timeout,
		Signal:  "SIGKILL",
	})
	if err != nil {
		return err
	}

	// Finally, remove the container
	err = pg.client.ContainerRemove(context.Background(), pg.id, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	// Remove the container's volumes
	for _, mount := range targetContainer.Mounts {
		err = pg.client.VolumeRemove(context.Background(), mount.Name, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func createPostgresStore(t *testing.T) (*PostgresStore, *postgresContainer, error) {
	container, err := createPostgresContainer(t)
	if err != nil {
		if container != nil {
			container.close()
			return nil, nil, err
		}
		return nil, nil, err
	}

	// Try this for up to five seconds, allowing the database
	// time to start in case that's the issue
	var store *PostgresStore
	for i := 0; i < 20; i++ {
		store, err = NewSqliteStore(
			container.username,
			container.password,
			"0.0.0.0",
			container.port,
			container.databaseName,
		)
		if err != nil {
			time.Sleep(250 * time.Millisecond)
		} else {
			break
		}
	}
	// If we still can't connect, abort and close the
	// container as sometihng is wrong
	if err != nil {
		container.close()
		return nil, nil, err
	}

	return store, container, nil
}

// func TestPostgres(t *testing.T) {
// 	store, container, err := createPostgresStore(t)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer container.close()

// 	err = store.Migrate()
// 	require.Nil(t, err)
// }

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
			store, container, err := createPostgresStore(t)
			require.Nil(t, err)
			defer container.close()

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
			store, container, err := createPostgresStore(t)
			require.Nil(t, err)
			defer container.close()

			err = store.Migrate()
			require.Nil(t, err)

			tests[name](t, store)
		})
	}
}
