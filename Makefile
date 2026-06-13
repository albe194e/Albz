SQLC_VERSION ?= v1.31.1
POWERSHELL ?= powershell.exe
FLUTTER ?= flutter
FRONTEND_FLUTTER_DIR := app/frontend_flutter
DEV_SERVER_SCRIPT := $(CURDIR)\scripts\run-server-window.ps1
DEV_CLIENT_SCRIPT := $(CURDIR)\scripts\run-flutter-client-window.ps1
DEV_STOP_SCRIPT := $(CURDIR)\scripts\stop-dev-windows.ps1

.PHONY: \
	sqlc \
	sqlc-core-go \
	sqlc-server \
	sqlc-verify-core-go \
	sqlc-verify-server \
	build-core-go-windows \
	build-core-go-android \
	run-server \
	run-client \
	run-client-android \
	run-client-profile \
	run-client-alice \
	run-client-bob \
	run-clients \
	run-dev \
	stop-dev \
	reset-client-data

sqlc-core-go: ## Generate sqlc code for core-go
	@echo "Generating core-go sqlc code..."
	@$(POWERSHELL) -NoProfile -Command "Set-Location 'app/core-go'; go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f 'db/sqlc/sqlc.yaml' generate"

sqlc-server: ## Generate sqlc code for server
	@echo "Generating server sqlc code..."
	@$(POWERSHELL) -NoProfile -Command "Set-Location 'server'; go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f 'db/sqlc.yaml' generate"

sqlc: sqlc-core-go sqlc-server ## Generate sqlc code for core-go and server

sqlc-verify-core-go: ## Verify generated core-go sqlc code is committed
	@echo "Verifying core-go sqlc output..."
	@$(POWERSHELL) -NoProfile -Command "Set-Location 'app/core-go'; go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f 'db/sqlc/sqlc.yaml' generate"
	@git diff --exit-code -- app/core-go/db/sqlc/sql

sqlc-verify-server: ## Verify generated server sqlc code is committed
	@echo "Verifying server sqlc output..."
	@$(POWERSHELL) -NoProfile -Command "Set-Location 'server'; go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f 'db/sqlc.yaml' generate"
	@git diff --exit-code -- server/db/generated

sqlc-verify: sqlc-verify-core-go sqlc-verify-server ## Verify all generated sqlc code is committed

build-core-go-windows: ## Build the core-go shared library for the Flutter Windows app
	@echo "Building core-go Windows shared library..."
	@$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File "app/core-go/build-windows.ps1"

build-core-go-android: ## Build Android shared libraries for the Flutter app
	@echo "Building core-go Android shared libraries..."
	@$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File "app/core-go/build-android.ps1"

run-server: ## Run the relay server in the current terminal
	@$(POWERSHELL) -NoProfile -Command '$$env:HADDLE_DEV_MODE="1"; go run ./server'

run-client: build-core-go-windows ## Run the Flutter desktop client on Windows
	@$(POWERSHELL) -NoProfile -Command "Set-Location '$(FRONTEND_FLUTTER_DIR)'; $(FLUTTER) run -d windows --dart-define=HADDLE_DEV_MODE=true"

run-client-android: build-core-go-android ## Run the Flutter client on a USB-connected Android device (optional: DEVICE=<flutter-device-id>)
	@$(POWERSHELL) -NoProfile -Command "Set-Location '$(FRONTEND_FLUTTER_DIR)'; if ([string]::IsNullOrWhiteSpace('$(DEVICE)')) { $(FLUTTER) run -d R3CX100N1WB --dart-define=HADDLE_DEV_MODE=true } else { $(FLUTTER) run -d $(DEVICE) --dart-define=HADDLE_DEV_MODE=true }"

run-client-profile: build-core-go-windows ## Run the Flutter desktop client with PROFILE=<name>
	@$(POWERSHELL) -NoProfile -Command "if ([string]::IsNullOrWhiteSpace('$(PROFILE)')) { Write-Error 'Usage: make run-client-profile PROFILE=alice'; exit 1 }; Set-Location '$(FRONTEND_FLUTTER_DIR)'; $(FLUTTER) run -d windows --dart-define=HADDLE_DEV_MODE=true --dart-define=HADDLE_PROFILE=$(PROFILE)"

run-client-alice: build-core-go-windows ## Run the Flutter desktop client with the alice profile
	@$(POWERSHELL) -NoProfile -Command "Set-Location '$(FRONTEND_FLUTTER_DIR)'; $(FLUTTER) run -d windows --dart-define=HADDLE_DEV_MODE=true --dart-define=HADDLE_PROFILE=alice"

run-client-bob: build-core-go-windows ## Run the Flutter desktop client with the bob profile
	@$(POWERSHELL) -NoProfile -Command "Set-Location '$(FRONTEND_FLUTTER_DIR)'; $(FLUTTER) run -d windows --dart-define=HADDLE_DEV_MODE=true --dart-define=HADDLE_PROFILE=bob"

run-clients: build-core-go-windows ## Launch alice and bob Flutter client windows together
	@$(POWERSHELL) -NoProfile -Command "Start-Process powershell.exe -WorkingDirectory '$(CURDIR)' -ArgumentList '-NoExit','-ExecutionPolicy','Bypass','-File','$(DEV_CLIENT_SCRIPT)','-WorkspaceRoot','$(CURDIR)','-Profile','alice','-Flutter','$(FLUTTER)'; Start-Process powershell.exe -WorkingDirectory '$(CURDIR)' -ArgumentList '-NoExit','-ExecutionPolicy','Bypass','-File','$(DEV_CLIENT_SCRIPT)','-WorkspaceRoot','$(CURDIR)','-Profile','bob','-Flutter','$(FLUTTER)'"

run-dev: build-core-go-windows ## Launch the server plus alice and bob Flutter client windows together
	@$(POWERSHELL) -NoProfile -Command "Start-Process powershell.exe -WorkingDirectory '$(CURDIR)' -ArgumentList '-NoExit','-ExecutionPolicy','Bypass','-File','$(DEV_SERVER_SCRIPT)','-WorkspaceRoot','$(CURDIR)'; Start-Process powershell.exe -WorkingDirectory '$(CURDIR)' -ArgumentList '-NoExit','-ExecutionPolicy','Bypass','-File','$(DEV_CLIENT_SCRIPT)','-WorkspaceRoot','$(CURDIR)','-Profile','alice','-Flutter','$(FLUTTER)'; Start-Process powershell.exe -WorkingDirectory '$(CURDIR)' -ArgumentList '-NoExit','-ExecutionPolicy','Bypass','-File','$(DEV_CLIENT_SCRIPT)','-WorkspaceRoot','$(CURDIR)','-Profile','bob','-Flutter','$(FLUTTER)'"

stop-dev: ## Stop the server and client windows started by run-dev/run-clients
	@$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File "$(DEV_STOP_SCRIPT)"

reset-client-data: ## Delete local development client data (repo-local plus legacy Flutter fallback)
	@$(POWERSHELL) -NoProfile -Command '$$paths = @("dev-local-db\local_storage"); if ($$env:LOCALAPPDATA) { $$paths += (Join-Path $$env:LOCALAPPDATA "Haddle\frontend_flutter") }; foreach ($$path in $$paths) { if (Test-Path -LiteralPath $$path) { Remove-Item -LiteralPath $$path -Recurse -Force } }'
