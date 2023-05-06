package sqlite

import (
	"database/sql"
	"time"

	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/memory"
	"github.com/wissance/stringFormatter"
)

func (store *SqliteStore) SaveSummary(summary *memory.Summary) error {
	query := `INSERT OR REPLACE INTO {0} (
		id,
		conversation,
        agent,
        user,
        keywords,
        summary,
		conversation_started_at,
		updated_at
	)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?)`

	summary.UpdatedAt = time.Now()

	query = stringFormatter.Format(query, SUMMARIES_TABLE)

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

func (store *SqliteStore) GetSummary(id string) (*memory.Summary, error) {
	query := `SELECT
		id,
		conversation,
		agent,
		user,
		keywords,
		summary,
		conversation_started_at,
		updated_at
	FROM {0} WHERE id = ?`

	query = stringFormatter.Format(query, SUMMARIES_TABLE)

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

func (store *SqliteStore) DeleteSummary(id string) error {
	query := `DELETE FROM {0} WHERE id = ?`

	query = stringFormatter.Format(query, SUMMARIES_TABLE)

	_, err := store.db.Exec(query, id)
	return err
}

func (store *SqliteStore) ListSummaries(filter store.Filter) ([]*memory.Summary, error) {
	queryFilters, params, err := filterToQueryParams(filter)
	if err != nil {
		return nil, err
	}

	query := `SELECT
		id,
		conversation,
		agent,
		user,
		keywords,
		summary,
		conversation_started_at,
		updated_at
	FROM
		{0}
	WHERE
		{1}
	`
	query = stringFormatter.Format(query, SUMMARIES_TABLE, queryFilters)

	rows, err := store.db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	summaries, err := store.sqlToSummmaries(rows)
	return summaries, err
}

func (store *SqliteStore) GetSummariesByAgentAndUser(agent string, user string) ([]*memory.Summary, error) {
	query := `SELECT
		id,
		conversation,
		agent,
		user,
		keywords,
		summary,
		conversation_started_at,
		updated_at
	FROM {0} WHERE agent = ? AND user = ?
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

func (store *SqliteStore) GetSummaryByConversation(conversation string) (*memory.Summary, error) {
	query := `SELECT
		id,
		conversation,
		agent,
		user,
		keywords,
		summary,
		conversation_started_at,
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

func (store *SqliteStore) GetConversationsToSummarize(minMessages int, minAge time.Duration, maxMessages int) ([]string, error) {
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
				messages.conversation
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
				(latest_message <= ? AND messages_count >= ?) OR
				(messages_since_summary.messages_since_update >= ?) OR
				(summary IS NULL AND messages_count >= ?)
			)
	`

	query = stringFormatter.Format(query, MESSAGES_TABLE, SUMMARIES_TABLE, SUMMARY_EXCLUSION_TABLE)

	rows, err := store.db.Query(query, ageTime, minMessages, maxMessages, maxMessages)
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

func (store *SqliteStore) ExcludeConversationFromSummary(conversation string) error {
	query := `
		INSERT INTO {0} (
			conversation,
			created_at
		) VALUES(?, ?)
	`

	query = stringFormatter.Format(query, SUMMARY_EXCLUSION_TABLE)

	_, err := store.db.Exec(query, conversation, time.Now())
	return err
}

func (store *SqliteStore) DeleteSummaryExclusion(conversation string) error {
	query := `DELETE FROM {0} WHERE conversation = ?`

	query = stringFormatter.Format(query, SUMMARY_EXCLUSION_TABLE)

	_, err := store.db.Exec(query, conversation)
	return err
}

func (store *SqliteStore) sqlToSummmaries(rows *sql.Rows) ([]*memory.Summary, error) {
	defer rows.Close()

	summaries := []*memory.Summary{}

	for rows.Next() {
		var summary memory.Summary
		var keywords string
		var updatedTime string
		var conversationStartTime string

		err := rows.Scan(
			&summary.ID,
			&summary.Conversation,
			&summary.Agent,
			&summary.User,
			&keywords,
			&summary.Summary,
			&conversationStartTime,
			&updatedTime,
		)
		if err != nil {
			return nil, err
		}
		timestamp, err := store.sqlTimestampToTime(updatedTime)
		if err != nil {
			return nil, err
		}
		summary.UpdatedAt = timestamp

		timestamp, err = store.sqlTimestampToTime(conversationStartTime)
		if err != nil {
			return nil, err
		}
		summary.ConversationStartedAt = timestamp

		summary.StringToKeywords(keywords)

		summaries = append(summaries, &summary)
	}

	return summaries, nil
}
