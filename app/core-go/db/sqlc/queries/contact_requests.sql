-- name: UpsertContactRequest :exec
INSERT INTO contact_requests (
  from_user_id,
  name,
  username,
  from_contact_code,
  created_at
) VALUES (
  ?, ?, ?, ?, ?
)
ON CONFLICT(from_user_id) DO UPDATE SET
  name = excluded.name,
  username = excluded.username,
  from_contact_code = excluded.from_contact_code,
  created_at = excluded.created_at;

-- name: ListContactRequests :many
SELECT *
FROM contact_requests
ORDER BY created_at ASC;

-- name: DeleteContactRequestByFromUserID :exec
DELETE FROM contact_requests
WHERE from_user_id = ?;
