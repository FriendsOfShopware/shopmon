current_dir := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))
.PHONY: help
help: # Show help
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

setup: # Setup the project
	@echo "Setting up the project"
	@echo "Installing dependencies"
	pnpm install --prefix api
	pnpm install --prefix frontend

migrate:
	cd api && bun run migrate.ts

dev: # Run the project locally
	npx concurrently -- 'npm run --prefix=api dev' 'npm --prefix frontend run dev:local'

dev-to-prod:
	npm --prefix frontend run dev

up: # Build and run the demo shop
	docker compose up -d
	@echo "Demo shop is running at http://localhost:3889"
	@echo "Integration: Key: SWIAUZL4OXRKEG1RR3PMCEVNMG, Secret: aXhNQ3NoRHZONmxPYktHT0c2c09rNkR0UHI0elZHOFIycjBzWks"
	@echo "Mailpit: http://localhost:8025"
