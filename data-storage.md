# Haddle data storage

This document describes the current data storage and transmission behavior in the codebase today.

## Local data stored on the client

The current Flutter client stores its local SQLite database under:

- default desktop path: `dev-local-db/local_storage/haddle.db`
- named desktop development profiles: `dev-local-db/local_storage/profiles/<profile>/haddle.db`
- Android path: the app support storage root plus `dev-local-db/local_storage/haddle.db`

The local schema currently stores:

- `local_identity`: one local identity row per app install, including the local `user_id`, the current `device_id`, the device public key, the encrypted device private key, the KDF salt and KDF parameters used to derive a local unlock key, the local display name, the optional local handle, the optional local profile picture path, the optional contact code, and the creation time
- `sessions`: the active local session identifier, linked `user_id`, linked `device_id`, creation time, and expiry time
- `conversations`: local conversation identifiers, names, conversation type values such as direct or room, creation time, and optional update time
- `conversation_participants`: local conversation membership data stored by `conversation_id` and participant `user_id`
- `messages`: local message history, globally unique message IDs, sender user/device identifiers, client message identifiers, timestamps, direction, and delivery state
- `contacts`: accepted local contact records, including local display names, optional local handles, optional local profile picture paths, contact codes, and creation times
- `contact_devices`: locally stored public keys for known contact devices so the client can prepare for future encrypted multi-device delivery without asking the relay server to own that data
- `contact_requests`: local contact request records, including sender user/device identifiers, local display data, optional sender profile picture paths, optional sender public keys, contact codes, invite payloads, request state, and creation time

If a user selects a profile picture, the image file itself is stored locally on the device in an `images/` directory under the same local client data directory that holds the SQLite database. The database stores a local filesystem path such as `local_identity.profile_picture_path` or `contacts.profile_picture_path`, not a server-owned profile URL.

The current client schema does not store the raw PIN/password and does not store a password hash for server-side login. Instead, the current Go core derives a local unlock key from the user-provided password plus the stored salt/KDF settings, and uses that key to decrypt the locally stored encrypted device private key on login.

After a successful local login, the current Go core keeps the decrypted device private key in process memory for the running app session so it can authenticate to the relay. That decrypted key is not written back into the SQLite database.

Current register/login behavior also requires a successful relay registration/authentication step before a new local session is considered complete. If relay registration fails, the local identity row still remains on the device so the user can retry later, but the login/register call returns an error instead of creating a fresh local session.

## Data stored on the server

The current relay server now persists a minimal device registry in its own SQLite database.

The current server-side database stores:

- `registered_devices`: the relay-side device registry, including `user_id`, `device_id`, `device_public_key`, `contact_code`, `created_at`, `last_seen`, and optional `revoked_at`

The running relay server also keeps some state in memory only while the process is running:

- the active WebSocket connection for an authenticated device
- the authenticated device's `user_id`
- the authenticated device's `device_id`
- the authenticated device's `contact_code`
- short-lived relay auth challenge state while a client is proving possession of its device private key

That in-memory relay state is removed when the client disconnects or the server process stops. The durable device registry remains in the relay database until it is explicitly changed or deleted by server-side maintenance code.

## Data transmitted through the server

The server relays live WebSocket events between connected clients. Those events currently include:

- device registration data sent to the relay before a device is authenticated: `user_id`, `device_id`, `device_public_key`, and `contact_code`
- relay auth challenge/response material used to prove possession of the local device private key without sending that private key to the server
- message bodies, conversation identifiers, and recipient user identifiers during `message.send` / `message.created`
- conversation creation events, including the selected participant user identifiers and any explicitly chosen conversation name
- contact request events, which now also include sender device identifiers, sender device public keys, and an explicitly shared profile picture as raw image bytes when the sender has chosen one
- contact request acceptance events, which can now also include an explicitly shared profile picture as raw image bytes
- contact request rejection events

The server needs to see enough event data to route it to the intended connected recipient in the current implementation.

When a contact request or contact acceptance event includes a shared profile picture, the receiving client saves those bytes into its own local `images/` storage. Pending requests store the resulting local filesystem path in `contact_requests.profile_picture_path`, and accepted contacts store it in `contacts.profile_picture_path`. The relay forwards those bytes in transit but does not store the image file durably.

## Data the server does not currently store durably

The current relay server does not durably store:

- message history
- conversation history
- contact lists
- contact requests
- registered device private keys
- device private keys
- PINs or passwords
- password hashes
- profile databases
- profile picture files
- local SQLite contents
- offline delivery queues
- pending auth challenge state

If the server restarts, connected-session routing state is lost, but the registered-device registry remains in the relay database.

## Encryption and transport notes

Local SQLite data is stored locally on disk and is not encrypted by this codebase today.

Network transport depends on the configured WebSocket URL:

- current code default: `wss://oncological-paroxysmal-karolyn.ngrok-free.dev/ws`
- optional override: `HADDLE_SERVER_WS_URL`

The current code default uses `wss://`, so transport encryption depends on that deployment. If this is overridden with a `ws://...` URL during development, that override uses plain WebSocket without TLS.

## Logging and diagnostics

The current codebase does not write a separate telemetry or analytics store.

Current development logging is controlled by `HADDLE_DEV_MODE`:

- current default behavior: local development commands enable dev mode by default
- when dev mode is enabled, Flutter, `app/core-go`, and the relay emit verbose startup and runtime diagnostics to standard output/error
- when dev mode is disabled, those runtimes limit themselves to breaking-error logging only

The current logging helpers are intended to avoid printing especially sensitive values such as:

- passwords
- decrypted private keys
- encrypted private-key blobs
- message bodies

Some log lines can still contain operational metadata such as filesystem paths, websocket URLs, or redacted user/device identifiers when that is needed for local debugging.

## Current versus future behavior

Current behavior:

- the client keeps identity, contact, conversation, message, and contact-device key data locally
- the server only relays live events for authenticated connected clients
- the relay currently stores a minimal registered-device database so it can remember `user_id`, `device_id`, `device_public_key`, and `contact_code` across restarts
- multi-participant conversations are live-only and require every participant to be connected when a conversation is created or a message is sent
- offline storage on the server is not implemented
- local session restore can still load local state, but relay authentication currently requires an unlocked device private key in memory, so a cold app restart still requires a fresh password unlock before reconnecting to the relay

Future work, if added later, should be documented here explicitly instead of being implied by current behavior.
