package sqlite

import (
	"github.com/hlfshell/coppermind/pkg/agents"
	"github.com/wissance/stringFormatter"
)

const agentSelectColumns = `id, name, identity`

func (store *SqliteStore) SaveAgent(agent *agents.Agent) error {
	query := `INSERT INTO {0} ({1}) VALUES(?, ?, ?)`

	query = stringFormatter.Format(query, AGENTS_TABLE, agentSelectColumns)

	_, err := store.db.Exec(
		query,
		agent.ID,
		agent.Name,
		agent.Identity,
	)

	return err
}

func (store *SqliteStore) GetAgent(id string) (*agents.Agent, error) {
	query := `SELECT {0} FROM {1} WHERE id = ?`

	query = stringFormatter.Format(query, agentSelectColumns, AGENTS_TABLE)

	rows, err := store.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var agent agents.Agent
		err = rows.Scan(
			&agent.ID,
			&agent.Name,
			&agent.Identity,
		)
		return &agent, err
	}

	return nil, nil
}

func (store *SqliteStore) DeleteAgent(id string) error {
	query := `DELETE FROM {0} WHERE id = ?`

	query = stringFormatter.Format(query, AGENTS_TABLE)

	_, err := store.db.Exec(query, id)
	return err
}

func (store *SqliteStore) ListAgents() ([]*agents.Agent, error) {
	query := `SELECT {0} FROM {1}`

	query = stringFormatter.Format(query, agentSelectColumns, AGENTS_TABLE)

	rows, err := store.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foundAgents []*agents.Agent

	for rows.Next() {
		var agent agents.Agent
		err = rows.Scan(
			&agent.ID,
			&agent.Name,
			&agent.Identity,
		)
		if err != nil {
			return nil, err
		}
		foundAgents = append(foundAgents, &agent)
	}

	return foundAgents, nil
}
