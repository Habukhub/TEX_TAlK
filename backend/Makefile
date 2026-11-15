.PHONY: docs-gen dev dev-win db-up db-up-win db-up-app db-up-app-win db-down migrate-diff migrate-apply migrate-apply-prod migrate-status migrate-status-prod migrate-hash test test-coverage tidy start

POWERSHELL := powershell.exe
POWERSHELL_FLAGS := -NoProfile -ExecutionPolicy Bypass -Command

# Swagger documentation
docs-gen:
	swag init -g src/cmd/api/main.go -o docs

# Testing
test:
	@echo "Running unit tests..."
	@go test ./...

test-coverage:
	@echo "Generating coverage report..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

# Development (Linux/macOS)
dev: db-up-app migrate-apply
	@trap 'echo "\nStopping Docker containers..."; docker compose stop; exit 0' INT TERM; \
	APP_ENV=local VERSION=$$(cat VERSION) go run ./src/cmd/api

# Development (Windows/PowerShell)
dev-win: db-up-app-win migrate-apply
	$(POWERSHELL) $(POWERSHELL_FLAGS) "$$projectRoot = '$(CURDIR)'; $$version = Get-Content (Join-Path $$projectRoot 'VERSION') -Raw; $$version = $$version.Trim(); $$exitCode = 0; Push-Location $$projectRoot; try { $$env:APP_ENV = 'local'; $$env:VERSION = $$version; go run ./src/cmd/api; $$exitCode = $$LASTEXITCODE } finally { Write-Host ''; Write-Host 'Stopping Docker containers...'; docker compose stop; Pop-Location }; exit $$exitCode"

# Database
db-up:
	docker compose up -d
	@echo "Waiting for database to be ready..."
	@until docker compose exec postgres pg_isready -U admin -d tex_db; do sleep 1; done

db-up-win:
	docker compose up -d
	$(POWERSHELL) $(POWERSHELL_FLAGS) "Write-Host 'Waiting for database to be ready...'; $$maxAttempts = 60; for ($$i = 0; $$i -lt $$maxAttempts; $$i++) { docker compose exec -T postgres pg_isready -U admin -d tex_db *> $$null; if ($$LASTEXITCODE -eq 0) { exit 0 }; Start-Sleep -Seconds 1 }; Write-Error 'Database did not become ready in time.'; exit 1"

db-up-app:
	docker compose up -d postgres
	@echo "Waiting for database to be ready..."
	@until docker compose exec postgres pg_isready -U admin -d tex_db; do sleep 1; done

db-up-app-win:
	docker compose up -d postgres
	$(POWERSHELL) $(POWERSHELL_FLAGS) "Write-Host 'Waiting for database to be ready...'; $$maxAttempts = 60; for ($$i = 0; $$i -lt $$maxAttempts; $$i++) { docker compose exec -T postgres pg_isready -U admin -d tex_db *> $$null; if ($$LASTEXITCODE -eq 0) { exit 0 }; Start-Sleep -Seconds 1 }; Write-Error 'Database did not become ready in time.'; exit 1"

db-down:
	docker compose down

# Atlas migrations
migrate-diff:
	atlas migrate diff $(name) --env local

migrate-apply:
	atlas migrate apply --env local

migrate-apply-prod:
	atlas migrate apply --env prod

migrate-status:
	atlas migrate status --env local

migrate-status-prod:
	atlas migrate status --env prod

migrate-hash:
	atlas migrate hash --env local
	atlas migrate hash --env prod

# Dependencies
tidy:
	go mod tidy

# All in one
start: tidy db-up-app migrate-apply dev
