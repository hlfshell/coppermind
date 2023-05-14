CREATE TABLE IF NOT EXISTS
    Artifacts_V1(
        id TEXT NOT NULL PRIMARY KEY,
        message TEXT NOT NULL,
        type TEXT NOT NULL,
        data BYTEA,
        created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS artifacts_id_v1 ON Artifacts_V1(id);
CREATE INDEX IF NOT EXISTS artifacts_message_v1 ON Artifacts_V1(message);
CREATE INDEX IF NOT EXISTS artifacts_time_v1 ON Artifacts_V1(created_at);