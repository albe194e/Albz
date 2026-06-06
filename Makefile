SQLC_VERSION ?= v1.31.1
POWERSHELL ?= powershell.exe
FYNE ?= fyne
APP_ID ?= com.albe194e.albz
APP_ICON ?= Icon.png
DEV_SERVER_CMD := title Albz Dev Server && go run ./server
DEV_CLIENT_ALICE_CMD := title Albz Dev Client Alice && go run ./client -profile alice
DEV_CLIENT_BOB_CMD := title Albz Dev Client Bob && go run ./client -profile bob

.PHONY: \
	sqlc \
	sqlc-client \
	sqlc-server \
	sqlc-verify-client \
	sqlc-verify-server \
	run-server \
	run-client \
	run-client-profile \
	run-client-alice \
	run-client-bob \
	run-clients \
	run-dev \
	stop-dev \
	package-android \
	package-ios \
	package-mobile

sqlc-client: ## Generate sqlc code for client
	@echo "Generating client sqlc code..."
	@$(POWERSHELL) -NoProfile -Command "Set-Location 'client'; go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f 'db/sqlc/sqlc.yaml' generate"

sqlc-server: ## Generate sqlc code for server
	@echo "Generating server sqlc code..."
	@$(POWERSHELL) -NoProfile -Command "Set-Location 'server'; go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f 'db/sqlc.yaml' generate"

sqlc: sqlc-client sqlc-server ## Generate sqlc code for client and server

sqlc-verify-client: ## Verify generated client sqlc code is committed
	@echo "Verifying client sqlc output..."
	@$(POWERSHELL) -NoProfile -Command "Set-Location 'client'; go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f 'db/sqlc/sqlc.yaml' generate"
	@git diff --exit-code -- client/db/generated

sqlc-verify-server: ## Verify generated server sqlc code is committed
	@echo "Verifying server sqlc output..."
	@$(POWERSHELL) -NoProfile -Command "Set-Location 'server'; go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f 'db/sqlc.yaml' generate"
	@git diff --exit-code -- server/db/generated

sqlc-verify: sqlc-verify-client sqlc-verify-server ## Verify all generated sqlc code is committed

run-server: ## Run the relay server in the current terminal
	@go run ./server

run-client: ## Run the desktop client with the default local database
	@go run ./client

run-client-profile: ## Run the desktop client with PROFILE=<name>
	@$(POWERSHELL) -NoProfile -Command "if ([string]::IsNullOrWhiteSpace('$(PROFILE)')) { Write-Error 'Usage: make run-client-profile PROFILE=alice'; exit 1 }; go run ./client -profile '$(PROFILE)'"

run-client-alice: ## Run the desktop client with the alice profile
	@go run ./client -profile alice

run-client-bob: ## Run the desktop client with the bob profile
	@go run ./client -profile bob

run-clients: ## Launch alice and bob client windows together
	@$(POWERSHELL) -NoProfile -Command "Start-Process cmd.exe -WorkingDirectory '$(CURDIR)' -ArgumentList '/c','$(DEV_CLIENT_ALICE_CMD)'; Start-Process cmd.exe -WorkingDirectory '$(CURDIR)' -ArgumentList '/c','$(DEV_CLIENT_BOB_CMD)'"

run-dev: ## Launch the server plus alice and bob client windows together
	@$(POWERSHELL) -NoProfile -Command "Start-Process cmd.exe -WorkingDirectory '$(CURDIR)' -ArgumentList '/c','$(DEV_SERVER_CMD)'; Start-Process cmd.exe -WorkingDirectory '$(CURDIR)' -ArgumentList '/c','$(DEV_CLIENT_ALICE_CMD)'; Start-Process cmd.exe -WorkingDirectory '$(CURDIR)' -ArgumentList '/c','$(DEV_CLIENT_BOB_CMD)'"

stop-dev: ## Stop the server and client windows started by run-dev/run-clients
	@$(POWERSHELL) -NoProfile -Command "$$targets = @('$(DEV_SERVER_CMD)', '$(DEV_CLIENT_ALICE_CMD)', '$(DEV_CLIENT_BOB_CMD)'); foreach ($$target in $$targets) { $$processes = Get-CimInstance Win32_Process | Where-Object { $$_.Name -eq 'cmd.exe' -and $$_.CommandLine -like ('*' + $$target + '*') }; foreach ($$process in $$processes) { cmd /c ('taskkill /PID ' + $$process.ProcessId + ' /T /F') | Out-Null } }"

package-android: ## Package the Fyne client as an Android APK
	@cd client && $(FYNE) package -os android -app-id $(APP_ID) -icon $(APP_ICON)

package-ios: ## Package the Fyne client as an iOS app bundle (requires macOS + Xcode)
	@cd client && $(FYNE) package -os ios -app-id $(APP_ID) -icon $(APP_ICON)

package-mobile: package-android package-ios ## Package the Fyne client for Android and iOS
