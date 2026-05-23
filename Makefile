SQLC_VERSION ?= v1.31.1

.PHONY: sqlc sqlc-client sqlc-server sqlc-verify-client sqlc-verify-server

sqlc-client: ## Generate sqlc code for client
	@echo "🧱 Generating client sqlc code..."
	@cd client && go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f db/sqlc/sqlc.yaml generate

sqlc-server: ## Generate sqlc code for server
	@echo "🧱 Generating server sqlc code..."
	@cd server && go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f db/sqlc/sqlc.yaml generate

sqlc: sqlc-client sqlc-server ## Generate sqlc code for client and server

sqlc-verify-client: ## Verify generated client sqlc code is committed
	@echo "🧱 Verifying client sqlc output..."
	@cd client && go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f db/sqlc/sqlc.yaml generate
	@git diff --exit-code -- client/db/generated

sqlc-verify-server: ## Verify generated server sqlc code is committed
	@echo "🧱 Verifying server sqlc output..."
	@cd server && go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) -f db/sqlc/sqlc.yaml generate
	@git diff --exit-code -- server/db/generated

sqlc-verify: sqlc-verify-client sqlc-verify-server ## Verify all generated sqlc code is committed