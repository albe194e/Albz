-- name: GetLocalIdentityByUserID :one
SELECT *
FROM local_identity
WHERE user_id = ?;
