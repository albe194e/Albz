CREATE TABLE IF NOT EXISTS users (
	id VARCHAR(36) PRIMARY KEY,
	name TEXT NOT NULL,
	username TEXT NOT NULL UNIQUE,
	hashed_password TEXT NOT NULL,
	friend_code TEXT NOT NULL UNIQUE
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
  client_message_id VARCHAR(36) NOT NULL,
  body TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  delivery_state TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_created
ON messages (conversation_id, created_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_client_message_id
ON messages (client_message_id);

CREATE TABLE IF NOT EXISTS conversation_participants (
	id INTEGER PRIMARY KEY,
	conversation_id VARCHAR(16) NOT NULL,
	participant_id VARCHAR(16) NOT NULL
);

CREATE TABLE IF NOT EXISTS friends (
	id INTEGER PRIMARY KEY,
	user_id VARCHAR(36) NOT NULL UNIQUE,
	name TEXT NOT NULL,
	username TEXT NOT NULL,
	friend_code TEXT NOT NULL,
	created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS friend_requests (
	id INTEGER PRIMARY KEY,
	from_user_id VARCHAR(36) NOT NULL UNIQUE,
	name TEXT NOT NULL,
	username TEXT NOT NULL,
	from_friend_code TEXT NOT NULL,
	created_at INTEGER NOT NULL
);
