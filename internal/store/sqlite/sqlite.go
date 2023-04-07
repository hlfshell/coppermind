package sqlite

import (
	"database/sql"
	"embed"
	"io/fs"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const MESSAGES_TABLE = "Messages_V1"
const SUMMARIES_TABLE = "Summaries_V1"
const SUMMARY_EXCLUSION_TABLE = "SummaryExclusion_V1"
const KNOWLEDGE_TABLE = "Knowledge_V1"
const KNOWLEDGE_EXTRACTION_TABLE = "KnowledgeExtraction_V1"

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
