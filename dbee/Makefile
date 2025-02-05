MIGRATION_DIR ?= migrations

.PHONY: create-migration
create-migration:
	@TIMESTAMP=$$(date +%s); \
	UP_FILE="$(MIGRATION_DIR)/$${TIMESTAMP}_$(name).up.sql"; \
	DOWN_FILE="$(MIGRATION_DIR)/$${TIMESTAMP}_$(name).down.sql"; \
	touch "$$UP_FILE" && touch "$$DOWN_FILE";