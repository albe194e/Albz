-- name: GetUserByID :one
SELECT
	id,
	name,
	username,
	hashed_password,
	friend_code
FROM users
WHERE id = ?;
