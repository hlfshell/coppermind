package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/wissance/stringFormatter"
)

const knowledgeColumns = `id, agent, user, source, content, metadata, vector, created_at, last_utilized`

func (store *SqliteStore) SaveKnowledge(fact *memory.Knowledge) error {
	query := `INSERT INTO {0} ({1}) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = stringFormatter.Format(query, KNOWLEDGE_TABLE, knowledgeColumns)

	// Convert the metadata map[string]string to bytes
	metadataBytes, err := json.Marshal(fact.Metadata)
	if err != nil {
		return err
	}
	vectorBytes, err := json.Marshal(fact.Vector)
	if err != nil {
		return err
	}

	_, err = store.db.Exec(
		query,
		fact.ID,
		fact.Agent,
		fact.User,
		fact.Source,
		fact.Content,
		metadataBytes,
		vectorBytes,
		fact.CreatedAt,
		fact.LastUtilized,
	)

	return err
}

func (store *SqliteStore) GetKnowledge(id string) (*memory.Knowledge, error) {
	query := `SELECT {0} FROM {1} WHERE id = ?`

	query = stringFormatter.Format(query, knowledgeColumns, KNOWLEDGE_TABLE)

	rows, err := store.db.Query(query, id)
	if err != nil {
		return nil, err
	}

	knowledge, err := store.sqlToKnowledge(rows)
	if err != nil {
		return nil, err
	} else if len(knowledge) == 0 {
		return nil, nil
	}

	return knowledge[0], nil
}

func (store *SqliteStore) DeleteKnowledge(id string) error {
	query := `DELETE FROM {0} WHERE id = ?`

	query = stringFormatter.Format(query, KNOWLEDGE_TABLE)

	_, err := store.db.Exec(query, id)
	return err
}

func (store *SqliteStore) ListKnowledge(filter *store.KnowledgeFilter) ([]*memory.Knowledge, error) {
	query := `SELECT {0} FROM {1} `

	params := []interface{}{}
	if filter != nil && !filter.Empty() {
		query += `WHERE `
		clauseCount := 0
		if filter.ID != nil {
			clause, values := filterToWhereClause("id", filter.ID.Operation, filter.ID.Value)
			query += clause
			clauseCount++
			params = append(params, values...)
		}
		if filter.Agent != nil {
			if clauseCount > 0 {
				query += `AND `
			}
			clause, values := filterToWhereClause("agent", filter.Agent.Operation, filter.Agent.Value)
			query += clause
			clauseCount++
			params = append(params, values...)
		}
		if filter.User != nil {
			if clauseCount > 0 {
				query += `AND `
			}
			clause, values := filterToWhereClause("user", filter.User.Operation, filter.User.Value)
			query += clause
			clauseCount++
			params = append(params, values...)
		}
		if filter.Source != nil {
			if clauseCount > 0 {
				query += `AND `
			}
			clause, values := filterToWhereClause("source", filter.Source.Operation, filter.Source.Value)
			query += clause
			clauseCount++
			params = append(params, values...)
		}
		if filter.CreatedAt != nil {
			if clauseCount > 0 {
				query += `AND `
			}
			clause, values := filterToWhereClause("created_at", filter.CreatedAt.Operation, filter.CreatedAt.Value)
			query += clause
			clauseCount++
			params = append(params, values...)
		}
		if filter.LastUtilized != nil {
			if clauseCount > 0 {
				query += `AND `
			}
			clause, values := filterToWhereClause("last_utilized", filter.LastUtilized.Operation, filter.LastUtilized.Value)
			query += clause
			params = append(params, values...)
		}
	}

	query += `ORDER BY created_at `
	if filter != nil && filter.OldestFirst {
		query += `DESC `
	} else {
		query += `ASC `
	}

	if filter != nil && filter.Limit > 0 {
		query += `LIMIT ? `
		params = append(params, filter.Limit)
	}

	query = stringFormatter.Format(query, knowledgeColumns, KNOWLEDGE_TABLE)
	fmt.Println(">>", query)
	fmt.Println(params)
	rows, err := store.db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	return store.sqlToKnowledge(rows)
}

func (store *SqliteStore) GetConversationsToExtractKnowledge() ([]string, error) {
	query := `
		SELECT
			messages.conversation
		FROM
			{0} AS messages LEFT JOIN
			{1} AS extraction ON messages.conversation = extraction.conversation
		WHERE
			extraction.conversation IS NULL OR
			extraction.updated_at < (
				SELECT MAX(created_at) FROM {0} WHERE conversation = messages.conversation
			)
		GROUP BY
			messages.conversation
	`

	query = stringFormatter.Format(query, MESSAGES_TABLE, KNOWLEDGE_EXTRACTION_TABLE)

	rows, err := store.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := []string{}
	for rows.Next() {
		var conversation string
		err = rows.Scan(&conversation)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conversation)
	}

	return conversations, nil
}

func (store *SqliteStore) SetConversationAsKnowledgeExtracted(conversation string) error {
	query := `
		INSERT INTO {0}
		(
			conversation,
			updated_at
		)
		VALUES(?, ?)
	`

	query = stringFormatter.Format(query, KNOWLEDGE_EXTRACTION_TABLE)

	_, err := store.db.Exec(query, conversation, time.Now())
	return err
}

func (store *SqliteStore) CompressKnowledge(agent string, user string, knowledge []*memory.Knowledge) error {
	// Start a sqlite transaction
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}

	// Delete all the knowledge that belongs to the agent and user, to be replaced by our incoming knowledge
	query := `DELETE FROM {0} WHERE agent = ? AND user = ?`
	query = stringFormatter.Format(query, KNOWLEDGE_TABLE)
	_, err = tx.Exec(query, agent, user)
	if err != nil {
		return err
	}

	// Insert the new knowledge with our transaction all in one query
	query = `
		INSERT INTO {0}
		(
			id,
			agent,
			user,
			subject,
			predicate,
			object,
			created_at,
			expires_at
		)
		VALUES 
		`
	query = stringFormatter.Format(query, KNOWLEDGE_TABLE)
	for i := 0; i < len(knowledge); i++ {
		query += "(?, ?, ?, ?, ?, ?, ?, ?)"
		if i < len(knowledge)-1 {
			query += ", "
		}
	}

	_, err = store.db.Exec(query, knowledge)
	return err
}

func (store *SqliteStore) sqlToKnowledge(rows *sql.Rows) ([]*memory.Knowledge, error) {
	defer rows.Close()

	knowledge := []*memory.Knowledge{}

	for rows.Next() {
		var fact memory.Knowledge
		var metadataBytes []byte
		var vectorBytes []byte
		var createdDatetime string
		var lastUtilizedDatetime string
		err := rows.Scan(
			&fact.ID,
			&fact.Agent,
			&fact.User,
			&fact.Source,
			&fact.Content,
			&metadataBytes,
			&vectorBytes,
			&createdDatetime,
			&lastUtilizedDatetime,
		)
		if err != nil {
			return nil, err
		}

		// Convert timestamps
		timestamp, err := store.sqlTimestampToTime(createdDatetime)
		if err != nil {
			return nil, err
		}
		fact.CreatedAt = timestamp
		timestamp, err = store.sqlTimestampToTime(lastUtilizedDatetime)
		if err != nil {
			return nil, err
		}
		fact.LastUtilized = timestamp

		// Convert bytes
		err = json.Unmarshal(metadataBytes, &fact.Metadata)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(vectorBytes, &fact.Vector)
		if err != nil {
			return nil, err
		}

		knowledge = append(knowledge, &fact)
	}

	return knowledge, nil
}

func filterToWhereClause(attribute string, operation string, value interface{}) (string, []interface{}) {
	placeholder := ""
	values := []interface{}{}
	if operation == "IN" {
		splitValue := strings.Split(value.(string), ",")
		placeholder += "("
		for i := 0; i < len(splitValue); i++ {
			placeholder += "?"
			if i < len(splitValue)-1 {
				placeholder += ", "
			}
			values = append(values, strings.TrimSpace(splitValue[i]))
		}
		placeholder += ")"
	} else {
		placeholder = "?"
		values = append(values, value)
	}

	return fmt.Sprintf("%s %s %s ", attribute, operation, placeholder), values
}
