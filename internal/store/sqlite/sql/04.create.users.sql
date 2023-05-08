CREATE TABLE IF NOT EXISTS
    Users_V1(
        id TEXT NOT NULL PRIMARY KEY,
        name TEXT NOT NULL,
        password TEXT NOT NULL,
        reset_token TEXT,
        reset_token_attempts INTEGER DEFAULT 0,
        reset_token_generated_at TIMESTAMP,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
    );

CREATE UNIQUE INDEX IF NOT EXISTS users_id_v1 ON Users_V1(id);