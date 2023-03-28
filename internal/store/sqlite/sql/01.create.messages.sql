CREATE TABLE IF NOT EXISTS
	Conversations_V1(
		id TEXT NOT NULL PRIMARY KEY,
		created_at TIMESTAMP DEFAULT NOW,
		agent TEXT NOT NULL,
		user TEXT NOT NULL
	);

CREATE UNIQUE INDEX IF NOT EXISTS conversations_id_v1 ON Conversations_V1(id);
CREATE INDEX IF NOT EXISTS conversations_agents_time_v1 ON Conversations_V1(agent, created_at); 
CREATE INDEX IF NOT EXISTS conversations_users_time_v1 ON Conversations_V1(user, created_at);
CREATE INDEX IF NOT EXISTS conversations_users_agents_time_v1 ON Conversations_V1(agent, user, created_at);

CREATE TABLE IF NOT EXISTS
	Messages_V1(
		id TEXT NOT NULL PRIMARY KEY,
		user TEXT NOT NULL,
		message TEXT,
		tone TEXT,
		created_at TIMESTAMP DEFAULT NOW, 
		conversation TEXT NOT NULL,

		FOREIGN KEY (conversation)
       		REFERENCES Conversations_V1(id)
	);

CREATE UNIQUE INDEX IF NOT EXISTS messages_id_v1 ON Messages_V1(id);
CREATE INDEX IF NOT EXISTS messages_conversation_time_v1 ON Messages_V1(conversation, created_at);
CREATE INDEX IF NOT EXISTS messages_users_time_v1 ON Messages_V1(user, created_at);