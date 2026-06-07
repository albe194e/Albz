-- name: UpsertFriendRequest :exec
INSERT INTO friend_requests (
  from_user_id,
  name,
  username,
  from_friend_code,
  created_at
) VALUES (
  ?, ?, ?, ?, ?
)
ON CONFLICT(from_user_id) DO UPDATE SET
  name = excluded.name,
  username = excluded.username,
  from_friend_code = excluded.from_friend_code,
  created_at = excluded.created_at;

-- name: ListFriendRequests :many
SELECT *
FROM friend_requests
ORDER BY created_at ASC;

-- name: DeleteFriendRequestByFromUserID :exec
DELETE FROM friend_requests
WHERE from_user_id = ?;
