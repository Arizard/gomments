CREATE TABLE IF NOT EXISTS reaction (
    article TEXT NOT NULL,
    kind TEXT NOT NULL,
    session_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
