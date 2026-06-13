CREATE TABLE IF NOT EXISTS registered_devices (
	id INTEGER PRIMARY KEY,
	user_id TEXT NOT NULL UNIQUE,
	device_id TEXT NOT NULL UNIQUE,
	device_public_key BLOB NOT NULL,
	contact_code TEXT NOT NULL UNIQUE,
	created_at INTEGER NOT NULL,
	last_seen INTEGER NOT NULL,
	revoked_at INTEGER
);

CREATE INDEX IF NOT EXISTS idx_registered_devices_contact_code
ON registered_devices (contact_code);
