CREATE TABLE IF NOT EXISTS
	Summaries_V1(
        id TEXT NOT NULL PRIMARY KEY,
        conversation TEXT,
        agent TEXT NOT NULL,
        user TEXT NOT NULL,
        keywords TEXT,
        summary TEXT NOT NULL,
        updated_at TIMESTAMP NOT NULL DEFAULT NOW,
        conversation_started_at TIMESTAMP NOT NULL
    );

CREATE UNIQUE INDEX IF NOT EXISTS summaries_id_v1 ON Summaries_V1(id);
CREATE UNIQUE INDEX IF NOT EXISTS summaries_conversation_v1 ON Summaries_V1(conversation);
CREATE INDEX IF NOT EXISTS summaries_conversation_time_V1 ON Summaries_V1(updated_at);
CREATE INDEX IF NOT EXISTS summaries_agent_user_time_V1 ON Summaries_V1(agent, user, updated_at);

CREATE TABLE IF NOT EXISTS
    SummaryExclusion_V1(
        conversation TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW
    );

CREATE UNIQUE INDEX IF NOT EXISTS summary_exclusion_v1 ON SummaryExclusion_V1(conversation);