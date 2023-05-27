package mysql

import (
	"database/sql"
	"time"

	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/wissance/stringFormatter"
)

func (store *MySQLStore) SaveKnowledge(fact *memory.Knowledge) error {
	query := `
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
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)
	`

	query = stringFormatter.Format(query, KNOWLEDGE_TABLE)

	_, err := store.db.Exec(
		query,
		fact.ID,
		fact.Agent,
		fact.User,
		fact.Subject,
		fact.Predicate,
		fact.Object,
		fact.CreatedAt,
		fact.ExpiresAt,
	)

	return err
}

func (store *MySQLStore) GetKnowledge(id string) (*memory.Knowledge, error) {
	query := `
		SELECT
			id,
			agent,
			user,
			subject,
			predicate,
			object,
			created_at,
			expires_at
		FROM
			{0}
		WHERE
			id = ?
	`

	query = stringFormatter.Format(query, KNOWLEDGE_TABLE)

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

func (store *MySQLStore) GetKnowlegeByAgentAndUser(agent string, user string) ([]*memory.Knowledge, error) {
	query := `
		SELECT 
			id,
			agent,
			user,
			subject,
			predicate,
			object,
			created_at,
			expires_at
		FROM
			{0}
		WHERE
			user = ? AND agent = ?
	`

	query = stringFormatter.Format(query, KNOWLEDGE_TABLE)

	rows, err := store.db.Query(query, user, agent)
	if err != nil {
		return nil, err
	}

	return store.sqlToKnowledge(rows)
}

func (store *MySQLStore) ExpireKnowledge() error {
	query := `
		DELETE FROM {0}
		WHERE
			expires_at < ?
	`

	query = stringFormatter.Format(query, KNOWLEDGE_TABLE)

	_, err := store.db.Exec(query, time.Now())
	return err
}

func (store *MySQLStore) GetKnowledgeGroupedByAgentAndUser(agent string, user string) (map[string]map[string][]*memory.Knowledge, error) {
	query := `
		SELECT
			id,
			agent,
			user,
			subject,
			predicate,
			object,
			created_at,
			expires_at
		FROM
			{0}
	`

	query = stringFormatter.Format(query, KNOWLEDGE_TABLE)

	rows, err := store.db.Query(query)
	if err != nil {
		return nil, err
	}

	facts, err := store.sqlToKnowledge(rows)
	if err != nil {
		return nil, err
	}

	knowledgebase := map[string]map[string][]*memory.Knowledge{}
	for _, fact := range facts {
		if _, ok := knowledgebase[fact.Agent]; !ok {
			knowledgebase[fact.Agent] = map[string][]*memory.Knowledge{}
		}
		if _, ok := knowledgebase[fact.Agent][fact.User]; !ok {
			knowledgebase[fact.Agent][fact.User] = []*memory.Knowledge{}
		}

		knowledgebase[fact.Agent][fact.User] = append(knowledgebase[fact.Agent][fact.User], fact)
	}

	return knowledgebase, nil
}

func (store *MySQLStore) GetConversationsToExtractKnowledge() ([]string, error) {
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

func (store *MySQLStore) SetConversationAsKnowledgeExtracted(conversation string) error {
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

func (store *MySQLStore) CompressKnowledge(agent string, user string, knowledge []*memory.Knowledge) error {
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

func (store *MySQLStore) sqlToKnowledge(rows *sql.Rows) ([]*memory.Knowledge, error) {
	defer rows.Close()

	knowledge := []*memory.Knowledge{}

	for rows.Next() {
		var fact memory.Knowledge
		var datetime string
		var expiration string
		err := rows.Scan(
			&fact.ID,
			&fact.Agent,
			&fact.User,
			&fact.Subject,
			&fact.Predicate,
			&fact.Object,
			&datetime,
			&expiration,
		)
		if err != nil {
			return nil, err
		}
		timestamp, err := store.sqlTimestampToTime(datetime)
		if err != nil {
			return nil, err
		}
		fact.CreatedAt = timestamp
		timestamp, err = store.sqlTimestampToTime(expiration)
		if err != nil {
			return nil, err
		}
		fact.ExpiresAt = timestamp
		knowledge = append(knowledge, &fact)
	}

	return knowledge, nil
}
