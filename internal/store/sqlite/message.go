package sqlite

import (
	"database/sql"
	"time"

	"github.com/hlfshell/coppermind/pkg/chat"
	"github.com/wissance/stringFormatter"
)

func (store *SqliteStore) SaveMessage(msg *chat.Message) error {
	query := `INSERT INTO {0} (id, user, agent, content, tone, created_at, conversation)
	VALUES(?, ?, ?, ?, ?, ?, ?)`

	query = stringFormatter.Format(query, MESSAGES_TABLE)

	_, err := store.db.Exec(
		query,
		msg.ID,
		msg.User,
		msg.Agent,
		msg.Content,
		msg.Tone,
		msg.CreatedAt,
		msg.Conversation,
	)

	return err
}

func (store *SqliteStore) GetConversation(conversation string) (*chat.Conversation, error) {
	query := `SELECT
		id,
		conversation,
		user,
		agent,
		content,
		tone,
		created_at
	FROM {0} WHERE conversation = ?
	`

	query = stringFormatter.Format(query, MESSAGES_TABLE)

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
			&msg.Content,
			&msg.Tone,
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
