package sqlite

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/wissance/stringFormatter"
)

const messageSelectColumns = `id, conversation, user, agent, author, content, created_at`

func (store *SqliteStore) SaveMessage(msg *chat.Message) error {
	query := `INSERT INTO {0} ({1}) VALUES(?, ?, ?, ?, ?, ?, ?)`

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

	return err
}

func (store *SqliteStore) GetMessage(id string) (*chat.Message, error) {
	query := `SELECT {0} FROM {1} WHERE id = ?`

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

	return messages[0], nil
}

func (store *SqliteStore) DeleteMessage(id string) error {
	query := `DELETE FROM {0} WHERE id = ?`

	query = stringFormatter.Format(query, MESSAGES_TABLE)

	_, err := store.db.Exec(query, id)
	return err
}

func (store *SqliteStore) ListMessages(filter store.Filter) ([]*chat.Message, error) {
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

	fmt.Println("Query", query)
	rows, err := store.db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	return store.sqlToMessages(rows)
}

func (store *SqliteStore) GetConversation(conversation string) (*chat.Conversation, error) {
	query := `SELECT {0} FROM {1} WHERE conversation = ?`

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

func (store *SqliteStore) DeleteConversation(conversation string) error {
	query := `DELETE FROM {0} WHERE conversation = ?`

	query = stringFormatter.Format(query, MESSAGES_TABLE)

	_, err := store.db.Exec(query, conversation)
	return err
}

func (db *SqliteStore) ListConversations(filter store.Filter) ([]*chat.Conversation, error) {
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
	fmt.Println("query", query)
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

func (store *SqliteStore) ListConversations2(filter store.Filter) ([]*chat.Conversation, error) {
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

func (store *SqliteStore) GetLatestConversation(agent string, user string) (string, time.Time, error) {
	query := `SELECT
		conversation,
		MAX(created_at) as latest_message
	FROM {0}
	WHERE
		agent = ? AND
		user = ?
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

	var timestring sql.NullString
	var timestamp time.Time
	var conversation sql.NullString

	for row.Next() {
		err = row.Scan(&conversation, &timestring)
		if err != nil {
			return "", time.Time{}, err
		}
		if !conversation.Valid || !timestring.Valid {
			return "", time.Time{}, nil
		}
		timestamp, err = store.sqlTimestampToTime(timestring.String)
		if err != nil {
			return "", time.Time{}, err
		}
	}

	return conversation.String, timestamp, nil
}

func (store *SqliteStore) sqlToMessages(rows *sql.Rows) ([]*chat.Message, error) {
	defer rows.Close()

	messages := []*chat.Message{}

	for rows.Next() {
		var msg chat.Message
		var datetime string
		err := rows.Scan(
			&msg.ID,
			&msg.Conversation,
			&msg.User,
			&msg.Agent,
			&msg.From,
			&msg.Content,
			&datetime,
		)
		if err != nil {
			return nil, err
		}
		timestamp, err := store.sqlTimestampToTime(datetime)
		if err != nil {
			return nil, err
		}
		msg.CreatedAt = timestamp
		messages = append(messages, &msg)
	}

	return messages, nil
}
