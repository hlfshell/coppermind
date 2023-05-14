package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/wissance/stringFormatter"
)

const summaryColumns = `id, conversation, agent, userId, keywords, summary, conversation_started_at, updated_at`

func (store *PostgresStore) SaveSummary(summary *memory.Summary) error {
	query := `INSERT INTO {0} ({1}) VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET {2}`

	summary.UpdatedAt = time.Now()

	// We need to build the update placeholders for each param
	summaryColumnsSplit := strings.Split(summaryColumns, ", ")
	// Drop the id column
	summaryColumnsSplit = summaryColumnsSplit[1:]
	updatePlaceholder := strings.Builder{}
	for i, column := range summaryColumnsSplit {
		if i > 0 {
			updatePlaceholder.WriteString(", ")
		}
		updatePlaceholder.WriteString(column)
		// It's + 2 as we dropped the id column
		updatePlaceholder.WriteString(fmt.Sprintf(" = $%d", i+2))
	}

	query = stringFormatter.Format(query, SUMMARIES_TABLE, summaryColumns, updatePlaceholder.String())

	_, err := store.db.Exec(
		query,
		summary.ID,
		summary.Conversation,
		summary.Agent,
		summary.User,
		summary.KeywordsToString(),
		summary.Summary,
		summary.ConversationStartedAt,
		summary.UpdatedAt,
	)

	return err
}

func (store *PostgresStore) GetSummary(id string) (*memory.Summary, error) {
	query := `SELECT {0} FROM {1} WHERE id = $1`

	query = stringFormatter.Format(query, summaryColumns, SUMMARIES_TABLE)

	rows, err := store.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	summaries, err := store.sqlToSummmaries(rows)
	if err != nil {
		return nil, nil
	} else if len(summaries) == 0 {
		return nil, nil
	}

	return summaries[0], nil
}

func (store *PostgresStore) DeleteSummary(id string) error {
	query := `DELETE FROM {0} WHERE id = $1`

	query = stringFormatter.Format(query, SUMMARIES_TABLE)

	_, err := store.db.Exec(query, id)
	return err
}

func (store *PostgresStore) ListSummaries(filter store.Filter) ([]*memory.Summary, error) {
	query := `SELECT {columns} FROM {table} `
	var filters string
	var params []interface{}
	var err error

	if !filter.Empty() {
		filters, params, err = filterToQueryParams(filter)
		if err != nil {
			return nil, err
		}
		query += `WHERE {filters} `
	}

	query += `ORDER BY conversation_started_at ASC `

	if filter.Limit > 0 {
		query += `LIMIT {limit} `
	}

	query = stringFormatter.FormatComplex(
		query,
		map[string]interface{}{
			"columns": summaryColumns,
			"table":   SUMMARIES_TABLE,
			"filters": filters,
			"limit":   filter.Limit,
		},
	)

	fmt.Println("list query", len(params), query)
	fmt.Println(params)
	rows, err := store.db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	summaries, err := store.sqlToSummmaries(rows)
	return summaries, err
}

func (store *PostgresStore) GetSummariesByAgentAndUser(agent string, user string) ([]*memory.Summary, error) {
	query := `SELECT
		id,
		conversation,
		agent,
		user,
		keywords,
		summary,
		conversation_started_at,
		updated_at
	FROM {0} WHERE agent = $1 AND user = $2
	`

	query = stringFormatter.Format(query, SUMMARIES_TABLE)

	rows, err := store.db.Query(query, agent, user)
	if err != nil {
		return nil, err
	}

	summaries, err := store.sqlToSummmaries(rows)
	if err != nil {
		return nil, err
	}

	return summaries, nil
}

func (store *PostgresStore) GetSummaryByConversation(conversation string) (*memory.Summary, error) {
	query := `SELECT
		id,
		conversation,
		agent,
		user,
		keywords,
		summary,
		conversation_started_at,
		updated_at
	FROM {0} WHERE conversation = $1`

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

func (store *PostgresStore) GetConversationsToSummarize(minMessages int, minAge time.Duration, maxMessages int) ([]string, error) {
	ageTime := time.Now().Add(-1 * minAge)

	query := `
		WITH target_conversations AS (
			SELECT
				messages.conversation as conversationId,
				MAX(messages.created_at) as latest_message,
				COUNT(messages.id) as messages_count,
				summaries.id as summary,
				summaries.updated_at as summary_updated_at
			FROM
				{0} AS messages LEFT JOIN
				{1} AS summaries ON messages.conversation = summaries.conversation LEFT JOIN
				{2} AS exclusion ON messages.conversation = exclusion.conversation
			WHERE
				exclusion.conversation IS NULL
			GROUP BY
				messages.conversation, summaries.id
		),
		messages_since_summary AS (
			SELECT
				conversationId,
				COUNT(messages.id) AS messages_since_update
			FROM
				{0} messages
				LEFT JOIN target_conversations ON
					messages.conversation = target_conversations.conversationId AND
					messages.created_at > target_conversations.summary_updated_at
			WHERE
				target_conversations.summary IS NOT NULL
			GROUP BY
				target_conversations.conversationId
		)
		SELECT
			target_conversations.conversationId
		FROM
			target_conversations LEFT JOIN
				messages_since_summary ON
				target_conversations.conversationId = messages_since_summary.conversationId
		WHERE
			(summary IS NULL OR latest_message > summary_updated_at) AND
			(
				(latest_message <= $1 AND messages_count >= $2) OR
				(messages_since_summary.messages_since_update >= $3) OR
				(summary IS NULL AND messages_count >= $3)
			)
	`

	query = stringFormatter.Format(query, MESSAGES_TABLE, SUMMARIES_TABLE, SUMMARY_EXCLUSION_TABLE)
	fmt.Println("query", query)
	rows, err := store.db.Query(query, ageTime, minMessages, maxMessages)
	if err != nil {
		fmt.Println("err on query", err)
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

func (store *PostgresStore) ExcludeConversationFromSummary(conversation string) error {
	query := `
		INSERT INTO {0} (
			conversation,
			created_at
		) VALUES($1, $2)
	`

	query = stringFormatter.Format(query, SUMMARY_EXCLUSION_TABLE)

	_, err := store.db.Exec(query, conversation, time.Now())
	return err
}

func (store *PostgresStore) DeleteSummaryExclusion(conversation string) error {
	query := `DELETE FROM {0} WHERE conversation = $1`

	query = stringFormatter.Format(query, SUMMARY_EXCLUSION_TABLE)

	_, err := store.db.Exec(query, conversation)
	return err
}

func (store *PostgresStore) sqlToSummmaries(rows *sql.Rows) ([]*memory.Summary, error) {
	defer rows.Close()

	summaries := []*memory.Summary{}

	for rows.Next() {
		var summary memory.Summary
		var keywords string

		err := rows.Scan(
			&summary.ID,
			&summary.Conversation,
			&summary.Agent,
			&summary.User,
			&keywords,
			&summary.Summary,
			&summary.ConversationStartedAt,
			&summary.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		summary.StringToKeywords(keywords)

		summaries = append(summaries, &summary)
	}

	return summaries, nil
}
