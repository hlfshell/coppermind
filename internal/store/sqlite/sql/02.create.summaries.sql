CREATE TABLE IF NOT EXISTS
	Summaries_V1(
        id VARCHAR,
        conversation VARCHAR,
        agent VARCHAR,
        user VARCHAR,
        keywords text,
        summary text,
        created_at timestamp
    );