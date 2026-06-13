-- name: UpsertLocalIdentity :exec
INSERT INTO local_identity (
	id,
	user_id,
	device_id,
	device_public_key,
	encrypted_device_private_key,
	kdf_salt,
	kdf_params,
	name,
	local_handle,
	profile_picture_path,
	contact_code,
	created_at
) VALUES (
	1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT(id) DO UPDATE SET
	user_id = excluded.user_id,
	device_id = excluded.device_id,
	device_public_key = excluded.device_public_key,
	encrypted_device_private_key = excluded.encrypted_device_private_key,
	kdf_salt = excluded.kdf_salt,
	kdf_params = excluded.kdf_params,
	name = excluded.name,
	local_handle = excluded.local_handle,
	profile_picture_path = excluded.profile_picture_path,
	contact_code = excluded.contact_code,
	created_at = excluded.created_at;

-- name: GetLocalIdentity :one
SELECT *
FROM local_identity
WHERE id = 1;

-- name: UpsertCurrentSession :exec
INSERT INTO sessions (
	id,
	session_id,
	user_id,
	device_id,
	created_at,
	expires_at
) VALUES (
	1, ?, ?, ?, ?, ?
)
ON CONFLICT(id) DO UPDATE SET
	session_id = excluded.session_id,
	user_id = excluded.user_id,
	device_id = excluded.device_id,
	created_at = excluded.created_at,
	expires_at = excluded.expires_at;

-- name: DeleteCurrentSession :exec
DELETE FROM sessions
WHERE id = 1;

-- name: GetCurrentSession :one
SELECT *
FROM sessions
WHERE id = 1;
