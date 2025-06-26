CREATE TABLE reply (
    reply_id INTEGER PRIMARY KEY AUTOINCREMENT,
    reply_idempotency_key TEXT UNIQUE NOT NULL,
    reply_signature TEXT NOT NULL,

    reply_article TEXT NOT NULL,
    reply_body TEXT NOT NULL,
    reply_deleted BOOLEAN DEFAULT FALSE NOT NULL,
    reply_created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,

    reply_author_name TEXT NOT NULL
);
