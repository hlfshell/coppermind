CREATE TABLE IF NOT EXISTS
    Knowledge_V1(
        id TEXT NOT NULL PRIMARY KEY,
        agent TEXT NOT NULL,
        user TEXT NOT NULL,
        subject TEXT NOT NULL,
        predicate TEXT NOT NULL,
        object TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW,
        expires_at TIMESTAMP
    );
        
CREATE UNIQUE INDEX IF NOT EXISTS knowledge_id_v1 ON Knowledge_V1(id);
CREATE INDEX IF NOT EXISTS knowledge_agent_user_time_v1 ON Knowledge_V1(agent, user, created_at);
CREATE INDEX IF NOT EXISTS knowledge_expiration_v1 ON Knowledge_V1(expires_at);

CREATE TABLE IF NOT EXISTS
    KnowledgeExtraction_V1(
        conversation TEXT NOT NULL,
        updated_at TIMESTAMP NOT NULL DEFAULT NOW
    );

CREATE UNIQUE INDEX IF NOT EXISTS knowledge_extraction_v1 ON KnowledgeExtraction_V1(conversation);