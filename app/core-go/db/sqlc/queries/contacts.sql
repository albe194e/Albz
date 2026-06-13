-- name: UpsertContact :exec
INSERT INTO contacts (
	user_id,
	display_name,
	local_handle,
	contact_code,
	profile_picture_path,
	created_at
) VALUES (
	?, ?, ?, ?, ?, ?
)
ON CONFLICT(user_id) DO UPDATE SET
	display_name = excluded.display_name,
	local_handle = excluded.local_handle,
	contact_code = excluded.contact_code,
	profile_picture_path = excluded.profile_picture_path,
	created_at = excluded.created_at;

-- name: ListContacts :many
SELECT *
FROM contacts
ORDER BY created_at ASC;
