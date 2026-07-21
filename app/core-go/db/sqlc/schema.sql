CREATE TABLE IF NOT EXISTS local_identity (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	user_id TEXT NOT NULL,
	device_id TEXT NOT NULL,
	device_public_key BLOB NOT NULL,
	encrypted_device_private_key BLOB NOT NULL,
	kdf_salt BLOB NOT NULL,
	kdf_params TEXT NOT NULL,
	name TEXT NOT NULL,
	local_handle TEXT,
	profile_picture_path TEXT,
	contact_code TEXT,
	created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	session_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	device_id TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	expires_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS conversations (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER
);

CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	conversation_id TEXT NOT NULL,
	sender_user_id TEXT NOT NULL,
	sender_device_id TEXT,
	client_message_id TEXT NOT NULL,
	body TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	received_at INTEGER,
	direction TEXT NOT NULL,
	delivery_state TEXT NOT NULL,

	FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_created
ON messages (conversation_id, created_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_client_message_id
ON messages (client_message_id);

CREATE TABLE IF NOT EXISTS conversation_participants (
	id INTEGER PRIMARY KEY,
	conversation_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	created_at INTEGER NOT NULL,

	FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_conversation_participants_unique
ON conversation_participants (conversation_id, user_id);

CREATE TABLE IF NOT EXISTS contacts (
	id INTEGER PRIMARY KEY,
	user_id TEXT NOT NULL UNIQUE,
	display_name TEXT NOT NULL,
	local_handle TEXT,
	profile_picture_path TEXT,
	contact_code TEXT,
	created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS contact_devices (
	id INTEGER PRIMARY KEY,
	contact_user_id TEXT NOT NULL,
	device_id TEXT NOT NULL,
	public_key BLOB NOT NULL,
	created_at INTEGER NOT NULL,
	revoked_at INTEGER
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_contact_devices_unique
ON contact_devices (contact_user_id, device_id);

CREATE TABLE IF NOT EXISTS contact_requests (
	id INTEGER PRIMARY KEY,
	from_user_id TEXT NOT NULL,
	from_device_id TEXT,
	display_name TEXT NOT NULL,
	local_handle TEXT,
	profile_picture_path TEXT,
	from_public_key BLOB,
	from_contact_code TEXT,
	invite_payload TEXT NOT NULL,
	state TEXT NOT NULL,
	created_at INTEGER NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_contact_requests_from_user
ON contact_requests (from_user_id);
