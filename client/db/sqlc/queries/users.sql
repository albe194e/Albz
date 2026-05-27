-- name: GetUserByID :one
SELECT
	id,
	name,
	username,
	hashed_password
FROM users
WHERE id = ?;
