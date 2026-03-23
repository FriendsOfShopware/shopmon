TASKS := setup lint lint-fix generate migrate migrate-status drop-db load-fixtures test build dev dev-worker up stop

.PHONY: help $(TASKS)

help: ## Show help
	@node dev.mjs help

$(TASKS):
	@node dev.mjs $@
