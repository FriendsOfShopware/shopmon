current_dir := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))
.PHONY: help
help: # Show help
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

setup: # Setup the project
	@echo "Setting up the project"
	@echo "Installing dependencies"
	pnpm install --prefix api
	pnpm install --prefix frontend

migrate: $(current_dir)/api/drizzle/*.sql # Run migrations
	for file in $^ ; do \
		echo "Running $$file" ; \
		pnpm --prefix api run wrangler d1 execute --local --yes shopmon --file $$file ; \
	done

dev: # Run the project locally
	npx concurrently -- 'npm --prefix api run  dev:local' 'npm --prefix frontend run dev:local'
