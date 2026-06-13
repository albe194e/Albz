# Frontend Flutter

`app/frontend_flutter/` is the current Flutter frontend for the local-first app.

Current status:

- desktop Windows proof-of-concept bridge to `app/core-go/` is implemented
- the app now has the first real UI scaffold for the current desktop/mobile direction
- landing, login, register, chat, and profile pages now exist in Flutter
- conversation and message lists render from the current core snapshot
- local session bootstrap, login, and register are now wired
- conversation open, group creation, start-chat, and message send are now wired
- contact request actions, profile image upload, and event polling are now wired

## Development flow

1. Install Flutter and verify the toolchain:
   - `flutter doctor`
2. Fetch Flutter dependencies:
   - `cd app/frontend_flutter`
   - `flutter pub get`
3. Build the Go shared library from the repository root:
   - `make build-core-go-windows`
4. Run the Flutter Windows app:
   - `cd app/frontend_flutter`
   - `flutter run -d windows`

Root shortcuts:

- `make run-client`
- `make run-client-android`
- `make run-client-profile PROFILE=alice`
- `make run-client-alice`
- `make run-client-bob`
- `make run-clients`
- `make run-dev`
- `make reset-client-data`

Android shortcut:

- `make run-client-android`
- if more than one Android target is connected: `make run-client-android DEVICE=<flutter-device-id>`

The named-profile shortcuts use `--dart-define=HADDLE_PROFILE=<name>` so the Flutter app initializes `core-go` with separate local profile roots.

The multi-window shortcuts (`make run-clients`, `make run-dev`, `make stop-dev`) are backed by PowerShell helper scripts in `scripts/` so Windows path quoting does not break the spawned Flutter terminals.

`make run-client` now uses the normal repo-local desktop development database path again. `make reset-client-data` removes that repo-local data plus the old legacy Flutter fallback directory under `%LOCALAPPDATA%\Haddle\frontend_flutter` so you can fully reset local accounts and sessions during development.

Windows plugin note:

- `file_picker` requires Windows Developer Mode / symlink support for plugin builds
- if Flutter reports missing symlink support, enable Developer Mode:
  - `start ms-settings:developers`

The Go build target writes:

- `app/frontend_flutter/native/windows/haddle_core.dll`
- `app/frontend_flutter/native/windows/haddle_core.h`

The Windows Flutter build installs `haddle_core.dll` next to the app executable so Dart FFI can load it with `DynamicLibrary.open("haddle_core.dll")`.

## When to rebuild what

- Flutter UI-only change:
  - save the Dart file and use hot reload
- Dart FFI wrapper change:
  - hot restart is usually enough
- Go runtime / bridge / FFI export change:
  - rerun `make build-core-go-windows`
  - restart the Flutter app

## Current scope

This first milestone proves both:

- Flutter can talk to the embedded Go core
- the current desktop chat/product flows can be carried by Flutter

Exposed desktop proof-of-concept calls:

- create the core runtime
- close the core runtime
- fetch runtime config JSON
- fetch runtime snapshot JSON
- try to load the current session
- login
- register
- open a conversation
- create a group conversation
- send a message
- start a direct conversation from a contact
- send / accept / reject contact requests
- save a profile image into local app storage
- poll queued core events

Future work:

- Android and iOS native bindings
