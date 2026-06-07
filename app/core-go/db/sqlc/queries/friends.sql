-- name: UpsertFriend :exec
INSERT INTO friends (
  user_id,
  name,
  username,
  friend_code,
	profile_picture_url,
  created_at
) VALUES (
  ?, ?, ?, ?, ?, ?
)
ON CONFLICT(user_id) DO UPDATE SET
  name = excluded.name,
  username = excluded.username,
  friend_code = excluded.friend_code,
  profile_picture_url = excluded.profile_picture_url,
  created_at = excluded.created_at;

-- name: ListFriends :many
SELECT *
FROM friends
ORDER BY created_at ASC;
