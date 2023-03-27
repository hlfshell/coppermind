package store

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
const SUMMARIES_TABLE = "Summaries_V1"

//go:embed sqlite/sql/*.sql
var sqlFolder embed.FS
var sqlFolderPath = "sqlite/sql"

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
	query := `INSERT INTO {0} (id, user, message, tone, time, conversation)
	VALUES(?, ?, ?, ?, ?, ?)`

	query = stringFormatter.Format(query, MESSAGES_TABLE)

	_, err := store.db.Exec(
		query,
		msg.ID,
		msg.User,
		msg.Content,
		msg.Tone,
		msg.Time,
		msg.Conversation,
	)

	return err
}

func (store *SqliteStore) GetLatestConversation(user string) (string, time.Time, error) {
	query := `SELECT conversation, time FROM {0} WHERE user = ?
	ORDER BY time DESC LIMIT 1`

	query = stringFormatter.Format(query, MESSAGES_TABLE)

	row, err := store.db.Query(query, user)

	if err != nil {
		return "", time.Time{}, err
	}

	defer row.Close()

	timestamp := time.Time{}
	var conversation string

	for row.Next() {
		err = row.Scan(&conversation, &timestamp)
		if err != nil {
			return "", time.Time{}, err
		}
	}

	return conversation, timestamp, nil
}

func (store *SqliteStore) LoadConversationHistory(conversation string) ([]*chat.Message, error) {
	query := `SELECT
	user,
	message,
	tone,
	conversation,
	time
	FROM {0}
	WHERE conversation = ?`

	query = stringFormatter.Format(query, MESSAGES_TABLE)

	results, err := store.db.Query(query, conversation)

	if err != nil {
		return nil, err
	}

	return store.sqlToMessages(results)
}

func (store *SqliteStore) GetSummariesByUser(user string) ([]*memory.Summary, error) {
	query := `SELECT (
		id,
		conversation,
		agent,
		user,
		keywords,
		summary,
		created_at
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
		created_at
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
		created_at
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
		summary.Brief,
		summary.CreatedAt,
	)

	return err
}

func (store *SqliteStore) sqlToMessages(rows *sql.Rows) ([]*chat.Message, error) {
	defer rows.Close()

	messages := []*chat.Message{}

	for rows.Next() {
		var msg chat.Message
		err := rows.Scan(
			&msg.User,
			&msg.Content,
			&msg.Tone,
			&msg.Conversation,
			&msg.Time,
		)
		if err != nil {
			return nil, err
		}
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

		err := rows.Scan(
			&summary.ID,
			&summary.Agent,
			&keywords,
			&summary.Brief,
			&summary.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		summary.StringToKeywords(keywords)

		summaries = append(summaries, &summary)
	}

	return summaries, nil
}
