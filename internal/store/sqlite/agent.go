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

	var agent agents.Agent
	for rows.Next() {
		err = rows.Scan(
			&agent.ID,
			&agent.Name,
			&agent.Identity,
		)
		if err != nil {
			return nil, err
		}
	}

	return &agent, nil
}
