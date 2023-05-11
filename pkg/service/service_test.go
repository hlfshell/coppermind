package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/llm"
	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/internal/store/sqlite"
	"github.com/hlfshell/coppermind/pkg/agents"
	"github.com/hlfshell/coppermind/pkg/config"
	"github.com/hlfshell/coppermind/pkg/users"
)

var testAgent agents.Agent = agents.Agent{
	ID:       uuid.New().String(),
	Name:     "Rose",
	Identity: `Sassy and cyncial at every chance, Rose still aims to help`,
}

var testUser users.User = users.User{
	ID:        uuid.New().String(),
	Name:      "Keith",
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

func createMockService(llm llm.LLM) (*Service, store.Store, error) {
	// Create a new sqlite store in-memory for tests
	store, err := sqlite.NewSqliteStore(":memory:")
	if err != nil {
		return nil, nil, err
	}
	if err = store.Migrate(); err != nil {
		return nil, nil, err
	}

	// Prepare a user and agent to work with
	err = store.SaveAgent(&testAgent)
	if err != nil {
		return nil, nil, err
	}
	err = store.CreateUser(&testUser, "super duper secret shhh")
	if err != nil {
		return nil, nil, err
	}

	return NewService(store, llm, &config.DefaultConfig), store, nil
}
