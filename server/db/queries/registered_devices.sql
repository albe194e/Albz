-- name: CreateRegisteredDevice :exec
INSERT INTO registered_devices (
	user_id,
	device_id,
	device_public_key,
	contact_code,
	created_at,
	last_seen,
	revoked_at
) VALUES (
	?, ?, ?, ?, ?, ?, ?
);

-- name: GetRegisteredDeviceByUserID :one
SELECT *
FROM registered_devices
WHERE user_id = ?;

-- name: UpdateRegisteredDeviceLastSeen :exec
UPDATE registered_devices
SET last_seen = ?
WHERE user_id = ?;
