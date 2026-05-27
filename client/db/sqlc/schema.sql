CREATE TABLE IF NOT EXISTS users (
	id VARCHAR(36) PRIMARY KEY,
	name TEXT NOT NULL,
	username TEXT NOT NULL UNIQUE,
	hashed_password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	session_id VARCHAR(36) NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	created_at INTEGER NOT NULL,
	expires_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS conversations (
	id VARCHAR(16) PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
  id INTEGER PRIMARY KEY,
  conversation_id VARCHAR(16) NOT NULL,
  sender_id VARCHAR(16) NOT NULL,
  body TEXT NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_created
ON messages (conversation_id, created_at);

CREATE TABLE IF NOT EXISTS conversation_participants (
	id INTEGER PRIMARY KEY,
	conversation_id VARCHAR(16) NOT NULL,
	participant_id VARCHAR(16) NOT NULL
);
