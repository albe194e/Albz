-- name: UpsertContactDevice :exec
INSERT INTO contact_devices (
	contact_user_id,
	device_id,
	public_key,
	created_at,
	revoked_at
) VALUES (
	?, ?, ?, ?, ?
)
ON CONFLICT(contact_user_id, device_id) DO UPDATE SET
	public_key = excluded.public_key,
	created_at = excluded.created_at,
	revoked_at = excluded.revoked_at;
