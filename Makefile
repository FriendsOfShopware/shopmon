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
	cd api && ../node_modules/.bin/tsc --noEmit
	bun oxlint
	bun oxfmt --check

lint-fix:
	bun oxfmt

generate-migration:
	cd api && npx drizzle-kit generate

generate-custom-migration:
	cd api && npx drizzle-kit generate --custom

generate-email:
	cd api && bun generate-mjml.js

migrate:
	cd api && bun migrate.ts

drop-db: # Drop the PostgreSQL database
	docker compose exec -T db psql -U shopmon -c "DROP SCHEMA IF EXISTS drizzle CASCADE; DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

load-fixtures: # Load fixtures
	@echo "Loading fixtures"
	@echo "Create api/.env if not exists"
	@test -f api/.env || cp api/.env.example api/.env
	$(MAKE) drop-db
	cd api && bun migrate.ts
	cd api && bun apply-fixtures.ts

dev: # Run the project locally
	@bun run --parallel --workspaces dev

up: # Build and run the demo shop
	docker compose up -d
	@echo "Demo shop is running at http://localhost:3889"
	@echo "Integration: Key: SWIAUZL4OXRKEG1RR3PMCEVNMG, Secret: aXhNQ3NoRHZONmxPYktHT0c2c09rNkR0UHI0elZHOFIycjBzWks"
	@echo "Mailpit: http://localhost:8025"
