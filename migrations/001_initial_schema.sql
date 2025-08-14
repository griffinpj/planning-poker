-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    owner_id TEXT NOT NULL REFERENCES users(id),
    current_ticket_id INTEGER,
    is_voting_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tickets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    final_estimate INTEGER,
    position INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ticket_id INTEGER NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id),
    vote_value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ticket_id, user_id)
);

CREATE TABLE participants (
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (session_id, user_id)
);

CREATE TABLE recent_emojis (
    user_id TEXT NOT NULL REFERENCES users(id),
    emoji TEXT NOT NULL,
    used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, emoji)
);

CREATE INDEX idx_sessions_owner ON sessions(owner_id);
CREATE INDEX idx_tickets_session ON tickets(session_id);
CREATE INDEX idx_votes_ticket ON votes(ticket_id);
CREATE INDEX idx_votes_user ON votes(user_id);
CREATE INDEX idx_participants_session ON participants(session_id);
CREATE INDEX idx_recent_emojis_user ON recent_emojis(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_recent_emojis_user;
DROP INDEX IF EXISTS idx_participants_session;
DROP INDEX IF EXISTS idx_votes_user;
DROP INDEX IF EXISTS idx_votes_ticket;
DROP INDEX IF EXISTS idx_tickets_session;
DROP INDEX IF EXISTS idx_sessions_owner;

DROP TABLE IF EXISTS recent_emojis;
DROP TABLE IF EXISTS participants;
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd