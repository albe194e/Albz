# Albz data storage

This document describes the current data storage and transmission behavior in the codebase today.

## Local data stored on the client

The current Flutter client stores its local SQLite database under:

- default desktop path: `dev-local-db/local_storage/albz.db`
- named desktop development profiles: `dev-local-db/local_storage/profiles/<profile>/albz.db`
- Android path: the app support storage root plus `dev-local-db/local_storage/albz.db`

The local schema currently stores:

- `users`: local account records, including `id`, `name`, `username`, `contact_code`, `profile_picture_url`, and the value stored in `hashed_password`
- `sessions`: the active local session identifier, linked user, creation time, and expiry time
- `conversations`: local conversation identifiers and names
- `messages`: local message history, sender identifiers, client message identifiers, timestamps, and delivery state
- `conversation_participants`: local conversation membership data
- `contacts`: accepted contact records, including names, usernames, contact codes, `profile_picture_url`, and creation times
- `contact_requests`: pending contact requests, including sender identifiers, names, usernames, contact codes, and creation times

If a user selects a profile picture during registration, the image file itself is also stored locally on the device in an `images/` directory under the same local client data directory that holds the SQLite database. The `users.profile_picture_url` value points to that local file path.

Important current limitation: the field named `hashed_password` is not hashed in the current implementation. The raw password string is stored locally in that column today.

## Data stored on the server

The current server does not persist application data to a database.

There is no implemented server-side SQL schema in active use, and the running relay server keeps its state in memory only while the process is running.

The server currently keeps in memory:

- the active WebSocket connection for a connected user
- the connected user's `user_id`
- the connected user's `contact_code`

That in-memory routing state is removed when the client disconnects or the server process stops.

## Data transmitted through the server

The server relays live WebSocket events between connected clients. Those events currently include:

- message bodies, conversation identifiers, and recipient user identifiers during `message.send` / `message.created`
- conversation creation events, including the selected participant user identifiers and any explicitly chosen conversation name
- contact request events
- contact request acceptance and rejection events

The server needs to see enough event data to route it to the intended connected recipient in the current implementation.

## Data the server does not currently store durably

The current relay server does not durably store:

- message history
- conversation history
- contact lists
- contact requests
- profile databases
- profile picture files
- local SQLite contents
- offline delivery queues

If the server restarts, connected-session routing state is lost.

## Encryption and transport notes

Local SQLite data is stored locally on disk and is not encrypted by this codebase today.

Network transport depends on the configured WebSocket URL:

- current code default: `wss://oncological-paroxysmal-karolyn.ngrok-free.dev/ws`
- optional override: `ALBZ_SERVER_WS_URL`

The current code default uses `wss://`, so transport encryption depends on that deployment. If this is overridden with a `ws://...` URL during development, that override uses plain WebSocket without TLS.

## Current versus future behavior

Current behavior:

- the client keeps message and social data locally
- the server only relays live events for connected clients
- multi-participant conversations are live-only and require every participant to be connected when a conversation is created or a message is sent
- offline storage on the server is not implemented

Future work, if added later, should be documented here explicitly instead of being implied by current behavior.
