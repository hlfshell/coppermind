CREATE TABLE IF NOT EXISTS
	Summaries_V1(
        id TEXT NOT NULL PRIMARY KEY,
        conversation TEXT,
        agent TEXT NOT NULL,
        user VARCHAR NOT NULL,
        keywords TEXT,
        summary TEXT,
        created_at TIMESTAMP NOT NULL DEFAULT NOW,

        -- FOREIGN KEY (agent)
        --     REFERENCES Agents_V1(id),
        
        FOREIGN KEY (conversation)
            REFERENCES Conversations_V1(id)
    );

CREATE UNIQUE INDEX IF NOT EXISTS summaries_id_v1 ON Summaries_V1(id);
CREATE INDEX IF NOT EXISTS summaries_conversation_time_V1 ON Summaries_V1(created_at);
CREATE INDEX IF NOT EXISTS summaries_agent_user_time_V1 ON Summaries_V1(agent, user, created_at);
