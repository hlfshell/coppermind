CREATE TABLE IF NOT EXISTS
	Messages_V1(
		id TEXT NOT NULL PRIMARY KEY,
		user TEXT NOT NULL,
		agent TEXT NOT NULL,
		content TEXT,
		tone TEXT,
		created_at TIMESTAMP DEFAULT NOW, 
		conversation TEXT NOT NULL
	);

CREATE UNIQUE INDEX IF NOT EXISTS messages_id_v1 ON Messages_V1(id);
CREATE INDEX IF NOT EXISTS messages_conversation_time_v1 ON Messages_V1(conversation, created_at);
CREATE INDEX IF NOT EXISTS messages_users_time_v1 ON Messages_V1(user, created_at);
CREATE INDEX IF NOT EXISTS messages_agent_user_time_v1 ON Messages_V1(agent, user, created_at);