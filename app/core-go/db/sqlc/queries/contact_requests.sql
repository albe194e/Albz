-- name: UpsertContactRequest :exec
INSERT INTO contact_requests (
	from_user_id,
	from_device_id,
	display_name,
	local_handle,
	from_public_key,
	from_contact_code,
	invite_payload,
	state,
	created_at
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT(from_user_id) DO UPDATE SET
	from_device_id = excluded.from_device_id,
	display_name = excluded.display_name,
	local_handle = excluded.local_handle,
	from_public_key = excluded.from_public_key,
	from_contact_code = excluded.from_contact_code,
	invite_payload = excluded.invite_payload,
	state = excluded.state,
	created_at = excluded.created_at;

-- name: ListContactRequests :many
SELECT *
FROM contact_requests
ORDER BY created_at ASC;

-- name: DeleteContactRequestByFromUserID :exec
DELETE FROM contact_requests
WHERE from_user_id = ?;
