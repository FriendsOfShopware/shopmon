current_dir := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))
.PHONY: help
help: # Show help
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

setup: # Setup the project
	@echo "Setting up the project"
	@echo "Create api/.env if not exists"
	@test -f api/.env || cp api/.env.example api/.env
	@echo "Installing dependencies"
	pnpm install

tslint:
	cd api && node ../node_modules/.bin/tsc --noEmit

lint:
	cd api && node_modules/.bin/tsc --noEmit
	cd frontend && node_modules/.bin/eslint --cache
	node_modules/.bin/biome ci

lint-fix:
	cd frontend && node_modules/.bin/eslint --cache --fix
	node_modules/.bin/biome check --fix --unsafe
	node_modules/.bin/biome format --write

generate-migration:
	cd api && npx drizzle-kit generate

generate-custom-migration:
	cd api && npx drizzle-kit generate --custom

generate-email:
	cd api && node generate-mjml.js

migrate:
	cd api && node --env-file-if-exists=.env migrate.ts

load-fixtures: # Load fixtures
	@echo "Loading fixtures"
	@echo "Create api/.env if not exists"
	@test -f api/.env || cp api/.env.example api/.env
	cd api && rm -rf shopmon.db*
	cd api && node --env-file-if-exists=.env migrate.ts
	cd api && node --env-file-if-exists=.env apply-fixtures.ts

dev: # Run the project locally
	npx concurrently -- 'cd api && PORT=5789 node --env-file-if-exists=.env --no-warnings=ExperimentalWarning --watch src/index.ts' 'npm --prefix frontend run dev:local'

dev-to-prod:
	npm --prefix frontend run dev

up: # Build and run the demo shop
	docker compose up -d
	@echo "Demo shop is running at http://localhost:3889"
	@echo "Integration: Key: SWIAUZL4OXRKEG1RR3PMCEVNMG, Secret: aXhNQ3NoRHZONmxPYktHT0c2c09rNkR0UHI0elZHOFIycjBzWks"
	@echo "Mailpit: http://localhost:8025"
