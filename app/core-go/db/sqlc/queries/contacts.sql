-- name: UpsertContact :exec
INSERT INTO contacts (
  user_id,
  name,
  username,
  contact_code,
	profile_picture_url,
  created_at
) VALUES (
  ?, ?, ?, ?, ?, ?
)
ON CONFLICT(user_id) DO UPDATE SET
  name = excluded.name,
  username = excluded.username,
  contact_code = excluded.contact_code,
  profile_picture_url = excluded.profile_picture_url,
  created_at = excluded.created_at;

-- name: ListContacts :many
SELECT *
FROM contacts
ORDER BY created_at ASC;
