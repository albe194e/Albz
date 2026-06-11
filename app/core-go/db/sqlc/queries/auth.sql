-- name: CreateUser :exec
INSERT INTO users (
	id,
	name,
	username,
	hashed_password,
	profile_picture_url,
	contact_code
) VALUES (
	?, ?, ?, ?, ?, ?
);

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = ?;

-- name: UpsertCurrentSession :exec
INSERT INTO sessions (
	id,
	session_id,
	user_id,
	created_at,
	expires_at
) VALUES (
	1, ?, ?, ?, ?
)
ON CONFLICT(id) DO UPDATE SET
	session_id = excluded.session_id,
	user_id = excluded.user_id,
	created_at = excluded.created_at,
	expires_at = excluded.expires_at;

-- name: DeleteCurrentSession :exec
DELETE FROM sessions
WHERE id = 1;

-- name: GetCurrentSession :one
SELECT *
FROM sessions
WHERE id = 1;
