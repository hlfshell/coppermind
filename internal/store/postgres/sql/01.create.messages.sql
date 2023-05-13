CREATE TABLE IF NOT EXISTS
	Messages_V1(
		id TEXT NOT NULL PRIMARY KEY,
		userId TEXT NOT NULL,
		agent TEXT NOT NULL,
		author TEXT NOT NULL,
		content TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		conversation TEXT NOT NULL
	);

CREATE UNIQUE INDEX IF NOT EXISTS messages_id_v1 ON Messages_V1 USING BTREE(id);
CREATE INDEX IF NOT EXISTS messages_conversation_time_v1 ON Messages_V1 USING BTREE(conversation, created_at ASC);
CREATE INDEX IF NOT EXISTS messages_users_time_v1 ON Messages_V1 USING BTREE(userId, created_at ASC);
CREATE INDEX IF NOT EXISTS messages_agent_user_time_v1 ON Messages_V1 USING BTREE(agent, userId, created_at ASC);