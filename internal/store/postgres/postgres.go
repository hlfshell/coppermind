package postgres

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

const USERS_TABLE = "Users_V1"
const AGENTS_TABLE = "Agents_V1"
const MESSAGES_TABLE = "Messages_V1"
const ARTIFACTS_TABLE = "Artifacts_V1"
const SUMMARIES_TABLE = "Summaries_V1"
const SUMMARY_EXCLUSION_TABLE = "SummaryExclusion_V1"
const KNOWLEDGE_TABLE = "Knowledge_V1"
const KNOWLEDGE_EXTRACTION_TABLE = "KnowledgeExtraction_V1"

//go:embed sql/*.sql
var sqlFolder embed.FS
var sqlFolderPath = "sql"

type PostgresStore struct {
	db *sql.DB
}

func NewSqliteStore(username string, password, host string, port string, database string) (*PostgresStore, error) {
	connectionString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (store *PostgresStore) Migrate() error {
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

func (store *PostgresStore) sqlTimestampToTime(timestamp string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		parsed, err = time.Parse("2006-01-02 15:04:05.999999999-07:00", timestamp)
		if err != nil {
			return time.Time{}, err
		}
	}
	return parsed, nil
}
