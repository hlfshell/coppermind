package postgres

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/artifacts"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/wissance/stringFormatter"
)

const messageSelectColumns = `id, conversation, userId, agent, author, content, created_at`
const artifactDataSelectColumns = `id, message, type, data, created_at`

func (store *PostgresStore) SaveMessage(msg *chat.Message) error {
	query := `INSERT INTO {0} ({1}) VALUES($1, $2, $3, $4, $5, $6, $7)`

	query = stringFormatter.Format(query, MESSAGES_TABLE, messageSelectColumns)

	_, err := store.db.Exec(
		query,
		msg.ID,
		msg.Conversation,
		msg.User,
		msg.Agent,
		msg.From,
		msg.Content,
		msg.CreatedAt,
	)
	if err != nil {
		return err
	}

	// Finally save any artifact data included
	return store.saveArtifactData(msg.Artifacts)
}

func (store *PostgresStore) saveArtifactData(data []*artifacts.ArtifactData) error {
	// If we have no data, we can just return
	if len(data) == 0 {
		return nil
	}

	query := `INSERT INTO {0} ({1}) VALUES {2}`

	// We need a ($1, $2, $3, $4, $5) for each artifact data, with a comma
	// between each set of values and obviously increasing over time
	placeholders := ""
	for i := 0; i < len(data); i++ {
		if i > 0 {
			placeholders += ", "
		}

		// We need to do $# for each, starting at 1
		// and jumping by 5 for each set
		placeholders += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5)
	}

	query = stringFormatter.Format(query, ARTIFACTS_TABLE, artifactDataSelectColumns, placeholders)

	// Now we need to flatten the data into a single array of values
	values := []interface{}{}
	for _, artifact := range data {
		values = append(
			values,
			artifact.ID,
			artifact.Message,
			artifact.Type,
			artifact.Data,
			artifact.CreatedAt,
		)
	}

	_, err := store.db.Exec(query, values...)
	return err
}

func (store *PostgresStore) GetMessage(id string) (*chat.Message, error) {
	query := `SELECT {0} FROM {1} WHERE id = $1`

	query = stringFormatter.Format(query, messageSelectColumns, MESSAGES_TABLE)

	rows, err := store.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	messages, err := store.sqlToMessages(rows)
	if err != nil {
		return nil, nil
	} else if len(messages) == 0 {
		return nil, nil
	}

	messages, err = store.populateArtifacts(messages)
	if err != nil {
		return nil, err
	}

	return messages[0], nil
}

func (store *PostgresStore) populateArtifacts(messages []*chat.Message) ([]*chat.Message, error) {
	messageIndexes := map[string]int{}
	messageIds := []interface{}{}
	paramString := "("
	for i, msg := range messages {
		if i > 0 {
			paramString += ", "
		}
		messageIds = append(messageIds, msg.ID)
		paramString += fmt.Sprintf("$%d", i+1)
		messageIndexes[msg.ID] = i
	}
	paramString += ")"

	query := `SELECT {0} FROM {1} WHERE message IN {2}`
	query = stringFormatter.Format(
		query,
		artifactDataSelectColumns,
		ARTIFACTS_TABLE,
		paramString,
	)
	rows, err := store.db.Query(query, messageIds...)
	if err != nil {
		return nil, err
	}
	artifactData, err := store.sqlToArtifacts(rows)
	if err != nil {
		return nil, err
	}
	for index, artifact := range artifactData {
		messages[messageIndexes[artifact.Message]].Artifacts = append(
			messages[messageIndexes[artifact.Message]].Artifacts,
			&artifactData[index],
		)
	}

	return messages, nil
}

func (store *PostgresStore) DeleteMessage(id string) error {
	query := `DELETE FROM {0} WHERE id = $1`

	query = stringFormatter.Format(query, MESSAGES_TABLE)

	_, err := store.db.Exec(query, id)
	if err != nil {
		return err
	}

	query = `DELETE FROM {0} WHERE message = $1`
	query = stringFormatter.Format(query, ARTIFACTS_TABLE)
	_, err = store.db.Exec(query, id)
	return err
}

func (store *PostgresStore) ListMessages(filter store.Filter) ([]*chat.Message, error) {
	query := `SELECT {columns} FROM {table} `
	var filters string
	var params []interface{}

	if !filter.Empty() {
		query += `WHERE {where} `

		var err error

		filters, params, err = filterToQueryParams(filter)
		if err != nil {
			return nil, err
		}
	}
	query += `ORDER BY created_at ASC `

	if filter.Limit > 0 {
		query += `LIMIT {limit}`
	}

	query = stringFormatter.FormatComplex(
		query,
		map[string]interface{}{
			"columns": messageSelectColumns,
			"table":   MESSAGES_TABLE,
			"where":   filters,
			"limit":   filter.Limit,
		},
	)

	rows, err := store.db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	messages, err := store.sqlToMessages(rows)
	if err != nil {
		return nil, err
	}

	return store.populateArtifacts(messages)
}

func (store *PostgresStore) GetConversation(conversation string) (*chat.Conversation, error) {
	query := `SELECT {0} FROM {1} WHERE conversation = $1`

	query = stringFormatter.Format(query, messageSelectColumns, MESSAGES_TABLE)

	rows, err := store.db.Query(query, conversation)
	if err != nil {
		return nil, err
	}
	messages, err := store.sqlToMessages(rows)
	if err != nil {
		return nil, nil
	} else if len(messages) == 0 {
		return nil, nil
	}

	return &chat.Conversation{
		ID:        conversation,
		User:      messages[0].User,
		Agent:     messages[0].Agent,
		CreatedAt: messages[0].CreatedAt,
		Messages:  messages,
	}, nil
}

func (store *PostgresStore) DeleteConversation(conversation string) error {
	query := `DELETE FROM {0} WHERE conversation = $1`

	query = stringFormatter.Format(query, MESSAGES_TABLE)

	_, err := store.db.Exec(query, conversation)
	return err
}

func (db *PostgresStore) ListConversations(filter store.Filter) ([]*chat.Conversation, error) {
	// First we find the conversations via a set query, then find all messages
	// within that conversation
	query := `WITH conversations AS
		(
			SELECT
				conversation,
				user,
				agent,
				MIN(created_at) as started_at
			FROM
				{table}
			GROUP BY
				conversation
		)
		SELECT
			conversation as id,
			user,
			agent,
			started_at as created_at
		FROM
			conversations `

	// const conversationColumns = `id, created_at, agent, user`

	if !filter.Empty() {
		query += `WHERE {filter} `
	}

	query += `GROUP BY conversation ORDER BY {orderBy} `

	if filter.Limit > 0 {
		query += `LIMIT {limit}`
	}

	whereFilter, params, err := filterToQueryParams(filter)
	if err != nil {
		return nil, err
	}

	var orderBy string
	if filter.OrderBy.Nil() {
		orderBy = `created_at ASC `
	} else {
		var dir string
		if filter.OrderBy.Ascending {
			dir = "ASC"
		} else {
			dir = "DESC"
		}
		orderBy = fmt.Sprintf(`%s %s `, filter.OrderBy.Attribute, dir)
	}

	query = stringFormatter.FormatComplex(
		query,
		map[string]interface{}{
			"table":   MESSAGES_TABLE,
			"filter":  whereFilter,
			"limit":   filter.Limit,
			"orderBy": orderBy,
		},
	)

	rows, err := db.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	conversationIds := []string{}
	for rows.Next() {
		var conversation string
		var datetime string
		var user string
		var agent string
		err = rows.Scan(
			&conversation,
			&user,
			&agent,
			&datetime,
		)
		if err != nil {
			return nil, err
		}
		conversationIds = append(conversationIds, conversation)
	}

	// Abort if we found no matching conversations according to our filter
	if len(conversationIds) == 0 {
		return []*chat.Conversation{}, nil
	}

	// Now for each conversations, query the messages
	messages, err := db.ListMessages(store.Filter{
		Attributes: []*store.FilterAttribute{
			{
				Attribute: "conversation",
				Value:     conversationIds,
				Operation: store.IN,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Now sort through the messages and group them by conversation. Then
	// create a conversation object for each
	conversationMap := map[string]*chat.Conversation{}
	for _, msg := range messages {
		if _, ok := conversationMap[msg.Conversation]; !ok {
			conversationMap[msg.Conversation] = &chat.Conversation{
				ID:        msg.Conversation,
				User:      msg.User,
				Agent:     msg.Agent,
				CreatedAt: msg.CreatedAt,
				Messages:  []*chat.Message{},
			}
		}
		conversationMap[msg.Conversation].Messages = append(conversationMap[msg.Conversation].Messages, msg)
	}

	// Finally return the generated conversations. Order them by the conversation
	// Createdat time
	conversations := []*chat.Conversation{}
	for _, conversation := range conversationMap {
		conversations = append(conversations, conversation)
	}
	return orderConversations(conversations), nil
}

func (store *PostgresStore) ListConversations2(filter store.Filter) ([]*chat.Conversation, error) {
	messages, err := store.ListMessages(filter)
	if err != nil {
		return nil, err
	}

	// Now sort through the messages and group them by conversation. Then
	// create a conversation object for each
	conversationMap := map[string]*chat.Conversation{}
	for _, msg := range messages {
		if _, ok := conversationMap[msg.Conversation]; !ok {
			conversationMap[msg.Conversation] = &chat.Conversation{
				ID:        msg.Conversation,
				User:      msg.User,
				Agent:     msg.Agent,
				CreatedAt: msg.CreatedAt,
				Messages:  []*chat.Message{},
			}
		}
		conversationMap[msg.Conversation].Messages = append(conversationMap[msg.Conversation].Messages, msg)
	}

	// Finally return the generated conversations. Order them by the conversation
	// Createdat time
	conversations := []*chat.Conversation{}
	for _, conversation := range conversationMap {
		conversations = append(conversations, conversation)
	}
	return orderConversations(conversations), nil
}

/*
orderConversations will order a slice of conversations by their
CreatedAtTime; oldest first
*/
func orderConversations(conversations []*chat.Conversation) []*chat.Conversation {
	comparator := func(i, j int) bool {
		return conversations[i].CreatedAt.Before(conversations[j].CreatedAt)
	}
	sort.Slice(conversations, comparator)
	return conversations
}

func (store *PostgresStore) GetLatestConversation(agent string, user string) (string, time.Time, error) {
	query := `SELECT
		conversation,
		MAX(created_at) as latest_message
	FROM {0}
	WHERE
		agent = $1 AND
		userId = $2
	GROUP BY conversation
	ORDER BY latest_message DESC
	LIMIT 1;
	`

	query = stringFormatter.Format(query, MESSAGES_TABLE)

	row, err := store.db.Query(query, agent, user)
	if err != nil {
		return "", time.Time{}, err
	}

	defer row.Close()

	var timestamp time.Time
	var conversation sql.NullString

	for row.Next() {
		err = row.Scan(&conversation, &timestamp)
		if err != nil {
			return "", time.Time{}, err
		}
	}

	return conversation.String, timestamp, nil
}

func (store *PostgresStore) sqlToMessages(rows *sql.Rows) ([]*chat.Message, error) {
	defer rows.Close()

	messages := []*chat.Message{}

	for rows.Next() {
		var msg chat.Message
		err := rows.Scan(
			&msg.ID,
			&msg.Conversation,
			&msg.User,
			&msg.Agent,
			&msg.From,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

func (store *PostgresStore) sqlToArtifacts(rows *sql.Rows) ([]artifacts.ArtifactData, error) {
	defer rows.Close()

	artifactData := []artifacts.ArtifactData{}

	for rows.Next() {
		var data artifacts.ArtifactData
		err := rows.Scan(
			&data.ID,
			&data.Message,
			&data.Type,
			&data.Data,
			&data.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		artifactData = append(artifactData, data)
	}
	return artifactData, nil
}
