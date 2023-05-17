package mysql

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(username string, password string, host string, port string, database string) (*MySQLStore, error) {
	// connectionString := fmt.Sprintf("mysql://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		username,
		password,
		host,
		port,
		database,
	)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		fmt.Println("Couldn't open connection", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return &MySQLStore{
		db: db,
	}, nil
}

func (store *MySQLStore) Migrate() error {
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

func (store *MySQLStore) sqlTimestampToTime(timestamp string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		parsed, err = time.Parse("2006-01-02 15:04:05.999999999-07:00", timestamp)
		if err != nil {
			return time.Time{}, err
		}
	}
	return parsed, nil
}
