current_dir := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))
.PHONY: help
help: # Show help
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

setup: # Setup the project
	@echo "Setting up the project"
	@echo "Create api/.env if not exists"
	@test -f api/.env || cp api/.env.example api/.env
	@echo "Installing dependencies"
	bun install

tslint:
	cd api && bun ../node_modules/.bin/tsc --noEmit

lint:
	cd api && bun ../node_modules/.bin/tsc --noEmit
	cd frontend && bun ../node_modules/.bin/eslint --cache
	bun node_modules/.bin/biome ci

lint-fix:
	cd frontend && bun ../node_modules/.bin/eslint --cache --fix
	bun node_modules/.bin/biome check --fix --unsafe
	bun node_modules/.bin/biome format --write

generate-migration:
	cd api && bunx --bun drizzle-kit generate

generate-custom-migration:
	cd api && bunx --bun drizzle-kit generate --custom

generate-email:
	cd api && bun generate-mjml.js

migrate:
	cd api && bun run migrate.ts

load-fixtures: # Load fixtures
	@echo "Loading fixtures"
	@echo "Create api/.env if not exists"
	@test -f api/.env || cp api/.env.example api/.env
	cd api && rm -rf shopmon.db*
	cd api && bun run migrate.ts
	cd api && bun run apply-fixtures.ts

dev: # Run the project locally
	npx concurrently -- 'cd api && bun --watch --port 5789 src/index.ts' 'npm --prefix frontend run dev:local'

dev-to-prod:
	npm --prefix frontend run dev

up: # Build and run the demo shop
	docker compose up -d
	@echo "Demo shop is running at http://localhost:3889"
	@echo "Integration: Key: SWIAUZL4OXRKEG1RR3PMCEVNMG, Secret: aXhNQ3NoRHZONmxPYktHT0c2c09rNkR0UHI0elZHOFIycjBzWks"
	@echo "Mailpit: http://localhost:8025"

sitespeed: # Run sitespeed analysis for a shop (usage: make sitespeed SHOP_ID=1 URL=https://example.com)
ifndef SHOP_ID
	@echo "Error: SHOP_ID is required. Usage: make sitespeed SHOP_ID=1 URL=https://example.com"
	@exit 1
endif
ifndef URL
	@echo "Error: URL is required. Usage: make sitespeed SHOP_ID=1 URL=https://example.com"
	@exit 1
endif
	@echo "Running sitespeed analysis for shop $(SHOP_ID) on $(URL)..."
	@curl -X POST $${APP_SITESPEED_ENDPOINT:-http://localhost:3001}/analyze \
		-H "Content-Type: application/json" \
		-d '{"shopId": $(SHOP_ID), "url": "$(URL)"}' \
		--silent --show-error --fail \
		| jq '.' && echo "✅ Sitespeed analysis completed successfully!" || echo "❌ Failed to run sitespeed analysis. Make sure the sitespeed service is running (make up)"

sitespeed-demo: # Run sitespeed analysis for demo shop (shop ID 1)
	@echo "Running sitespeed analysis for demo shop..."
	@curl -X POST $${APP_SITESPEED_ENDPOINT:-http://localhost:3001}/analyze \
		-H "Content-Type: application/json" \
		-d '{"shopId": 1, "url": "http://localhost:3889"}' \
		--silent --show-error --fail \
		| jq '.' && echo "✅ Demo shop sitespeed analysis completed successfully!" || echo "❌ Failed to run sitespeed analysis. Make sure the sitespeed service is running (make up)"
