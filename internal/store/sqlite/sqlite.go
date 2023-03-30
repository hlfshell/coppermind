package sqlite

import (
	"database/sql"
	"embed"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/hlfshell/coppermind/internal/chat"
	"github.com/hlfshell/coppermind/internal/memory"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wissance/stringFormatter"
)

const MESSAGES_TABLE = "Messages_V1"
const CONVERSATIONS_TABLE = "Conversations_V1"
const SUMMARIES_TABLE = "Summaries_V1"

//go:embed sql/*.sql
var sqlFolder embed.FS
var sqlFolderPath = "sql"

type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore(dbFilePath string) (*SqliteStore, error) {
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	return &SqliteStore{
		db: db,
	}, nil
}

func (store *SqliteStore) Migrate() error {
	entries, err := fs.ReadDir(sqlFolder, sqlFolderPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(sqlFolderPath, entry.Name())
		bytes, err := fs.ReadFile(sqlFolder, path)
		if err != nil {
			return err
		}
		migrationQuery := string(bytes)

		_, err = store.db.Exec(migrationQuery)
		if err != nil {
			return err
		}
	}

	return nil
}

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

// func (store *SqliteStore) GetConversation(conversation string) (*chat.Conversation, error) {
// 	query := `
// 		SELECT id, agent, user, created_at
// 		FROM {0} WHERE id = ?
// 	`
// 	query = stringFormatter.Format(query, CONVERSATIONS_TABLE)

// 	rows, err := store.db.Query(query, conversation)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer rows.Close()

// 	var result *chat.Conversation
// 	var datetime string
// 	for rows.Next() {
// 		result = &chat.Conversation{}
// 		err = rows.Scan(
// 			&result.ID,
// 			&result.Agent,
// 			&result.User,
// 			&datetime,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		timestamp, err := store.sqlTimestampToTime(datetime)
// 		if err != nil {
// 			return nil, err
// 		}
// 		result.CreatedAt = timestamp
// 	}
// 	return result, nil
// }

// func (store *SqliteStore) SaveConversation(conversation *chat.Conversation) error {
// 	query := `INSERT INTO {0} (id, agent, user, created_at)
// 	VALUES(?, ?, ?, ?)`

// 	query = stringFormatter.Format(query, CONVERSATIONS_TABLE)

// 	_, err := store.db.Exec(
// 		query,
// 		conversation.ID,
// 		conversation.Agent,
// 		conversation.User,
// 		conversation.CreatedAt,
// 	)

// 	return err
// }

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

func (store *SqliteStore) GetSummariesByUser(user string) ([]*memory.Summary, error) {
	query := `SELECT (
		id,
		conversation,
		agent,
		user,
		keywords,
		summary,
		updated_at
	) FROM {0} WHERE user = ?
	`

	query = stringFormatter.Format(query, SUMMARIES_TABLE)

	rows, err := store.db.Query(query, user)
	if err != nil {
		return nil, err
	}

	summaries, err := store.sqlToSummmaries(rows)
	if err != nil {
		return nil, err
	}

	return summaries, nil
}

func (store *SqliteStore) GetSummaryByConversation(conversation string) (*memory.Summary, error) {
	query := `SELECT
		id,
		conversation,
		agent,
		user,
		keywords,
		summary,
		updated_at
	FROM {0} WHERE conversation = ?`

	query = stringFormatter.Format(query, SUMMARIES_TABLE)

	rows, err := store.db.Query(
		query,
		conversation,
	)
	if err != nil {
		return nil, err
	}
	summaries, err := store.sqlToSummmaries(rows)
	if err != nil {
		return nil, err
	}
	if len(summaries) == 0 {
		return nil, nil
	} else {
		return summaries[0], nil
	}
}

func (store *SqliteStore) SaveSummary(summary *memory.Summary) error {
	query := `INSERT INTO {0} (
		id,
		conversation,
        agent,
        user,
        keywords,
        summary,
		updated_at
	)
	VALUES(?, ?, ?, ?, ?, ?, ?)`

	query = stringFormatter.Format(query, SUMMARIES_TABLE)

	_, err := store.db.Exec(
		query,
		summary.ID,
		summary.Conversation,
		summary.Agent,
		summary.User,
		summary.KeywordsToString(),
		summary.Summary,
		summary.UpdatedAt,
	)

	return err
}

func (store *SqliteStore) GetConversationsToSummarize() ([]string, error) {
	query := `
		WITH target_conversations AS(
		SELECT
			messages.conversation as conversationId,
			MAX(messages.created_at) as latest_message,
			summaries.id as summary,
			summaries.updated_at as summary_updated_at
		FROM {0} AS messages LEFT JOIN {1} AS summaries
			ON messages.conversation = summaries.conversation
		GROUP BY messages.conversation
		HAVING
			summary IS NULL OR
			MAX(messages.created_at) > summaries.updated_at
		)
		SELECT conversationId FROM target_conversations;
	`

	query = stringFormatter.Format(query, MESSAGES_TABLE, SUMMARIES_TABLE, MESSAGES_TABLE)

	rows, err := store.db.Query(query)
	if err != nil {
		return nil, err
	}

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

func (store *SqliteStore) sqlToSummmaries(rows *sql.Rows) ([]*memory.Summary, error) {
	defer rows.Close()

	summaries := []*memory.Summary{}

	for rows.Next() {
		var summary memory.Summary
		var keywords string
		var datetime string

		err := rows.Scan(
			&summary.ID,
			&summary.Conversation,
			&summary.Agent,
			&summary.User,
			&keywords,
			&summary.Summary,
			&datetime,
		)
		if err != nil {
			return nil, err
		}
		timestamp, err := store.sqlTimestampToTime(datetime)
		if err != nil {
			return nil, err
		}
		summary.UpdatedAt = timestamp

		summary.StringToKeywords(keywords)

		summaries = append(summaries, &summary)
	}

	return summaries, nil
}

func (store *SqliteStore) sqlTimestampToTime(timestamp string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		parsed, err = time.Parse("2006-01-02 15:04:05.999999999-07:00", timestamp)
		if err != nil {
			return time.Time{}, err
		}
	}
	return parsed, nil
}
